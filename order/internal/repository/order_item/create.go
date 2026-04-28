package orderitem

import (
	"context"
	"fmt"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
)

func (r *repository) Create(ctx context.Context, items []model.OrderItem) error {
	const query = `INSERT INTO order_items (uuid, order_uuid, part_type, part_uuid, price) VALUES ($1, $2, $3, $4, $5)`

	for _, item := range items {
		_, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(ctx, query, item.UUID, item.OrderUUID, string(item.PartType), item.PartUUID, item.Price)
		if err != nil {
			return fmt.Errorf("создать деталь заказа: %w", err)
		}
	}
	return nil
}
