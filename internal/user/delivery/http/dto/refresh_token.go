package dto

type UserRefreshTokenDto struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type UserRefreshTokenResponseDto struct {
	AccessToken  string `json:"access_token" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}
