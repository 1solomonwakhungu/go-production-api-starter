package model

import (
	"errors"
	"regexp"
	"time"
)

var (
	ErrInvalidEmail = errors.New("invalid email format")
	ErrNameTooShort = errors.New("name must be at least 2 characters")
)

// User represents the domain model for a user in the system.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserInput is the payload for creating a new user.
type CreateUserInput struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UpdateUserInput is the payload for updating an existing user.
type UpdateUserInput struct {
	Email *string `json:"email,omitempty"`
	Name  *string `json:"name,omitempty"`
}

// Validate checks the CreateUserInput for correctness.
func (i CreateUserInput) Validate() error {
	if len(i.Name) < 2 {
		return ErrNameTooShort
	}
	if !isValidEmail(i.Email) {
		return ErrInvalidEmail
	}
	return nil
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}
