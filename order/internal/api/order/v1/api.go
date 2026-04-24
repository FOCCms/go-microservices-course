package v1

import orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"

type api struct {
	orderv1.UnimplementedHandler

	orderService OrderService
}

func NewAPI(orderService OrderService) *api {
	return &api{
		orderService: orderService,
	}
}

// SetupServer создаёт OpenAPI сервер на основе обработчика.
func SetupServer(a *api) (*orderv1.Server, error) {
	return orderv1.NewServer(a)
}
