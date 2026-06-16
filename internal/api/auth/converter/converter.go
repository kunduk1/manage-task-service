// Package converter переводит между транспортными DTO (pkg/auth/v1) и service-моделями (internal/model).
package converter

import (
	"github.com/kunduk1/manage-task-service/internal/model"
	authv1 "github.com/kunduk1/manage-task-service/pkg/auth/v1"
)

func ToRegisterInput(req authv1.RegisterRequest) model.RegisterInput {
	return model.RegisterInput{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	}
}

func ToRegisterResponse(u *model.User) authv1.RegisterResponse {
	return authv1.RegisterResponse{
		ID:    u.ID,
		Email: u.Email,
		Name:  u.Name,
	}
}

func ToLoginInput(req authv1.LoginRequest) model.LoginInput {
	return model.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}
}

func ToLoginResponse(t *model.AuthTokens) authv1.LoginResponse {
	return authv1.LoginResponse{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    t.ExpiresIn,
	}
}
