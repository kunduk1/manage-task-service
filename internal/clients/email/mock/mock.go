// Package mock — имитация внешнего почтового сервиса: искусственная задержка
// «сетевого» вызова плюс настраиваемая доля сбоев (чтобы брейкер можно
// было наблюдать в действии). Это НЕ gomock-мок для тестов — тот лежит в mocks/.
package mock

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"go.uber.org/zap"

	"github.com/kunduk1/manage-task-service/internal/clients/email"
	"github.com/kunduk1/manage-task-service/internal/config"
	"github.com/kunduk1/manage-task-service/internal/logger"
)

type sender struct {
	failRate float64       // доля искусственных сбоев, 0.0..1.0
	latency  time.Duration // искусственная задержка «сетевого» вызова
}

// NewSender создаёт имитацию почтового клиента по конфигурации.
func NewSender(cfg config.EmailConfig) email.Client {
	return &sender{
		failRate: cfg.FailRate(),
		latency:  cfg.Latency(),
	}
}

func (s *sender) SendInvite(ctx context.Context, in email.Invite) error {
	// Имитируем сетевую задержку, уважая отмену контекста.
	if s.latency > 0 {
		select {
		case <-time.After(s.latency):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// G404: это имитация сбоев бэкенда для демонстрации брейкера,
	// а не криптография — слабый ГПСЧ здесь уместен.
	if s.failRate > 0 && rand.Float64() < s.failRate { //nolint:gosec
		return fmt.Errorf("email backend unavailable")
	}

	logger.Info("invitation email sent (simulated)",
		zap.String("to", in.ToEmail),
		zap.Int64("team_id", in.TeamID),
		zap.String("role", in.Role),
	)
	return nil
}

func (s *sender) Close() error { return nil }
