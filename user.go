package main

import (
	"crypto/rand"

	"github.com/go-webauthn/webauthn/webauthn"
)

type User struct {
	id               []byte
	name             string
	displayName      string
	session          *webauthn.SessionData
	lastLoginSession *webauthn.SessionData
	credentials      []webauthn.Credential
}

// WebAuthnID provides the user handle of the user account. A user handle is an opaque byte sequence with a maximum
// size of 64 bytes, and is not meant to be displayed to the user.
//
// To ensure secure operation, authentication and authorization decisions MUST be made on the basis of this id
// member, not the displayName nor name members. See Section 6.1 of [RFC8266].
//
// It's recommended this value is completely random and uses the entire 64 bytes.
//
// Specification: §5.4.3. User Account Parameters for Credential Generation (https://w3c.github.io/webauthn/#dom-publickeycredentialuserentity-id)
func (u User) WebAuthnID() []byte {
	return u.id
}

// WebAuthnName provides the name attribute of the user account during registration and is a human-palatable name for the user
// account, intended only for display. For example, "Alex Müller" or "田中倫". The Relying Party SHOULD let the user
// choose this, and SHOULD NOT restrict the choice more than necessary.
//
// Specification: §5.4.3. User Account Parameters for Credential Generation (https://w3c.github.io/webauthn/#dictdef-publickeycredentialuserentity)
func (u User) WebAuthnName() string {
	return u.name
}

// WebAuthnDisplayName provides the name attribute of the user account during registration and is a human-palatable
// name for the user account, intended only for display. For example, "Alex Müller" or "田中倫". The Relying Party
// SHOULD let the user choose this, and SHOULD NOT restrict the choice more than necessary.
//
// Specification: §5.4.3. User Account Parameters for Credential Generation (https://www.w3.org/TR/webauthn/#dom-publickeycredentialuserentity-displayname)
func (u User) WebAuthnDisplayName() string {
	return u.displayName
}

// WebAuthnCredentials provides the list of Credential objects owned by the user.
func (u User) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

func (u User) WebAuthnIcon() string {
	return "icon"
}

// generateID generates a new ID for the user
func (u *User) generateID() error {
	id := make([]byte, 32)
	if _, err := rand.Read(id); err != nil {
		return err
	}
	u.id = id
	return nil
}

// NewUser creates a new user with the given name and displayName, and generates an ID for the user
func NewUser(name, displayName string) (User, error) {
	user := User{name: name, displayName: displayName}
	if err := user.generateID(); err != nil {
		return User{}, err
	}
	return user, nil
}
