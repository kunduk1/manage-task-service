package email

import (
	"context"

	"github.com/kunduk1/manage-task-service/internal/circuitbreaker"
)

// cbClient — декоратор почтового клиента, прогоняющий вызовы через брейкер.
// Реализует тот же интерфейс Client, поэтому для сервисного слоя брейкер прозрачен.
type cbClient struct {
	inner   Client
	breaker *circuitbreaker.Breaker
}

func NewCircuitBreaker(inner Client, breaker *circuitbreaker.Breaker) Client {
	return &cbClient{inner: inner, breaker: breaker}
}

func (c *cbClient) SendInvite(ctx context.Context, in Invite) error {
	return c.breaker.Execute(ctx, func() error {
		return c.inner.SendInvite(ctx, in)
	})
}

func (c *cbClient) Close() error { return c.inner.Close() }
