package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type Storer interface {
	SaveUser(User) error
	GetUser(string) (User, error)
}

type WebAuthn interface {
	BeginRegistration(user webauthn.User, opts ...webauthn.RegistrationOption) (creation *protocol.CredentialCreation, session *webauthn.SessionData, err error)
	FinishRegistration(user webauthn.User, session webauthn.SessionData, response *http.Request) (*webauthn.Credential, error)
	BeginLogin(user webauthn.User, opts ...webauthn.LoginOption) (*protocol.CredentialAssertion, *webauthn.SessionData, error)
	FinishLogin(user webauthn.User, session webauthn.SessionData, response *http.Request) (*webauthn.Credential, error)
}

type WebAuthnHandler struct {
	storer   Storer
	webauthn WebAuthn
}

type RegistrationRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type LoginRequest struct {
	Name string `json:"name"`
}

func setCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*") // Adjust as necessary
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

// handleOptionsRequest checks if the request is an OPTIONS request and handles it.
func handleOptionsRequest(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return true
	}
	return false
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (h WebAuthnHandler) BeginRegistration(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	var request RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := NewUser(request.Name, request.DisplayName)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to generate user ID")
		return
	}

	options, session, err := h.webauthn.BeginRegistration(user)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to begin registration")
		return
	}

	user.session = session
	if err := h.storer.SaveUser(user); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to save user session")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(options)
}

func (h WebAuthnHandler) FinishRegistration(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	name := r.URL.Query().Get("userName")
	user, err := h.storer.GetUser(name)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "user not found")
		return
	}

	credential, err := h.webauthn.FinishRegistration(user, *user.session, r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	user.credentials = append(user.credentials, *credential)
	if err := h.storer.SaveUser(user); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to save user credentials")
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Registration Success"))
}

func (h WebAuthnHandler) BeginLogin(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if handleOptionsRequest(w, r) {
		return
	}

	var request LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.storer.GetUser(request.Name)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	options, session, err := h.webauthn.BeginLogin(user)
	if err != nil {
		http.Error(w, "Error initiating login", http.StatusInternalServerError)
		return
	}

	user.lastLoginSession = session
	if err := h.storer.SaveUser(user); err != nil {
		http.Error(w, "Error saving user session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(options)
}

func (h WebAuthnHandler) FinishLogin(w http.ResponseWriter, r *http.Request) {
	setCorsHeaders(w)
	if handleOptionsRequest(w, r) {
		return
	}

	name := r.URL.Query().Get("userName")
	user, err := h.storer.GetUser(name)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	session := user.lastLoginSession
	credential, err := h.webauthn.FinishLogin(user, *session, r)
	if err != nil {
		http.Error(w, "Error finishing login", http.StatusInternalServerError)
		return
	}

	user.credentials = append(user.credentials, *credential)
	if err := h.storer.SaveUser(user); err != nil {
		http.Error(w, "Error updating user credentials", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login Success"))
}
