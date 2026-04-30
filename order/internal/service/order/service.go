package order

type service struct {
	orderRepository OrderRepository
	paymentClient   PaymentClient
	inventoryClient InventoryClient
}

func NewService(orderRepository OrderRepository, paymentClient PaymentClient, inventoryClient InventoryClient) *service {
	return &service{
		orderRepository: orderRepository,
		paymentClient:   paymentClient,
		inventoryClient: inventoryClient,
	}
}
