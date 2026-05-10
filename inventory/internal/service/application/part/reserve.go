package part

import (
	"context"
	"fmt"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
)

func (s *service) Reserve(ctx context.Context, uuids []string) error {
	parts, err := s.List(ctx, valueobject.PartFilter{
		UUIDs: uuids,
	})
	if err != nil {
		return fmt.Errorf("зарезервировать детали: %w", err)
	}

	for i := range parts {
		if err = parts[i].Reserve(); err != nil {
			return fmt.Errorf("зарезервировать детали: %w", errs.ErrOutOfStock)
		}
	}

	err = s.partRepository.UpdateReservationsBatch(ctx, parts)
	if err != nil {
		return fmt.Errorf("зарезервировать детали: %w", err)
	}

	return nil
}
