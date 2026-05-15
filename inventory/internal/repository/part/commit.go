package part

import (
	"context"
	"fmt"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
)

func (r *repository) Commit(ctx context.Context, uuids []string) error {
	const query = `UPDATE parts
		SET stock_quantity = stock_quantity - 1,
			reserved = reserved - 1
		WHERE uuid = ANY($1)
		  AND stock_quantity > 0
		  AND reserved > 0
	`

	tag, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(ctx, query, uuids)
	if err != nil {
		return fmt.Errorf("обновить строки при списании деталей: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("обновить строки при списании деталей: %w", errs.ErrCommitParts)
	}

	return nil
}
