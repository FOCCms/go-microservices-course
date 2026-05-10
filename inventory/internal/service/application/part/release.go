package part

import (
	"context"
	"fmt"

	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
)

func (s *service) Release(ctx context.Context, uuids []string) error {
	parts, err := s.List(ctx, valueobject.PartFilter{
		UUIDs: uuids,
	})
	if err != nil {
		return fmt.Errorf("освободить детали: %w", err)
	}

	for i := range parts {
		if err = parts[i].Release(); err != nil {
			return fmt.Errorf("освободить детали: %w", err)
		}
	}

	err = s.partRepository.UpdateReservationsBatch(ctx, parts)
	if err != nil {
		return fmt.Errorf("освободить детали: %w", err)
	}

	return nil
}
