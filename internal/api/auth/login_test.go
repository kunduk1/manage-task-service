package auth

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	require.Equal(t, http.StatusOK, rec.Code)
	var resp authv1.LoginResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	assert.Equal(t, "access-x", resp.AccessToken)
	assert.Equal(t, "refresh-y", resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, int64(900), resp.ExpiresIn)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	h, svc := newAuthHandler(t)
	svc.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, errors.ErrInvalidCredentials)

	rec := doPost(h.Login, `{"email":"user@example.com","password":"wrong"}`)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestLogin_MalformedJSON(t *testing.T) {
	// Никаких EXPECT: при ошибке декодирования сервис не вызывается.
	h, _ := newAuthHandler(t)

	rec := doPost(h.Login, `{`)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
