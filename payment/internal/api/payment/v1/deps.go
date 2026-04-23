package v1

import (
	"context"

	"github.com/FOCCms/go-microservices-course/payment/internal/model"
)

type PaymentService interface {
	Pay(ctx context.Context, req model.PayRequest) (string, error)
}
