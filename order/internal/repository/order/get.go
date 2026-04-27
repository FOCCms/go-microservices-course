package order

import (
	"context"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
	"github.com/FOCCms/go-microservices-course/order/internal/repository/converter"
)

func (r *repository) Get(_ context.Context, uuid uuid.UUID) (model.Order, error) {
	r.mu.RLock()
	order, ok := r.orders[uuid]
	r.mu.RUnlock()

	if !ok {
		return model.Order{}, errs.ErrOrderNotFound
	}

	return converter.OrderToModel(order), nil
}
