package part

import (
	"context"
	"fmt"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
)

func (s *service) Release(ctx context.Context, uuids []string) error {
	parts, err := s.List(ctx, valueobject.PartFilter{
		UUIDs: uuids,
	})
	if err != nil {
		return fmt.Errorf("освободить детали: %w", err)
	}

	for _, part := range parts {
		if err = part.Release(); err != nil {
			return fmt.Errorf("освободить детали: %w", errs.ErrOutOfStock)
		}
	}

	err = s.partRepository.UpdateReservedBatch(ctx, parts)
	if err != nil {
		return fmt.Errorf("освободить детали: %w", err)
	}

	return nil
}
