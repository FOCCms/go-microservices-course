package part

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
)

func (s *service) Commit(ctx context.Context, uuids []string) error {
	err := s.txManager.Do(ctx, func(ctx context.Context) error {
		_, err := s.listForUpdate(ctx, valueobject.PartFilter{
			UUIDs: uuids,
		})
		if err != nil {
			return fmt.Errorf("списать детали: %w", err)
		}

		err = s.partRepository.Commit(ctx, uuids)
		if err != nil {
			return fmt.Errorf("списать детали: %w", err)
		}

		slog.Info("списание деталей со склада завершено",
			slog.Int("parts_count", len(uuids)), // Кол-во деталей
			slog.Any("part_uuids", uuids),       // Список деталей
		)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
