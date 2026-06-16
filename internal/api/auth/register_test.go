package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/model"
	"github.com/kunduk1/manage-task-service/internal/service/mocks"
	authv1 "github.com/kunduk1/manage-task-service/pkg/auth/v1"
	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// newAuthHandler собирает хендлер с замоканным сервисом.
// gomock.NewController(t) сам проверяет ожидания по завершении теста.
func newAuthHandler(t *testing.T) (*Handler, *mocks.MockAuthService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockAuthService(ctrl)
	return NewHandler(svc), svc
}

func doPost(handler http.HandlerFunc, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	handler(rec, req)
	return rec
}

func TestRegister_Success(t *testing.T) {
	h, svc := newAuthHandler(t)

	// Конвертер прокидывает поля DTO напрямую — проверяем, что сервис получает их без искажений.
	svc.EXPECT().
		Register(gomock.Any(), model.RegisterInput{Email: "user@example.com", Name: "Alice", Password: "secret123"}).
		Return(&model.User{ID: 42, Email: "user@example.com", Name: "Alice"}, nil)

	rec := doPost(h.Register, `{"email":"user@example.com","name":"Alice","password":"secret123"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (body=%q)", rec.Code, rec.Body.String())
	}
	var resp authv1.RegisterResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if resp.ID != 42 || resp.Email != "user@example.com" || resp.Name != "Alice" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestRegister_Conflict(t *testing.T) {
	h, svc := newAuthHandler(t)
	svc.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil, errors.ErrUserExists)

	rec := doPost(h.Register, `{"email":"dup@example.com","password":"secret123"}`)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rec.Code)
	}
}

func TestRegister_MalformedJSON(t *testing.T) {
	// Никаких EXPECT: тело не парсится, сервис не должен вызываться (иначе gomock провалит тест).
	h, _ := newAuthHandler(t)

	rec := doPost(h.Register, `{`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed JSON, got %d", rec.Code)
	}
}
