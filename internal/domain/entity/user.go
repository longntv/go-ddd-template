package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserID is a unique identifier for a user.
type UserID = uuid.UUID

// User represents a user entity in the domain.
type User struct {
	ID    uuid.UUID
	Name  string
	Email string
	// PasswordHash is a one-way hash (see gateway.PasswordHasher), never the
	// plain-text password.
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUser creates a new User entity from an already hashed password.
func NewUser(name, email, passwordHash string) *User {
	now := time.Now()
	return &User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// Update updates the user's information. passwordHash must already be hashed.
func (u *User) Update(name, email, passwordHash string) {
	u.Name = name
	u.Email = email
	u.PasswordHash = passwordHash
	u.UpdatedAt = time.Now()
}
