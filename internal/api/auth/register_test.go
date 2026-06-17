package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	require.Equal(t, http.StatusCreated, rec.Code)
	var resp authv1.RegisterResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, int64(42), resp.ID)
	assert.Equal(t, "user@example.com", resp.Email)
	assert.Equal(t, "Alice", resp.Name)
}

func TestRegister_Conflict(t *testing.T) {
	h, svc := newAuthHandler(t)
	svc.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil, errors.ErrUserExists)

	rec := doPost(h.Register, `{"email":"dup@example.com","password":"secret123"}`)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestRegister_MalformedJSON(t *testing.T) {
	// Никаких EXPECT: тело не парсится, сервис не должен вызываться (иначе gomock провалит тест).
	h, _ := newAuthHandler(t)

	rec := doPost(h.Register, `{`)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
