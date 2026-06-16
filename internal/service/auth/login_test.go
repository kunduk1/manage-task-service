package auth

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func TestLogin_Success(t *testing.T) {
	svc, userRepo, tokenRepo := newTestService(t)

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Errorf("expected non-empty tokens, got %+v", tokens)
	}
	if tokens.ExpiresIn <= 0 {
		t.Errorf("expected positive expires_in, got %d", tokens.ExpiresIn)
	}
	if savedToken != tokens.RefreshToken {
		t.Errorf("refresh token saved (%q) does not match returned token (%q)", savedToken, tokens.RefreshToken)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	svc, userRepo, _ := newTestService(t)

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	userRepo.EXPECT().GetByEmail(gomock.Any(), "wp@example.com").
		Return(&model.User{ID: 1, Email: "wp@example.com", PasswordHash: string(hash)}, nil)

	_, err = svc.Login(context.Background(), model.LoginInput{
		Email: "wp@example.com", Password: "wrong-password",
	})
	if !stderrors.Is(err, errors.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	svc, userRepo, _ := newTestService(t)

	userRepo.EXPECT().GetByEmail(gomock.Any(), "ghost@example.com").Return(nil, errors.ErrUserNotFound)

	_, err := svc.Login(context.Background(), model.LoginInput{
		Email: "ghost@example.com", Password: "secret123",
	})
	if !stderrors.Is(err, errors.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_Validation(t *testing.T) {
	// Пустой email/пароль отсекается до репозитория — EXPECT не настраиваем.
	svc, _, _ := newTestService(t)

	cases := []model.LoginInput{
		{Email: "", Password: "secret123"},
		{Email: "a@b.c", Password: ""},
	}
	for i, in := range cases {
		if _, err := svc.Login(context.Background(), in); !stderrors.Is(err, errors.ErrValidation) {
			t.Errorf("case %d: expected ErrValidation, got %v", i, err)
		}
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
	if !stderrors.Is(err, errDB) {
		t.Errorf("expected propagated repo error, got %v", err)
	}
}

func TestLogin_SaveRefreshError(t *testing.T) {
	// Пароль верный, access выпущен, но сохранение refresh-токена падает — ошибка наружу.
	svc, userRepo, tokenRepo := newTestService(t)
	errDB := stderrors.New("redis down")

	hash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	userRepo.EXPECT().GetByEmail(gomock.Any(), "save@example.com").
		Return(&model.User{ID: 7, Email: "save@example.com", PasswordHash: string(hash)}, nil)
	tokenRepo.EXPECT().SaveRefresh(gomock.Any(), gomock.Any(), int64(7), gomock.Any()).Return(errDB)

	_, err = svc.Login(context.Background(), model.LoginInput{
		Email: "save@example.com", Password: "secret123",
	})
	if !stderrors.Is(err, errDB) {
		t.Errorf("expected propagated save-refresh error, got %v", err)
	}
}
