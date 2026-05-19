package order

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
	"github.com/FOCCms/go-microservices-course/order/internal/repository/converter"
	"github.com/FOCCms/go-microservices-course/order/internal/repository/record"
)

func (r *repository) Get(ctx context.Context, orderUUID string) (model.Order, error) {
	const query = `
		SELECT uuid, total_price, status, transaction_uuid, payment_method, created_at, updated_at, user_uuid 
		FROM orders
		WHERE uuid = $1
	`
	return r.get(ctx, query, orderUUID)
}

func (r *repository) GetForUpdate(ctx context.Context, orderUUID string) (model.Order, error) {
	const query = `
		SELECT uuid, total_price, status, transaction_uuid, payment_method, created_at, updated_at, user_uuid 
		FROM orders
		WHERE uuid = $1
		FOR UPDATE
	`
	return r.get(ctx, query, orderUUID)
}

func (r *repository) get(ctx context.Context, query, orderUUID string) (model.Order, error) {
	rows, err := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, query, orderUUID)
	if err != nil {
		return model.Order{}, fmt.Errorf("считать заказ: %w", err)
	}
	defer rows.Close()

	order, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[record.Order])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, fmt.Errorf("считать заказ: %w", errs.ErrOrderNotFound)
		}
		return model.Order{}, fmt.Errorf("считать заказ: %w", err)
	}

	orderItems, err := r.getOrderItems(ctx, orderUUID)
	if err != nil {
		return model.Order{}, fmt.Errorf("считать заказ: %w", err)
	}

	return converter.OrderToModel(order, orderItems), nil
}

func (r *repository) getOrderItems(ctx context.Context, orderUUID string) ([]record.OrderItem, error) {
	const query = `
		SELECT uuid, order_uuid, part_uuid, part_type, price, created_at
		FROM order_items
		WHERE order_uuid = $1`

	rows, err := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, query, orderUUID)
	if err != nil {
		return []record.OrderItem{}, fmt.Errorf("считать список деталей: %w", err)
	}
	defer rows.Close()

	orderItems, err := pgx.CollectRows(rows, pgx.RowToStructByName[record.OrderItem])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return []record.OrderItem{}, fmt.Errorf("считать список деталей: %w", errs.ErrOrderNotFound)
		}
		return []record.OrderItem{}, fmt.Errorf("считать список деталей: %w", err)
	}

	return orderItems, nil
}
