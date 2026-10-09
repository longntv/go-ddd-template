package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserID is a unique identifier for a user.
type UserID = uuid.UUID

// User represents a user entity in the domain.
type User struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewUser creates a new User entity.
func NewUser(name, email, password string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Name:      name,
		Email:     email,
		Password:  password,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Update updates the user's information.
func (u *User) Update(name, email, password string) {
	u.Name = name
	u.Email = email
	u.Password = password
	u.UpdatedAt = time.Now()
}
