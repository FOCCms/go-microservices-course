package part

import (
	"context"
	"fmt"
	"log/slog"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
)

func (s *service) Reserve(ctx context.Context, uuids []string) error {
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		parts, err := s.listForUpdate(ctx, valueobject.PartFilter{
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

		slog.Info("детали успешно зарезервированы",
			slog.Int("parts_count", len(parts)),
			slog.Any("part_uuids", uuids),
		)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
