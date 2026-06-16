// Package v1 — публичный контракт транспортного уровня для домена auth (HTTP/JSON).
// Это модели 1-го уровня: их видит только API-слой и (потенциально) внешние клиенты.
package v1

type RegisterRequest struct {
	Email    string `json:"email"    example:"user@example.com"`
	Name     string `json:"name"     example:"Jane Doe"`
	Password string `json:"password" example:"s3cret-pass"`
}

type RegisterResponse struct {
	ID    int64  `json:"id"    example:"42"`
	Email string `json:"email" example:"user@example.com"`
	Name  string `json:"name"  example:"Jane Doe"`
}

type LoginRequest struct {
	Email    string `json:"email"    example:"user@example.com"`
	Password string `json:"password" example:"s3cret-pass"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"  example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType    string `json:"token_type"    example:"Bearer"`
	ExpiresIn    int64  `json:"expires_in"    example:"900"`
}
