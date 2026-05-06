package app

import (
	"context"
	"fmt"

	paymentV1API "github.com/FOCCms/go-microservices-course/payment/internal/api/payment/v1"
	paymentSrv "github.com/FOCCms/go-microservices-course/payment/internal/service/payment"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

type diContainer struct {
	paymentService   paymentV1API.PaymentService
	paymentV1Handler paymentv1.PaymentServiceServer
}

func (d *diContainer) PaymentService(_ context.Context) (paymentV1API.PaymentService, error) {
	if d.paymentService == nil {
		d.paymentService = paymentSrv.NewService()
	}
	return d.paymentService, nil
}

func (d *diContainer) PaymentV1API(ctx context.Context) (paymentv1.PaymentServiceServer, error) {
	if d.paymentV1Handler == nil {
		service, err := d.PaymentService(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать хендлер: %w", err)
		}
		d.paymentV1Handler = paymentV1API.NewAPI(service)
	}
	return d.paymentV1Handler, nil
}
