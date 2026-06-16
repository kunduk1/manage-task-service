package response

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kunduk1/manage-task-service/pkg/errors"
)

// TestServiceError_StatusMapping проверяет маппинг доменных ошибок в HTTP-статусы.
func TestServiceError_StatusMapping(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "validation",
			err:        fmt.Errorf("%w: email and password are required", errors.ErrValidation),
			wantStatus: http.StatusBadRequest,
			wantMsg:    "validation error: email and password are required", // отдаём полный текст ошибки
		},
		{
			name:       "user exists",
			err:        errors.ErrUserExists,
			wantStatus: http.StatusConflict,
			wantMsg:    "user already exists",
		},
		{
			name:       "invalid credentials",
			err:        errors.ErrInvalidCredentials,
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "invalid email or password",
		},
		{
			name:       "forbidden",
			err:        errors.ErrForbidden,
			wantStatus: http.StatusForbidden,
			wantMsg:    "forbidden",
		},
		{
			name:       "team not found",
			err:        errors.ErrTeamNotFound,
			wantStatus: http.StatusNotFound,
			wantMsg:    "team not found",
		},
		{
			name:       "user not found",
			err:        errors.ErrUserNotFound,
			wantStatus: http.StatusNotFound,
			wantMsg:    "user not found",
		},
		{
			name:       "member exists",
			err:        errors.ErrMemberExists,
			wantStatus: http.StatusConflict,
			wantMsg:    "user is already a team member",
		},
		{
			name:       "unknown error falls back to 500",
			err:        stderrors.New("boom"),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "internal server error",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			ServiceError(rec, tc.err)

			if rec.Code != tc.wantStatus {
				t.Errorf("expected status %d, got %d", tc.wantStatus, rec.Code)
			}

			var body ErrorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("response body is not valid ErrorBody JSON: %v (body=%q)", err, rec.Body.String())
			}
			if body.Error != tc.wantMsg {
				t.Errorf("expected error message %q, got %q", tc.wantMsg, body.Error)
			}
		})
	}
}

// TestServiceError_ContentType — единый JSON-формат ответа об ошибке.
func TestServiceError_ContentType(t *testing.T) {
	rec := httptest.NewRecorder()
	ServiceError(rec, errors.ErrUserExists)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}
}
