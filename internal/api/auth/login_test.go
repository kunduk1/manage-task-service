package auth

import (
	"encoding/json"
	"net/http"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	authv1 "github.com/kunduk1/manage-task-service/pkg/auth/v1"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

func TestLogin_Success(t *testing.T) {
	h, svc := newAuthHandler(t)

	svc.EXPECT().
		Login(gomock.Any(), model.LoginInput{Email: "user@example.com", Password: "secret123"}).
		Return(&model.AuthTokens{AccessToken: "access-x", RefreshToken: "refresh-y", ExpiresIn: 900}, nil)

	rec := doPost(h.Login, `{"email":"user@example.com","password":"secret123"}`)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (body=%q)", rec.Code, rec.Body.String())
	}
	var resp authv1.LoginResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if resp.AccessToken != "access-x" || resp.RefreshToken != "refresh-y" {
		t.Errorf("unexpected tokens in response: %+v", resp)
	}
	if resp.TokenType != "Bearer" {
		t.Errorf("expected token_type Bearer, got %q", resp.TokenType)
	}
	if resp.ExpiresIn != 900 {
		t.Errorf("expected expires_in 900, got %d", resp.ExpiresIn)
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	h, svc := newAuthHandler(t)
	svc.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, errors.ErrInvalidCredentials)

	rec := doPost(h.Login, `{"email":"user@example.com","password":"wrong"}`)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestLogin_MalformedJSON(t *testing.T) {
	// Никаких EXPECT: при ошибке декодирования сервис не вызывается.
	h, _ := newAuthHandler(t)

	rec := doPost(h.Login, `{`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed JSON, got %d", rec.Code)
	}
}
