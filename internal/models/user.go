package models

import (
	"time"

	"github.com/google/uuid"
)

// User model
type User struct {
	UserID    uuid.UUID `json:"user_id" db:"user_id" validate:"omitempty"`
	Email     string    `json:"email" db:"email" validate:"omitempty,lte=60,email"`
	FirstName string    `json:"first_name" db:"first_name" validate:"required,lte=30"`
	LastName  string    `json:"last_name" db:"last_name" validate:"required,lte=30"`
	Role      string    `json:"role" db:"role" validate:"required"`
	Password  string    `json:"password,omitempty" db:"password"`
	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
}