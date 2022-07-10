package dto

import (
	"github.com/google/uuid"
)

type LoginRequestDto struct {
	Email    string `json:"email" validate:"required,lte=60,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponseDto struct {
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	SessionID string    `json:"session_id" validate:"required"`
}
