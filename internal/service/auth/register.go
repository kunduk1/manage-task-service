package auth

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

const minPasswordLength = 6

func (s *serv) Register(ctx context.Context, in model.RegisterInput) (*model.User, error) {
	email := normalizeEmail(in.Email)
	if email == "" || in.Password == "" {
		return nil, fmt.Errorf("%w: email and password are required", errors.ErrValidation)
	}
	if len(in.Password) < minPasswordLength {
		return nil, fmt.Errorf("%w: password must be at least %d characters", errors.ErrValidation, minPasswordLength)
	}

	// Предварительная проверка существования (UNIQUE-индекс в БД — финальная страховка от гонки).
	switch _, err := s.userRepo.GetByEmail(ctx, email); {
	case err == nil:
		return nil, errors.ErrUserExists
	case !stderrors.Is(err, errors.ErrUserNotFound):
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		Email:        email,
		Name:         strings.TrimSpace(in.Name),
		PasswordHash: string(hash),
	}

	id, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}
	user.ID = id

	return user, nil
}
