package part

import (
	"context"
	"fmt"

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

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
