package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kunduk1/manage-task-service/internal/token"
)

func newToken(t *testing.T, mgr *token.Manager, userID int64) string {
	t.Helper()
	signed, _, err := mgr.GenerateAccess(userID, "u@e.com")
	require.NoError(t, err)
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

	require.True(t, nextCalled, "expected next handler to be called for a valid token")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, gotOK)
	assert.Equal(t, "42", gotUserID)
}

func TestAuth_MissingHeader(t *testing.T) {
	mgr := token.NewManager("test-secret", time.Minute)
	rec, nextCalled, _, _ := serve(mgr, "")

	assert.False(t, nextCalled, "next handler must not be called without an Authorization header")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_MalformedHeader(t *testing.T) {
	mgr := token.NewManager("test-secret", time.Minute)
	// Корректный токен, но без префикса "Bearer ".
	rec, nextCalled, _, _ := serve(mgr, "Token "+newToken(t, mgr, 1))

	assert.False(t, nextCalled, "next handler must not be called for a header without the Bearer prefix")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_InvalidToken(t *testing.T) {
	issuer := token.NewManager("secret-a", time.Minute)
	verifier := token.NewManager("secret-b", time.Minute) // другой секрет → подпись не сойдётся
	rec, nextCalled, _, _ := serve(verifier, "Bearer "+newToken(t, issuer, 1))

	assert.False(t, nextCalled, "next handler must not be called for an invalid token")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestAuth_ExpiredToken(t *testing.T) {
	mgr := token.NewManager("test-secret", -time.Minute) // exp в прошлом
	rec, nextCalled, _, _ := serve(mgr, "Bearer "+newToken(t, mgr, 1))

	assert.False(t, nextCalled, "next handler must not be called for an expired token")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestUserIDFromContext_Absent(t *testing.T) {
	v, ok := UserIDFromContext(t.Context())
	assert.False(t, ok)
	assert.Empty(t, v)
}
