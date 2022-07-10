package dto

import (
	"github.com/google/uuid"
)

type RegisterRequestDto struct {
	Email     string `json:"email" validate:"required,lte=60,email"`
	FirstName string `json:"first_name" validate:"required,lte=30"`
	LastName  string `json:"last_name" validate:"required,lte=30"`
	Password  string `json:"password" validate:"required"`
	Role      string `json:"role" validate:"required"`
	Avatar    string `json:"avatar"`
}

type RegisterResponseDto struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}
