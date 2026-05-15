package assembly

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/FOCCms/go-microservices-course/assembly/internal/converter"
	"github.com/FOCCms/go-microservices-course/assembly/internal/model"
)

func (s *service) Assemble(ctx context.Context, event model.OrderPaidEvent) error {
	d := randomDuration(1, 3)

	time.Sleep(d * time.Second)

	err := s.assemblyProducerService.ProduceShipAssembled(ctx, converter.ToShipAssembledEvent(event, d, time.Now()))
	if err != nil {
		return fmt.Errorf("сообщить об успешной сборке корабля: %w", err)
	}

	return nil
}

func randomDuration(minSec, maxSec int) time.Duration {
	if minSec >= maxSec {
		return time.Duration(minSec)
	}
	// Генерируем случайное число в диапазоне [min, max]
	return time.Duration(rand.Intn(maxSec-minSec+1) + minSec)
}
