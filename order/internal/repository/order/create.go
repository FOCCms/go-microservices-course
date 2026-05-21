package order

import (
	"context"
	"fmt"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
	"github.com/FOCCms/go-microservices-course/order/internal/repository/converter"
)

func (r *repository) Create(ctx context.Context, order model.Order, items []model.OrderItem) error {
	err := r.txManager.Do(ctx, func(ctx context.Context) error {
		if err := r.createOrder(ctx, order); err != nil {
			return err
		}
		if err := r.createOrderItems(ctx, items); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("создать заказ в транзакции: %w", err)
	}
	return nil
}

func (r *repository) createOrder(ctx context.Context, order model.Order) error {
	o := converter.ModelOrderToRecordOrder(order)
	const query = `INSERT INTO orders (uuid, total_price, status, created_at, user_uuid) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(ctx, query, o.UUID, o.TotalPrice, o.Status, o.CreatedAt, o.UserUUID)
	if err != nil {
		return fmt.Errorf("создать заказ: %w", err)
	}
	return nil
}

func (r *repository) createOrderItems(ctx context.Context, items []model.OrderItem) error {
	const query = `INSERT INTO order_items (uuid, order_uuid, part_type, part_uuid, price) VALUES ($1, $2, $3, $4, $5)`

	for _, item := range items {
		_, err := r.getter.DefaultTrOrDB(ctx, r.pool).Exec(ctx, query, item.UUID, item.OrderUUID, string(item.PartType), item.PartUUID, item.Price)
		if err != nil {
			return fmt.Errorf("создать деталь заказа: %w", err)
		}
	}
	return nil
}
