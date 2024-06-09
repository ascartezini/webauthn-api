package main

// curl 'https://zj5xojefyj.execute-api.us-east-1.amazonaws.com/dev/webauthn/registration/begin' \
// --data '{"username": "test", "displayName": "test"}' | jq
// import (
// 	"log"
// 	"net/http"

// 	"github.com/go-webauthn/webauthn/webauthn"
// )

// func main() {
// 	webAuthn, err := setupWebAuthn()
// 	if err != nil {
// 		log.Fatalf("Failed to initialize WebAuthn: %v", err)
// 	}

// 	s := NewStorer()

// 	setupRoutes(s, webAuthn)

// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }

// func setupWebAuthn() (*webauthn.WebAuthn, error) {
// 	wconfig := &webauthn.Config{
// 		RPDisplayName: "Go Webauthn",
// 		RPID:          "localhost",
// 		RPOrigins:     []string{"http://localhost:8000"},
// 	}

// 	webAuthn, err := webauthn.New(wconfig)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return webAuthn, nil
// }

// func setupRoutes(s Storer, webAuthn WebAuthn) {
// 	h := WebAuthnHandler{
// 		storer:   s,
// 		webauthn: webAuthn,
// 	}

// 	http.HandleFunc("/webauthn/registration/begin", h.BeginRegistration)
// 	http.HandleFunc("/webauthn/registration/finish", h.FinishRegistration)
// 	http.HandleFunc("/webauthn/login/begin", h.BeginLogin)
// 	http.HandleFunc("/webauthn/login/finish", h.FinishLogin)
// }

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/go-webauthn/webauthn/webauthn"
)

var webAuthn *webauthn.WebAuthn
var storer Storer

func init() {
	var err error
	webAuthn, err = setupWebAuthn()
	if err != nil {
		log.Fatalf("Failed to initialize WebAuthn: %v", err)
	}

	storer = NewStorer()
}

func setupWebAuthn() (*webauthn.WebAuthn, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: "Go Webauthn",
		RPID:          "localhost",
		RPOrigins:     []string{"http://localhost:8000"},
	}

	webAuthn, err := webauthn.New(wconfig)
	if err != nil {
		return nil, err
	}

	return webAuthn, nil
}

func setupRoutes() map[string]http.HandlerFunc {
	h := WebAuthnHandler{
		storer:   storer,
		webauthn: webAuthn,
	}

	routes := map[string]http.HandlerFunc{
		"/webauthn/registration/begin":  h.BeginRegistration,
		"/webauthn/registration/finish": h.FinishRegistration,
		"/webauthn/login/begin":         h.BeginLogin,
		"/webauthn/login/finish":        h.FinishLogin,
	}

	return routes
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	routes := setupRoutes()

	if route, exists := routes[req.Path]; exists {
		w := &responseWriter{
			headers: make(http.Header),
		}

		r := &http.Request{
			Method: req.HTTPMethod,
			URL:    &url.URL{Path: req.Path},
			Body:   io.NopCloser(strings.NewReader(req.Body)),
		}

		route(w, r)

		return events.APIGatewayProxyResponse{
			StatusCode: w.statusCode,
			Body:       w.body,
			Headers:    map[string]string{"Content-Type": "application/json"},
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNotFound,
		Body:       `{"message": "not found"}`,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}

type responseWriter struct {
	headers    http.Header
	statusCode int
	body       string
}

func (rw *responseWriter) Header() http.Header {
	return rw.headers
}

func (rw *responseWriter) Write(body []byte) (int, error) {
	rw.body = string(body)
	return len(body), nil
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

func main() {
	lambda.Start(handler)
}
