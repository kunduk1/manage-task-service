// Package errors содержит доменные ошибки, на которые завязан маппинг
// в HTTP-статусы на транспортном слое.
package errors

import "errors"

var (
	ErrUserExists           = errors.New("user already exists")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserNotFound         = errors.New("user not found")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrValidation           = errors.New("validation error")
	ErrCacheMiss            = errors.New("cache: key not found")
)
