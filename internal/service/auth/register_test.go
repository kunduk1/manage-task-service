package auth

import (
	"context"
	stderrors "errors"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func TestRegister_Success(t *testing.T) {
	svc, userRepo, _ := newTestService(t)

	// Пользователя ещё нет → сервис создаёт нового и берёт ID из репозитория.
	userRepo.EXPECT().GetByEmail(gomock.Any(), "user@example.com").Return(nil, errors.ErrUserNotFound)
	userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(1), nil)

	user, err := svc.Register(context.Background(), model.RegisterInput{
		Email:    "  USER@Example.com ",
		Name:     "Alice",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != 1 {
		t.Errorf("expected id from repository, got %d", user.ID)
	}
	if user.Email != "user@example.com" {
		t.Errorf("expected normalized email, got %q", user.Email)
	}
	if user.PasswordHash == "secret123" || user.PasswordHash == "" {
		t.Errorf("password must be hashed, got %q", user.PasswordHash)
	}
}

func TestRegister_Duplicate(t *testing.T) {
	svc, userRepo, _ := newTestService(t)
	in := model.RegisterInput{Email: "dup@example.com", Password: "secret123"}

	// Первая регистрация проходит, при второй пользователь уже существует.
	gomock.InOrder(
		userRepo.EXPECT().GetByEmail(gomock.Any(), "dup@example.com").Return(nil, errors.ErrUserNotFound),
		userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(1), nil),
		userRepo.EXPECT().GetByEmail(gomock.Any(), "dup@example.com").Return(&model.User{ID: 1, Email: "dup@example.com"}, nil),
	)

	if _, err := svc.Register(context.Background(), in); err != nil {
		t.Fatalf("unexpected error on first register: %v", err)
	}
	_, err := svc.Register(context.Background(), in)
	if !stderrors.Is(err, errors.ErrUserExists) {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestRegister_Validation(t *testing.T) {
	// Невалидный ввод отсекается до обращения к репозиторию — никаких EXPECT,
	// gomock провалит тест, если сервис всё же вызовет репозиторий.
	svc, _, _ := newTestService(t)

	cases := []model.RegisterInput{
		{Email: "", Password: "secret123"},
		{Email: "a@b.c", Password: ""},
		{Email: "a@b.c", Password: "123"}, // слишком короткий
	}
	for i, in := range cases {
		if _, err := svc.Register(context.Background(), in); !stderrors.Is(err, errors.ErrValidation) {
			t.Errorf("case %d: expected ErrValidation, got %v", i, err)
		}
	}
}

func TestRegister_LookupError(t *testing.T) {
	// Общая ошибка БД при предварительной проверке существования пробрасывается.
	svc, userRepo, _ := newTestService(t)
	errDB := stderrors.New("db unavailable")

	userRepo.EXPECT().GetByEmail(gomock.Any(), "look@example.com").Return(nil, errDB)

	_, err := svc.Register(context.Background(), model.RegisterInput{
		Email: "look@example.com", Password: "secret123",
	})
	if !stderrors.Is(err, errDB) {
		t.Errorf("expected propagated lookup error, got %v", err)
	}
}

func TestRegister_CreateError(t *testing.T) {
	// Пользователя нет, но вставка падает (например, гонка/UNIQUE) — ошибка наружу.
	svc, userRepo, _ := newTestService(t)
	errDB := stderrors.New("insert failed")

	gomock.InOrder(
		userRepo.EXPECT().GetByEmail(gomock.Any(), "create@example.com").Return(nil, errors.ErrUserNotFound),
		userRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(int64(0), errDB),
	)

	_, err := svc.Register(context.Background(), model.RegisterInput{
		Email: "create@example.com", Password: "secret123",
	})
	if !stderrors.Is(err, errDB) {
		t.Errorf("expected propagated create error, got %v", err)
	}
}

func TestRegister_HashError(t *testing.T) {
	// bcrypt не принимает пароль длиннее 72 байт — сервис оборачивает ошибку хеширования.
	svc, userRepo, _ := newTestService(t)

	userRepo.EXPECT().GetByEmail(gomock.Any(), "long@example.com").Return(nil, errors.ErrUserNotFound)

	longPassword := strings.Repeat("x", 73)
	_, err := svc.Register(context.Background(), model.RegisterInput{
		Email: "long@example.com", Password: longPassword,
	})
	if !stderrors.Is(err, bcrypt.ErrPasswordTooLong) {
		t.Errorf("expected wrapped bcrypt.ErrPasswordTooLong, got %v", err)
	}
}
