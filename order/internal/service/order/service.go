package order

type service struct {
	orderRepository      OrderRepository
	paymentClient        PaymentClient
	inventoryClient      InventoryClient
	txManager            TxManager
	orderProducerService OrderProducerService
}

func NewService(orderRepository OrderRepository, paymentClient PaymentClient, inventoryClient InventoryClient, txManager TxManager, orderProducerService OrderProducerService) *service {
	return &service{
		orderRepository:      orderRepository,
		paymentClient:        paymentClient,
		inventoryClient:      inventoryClient,
		txManager:            txManager,
		orderProducerService: orderProducerService,
	}
}
