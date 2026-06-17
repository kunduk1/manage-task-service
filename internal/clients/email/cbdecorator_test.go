package email_test

import (
	stderrors "errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/kunduk1/manage-task-service/internal/circuitbreaker"
	"github.com/kunduk1/manage-task-service/internal/clients/email"
	"github.com/kunduk1/manage-task-service/internal/clients/email/mocks"
)

var errBackend = stderrors.New("email backend unavailable")

func newDecorated(t *testing.T, maxFailures int) (email.Client, *mocks.MockClient) {
	t.Helper()
	ctrl := gomock.NewController(t)
	inner := mocks.NewMockClient(ctrl)
	cb := circuitbreaker.New(circuitbreaker.Settings{
		Name:        "email-test",
		MaxFailures: maxFailures,
		OpenTimeout: time.Hour, // в тесте таймаут не истекает
	})
	return email.NewCircuitBreaker(inner, cb), inner
}

func TestDecorator_OpensAfterNAndFastFails(t *testing.T) {
	const maxFailures = 3
	client, inner := newDecorated(t, maxFailures)

	// Ровно maxFailures обращений к бэкенду; на (maxFailures+1)-м брейкер
	// разомкнут и inner вызываться НЕ должен — .Times гарантирует это.
	inner.EXPECT().SendInvite(gomock.Any(), gomock.Any()).Return(errBackend).Times(maxFailures)

	for i := 0; i < maxFailures; i++ {
		if err := client.SendInvite(t.Context(), email.Invite{}); !stderrors.Is(err, errBackend) {
			t.Fatalf("call %d: expected backend error, got %v", i, err)
		}
	}

	err := client.SendInvite(t.Context(), email.Invite{})
	if !stderrors.Is(err, circuitbreaker.ErrOpen) {
		t.Fatalf("expected ErrOpen after breaker trips, got %v", err)
	}
}

func TestDecorator_PassesThroughSuccess(t *testing.T) {
	client, inner := newDecorated(t, 3)

	want := email.Invite{ToEmail: "u@example.com", TeamID: 1, Role: "member"}
	inner.EXPECT().SendInvite(gomock.Any(), want).Return(nil)

	if err := client.SendInvite(t.Context(), want); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDecorator_CloseDelegates(t *testing.T) {
	client, inner := newDecorated(t, 3)
	inner.EXPECT().Close().Return(nil)

	if err := client.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
