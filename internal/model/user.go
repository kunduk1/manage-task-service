package model

import (
	"time"
)

// User — доменная (service-level) сущность пользователя.
// Без тегов хранилища: персистентная модель живёт в internal/repository/user/model.
type User struct {
	ID           int64
	Email        string
	Name         string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64 // время жизни access-токена в секундах
}

type RegisterInput struct {
	Email    string
	Name     string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}
