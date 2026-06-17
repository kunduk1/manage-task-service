package auth

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func TestLogin_Success(t *testing.T) {
	svc, userRepo, tokenRepo := newTestService(t)

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	require.NoError(t, err)
	userRepo.EXPECT().GetByEmail(gomock.Any(), "login@example.com").
		Return(&model.User{ID: 1, Email: "login@example.com", PasswordHash: string(hash)}, nil)

	// Перехватываем refresh-токен, переданный в репозиторий, чтобы сверить его с ответом.
	var savedToken string
	tokenRepo.EXPECT().SaveRefresh(gomock.Any(), gomock.Any(), int64(1), gomock.Any()).
		DoAndReturn(func(_ context.Context, tkn string, _ int64, _ time.Duration) error {
			savedToken = tkn
			return nil
		})

	tokens, err := svc.Login(context.Background(), model.LoginInput{
		Email: "login@example.com", Password: "secret123",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Positive(t, tokens.ExpiresIn)
	assert.Equal(t, tokens.RefreshToken, savedToken)
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, userRepo, _ := newTestService(t)

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	require.NoError(t, err)
	userRepo.EXPECT().GetByEmail(gomock.Any(), "wp@example.com").
		Return(&model.User{ID: 1, Email: "wp@example.com", PasswordHash: string(hash)}, nil)

	_, err = svc.Login(context.Background(), model.LoginInput{
		Email: "wp@example.com", Password: "wrong-password",
	})
	assert.ErrorIs(t, err, errors.ErrInvalidCredentials)
}

func TestLogin_UnknownUser(t *testing.T) {
	svc, userRepo, _ := newTestService(t)

	userRepo.EXPECT().GetByEmail(gomock.Any(), "ghost@example.com").Return(nil, errors.ErrUserNotFound)

	_, err := svc.Login(context.Background(), model.LoginInput{
		Email: "ghost@example.com", Password: "secret123",
	})
	assert.ErrorIs(t, err, errors.ErrInvalidCredentials)
}

func TestLogin_Validation(t *testing.T) {
	// Пустой email/пароль отсекается до репозитория — EXPECT не настраиваем.
	svc, _, _ := newTestService(t)

	cases := []model.LoginInput{
		{Email: "", Password: "secret123"},
		{Email: "a@b.c", Password: ""},
	}
	for _, in := range cases {
		_, err := svc.Login(context.Background(), in)
		assert.ErrorIs(t, err, errors.ErrValidation)
	}
}

func TestLogin_RepoError(t *testing.T) {
	// Ошибка репозитория, отличная от ErrUserNotFound, пробрасывается как есть.
	svc, userRepo, _ := newTestService(t)
	errDB := stderrors.New("db unavailable")

	userRepo.EXPECT().GetByEmail(gomock.Any(), "err@example.com").Return(nil, errDB)

	_, err := svc.Login(context.Background(), model.LoginInput{
		Email: "err@example.com", Password: "secret123",
	})
	assert.ErrorIs(t, err, errDB)
}

func TestLogin_SaveRefreshError(t *testing.T) {
	// Пароль верный, access выпущен, но сохранение refresh-токена падает — ошибка наружу.
	svc, userRepo, tokenRepo := newTestService(t)
	errDB := stderrors.New("redis down")

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	require.NoError(t, err)
	userRepo.EXPECT().GetByEmail(gomock.Any(), "save@example.com").
		Return(&model.User{ID: 7, Email: "save@example.com", PasswordHash: string(hash)}, nil)
	tokenRepo.EXPECT().SaveRefresh(gomock.Any(), gomock.Any(), int64(7), gomock.Any()).Return(errDB)

	_, err = svc.Login(context.Background(), model.LoginInput{
		Email: "save@example.com", Password: "secret123",
	})
	assert.ErrorIs(t, err, errDB)
}
