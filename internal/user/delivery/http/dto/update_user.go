package dto

type UserUpdateRequestDto struct {
	FirstName       *string `json:"first_name" validate:"omitempty,lte=30"`
	LastName        *string `json:"last_name" validate:"omitempty,lte=30"`
	Password        *string `json:"password" validate:"omitempty"`
	Avatar          *string `json:"avatar" validate:"omitempty"`
}
