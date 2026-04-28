package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (r *repository) Get(ctx context.Context, orderUUID string) (model.Order, error) {
	const query = `
		SELECT
			o.uuid,
			o.total_price,
			o.status,
			o.transaction_uuid,
			o.payment_method,
			o.created_at,
			MAX(CASE WHEN oi.part_type = 'HULL' THEN oi.part_uuid::text END) AS hull_uuid,
			MAX(CASE WHEN oi.part_type = 'ENGINE' THEN oi.part_uuid::text END) AS engine_uuid,
			MAX(CASE WHEN oi.part_type = 'SHIELD' THEN oi.part_uuid::text END) AS shield_uuid,
			MAX(CASE WHEN oi.part_type = 'WEAPON' THEN oi.part_uuid::text END) AS weapon_uuid
		FROM orders o
		LEFT JOIN order_items oi ON o.uuid = oi.order_uuid
		WHERE o.uuid = $1
		GROUP BY o.uuid, o.total_price, o.status, o.transaction_uuid, o.payment_method, o.created_at
	`

	rows, err := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, query, orderUUID)
	if err != nil {
		return model.Order{}, err
	}
	defer rows.Close()

	order, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[model.Order])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, errs.ErrOrderNotFound
		}
		return model.Order{}, fmt.Errorf("считать заказ: %w", err)
	}

	return order, nil
}
