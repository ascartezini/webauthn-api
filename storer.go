package main

import (
	"errors"
)

// ErrUserNotFound is returned when a user cannot be found in the store.
var ErrUserNotFound = errors.New("user not found")

// Storer manages user storage.
type InMemoryStorer struct {
	users map[string]User
}

// NewStorer creates a new instance of Storer.
func NewStorer() *InMemoryStorer {
	return &InMemoryStorer{users: make(map[string]User)}
}

// GetUser retrieves a user by name. Returns ErrUserNotFound if the user does not exist.
func (s *InMemoryStorer) GetUser(name string) (User, error) {
	user, ok := s.users[name]
	if !ok {
		return User{}, ErrUserNotFound
	}
	return user, nil
}

// SaveUser saves a user to the store.
func (s *InMemoryStorer) SaveUser(user User) error {
	s.users[user.name] = user
	return nil
}
