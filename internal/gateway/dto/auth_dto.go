package dto

type RegisterHTTPRequest struct {
	FullName string `json:"full_name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginHTTPRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshHTTPRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
