package team

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/service/mocks"
	"github.com/kunduk1/manage-task-service/internal/token"
	"github.com/kunduk1/manage-task-service/internal/transport/middleware"
)

// newTeamHandler собирает хендлер с замоканным сервисом.
func newTeamHandler(t *testing.T) (*Handler, *mocks.MockTeamsService) {
	t.Helper()
	ctrl := gomock.NewController(t)
	svc := mocks.NewMockTeamsService(ctrl)
	return NewHandler(svc), svc
}

func newManager() *token.Manager {
	return token.NewManager("test-secret", time.Hour)
}

// authedRequest формирует запрос с валидным access-токеном пользователя uid.
func authedRequest(t *testing.T, method, target, body string, uid int64, mgr *token.Manager) *http.Request {
	t.Helper()
	tok, _, err := mgr.GenerateAccess(uid, "user@example.com")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	return req
}

// serveAuthed прогоняет запрос через middleware.Auth и хендлер.
func serveAuthed(h http.HandlerFunc, req *http.Request, mgr *token.Manager) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	middleware.Auth(mgr)(h).ServeHTTP(rec, req)
	return rec
}
