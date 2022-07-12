package dto

import (
	"github.com/google/uuid"
)

type UserLoginRequestDto struct {
	Email    string `json:"email" validate:"required,lte=60,email"`
	Password string `json:"password" validate:"required"`
}

type UserLoginResponseDto struct {
	UserID uuid.UUID                    `json:"user_id" validate:"required"`
	Tokens *UserRefreshTokenResponseDto `json:"tokens" validate:"required"`
}
