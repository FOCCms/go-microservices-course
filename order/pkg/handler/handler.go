package handler

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

// Order представляет заказ на постройку космического корабля.
type Order struct {
	OrderUUID       uuid.UUID
	HullUUID        uuid.UUID
	EngineUUID      uuid.UUID
	ShieldUUID      *uuid.UUID // опциональный
	WeaponUUID      *uuid.UUID // опциональный
	TotalPrice      int64      // в копейках
	TransactionUUID *uuid.UUID
	PaymentMethod   *string
	Status          string // PENDING_PAYMENT, PAID, CANCELLED
	CreatedAt       time.Time
}

// OrderStore — хранилище заказов (in-memory).
type OrderStore struct {
	mu     sync.RWMutex
	orders map[uuid.UUID]Order
}

// NewOrderStore создаёт новое пустое хранилище заказов.
func NewOrderStore() *OrderStore {
	return &OrderStore{
		orders: make(map[uuid.UUID]Order),
	}
}

// OrderHandler реализует интерфейс orderv1.Handler, сгенерированный ogen.
type OrderHandler struct {
	orderv1.UnimplementedHandler
	inventoryClient inventoryv1.InventoryServiceClient
	paymentClient   paymentv1.PaymentServiceClient
	store           *OrderStore
}

// NewOrderHandler создаёт новый обработчик заказов.
func NewOrderHandler(
	inventoryClient inventoryv1.InventoryServiceClient,
	paymentClient paymentv1.PaymentServiceClient,
	store *OrderStore,
) *OrderHandler {
	return &OrderHandler{
		inventoryClient: inventoryClient,
		paymentClient:   paymentClient,
		store:           store,
	}
}

// SetupServer создаёт OpenAPI сервер на основе обработчика.
func SetupServer(h *OrderHandler) (*orderv1.Server, error) {
	return orderv1.NewServer(h)
}

// GetOrder реализует операцию getOrder (пример реализации).
// GET /api/v1/orders/{order_uuid}.
func (h *OrderHandler) GetOrder(_ context.Context, params orderv1.GetOrderParams) (orderv1.GetOrderRes, error) {
	// 1. Найти заказ в store (с блокировкой для thread-safety).
	h.store.mu.RLock()
	order, ok := h.store.orders[params.OrderUUID]
	h.store.mu.RUnlock()

	// 2. Если не найден — вернуть 404.
	if !ok {
		return &orderv1.GetOrderNotFound{
			Code:    http.StatusNotFound,
			Message: "заказ не найден",
		}, nil
	}

	// 3. Преобразовать в DTO и вернуть.
	var shieldUUID orderv1.OptNilUUID
	if order.ShieldUUID != nil {
		shieldUUID = orderv1.NewOptNilUUID(*order.ShieldUUID)
	}

	var weaponUUID orderv1.OptNilUUID
	if order.WeaponUUID != nil {
		weaponUUID = orderv1.NewOptNilUUID(*order.WeaponUUID)
	}

	var transactionUUID orderv1.OptNilUUID
	if order.TransactionUUID != nil {
		transactionUUID = orderv1.NewOptNilUUID(*order.TransactionUUID)
	}

	var paymentMethod orderv1.OptNilPaymentMethod
	if order.PaymentMethod != nil {
		paymentMethod = orderv1.NewOptNilPaymentMethod(orderv1.PaymentMethod(*order.PaymentMethod))
	}

	return &orderv1.OrderDto{
		OrderUUID:       order.OrderUUID,
		HullUUID:        order.HullUUID,
		EngineUUID:      order.EngineUUID,
		ShieldUUID:      shieldUUID,
		WeaponUUID:      weaponUUID,
		TotalPrice:      order.TotalPrice,
		TransactionUUID: transactionUUID,
		PaymentMethod:   paymentMethod,
		Status:          orderv1.OrderStatus(order.Status),
		CreatedAt:       order.CreatedAt,
	}, nil
}

// CreateOrder реализует операцию createOrder.
// POST /api/v1/orders.
func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (orderv1.CreateOrderRes, error) {
	if req.HullUUID == uuid.Nil || req.EngineUUID == uuid.Nil {
		return &orderv1.CreateOrderBadRequest{
			Code:    http.StatusBadRequest,
			Message: "hull_uuid и engine_uuid обязательны",
		}, nil
	}

	// Собираем список UUID компонентов.
	parts := []string{req.HullUUID.String(), req.EngineUUID.String()}
	if req.ShieldUUID.IsSet() {
		parts = append(parts, req.ShieldUUID.Value.String())
	}
	if req.WeaponUUID.IsSet() {
		parts = append(parts, req.WeaponUUID.Value.String())
	}

	// Получаем список компонентов.
	partsRes, err := h.inventoryClient.ListParts(ctx, &inventoryv1.ListPartsRequest{Uuids: parts})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return &orderv1.CreateOrderInternalServerError{
				Code:    http.StatusInternalServerError,
				Message: "внутренняя ошибка сервера",
			}, nil
		}

		switch st.Code() {
		case codes.NotFound:
			return &orderv1.CreateOrderNotFound{
				Code:    http.StatusNotFound,
				Message: "компонент не найден",
			}, nil
		case codes.InvalidArgument:
			return &orderv1.CreateOrderBadRequest{
				Code:    http.StatusBadRequest,
				Message: "передан некорректный UUID компонента",
			}, nil
		default:
			return &orderv1.CreateOrderInternalServerError{
				Code:    http.StatusInternalServerError,
				Message: "ошибка inventory сервиса: " + st.Message(),
			}, nil
		}
	}

	// Проверяем, что все компоненты найдены.
	if len(partsRes.GetParts()) != len(parts) {
		return &orderv1.CreateOrderNotFound{
			Code:    http.StatusNotFound,
			Message: "компонент не найден",
		}, nil
	}

	// Вычисляем общую стоимость.
	var totalPrice int64 = 0
	for _, part := range partsRes.GetParts() {
		stockQuantity := part.GetStockQuantity()

		if stockQuantity <= 0 {
			return &orderv1.CreateOrderConflict{
				Code:    http.StatusConflict,
				Message: "компонент отсутствует на складе: " + part.GetName(),
			}, nil
		}

		totalPrice += part.GetPrice()
	}

	// Создаем заказ
	orderUUID := uuid.New()

	order := &Order{
		OrderUUID:  orderUUID,
		HullUUID:   req.HullUUID,
		EngineUUID: req.EngineUUID,
		TotalPrice: totalPrice,
		Status:     "PENDING_PAYMENT",
		CreatedAt:  time.Now(),
	}

	if req.ShieldUUID.IsSet() {
		order.ShieldUUID = new(req.ShieldUUID.Value)
	}
	if req.WeaponUUID.IsSet() {
		order.WeaponUUID = new(req.WeaponUUID.Value)
	}

	h.store.mu.Lock()
	h.store.orders[orderUUID] = *order
	h.store.mu.Unlock()

	return &orderv1.CreateOrderResponse{
		OrderUUID:  orderUUID,
		TotalPrice: totalPrice,
	}, nil
}

// PayOrder реализует операцию payOrder.
// POST /api/v1/orders/{order_uuid}/pay.
func (h *OrderHandler) PayOrder(ctx context.Context, req *orderv1.PayOrderRequest, params orderv1.PayOrderParams) (orderv1.PayOrderRes, error) {
	h.store.mu.Lock()
	order, ok := h.store.orders[params.OrderUUID]
	h.store.mu.Unlock()

	if !ok {
		return &orderv1.PayOrderNotFound{
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("заказ %s не найден", params.OrderUUID),
		}, nil
	}

	if order.Status != "PENDING_PAYMENT" {
		return &orderv1.PayOrderConflict{
			Code:    http.StatusConflict,
			Message: fmt.Sprintf("невозможно оплатить заказ, статус %s не PENDING_PAYMENT", order.Status),
		}, nil
	}

	transaction, err := h.paymentClient.PayOrder(ctx, &paymentv1.PayOrderRequest{
		OrderUuid:     params.OrderUUID.String(),
		PaymentMethod: orderPaymentMethodToPaymentMethod(req.PaymentMethod),
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return &orderv1.PayOrderInternalServerError{
				Code:    http.StatusInternalServerError,
				Message: "внутренняя ошибка сервера",
			}, nil
		}

		return &orderv1.PayOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "ошибка payment сервиса: %s" + st.Message(),
		}, nil
	}

	transactionUUID, err := uuid.Parse(transaction.GetTransactionUuid())
	if err != nil {
		return &orderv1.PayOrderInternalServerError{
			Code:    http.StatusInternalServerError,
			Message: "внутренняя ошибка сервера",
		}, nil
	}

	h.store.mu.Lock()
	defer h.store.mu.Unlock()

	currentOrder, ok := h.store.orders[params.OrderUUID]
	if !ok || currentOrder.Status != "PENDING_PAYMENT" {
		return &orderv1.PayOrderConflict{
			Code:    http.StatusConflict,
			Message: fmt.Sprintf("заказ %s был изменён во время оплаты", params.OrderUUID),
		}, nil
	}

	pm := string(req.PaymentMethod)
	currentOrder.PaymentMethod = &pm

	currentOrder.Status = "PAID"

	currentOrder.TransactionUUID = &transactionUUID

	h.store.orders[params.OrderUUID] = currentOrder

	return &orderv1.PayOrderResponse{
		TransactionUUID: transactionUUID,
	}, nil
}

// CancelOrder реализует операцию cancelOrder.
// POST /api/v1/orders/{order_uuid}/cancel.
func (h *OrderHandler) CancelOrder(ctx context.Context, params orderv1.CancelOrderParams) (orderv1.CancelOrderRes, error) {
	h.store.mu.Lock()
	defer h.store.mu.Unlock()
	order, ok := h.store.orders[params.OrderUUID]

	if !ok {
		return &orderv1.CancelOrderNotFound{
			Code:    http.StatusNotFound,
			Message: "заказ не найден",
		}, nil
	}

	if order.Status != "PENDING_PAYMENT" {
		return &orderv1.CancelOrderConflict{
			Code:    http.StatusConflict,
			Message: fmt.Sprintf("невозможно отменить заказ, статус %s не PENDING_PAYMENT", order.Status),
		}, nil
	}

	order.Status = "CANCELLED"
	h.store.orders[params.OrderUUID] = order

	return &orderv1.CancelOrderResponse{}, nil
}

func orderPaymentMethodToPaymentMethod(method orderv1.PaymentMethod) paymentv1.PaymentMethod {
	switch method {
	case orderv1.PaymentMethodCARD:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_CARD
	case orderv1.PaymentMethodSBP:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_SBP
	case orderv1.PaymentMethodCREDITCARD:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD
	case orderv1.PaymentMethodINVESTORMONEY:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_INVESTOR_MONEY
	default:
		return paymentv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED
	}
}
