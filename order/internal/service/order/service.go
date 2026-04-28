package order

type service struct {
	orderRepository     OrderRepository
	orderItemRepository OrderItemRepository
	paymentClient       PaymentClient
	inventoryClient     InventoryClient
	txManager           TxManager
}

func NewService(orderRepository OrderRepository, orderItemRepository OrderItemRepository, paymentClient PaymentClient, inventoryClient InventoryClient, txManager TxManager) *service {
	return &service{
		orderRepository:     orderRepository,
		orderItemRepository: orderItemRepository,
		paymentClient:       paymentClient,
		inventoryClient:     inventoryClient,
		txManager:           txManager,
	}
}
