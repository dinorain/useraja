package dto

type RefreshTokenDto struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshTokenResponseDto struct {
	AccessToken  string `json:"access_token" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}
