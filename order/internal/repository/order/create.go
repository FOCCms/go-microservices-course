package order

import (
	"context"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
	"github.com/FOCCms/go-microservices-course/order/internal/repository/converter"
)

func (r *repository) Create(ctx context.Context, order model.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.UUID] = converter.ModelOrderToRecordOrder(order)
	return nil
}
