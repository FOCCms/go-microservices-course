package assembly

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/FOCCms/go-microservices-course/assembly/internal/converter"
	"github.com/FOCCms/go-microservices-course/assembly/internal/model"
)

func (s *service) Assemble(ctx context.Context, event model.OrderPaidEvent) error {
	slog.Info("сборка космического корабля: начало",
		slog.String("order_uuid", event.OrderUUID.String()),
		slog.String("user_uuid", event.UserUUID.String()),
	)

	d := randomDurationSec(1, 3)

	select {
	case <-time.After(d):
	case <-ctx.Done():
		return fmt.Errorf("сборка корабля прервана: %w", ctx.Err())
	}

	err := s.assemblyProducerService.ProduceShipAssembled(ctx, converter.ToShipAssembledEvent(event, d, time.Now()))
	if err != nil {
		return fmt.Errorf("сообщить об успешной сборке корабля: %w", err)
	}

	slog.Info("сборка космического корабля: конец",
		slog.String("order_uuid", event.OrderUUID.String()),
		slog.Duration("assembly_duration", d),
	)

	return nil
}

func randomDurationSec(minSec, maxSec int) time.Duration {
	if minSec >= maxSec {
		return time.Duration(minSec) * time.Second
	}
	//nolint:gosec // math/rand/v2 используется осознанно для имитации задержки сборки, криптографическая стойкость не требуется.
	val := rand.IntN(maxSec-minSec+1) + minSec

	return time.Duration(val) * time.Second
}
