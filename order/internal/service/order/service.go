package order

type Service struct {
	orderRepository      OrderRepository
	paymentClient        PaymentClient
	inventoryClient      InventoryClient
	txManager            TxManager
	orderProducerService OrderProducerService
}

func NewService(orderRepository OrderRepository, paymentClient PaymentClient, inventoryClient InventoryClient, txManager TxManager, orderProducerService OrderProducerService) *Service {
	return &Service{
		orderRepository:      orderRepository,
		paymentClient:        paymentClient,
		inventoryClient:      inventoryClient,
		txManager:            txManager,
		orderProducerService: orderProducerService,
	}
}
