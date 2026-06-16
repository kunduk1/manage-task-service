package auth

import (
	"context"
	stderrors "errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func (s *serv) Login(ctx context.Context, in model.LoginInput) (*model.AuthTokens, error) {
	email := normalizeEmail(in.Email)
	if email == "" || in.Password == "" {
		return nil, fmt.Errorf("%w: email and password are required", errors.ErrValidation)
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if stderrors.Is(err, errors.ErrUserNotFound) {
			return nil, errors.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)); err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	accessToken, expiresIn, err := s.jwtManager.GenerateAccess(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	refreshToken := uuid.NewString()
	if err := s.tokenRepo.SaveRefresh(ctx, refreshToken, user.ID, s.refreshTTL); err != nil {
		return nil, err
	}

	return &model.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	}, nil
}
