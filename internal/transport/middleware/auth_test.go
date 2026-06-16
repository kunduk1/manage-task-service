package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kunduk1/manage-task-service/internal/token"
)

func newToken(t *testing.T, mgr *token.Manager, userID int64) string {
	t.Helper()
	signed, _, err := mgr.GenerateAccess(userID, "u@e.com")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	return signed
}

// serve прогоняет запрос через Auth-middleware и репортит, дошёл ли он до next.
func serve(mgr *token.Manager, authHeader string) (rec *httptest.ResponseRecorder, nextCalled bool, gotUserID string, gotOK bool) {
	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	next := http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		nextCalled = true
		gotUserID, gotOK = UserIDFromContext(r.Context())
	})

	Auth(mgr)(next).ServeHTTP(rec, req)
	return rec, nextCalled, gotUserID, gotOK
}

func TestAuth_ValidToken(t *testing.T) {
	mgr := token.NewManager("test-secret", time.Minute)
	rec, nextCalled, gotUserID, gotOK := serve(mgr, "Bearer "+newToken(t, mgr, 42))

	if !nextCalled {
		t.Fatal("expected next handler to be called for a valid token")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if !gotOK || gotUserID != "42" {
		t.Errorf("expected userID %q in context, got %q (ok=%v)", "42", gotUserID, gotOK)
	}
}

func TestAuth_MissingHeader(t *testing.T) {
	mgr := token.NewManager("test-secret", time.Minute)
	rec, nextCalled, _, _ := serve(mgr, "")

	if nextCalled {
		t.Error("next handler must not be called without an Authorization header")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_MalformedHeader(t *testing.T) {
	mgr := token.NewManager("test-secret", time.Minute)
	// Корректный токен, но без префикса "Bearer ".
	rec, nextCalled, _, _ := serve(mgr, "Token "+newToken(t, mgr, 1))

	if nextCalled {
		t.Error("next handler must not be called for a header without the Bearer prefix")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	issuer := token.NewManager("secret-a", time.Minute)
	verifier := token.NewManager("secret-b", time.Minute) // другой секрет → подпись не сойдётся
	rec, nextCalled, _, _ := serve(verifier, "Bearer "+newToken(t, issuer, 1))

	if nextCalled {
		t.Error("next handler must not be called for an invalid token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	mgr := token.NewManager("test-secret", -time.Minute) // exp в прошлом
	rec, nextCalled, _, _ := serve(mgr, "Bearer "+newToken(t, mgr, 1))

	if nextCalled {
		t.Error("next handler must not be called for an expired token")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestUserIDFromContext_Absent(t *testing.T) {
	if v, ok := UserIDFromContext(t.Context()); ok || v != "" {
		t.Errorf("expected (\"\", false) for a context without userID, got (%q, %v)", v, ok)
	}
}
