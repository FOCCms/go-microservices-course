package order

type service struct {
	orderRepository OrderRepository
	paymentClient   PaymentClient
	inventoryClient InventoryClient
	txManager       TxManager
}

func NewService(orderRepository OrderRepository, paymentClient PaymentClient, inventoryClient InventoryClient, txManager TxManager) *service {
	return &service{
		orderRepository: orderRepository,
		paymentClient:   paymentClient,
		inventoryClient: inventoryClient,
		txManager:       txManager,
	}
}
