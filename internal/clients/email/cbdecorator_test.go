package email_test

import (
	stderrors "errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
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
		err := client.SendInvite(t.Context(), email.Invite{})
		require.ErrorIs(t, err, errBackend)
	}

	err := client.SendInvite(t.Context(), email.Invite{})
	require.ErrorIs(t, err, circuitbreaker.ErrOpen)
}

func TestDecorator_PassesThroughSuccess(t *testing.T) {
	client, inner := newDecorated(t, 3)

	want := email.Invite{ToEmail: "u@example.com", TeamID: 1, Role: "member"}
	inner.EXPECT().SendInvite(gomock.Any(), want).Return(nil)

	require.NoError(t, client.SendInvite(t.Context(), want))
}

func TestDecorator_CloseDelegates(t *testing.T) {
	client, inner := newDecorated(t, 3)
	inner.EXPECT().Close().Return(nil)

	require.NoError(t, client.Close())
}
