package order

import (
	"context"
	"fmt"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
	"github.com/FOCCms/go-microservices-course/order/internal/repository/converter"
)

func (r *repository) Create(ctx context.Context, order model.Order) error {
	const query = `INSERT INTO orders (uuid, total_price, status, created_at) VALUES ($1, $2, $3, $4)`
	o := converter.ModelOrderToRecordOrder(order)

	_, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(ctx, query, o.UUID, o.TotalPrice, o.Status, o.CreatedAt)
	if err != nil {
		return fmt.Errorf("создать заказ: %w", err)
	}

	return nil
}
