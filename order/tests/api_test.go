package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	invApp "github.com/FOCCms/go-microservices-course/inventory/pkg/app"
	"github.com/FOCCms/go-microservices-course/order/pkg/app"
	"github.com/FOCCms/go-microservices-course/order/tests/testutil"
	payApp "github.com/FOCCms/go-microservices-course/payment/pkg/app"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

// Предзагруженные UUID и цены деталей (из migrations/inventory/00002_seed_parts.sql).
const (
	HullAluminumUUID   = "550e8400-e29b-41d4-a716-446655440001" // 500000 kopecks (5000 RUB)
	HullTitaniumUUID   = "550e8400-e29b-41d4-a716-446655440002" // 1500000 kopecks (15000 RUB)
	EngineIonCUUID     = "550e8400-e29b-41d4-a716-446655440003" // 300000 kopecks (3000 RUB)
	EngineIonBUUID     = "550e8400-e29b-41d4-a716-446655440004" // 800000 kopecks (8000 RUB)
	ShieldEnergyUUID   = "550e8400-e29b-41d4-a716-446655440005" // 400000 kopecks (4000 RUB)
	WeaponLaserUUID    = "550e8400-e29b-41d4-a716-446655440006" // 250000 kopecks (2500 RUB)
	HullOutOfStockUUID = "550e8400-e29b-41d4-a716-446655440007" // 2000000 kopecks (20000 RUB), stock=0

	// Цены в копейках.
	HullAluminumPrice   = 500000
	HullTitaniumPrice   = 1500000
	EngineIonCPrice     = 300000
	EngineIonBPrice     = 800000
	ShieldEnergyPrice   = 400000
	WeaponLaserPrice    = 250000
	HullOutOfStockPrice = 2000000
)

const bufSize = 1024 * 1024

var (
	invLis *bufconn.Listener
	payLis *bufconn.Listener

	inventoryClient inventoryv1.InventoryServiceClient
	paymentClient   paymentv1.PaymentServiceClient
	httpClient      = testutil.NewHTTPClient()
	ts              *httptest.Server
)

func invBufDialer(context.Context, string) (net.Conn, error) {
	return invLis.Dial()
}

func payBufDialer(context.Context, string) (net.Conn, error) {
	return payLis.Dial()
}

// orderBaseURL возвращает базовый URL для HTTP тестов заказов.
func orderBaseURL() string {
	return ts.URL
}

// startPostgres запускает PostgreSQL контейнер и возвращает DSN подключения.
func startPostgres(ctx context.Context, dbName, user, password string) (*tcpostgres.PostgresContainer, string, error) {
	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase(dbName),
		tcpostgres.WithUsername(user),
		tcpostgres.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, "", err
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, "", err
	}

	return container, dsn, nil
}

// runMigrations запускает goose-миграции из указанной директории.
func runMigrations(dsn, migrationsDir string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		return err
	}

	return goose.Up(db, absDir)
}

// TestMain запускает все сервисы перед тестами и останавливает после.
func TestMain(m *testing.M) {
	ctx := context.Background()

	// 1. Запускаем PostgreSQL контейнер для order-сервиса
	orderContainer, orderDSN, err := startPostgres(ctx,
		"order-service",
		"order-service-user",
		"order-service-password",
	)
	if err != nil {
		panic(err)
	}

	// 2. Запускаем PostgreSQL контейнер для inventory-сервиса
	inventoryContainer, inventoryDSN, err := startPostgres(ctx,
		"inventory-service",
		"inventory-service-user",
		"inventory-service-password",
	)
	if err != nil {
		panic(err)
	}

	// 3. Накатываем миграции для order-сервиса
	if err = runMigrations(orderDSN, "../../migrations/order"); err != nil {
		panic(err)
	}

	// 4. Накатываем миграции для inventory-сервиса
	if err = runMigrations(inventoryDSN, "../../migrations/inventory"); err != nil {
		panic(err)
	}

	// 5. Создаём pgxpool для order-сервиса
	orderPool, err := pgxpool.New(ctx, orderDSN)
	if err != nil {
		panic(err)
	}

	// 6. Создаём pgxpool для inventory-сервиса
	inventoryPool, err := pgxpool.New(ctx, inventoryDSN)
	if err != nil {
		panic(err)
	}

	// 7. Создаём TxManager для order-сервиса
	txManager, err := manager.New(trmpgx.NewDefaultFactory(orderPool))
	if err != nil {
		panic(err)
	}

	// 8. Inventory gRPC через bufconn
	invLis = bufconn.Listen(bufSize)
	invGRPCServer := grpc.NewServer(invApp.Interceptors()...)
	invApp.RegisterServices(invGRPCServer, inventoryPool)
	go func() {
		if invServeErr := invGRPCServer.Serve(invLis); invServeErr != nil {
			panic(invServeErr)
		}
	}()

	invConn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(invBufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	inventoryClient = inventoryv1.NewInventoryServiceClient(invConn)

	// 9. Payment gRPC через bufconn
	payLis = bufconn.Listen(bufSize)
	payGRPCServer := grpc.NewServer(payApp.Interceptors()...)
	payApp.RegisterServices(payGRPCServer)
	go func() {
		if payServeErr := payGRPCServer.Serve(payLis); payServeErr != nil {
			panic(payServeErr)
		}
	}()

	payConn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(payBufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	paymentClient = paymentv1.NewPaymentServiceClient(payConn)

	// 10. Order HTTP через httptest
	orderServer, err := app.NewHTTPHandler(orderPool, txManager, inventoryClient, paymentClient)
	if err != nil {
		panic(err)
	}
	ts = httptest.NewServer(orderServer)

	code := m.Run()

	ts.Close()
	if err = invConn.Close(); err != nil {
		panic(err)
	}
	if err = payConn.Close(); err != nil {
		panic(err)
	}
	invGRPCServer.Stop()
	payGRPCServer.Stop()

	orderPool.Close()
	inventoryPool.Close()

	if err = orderContainer.Terminate(ctx); err != nil {
		panic(err)
	}
	if err = inventoryContainer.Terminate(ctx); err != nil {
		panic(err)
	}

	os.Exit(code)
}

// HTTP типы запросов/ответов.

// CreateOrderRequest представляет тело запроса для создания заказа.
type CreateOrderRequest struct {
	HullUUID   string  `json:"hull_uuid"`
	EngineUUID string  `json:"engine_uuid"`
	ShieldUUID *string `json:"shield_uuid,omitempty"`
	WeaponUUID *string `json:"weapon_uuid,omitempty"`
}

// CreateOrderResponse представляет ответ на создание заказа.
type CreateOrderResponse struct {
	OrderUUID  string `json:"order_uuid"`
	TotalPrice int64  `json:"total_price"`
}

// PayOrderRequest представляет тело запроса для оплаты заказа.
type PayOrderRequest struct {
	PaymentMethod string `json:"payment_method"`
}

// PayOrderResponse представляет ответ на оплату заказа.
type PayOrderResponse struct {
	TransactionUUID string `json:"transaction_uuid"`
}

// CancelOrderResponse представляет ответ на отмену заказа (пустой).
type CancelOrderResponse struct{}

// OrderDTO представляет заказ в ответе API.
type OrderDTO struct {
	OrderUUID       string  `json:"order_uuid"`
	HullUUID        string  `json:"hull_uuid"`
	EngineUUID      string  `json:"engine_uuid"`
	ShieldUUID      *string `json:"shield_uuid"`
	WeaponUUID      *string `json:"weapon_uuid"`
	TotalPrice      int64   `json:"total_price"`
	TransactionUUID *string `json:"transaction_uuid"`
	PaymentMethod   *string `json:"payment_method"`
	Status          string  `json:"status"`
	CreatedAt       string  `json:"created_at"`
}

// ErrorResponse представляет ответ с ошибкой от API.
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Вспомогательные HTTP функции.

func createOrder(t *testing.T, req *CreateOrderRequest) (*CreateOrderResponse, *http.Response) {
	t.Helper()

	jsonBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq, err := http.NewRequest(http.MethodPost, orderBaseURL()+"/api/v1/orders", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)

	if resp.StatusCode == http.StatusCreated {
		var result CreateOrderResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		return &result, resp
	}

	return nil, resp
}

func getOrder(t *testing.T, orderUUID string) (*OrderDTO, *http.Response) {
	t.Helper()

	resp, err := httpClient.Get(orderBaseURL() + "/api/v1/orders/" + orderUUID)
	require.NoError(t, err)

	if resp.StatusCode == http.StatusOK {
		var result OrderDTO
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		return &result, resp
	}

	return nil, resp
}

func payOrder(t *testing.T, orderUUID string, req *PayOrderRequest) (*PayOrderResponse, *http.Response) {
	t.Helper()

	jsonBody, err := json.Marshal(req)
	require.NoError(t, err)

	httpReq, err := http.NewRequest(http.MethodPost, orderBaseURL()+"/api/v1/orders/"+orderUUID+"/pay", bytes.NewReader(jsonBody))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)

	if resp.StatusCode == http.StatusOK {
		var result PayOrderResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		return &result, resp
	}

	return nil, resp
}

func cancelOrder(t *testing.T, orderUUID string) (*CancelOrderResponse, *http.Response) {
	t.Helper()

	httpReq, err := http.NewRequest(http.MethodPost, orderBaseURL()+"/api/v1/orders/"+orderUUID+"/cancel", nil)
	require.NoError(t, err)

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)

	if resp.StatusCode == http.StatusOK {
		var result CancelOrderResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		return &result, resp
	}

	return nil, resp
}

// Тесты InventoryService (gRPC).

func TestInventory_GetPart_Success(t *testing.T) {
	resp, err := inventoryClient.GetPart(context.Background(), &inventoryv1.GetPartRequest{
		Uuid: HullAluminumUUID,
	})
	require.NoError(t, err)

	part := resp.GetPart()
	assert.Equal(t, HullAluminumUUID, part.GetUuid())
	assert.Equal(t, int64(HullAluminumPrice), part.GetPrice())
	assert.Equal(t, inventoryv1.PartType_PART_TYPE_HULL, part.GetPartType())
	assert.NotEmpty(t, part.GetName())
	assert.NotEmpty(t, part.GetDescription(), "description должен быть заполнен")
	assert.NotNil(t, part.GetCreatedAt())
}

func TestInventory_GetPart_AllTypes(t *testing.T) {
	testCases := []struct {
		name        string
		uuid        string
		price       int64
		partType    inventoryv1.PartType
		description string
	}{
		{"Hull Aluminum", HullAluminumUUID, HullAluminumPrice, inventoryv1.PartType_PART_TYPE_HULL, "Лёгкий корпус для небольших кораблей"},
		{"Hull Titanium", HullTitaniumUUID, HullTitaniumPrice, inventoryv1.PartType_PART_TYPE_HULL, "Прочный корпус для средних кораблей"},
		{"Engine Ion C", EngineIonCUUID, EngineIonCPrice, inventoryv1.PartType_PART_TYPE_ENGINE, "Базовый ионный двигатель класса C"},
		{"Engine Ion B", EngineIonBUUID, EngineIonBPrice, inventoryv1.PartType_PART_TYPE_ENGINE, "Улучшенный ионный двигатель класса B"},
		{"Shield Energy", ShieldEnergyUUID, ShieldEnergyPrice, inventoryv1.PartType_PART_TYPE_SHIELD, "Стандартный энергетический щит"},
		{"Weapon Laser", WeaponLaserUUID, WeaponLaserPrice, inventoryv1.PartType_PART_TYPE_WEAPON, "Точная лазерная пушка"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := inventoryClient.GetPart(context.Background(), &inventoryv1.GetPartRequest{
				Uuid: tc.uuid,
			})
			require.NoError(t, err)

			part := resp.GetPart()
			assert.Equal(t, tc.uuid, part.GetUuid())
			assert.Equal(t, tc.price, part.GetPrice())
			assert.Equal(t, tc.partType, part.GetPartType())
			assert.Equal(t, tc.description, part.GetDescription())
		})
	}
}

func TestInventory_GetPart_NotFound(t *testing.T) {
	_, err := inventoryClient.GetPart(context.Background(), &inventoryv1.GetPartRequest{
		Uuid: uuid.New().String(),
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.NotFound)
}

func TestInventory_GetPart_EmptyUUID(t *testing.T) {
	_, err := inventoryClient.GetPart(context.Background(), &inventoryv1.GetPartRequest{
		Uuid: "",
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.InvalidArgument)
}

func TestInventory_GetPart_InvalidUUID(t *testing.T) {
	_, err := inventoryClient.GetPart(context.Background(), &inventoryv1.GetPartRequest{
		Uuid: "invalid-uuid-format",
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.InvalidArgument)
}

func TestInventory_ListParts_All(t *testing.T) {
	resp, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		PartType: inventoryv1.PartType_PART_TYPE_UNSPECIFIED,
	})
	require.NoError(t, err)
	assert.Len(t, resp.GetParts(), 7)
}

func TestInventory_ListParts_ByType_Hull(t *testing.T) {
	resp, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		PartType: inventoryv1.PartType_PART_TYPE_HULL,
	})
	require.NoError(t, err)
	assert.Len(t, resp.GetParts(), 3) // Алюминиевый, Титановый, Плазменный (stock=0)

	for _, part := range resp.GetParts() {
		assert.Equal(t, inventoryv1.PartType_PART_TYPE_HULL, part.GetPartType())
	}
}

func TestInventory_ListParts_ByType_Engine(t *testing.T) {
	resp, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		PartType: inventoryv1.PartType_PART_TYPE_ENGINE,
	})
	require.NoError(t, err)
	assert.Len(t, resp.GetParts(), 2)

	for _, part := range resp.GetParts() {
		assert.Equal(t, inventoryv1.PartType_PART_TYPE_ENGINE, part.GetPartType())
	}
}

func TestInventory_ListParts_ByType_Shield(t *testing.T) {
	resp, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		PartType: inventoryv1.PartType_PART_TYPE_SHIELD,
	})
	require.NoError(t, err)
	assert.Len(t, resp.GetParts(), 1)
	assert.Equal(t, ShieldEnergyUUID, resp.GetParts()[0].GetUuid())
}

func TestInventory_ListParts_ByType_Weapon(t *testing.T) {
	resp, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		PartType: inventoryv1.PartType_PART_TYPE_WEAPON,
	})
	require.NoError(t, err)
	assert.Len(t, resp.GetParts(), 1)
	assert.Equal(t, WeaponLaserUUID, resp.GetParts()[0].GetUuid())
}

func TestInventory_ListParts_SortedByName(t *testing.T) {
	resp, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		PartType: inventoryv1.PartType_PART_TYPE_UNSPECIFIED,
	})
	require.NoError(t, err)

	parts := resp.GetParts()
	for i := 1; i < len(parts); i++ {
		assert.LessOrEqual(t, parts[i-1].GetName(), parts[i].GetName(),
			"детали должны быть отсортированы по имени в алфавитном порядке")
	}
}

// Тесты ListParts.uuids.

func TestInventory_ListParts_ByUuids_Success(t *testing.T) {
	uuids := []string{HullAluminumUUID, EngineIonCUUID, ShieldEnergyUUID}

	resp, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		Uuids: uuids,
	})
	require.NoError(t, err)
	assert.Len(t, resp.GetParts(), 3)

	// Проверяем, что вернулись нужные детали
	returnedUUIDs := make([]string, len(resp.GetParts()))
	for i, part := range resp.GetParts() {
		returnedUUIDs[i] = part.GetUuid()
	}
	assert.ElementsMatch(t, uuids, returnedUUIDs)
}

func TestInventory_ListParts_ByUuids_PreservesOrder(t *testing.T) {
	// Запрос в определённом порядке: Engine, Hull, Weapon
	uuids := []string{EngineIonCUUID, HullAluminumUUID, WeaponLaserUUID}

	resp, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		Uuids: uuids,
	})
	require.NoError(t, err)
	assert.Len(t, resp.GetParts(), 3)

	// Проверяем, что порядок сохранён как в запросе
	for i, part := range resp.GetParts() {
		assert.Equal(t, uuids[i], part.GetUuid(),
			"деталь с индексом %d должна соответствовать порядку запрошенных UUID", i)
	}
}

func TestInventory_ListParts_ByUuids_IgnoresPartType(t *testing.T) {
	// Запрос с uuids И part_type — part_type должен быть проигнорирован
	uuids := []string{HullAluminumUUID, EngineIonCUUID}

	resp, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		Uuids:    uuids,
		PartType: inventoryv1.PartType_PART_TYPE_WEAPON, // Должен быть проигнорирован
	})
	require.NoError(t, err)
	assert.Len(t, resp.GetParts(), 2)

	// Проверяем, что получили Hull и Engine, а не Weapons
	assert.Equal(t, HullAluminumUUID, resp.GetParts()[0].GetUuid())
	assert.Equal(t, EngineIonCUUID, resp.GetParts()[1].GetUuid())
}

func TestInventory_ListParts_ByUuids_NotFound(t *testing.T) {
	// Включаем один несуществующий UUID
	nonExistentUUID := uuid.New().String()
	uuids := []string{HullAluminumUUID, nonExistentUUID, EngineIonCUUID}

	_, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		Uuids: uuids,
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.NotFound)
}

func TestInventory_ListParts_ByUuids_InvalidUUID(t *testing.T) {
	uuids := []string{HullAluminumUUID, "invalid-uuid-format"}

	_, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		Uuids: uuids,
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.InvalidArgument)
}

func TestInventory_ListParts_ByUuids_SingleUUID(t *testing.T) {
	uuids := []string{WeaponLaserUUID}

	resp, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		Uuids: uuids,
	})
	require.NoError(t, err)
	assert.Len(t, resp.GetParts(), 1)
	assert.Equal(t, WeaponLaserUUID, resp.GetParts()[0].GetUuid())
	assert.Equal(t, int64(WeaponLaserPrice), resp.GetParts()[0].GetPrice())
}

func TestInventory_ListParts_ByUuids_AllParts(t *testing.T) {
	// Запрашиваем все 6 деталей по UUID
	uuids := []string{
		HullAluminumUUID, HullTitaniumUUID,
		EngineIonCUUID, EngineIonBUUID,
		ShieldEnergyUUID, WeaponLaserUUID,
	}

	resp, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		Uuids: uuids,
	})
	require.NoError(t, err)
	assert.Len(t, resp.GetParts(), 6)

	// Проверяем, что порядок совпадает с порядком запроса
	for i, part := range resp.GetParts() {
		assert.Equal(t, uuids[i], part.GetUuid())
	}
}

func TestInventory_ListParts_ByUuids_EmptyList(t *testing.T) {
	// Пустой список UUID — должен вернуть все детали (фильтрация по типу UNSPECIFIED)
	resp, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		Uuids: []string{},
	})
	require.NoError(t, err)
	assert.Len(t, resp.GetParts(), 7)
}

// Тесты PaymentService (gRPC).

func TestPayment_PayOrder_Success_Card(t *testing.T) {
	resp, err := paymentClient.PayOrder(context.Background(), &paymentv1.PayOrderRequest{
		OrderUuid:     uuid.New().String(),
		PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_CARD,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.GetTransactionUuid())

	// Проверяем, что UUID транзакции валиден
	_, err = uuid.Parse(resp.GetTransactionUuid())
	assert.NoError(t, err)
}

func TestPayment_PayOrder_Success_SBP(t *testing.T) {
	resp, err := paymentClient.PayOrder(context.Background(), &paymentv1.PayOrderRequest{
		OrderUuid:     uuid.New().String(),
		PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_SBP,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.GetTransactionUuid())
}

func TestPayment_PayOrder_Success_CreditCard(t *testing.T) {
	resp, err := paymentClient.PayOrder(context.Background(), &paymentv1.PayOrderRequest{
		OrderUuid:     uuid.New().String(),
		PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.GetTransactionUuid())
}

func TestPayment_PayOrder_Success_InvestorMoney(t *testing.T) {
	resp, err := paymentClient.PayOrder(context.Background(), &paymentv1.PayOrderRequest{
		OrderUuid:     uuid.New().String(),
		PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_INVESTOR_MONEY,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, resp.GetTransactionUuid())
}

func TestPayment_PayOrder_EmptyOrderUUID(t *testing.T) {
	_, err := paymentClient.PayOrder(context.Background(), &paymentv1.PayOrderRequest{
		OrderUuid:     "",
		PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_CARD,
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.InvalidArgument)
}

func TestPayment_PayOrder_UnspecifiedMethod(t *testing.T) {
	_, err := paymentClient.PayOrder(context.Background(), &paymentv1.PayOrderRequest{
		OrderUuid:     uuid.New().String(),
		PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED,
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.InvalidArgument)
}

func TestPayment_PayOrder_UniqueTransactions(t *testing.T) {
	orderUUID := uuid.New().String()

	resp1, err := paymentClient.PayOrder(context.Background(), &paymentv1.PayOrderRequest{
		OrderUuid:     orderUUID,
		PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_CARD,
	})
	require.NoError(t, err)

	resp2, err := paymentClient.PayOrder(context.Background(), &paymentv1.PayOrderRequest{
		OrderUuid:     orderUUID,
		PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_CARD,
	})
	require.NoError(t, err)

	assert.NotEqual(t, resp1.GetTransactionUuid(), resp2.GetTransactionUuid(),
		"каждый платёж должен генерировать уникальный UUID транзакции")
}

// Тесты OrderService (HTTP).

func TestOrder_Create_Success_MinimalParts(t *testing.T) {
	req := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}

	result, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusCreated)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.OrderUUID)
	assert.Equal(t, int64(HullAluminumPrice+EngineIonCPrice), result.TotalPrice)
}

func TestOrder_Create_Success_AllParts(t *testing.T) {
	shieldUUID := ShieldEnergyUUID
	weaponUUID := WeaponLaserUUID
	req := &CreateOrderRequest{
		HullUUID:   HullTitaniumUUID,
		EngineUUID: EngineIonBUUID,
		ShieldUUID: &shieldUUID,
		WeaponUUID: &weaponUUID,
	}

	result, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusCreated)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.OrderUUID)

	expectedTotal := int64(HullTitaniumPrice + EngineIonBPrice + ShieldEnergyPrice + WeaponLaserPrice)
	assert.Equal(t, expectedTotal, result.TotalPrice)
}

func TestOrder_Create_VerifyTotalPrice(t *testing.T) {
	req := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID, // 500000
		EngineUUID: EngineIonCUUID,   // 300000
	}

	result, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusCreated)
	assert.Equal(t, int64(800000), result.TotalPrice, "500000 + 300000 = 800000")
}

func TestOrder_Create_HullNotFound(t *testing.T) {
	req := &CreateOrderRequest{
		HullUUID:   uuid.New().String(),
		EngineUUID: EngineIonCUUID,
	}

	_, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusNotFound)
}

func TestOrder_Create_EngineNotFound(t *testing.T) {
	req := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: uuid.New().String(),
	}

	_, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusNotFound)
}

func TestOrder_Create_ShieldNotFound(t *testing.T) {
	invalidShield := uuid.New().String()
	req := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
		ShieldUUID: &invalidShield,
	}

	_, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusNotFound)
}

func TestOrder_Create_WeaponNotFound(t *testing.T) {
	invalidWeapon := uuid.New().String()
	req := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
		WeaponUUID: &invalidWeapon,
	}

	_, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusNotFound)
}

func TestOrder_Get_Success(t *testing.T) {
	// Сначала создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// Получаем заказ
	order, resp := getOrder(t, createResult.OrderUUID)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusOK)
	require.NotNil(t, order)
	assert.Equal(t, createResult.OrderUUID, order.OrderUUID)
	assert.Equal(t, HullAluminumUUID, order.HullUUID)
	assert.Equal(t, EngineIonCUUID, order.EngineUUID)
	assert.Equal(t, createResult.TotalPrice, order.TotalPrice)
}

func TestOrder_Get_VerifyStatus_PendingPayment(t *testing.T) {
	// Создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// Получаем и проверяем статус
	order, resp := getOrder(t, createResult.OrderUUID)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusOK)
	assert.Equal(t, "PENDING_PAYMENT", order.Status)
}

func TestOrder_Get_NotFound(t *testing.T) {
	_, resp := getOrder(t, uuid.New().String())
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusNotFound)
}

func TestOrder_Pay_Success_Card(t *testing.T) {
	// Создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// Оплачиваем заказ
	payReq := &PayOrderRequest{PaymentMethod: "CARD"}
	payResult, resp := payOrder(t, createResult.OrderUUID, payReq)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusOK)
	require.NotNil(t, payResult)
	assert.NotEmpty(t, payResult.TransactionUUID)
}

func TestOrder_Pay_VerifyStatusChange(t *testing.T) {
	// Создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// Оплачиваем заказ
	payReq := &PayOrderRequest{PaymentMethod: "CARD"}
	_, payResp := payOrder(t, createResult.OrderUUID, payReq)
	_ = payResp.Body.Close()

	// Получаем и проверяем статус changed to PAID
	order, getResp := getOrder(t, createResult.OrderUUID)
	defer func() { _ = getResp.Body.Close() }()

	testutil.AssertHTTPStatus(t, getResp, http.StatusOK)
	assert.Equal(t, "PAID", order.Status)
	assert.NotNil(t, order.TransactionUUID)
	assert.NotNil(t, order.PaymentMethod)
	assert.Equal(t, "CARD", *order.PaymentMethod)
}

func TestOrder_Pay_NotFound(t *testing.T) {
	payReq := &PayOrderRequest{PaymentMethod: "CARD"}
	_, resp := payOrder(t, uuid.New().String(), payReq)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusNotFound)
}

func TestOrder_Pay_AlreadyPaid(t *testing.T) {
	// Создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// Оплачиваем заказ в первый раз
	payReq := &PayOrderRequest{PaymentMethod: "CARD"}
	_, payResp1 := payOrder(t, createResult.OrderUUID, payReq)
	_ = payResp1.Body.Close()

	// Пытаемся оплатить повторно — должна быть ошибка конфликта
	_, payResp2 := payOrder(t, createResult.OrderUUID, payReq)
	defer func() { _ = payResp2.Body.Close() }()

	testutil.AssertHTTPStatus(t, payResp2, http.StatusConflict)
}

func TestOrder_Pay_AlreadyCancelled(t *testing.T) {
	// Создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// Отменяем заказ
	_, cancelResp := cancelOrder(t, createResult.OrderUUID)
	_ = cancelResp.Body.Close()

	// Пытаемся оплатить отменённый заказ — должна быть ошибка конфликта
	payReq := &PayOrderRequest{PaymentMethod: "CARD"}
	_, payResp := payOrder(t, createResult.OrderUUID, payReq)
	defer func() { _ = payResp.Body.Close() }()

	testutil.AssertHTTPStatus(t, payResp, http.StatusConflict)
}

func TestOrder_Cancel_Success(t *testing.T) {
	// Создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// Отменяем заказ
	_, resp := cancelOrder(t, createResult.OrderUUID)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusOK)
}

func TestOrder_Cancel_VerifyStatusChange(t *testing.T) {
	// Создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// Отменяем заказ
	_, cancelResp := cancelOrder(t, createResult.OrderUUID)
	_ = cancelResp.Body.Close()

	// Получаем и проверяем статус changed to CANCELLED
	order, getResp := getOrder(t, createResult.OrderUUID)
	defer func() { _ = getResp.Body.Close() }()

	testutil.AssertHTTPStatus(t, getResp, http.StatusOK)
	assert.Equal(t, "CANCELLED", order.Status)
}

func TestOrder_Cancel_NotFound(t *testing.T) {
	_, resp := cancelOrder(t, uuid.New().String())
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusNotFound)
}

func TestOrder_Cancel_AlreadyPaid(t *testing.T) {
	// Создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// Оплачиваем заказ
	payReq := &PayOrderRequest{PaymentMethod: "CARD"}
	_, payResp := payOrder(t, createResult.OrderUUID, payReq)
	_ = payResp.Body.Close()

	// Пытаемся отменить оплаченный заказ — должна быть ошибка конфликта
	_, cancelResp := cancelOrder(t, createResult.OrderUUID)
	defer func() { _ = cancelResp.Body.Close() }()

	testutil.AssertHTTPStatus(t, cancelResp, http.StatusConflict)
}

func TestOrder_Cancel_AlreadyCancelled(t *testing.T) {
	// Создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// Отменяем заказ first time
	_, cancelResp1 := cancelOrder(t, createResult.OrderUUID)
	_ = cancelResp1.Body.Close()

	// Пытаемся отменить повторно — должна быть ошибка конфликта
	_, cancelResp2 := cancelOrder(t, createResult.OrderUUID)
	defer func() { _ = cancelResp2.Body.Close() }()

	testutil.AssertHTTPStatus(t, cancelResp2, http.StatusConflict)
}

// Дополнительные тесты валидации.

func TestOrder_Create_WithWeaponOnly(t *testing.T) {
	weaponUUID := WeaponLaserUUID
	req := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
		WeaponUUID: &weaponUUID,
	}

	result, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusCreated)
	require.NotNil(t, result)
	expectedTotal := int64(HullAluminumPrice + EngineIonCPrice + WeaponLaserPrice)
	assert.Equal(t, expectedTotal, result.TotalPrice)
}

func TestOrder_Pay_AllMethods(t *testing.T) {
	methods := []string{"CARD", "SBP", "CREDIT_CARD", "INVESTOR_MONEY"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Создаём заказ
			createReq := &CreateOrderRequest{
				HullUUID:   HullAluminumUUID,
				EngineUUID: EngineIonCUUID,
			}
			createResult, createResp := createOrder(t, createReq)
			_ = createResp.Body.Close()
			require.NotNil(t, createResult)

			// Оплачиваем этим методом
			payReq := &PayOrderRequest{PaymentMethod: method}
			payResult, resp := payOrder(t, createResult.OrderUUID, payReq)
			defer func() { _ = resp.Body.Close() }()

			testutil.AssertHTTPStatus(t, resp, http.StatusOK)
			require.NotNil(t, payResult)
			assert.NotEmpty(t, payResult.TransactionUUID)

			// Проверяем, что метод оплаты сохранён
			order, getResp := getOrder(t, createResult.OrderUUID)
			_ = getResp.Body.Close()
			require.NotNil(t, order.PaymentMethod)
			assert.Equal(t, method, *order.PaymentMethod)
		})
	}
}

func TestOrder_Get_WithOptionalParts(t *testing.T) {
	shieldUUID := ShieldEnergyUUID
	weaponUUID := WeaponLaserUUID
	req := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
		ShieldUUID: &shieldUUID,
		WeaponUUID: &weaponUUID,
	}

	createResult, createResp := createOrder(t, req)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// Получаем заказ и проверяем, что опциональные детали сохранены
	order, resp := getOrder(t, createResult.OrderUUID)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusOK)
	require.NotNil(t, order.ShieldUUID)
	require.NotNil(t, order.WeaponUUID)
	assert.Equal(t, shieldUUID, *order.ShieldUUID)
	assert.Equal(t, weaponUUID, *order.WeaponUUID)
}

func TestPayment_PayOrder_InvalidUUIDFormat(t *testing.T) {
	_, err := paymentClient.PayOrder(context.Background(), &paymentv1.PayOrderRequest{
		OrderUuid:     "invalid-uuid-format",
		PaymentMethod: paymentv1.PaymentMethod_PAYMENT_METHOD_CARD,
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.InvalidArgument)
}

// Тесты полного жизненного цикла.

func TestOrder_FullLifecycle_CreatePayGet(t *testing.T) {
	// 1. Создаём заказ
	shieldUUID := ShieldEnergyUUID
	createReq := &CreateOrderRequest{
		HullUUID:   HullTitaniumUUID,
		EngineUUID: EngineIonBUUID,
		ShieldUUID: &shieldUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)
	assert.NotEmpty(t, createResult.OrderUUID)

	expectedTotal := int64(HullTitaniumPrice + EngineIonBPrice + ShieldEnergyPrice)
	assert.Equal(t, expectedTotal, createResult.TotalPrice)

	// 2. Получаем заказ — проверяем PENDING_PAYMENT
	order1, getResp1 := getOrder(t, createResult.OrderUUID)
	_ = getResp1.Body.Close()
	assert.Equal(t, "PENDING_PAYMENT", order1.Status)
	assert.Nil(t, order1.TransactionUUID)

	// 3. Оплачиваем заказ
	payReq := &PayOrderRequest{PaymentMethod: "SBP"}
	payResult, payResp := payOrder(t, createResult.OrderUUID, payReq)
	_ = payResp.Body.Close()
	require.NotNil(t, payResult)
	assert.NotEmpty(t, payResult.TransactionUUID)

	// 4. Получаем заказ — проверяем PAID
	order2, getResp2 := getOrder(t, createResult.OrderUUID)
	defer func() { _ = getResp2.Body.Close() }()

	assert.Equal(t, "PAID", order2.Status)
	require.NotNil(t, order2.TransactionUUID)
	assert.Equal(t, payResult.TransactionUUID, *order2.TransactionUUID)
	require.NotNil(t, order2.PaymentMethod)
	assert.Equal(t, "SBP", *order2.PaymentMethod)
}

func TestOrder_FullLifecycle_CreateCancelGet(t *testing.T) {
	// 1. Создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// 2. Получаем заказ — проверяем PENDING_PAYMENT
	order1, getResp1 := getOrder(t, createResult.OrderUUID)
	_ = getResp1.Body.Close()
	assert.Equal(t, "PENDING_PAYMENT", order1.Status)

	// 3. Отменяем заказ
	_, cancelResp := cancelOrder(t, createResult.OrderUUID)
	_ = cancelResp.Body.Close()

	// 4. Получаем заказ — проверяем CANCELLED
	order2, getResp2 := getOrder(t, createResult.OrderUUID)
	defer func() { _ = getResp2.Body.Close() }()

	assert.Equal(t, "CANCELLED", order2.Status)
	assert.Nil(t, order2.TransactionUUID)
}

func TestOrder_FullLifecycle_AllPartsPayGet(t *testing.T) {
	// Полный жизненный цикл со всеми 4 деталями: hull + engine + shield + weapon
	shieldUUID := ShieldEnergyUUID
	weaponUUID := WeaponLaserUUID
	createReq := &CreateOrderRequest{
		HullUUID:   HullTitaniumUUID,
		EngineUUID: EngineIonBUUID,
		ShieldUUID: &shieldUUID,
		WeaponUUID: &weaponUUID,
	}

	// 1. Создаём заказ
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	expectedTotal := int64(HullTitaniumPrice + EngineIonBPrice + ShieldEnergyPrice + WeaponLaserPrice)
	assert.Equal(t, expectedTotal, createResult.TotalPrice)

	// 2. Проверяем все детали в GET ответе
	order1, getResp1 := getOrder(t, createResult.OrderUUID)
	_ = getResp1.Body.Close()
	assert.Equal(t, HullTitaniumUUID, order1.HullUUID)
	assert.Equal(t, EngineIonBUUID, order1.EngineUUID)
	require.NotNil(t, order1.ShieldUUID)
	assert.Equal(t, shieldUUID, *order1.ShieldUUID)
	require.NotNil(t, order1.WeaponUUID)
	assert.Equal(t, weaponUUID, *order1.WeaponUUID)

	// 3. Оплачиваем заказ
	payReq := &PayOrderRequest{PaymentMethod: "CREDIT_CARD"}
	payResult, payResp := payOrder(t, createResult.OrderUUID, payReq)
	_ = payResp.Body.Close()
	require.NotNil(t, payResult)

	// 4. Проверяем финальное состояние
	order2, getResp2 := getOrder(t, createResult.OrderUUID)
	defer func() { _ = getResp2.Body.Close() }()

	assert.Equal(t, "PAID", order2.Status)
	require.NotNil(t, order2.PaymentMethod)
	assert.Equal(t, "CREDIT_CARD", *order2.PaymentMethod)
}

// Тесты ogen-валидации (400 Bad Request).

func TestOrder_Create_InvalidBody_EmptyJSON(t *testing.T) {
	httpReq, err := http.NewRequest(http.MethodPost, orderBaseURL()+"/api/v1/orders", bytes.NewReader([]byte("{}")))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusBadRequest)
}

func TestOrder_Create_InvalidBody_NotJSON(t *testing.T) {
	httpReq, err := http.NewRequest(http.MethodPost, orderBaseURL()+"/api/v1/orders", bytes.NewReader([]byte("not json")))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusBadRequest)
}

func TestOrder_Create_InvalidBody_MissingHullUUID(t *testing.T) {
	body := `{"engine_uuid": "` + EngineIonCUUID + `"}`
	httpReq, err := http.NewRequest(http.MethodPost, orderBaseURL()+"/api/v1/orders", bytes.NewReader([]byte(body)))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusBadRequest)
}

func TestOrder_Create_InvalidBody_MissingEngineUUID(t *testing.T) {
	body := `{"hull_uuid": "` + HullAluminumUUID + `"}`
	httpReq, err := http.NewRequest(http.MethodPost, orderBaseURL()+"/api/v1/orders", bytes.NewReader([]byte(body)))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusBadRequest)
}

func TestOrder_Create_InvalidBody_InvalidHullUUID(t *testing.T) {
	body := `{"hull_uuid": "not-a-uuid", "engine_uuid": "` + EngineIonCUUID + `"}`
	httpReq, err := http.NewRequest(http.MethodPost, orderBaseURL()+"/api/v1/orders", bytes.NewReader([]byte(body)))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusBadRequest)
}

func TestOrder_Get_InvalidUUIDInPath(t *testing.T) {
	resp, err := httpClient.Get(orderBaseURL() + "/api/v1/orders/not-a-uuid")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusBadRequest)
}

func TestOrder_Pay_InvalidUUIDInPath(t *testing.T) {
	body := `{"payment_method": "CARD"}`
	httpReq, err := http.NewRequest(http.MethodPost, orderBaseURL()+"/api/v1/orders/not-a-uuid/pay", bytes.NewReader([]byte(body)))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusBadRequest)
}

func TestOrder_Pay_InvalidPaymentMethod(t *testing.T) {
	// Создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// Пытаемся оплатить невалидным методом — ogen отклонит
	body := `{"payment_method": "BITCOIN"}`
	httpReq, err := http.NewRequest(http.MethodPost,
		orderBaseURL()+"/api/v1/orders/"+createResult.OrderUUID+"/pay",
		bytes.NewReader([]byte(body)))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusBadRequest)
}

func TestOrder_Pay_MissingPaymentMethod(t *testing.T) {
	// Создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	// Пытаемся оплатить без payment_method
	body := `{}`
	httpReq, err := http.NewRequest(http.MethodPost,
		orderBaseURL()+"/api/v1/orders/"+createResult.OrderUUID+"/pay",
		bytes.NewReader([]byte(body)))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusBadRequest)
}

func TestOrder_Pay_EmptyBody(t *testing.T) {
	// Создаём заказ
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	httpReq, err := http.NewRequest(http.MethodPost,
		orderBaseURL()+"/api/v1/orders/"+createResult.OrderUUID+"/pay",
		bytes.NewReader([]byte("")))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusBadRequest)
}

func TestOrder_Cancel_InvalidUUIDInPath(t *testing.T) {
	httpReq, err := http.NewRequest(http.MethodPost, orderBaseURL()+"/api/v1/orders/not-a-uuid/cancel", nil)
	require.NoError(t, err)

	resp, err := httpClient.Do(httpReq)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusBadRequest)
}

// Тесты out of stock.

func TestOrder_Create_OutOfStock_Hull(t *testing.T) {
	// Плазменный корпус — stock_quantity=0, заказ должен быть отклонён.
	req := &CreateOrderRequest{
		HullUUID:   HullOutOfStockUUID,
		EngineUUID: EngineIonCUUID,
	}

	_, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusConflict)
}

func TestOrder_Create_OutOfStock_WithOptionalParts(t *testing.T) {
	// Out of stock деталь среди опциональных — shield.
	outOfStock := HullOutOfStockUUID
	req := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
		ShieldUUID: &outOfStock, // Передаём hull UUID как shield — тип не совпадёт, но out of stock проверяется первым.
	}

	_, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	// Либо Conflict (out of stock), либо другая ошибка — не 201.
	assert.NotEqual(t, http.StatusCreated, resp.StatusCode)
}

// Тесты Inventory: out of stock деталь присутствует в списке.

func TestInventory_GetPart_OutOfStock(t *testing.T) {
	resp, err := inventoryClient.GetPart(context.Background(), &inventoryv1.GetPartRequest{
		Uuid: HullOutOfStockUUID,
	})
	require.NoError(t, err)

	part := resp.GetPart()
	assert.Equal(t, HullOutOfStockUUID, part.GetUuid())
	assert.Equal(t, int64(HullOutOfStockPrice), part.GetPrice())
	assert.Equal(t, inventoryv1.PartType_PART_TYPE_HULL, part.GetPartType())
	assert.Equal(t, int64(0), part.GetStockQuantity())
	assert.Equal(t, "Экспериментальный корпус (нет на складе)", part.GetDescription())
}

func TestInventory_ListParts_ByUuids_IncludesOutOfStock(t *testing.T) {
	uuids := []string{HullAluminumUUID, HullOutOfStockUUID}

	resp, err := inventoryClient.ListParts(context.Background(), &inventoryv1.ListPartsRequest{
		Uuids: uuids,
	})
	require.NoError(t, err)
	assert.Len(t, resp.GetParts(), 2)

	// Out of stock деталь возвращается — фильтрации по наличию нет.
	assert.Equal(t, HullOutOfStockUUID, resp.GetParts()[1].GetUuid())
	assert.Equal(t, int64(0), resp.GetParts()[1].GetStockQuantity())
}

// Тесты Order: проверка created_at.

func TestOrder_Get_VerifyCreatedAt(t *testing.T) {
	createReq := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
	}
	createResult, createResp := createOrder(t, createReq)
	_ = createResp.Body.Close()
	require.NotNil(t, createResult)

	order, resp := getOrder(t, createResult.OrderUUID)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusOK)
	require.NotEmpty(t, order.CreatedAt, "created_at должен быть заполнен")

	// Парсим время — проверяем, что строка валидна и время не нулевое.
	createdAt, err := time.Parse(time.RFC3339Nano, order.CreatedAt)
	if err != nil {
		createdAt, err = time.Parse(time.RFC3339, order.CreatedAt)
	}
	if err != nil {
		createdAt, err = time.Parse("2006-01-02T15:04:05Z", order.CreatedAt)
	}
	require.NoError(t, err, "не удалось распарсить created_at: %s", order.CreatedAt)
	assert.False(t, createdAt.IsZero(), "created_at не должен быть нулевым")
}

// Тесты с shield only (без weapon).

func TestOrder_Create_WithShieldOnly(t *testing.T) {
	shieldUUID := ShieldEnergyUUID
	req := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonCUUID,
		ShieldUUID: &shieldUUID,
	}

	result, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusCreated)
	require.NotNil(t, result)
	expectedTotal := int64(HullAluminumPrice + EngineIonCPrice + ShieldEnergyPrice)
	assert.Equal(t, expectedTotal, result.TotalPrice)
}

// Тесты кросс-типовой валидации (хэндлер не проверяет тип детали).

func TestOrder_Create_WrongPartType_WeaponAsHull(t *testing.T) {
	// Хэндлер не валидирует, что UUID корпуса действительно является корпусом.
	// Передаём UUID оружия вместо корпуса — заказ должен создаться.
	req := &CreateOrderRequest{
		HullUUID:   WeaponLaserUUID,
		EngineUUID: EngineIonCUUID,
	}

	result, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusCreated)
	require.NotNil(t, result)
	expectedTotal := int64(WeaponLaserPrice + EngineIonCPrice)
	assert.Equal(t, expectedTotal, result.TotalPrice)
}

func TestOrder_Create_WrongPartType_HullAsEngine(t *testing.T) {
	// Аналогично — UUID корпуса вместо двигателя
	req := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: HullTitaniumUUID,
	}

	result, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusCreated)
	require.NotNil(t, result)
	expectedTotal := int64(HullAluminumPrice + HullTitaniumPrice)
	assert.Equal(t, expectedTotal, result.TotalPrice)
}

// Тесты с дубликатами UUID.

func TestOrder_Create_DuplicateUUID_HullAndEngine(t *testing.T) {
	// Передаём один и тот же UUID для hull и engine.
	// PostgreSQL WHERE uuid = ANY(array) дедуплицирует UUID, но listByUUIDs
	// итерирует по запрошенным UUID и находит каждый в recordMap — заказ создаётся
	// с одной и той же деталью дважды. Цена удваивается.
	req := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: HullAluminumUUID,
	}

	result, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusCreated)
	require.NotNil(t, result)
	assert.Equal(t, int64(HullAluminumPrice+HullAluminumPrice), result.TotalPrice)
}

// Тесты ValidateCompatibility (gRPC).

func TestInventory_ValidateCompatibility_Success_Compatible(t *testing.T) {
	// Алюминиевый корпус (strength=50) + Ионный двигатель C (required_strength=30) — совместимы.
	_, err := inventoryClient.ValidateCompatibility(context.Background(), &inventoryv1.ValidateCompatibilityRequest{
		Uuids: []string{HullAluminumUUID, EngineIonCUUID},
	})
	require.NoError(t, err)
}

func TestInventory_ValidateCompatibility_Success_StrongHull(t *testing.T) {
	// Титановый корпус (strength=150) + Ионный двигатель B (required_strength=70) — совместимы.
	_, err := inventoryClient.ValidateCompatibility(context.Background(), &inventoryv1.ValidateCompatibilityRequest{
		Uuids: []string{HullTitaniumUUID, EngineIonBUUID},
	})
	require.NoError(t, err)
}

func TestInventory_ValidateCompatibility_Success_AllParts(t *testing.T) {
	// Титановый корпус + Ion B + Energy shield + Laser — всё совместимо.
	_, err := inventoryClient.ValidateCompatibility(context.Background(), &inventoryv1.ValidateCompatibilityRequest{
		Uuids: []string{HullTitaniumUUID, EngineIonBUUID, ShieldEnergyUUID, WeaponLaserUUID},
	})
	require.NoError(t, err)
}

func TestInventory_ValidateCompatibility_Fail_WeakHull(t *testing.T) {
	// Алюминиевый корпус (strength=50) + Ионный двигатель B (required_strength=70) — несовместимы.
	// Корпус слишком слаб для двигателя класса B.
	_, err := inventoryClient.ValidateCompatibility(context.Background(), &inventoryv1.ValidateCompatibilityRequest{
		Uuids: []string{HullAluminumUUID, EngineIonBUUID},
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.FailedPrecondition)
}

func TestInventory_ValidateCompatibility_Success_HullOnly(t *testing.T) {
	// Только корпус — нет пары для проверки, ошибки не должно быть.
	_, err := inventoryClient.ValidateCompatibility(context.Background(), &inventoryv1.ValidateCompatibilityRequest{
		Uuids: []string{HullAluminumUUID},
	})
	require.NoError(t, err)
}

func TestInventory_ValidateCompatibility_Success_EngineOnly(t *testing.T) {
	// Только двигатель — нет пары для проверки, ошибки не должно быть.
	_, err := inventoryClient.ValidateCompatibility(context.Background(), &inventoryv1.ValidateCompatibilityRequest{
		Uuids: []string{EngineIonBUUID},
	})
	require.NoError(t, err)
}

func TestInventory_ValidateCompatibility_Success_ShieldAndWeapon_EnergyLaser(t *testing.T) {
	// Энергетический щит + лазерная пушка — совместимы (конфликт только у плазменного щита с лазером).
	_, err := inventoryClient.ValidateCompatibility(context.Background(), &inventoryv1.ValidateCompatibilityRequest{
		Uuids: []string{ShieldEnergyUUID, WeaponLaserUUID},
	})
	require.NoError(t, err)
}

func TestInventory_ValidateCompatibility_NotFound(t *testing.T) {
	// Несуществующий UUID — ошибка NotFound.
	_, err := inventoryClient.ValidateCompatibility(context.Background(), &inventoryv1.ValidateCompatibilityRequest{
		Uuids: []string{HullAluminumUUID, uuid.New().String()},
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.NotFound)
}

// Тесты ReserveParts (gRPC).

func TestInventory_ReserveParts_Success(t *testing.T) {
	// Резервируем доступные детали — ожидаем успех.
	_, err := inventoryClient.ReserveParts(context.Background(), &inventoryv1.ReservePartsRequest{
		Uuids: []string{HullAluminumUUID, EngineIonCUUID},
	})
	require.NoError(t, err)

	// Освобождаем обратно, чтобы не ломать другие тесты.
	_, err = inventoryClient.ReleaseParts(context.Background(), &inventoryv1.ReleasePartsRequest{
		Uuids: []string{HullAluminumUUID, EngineIonCUUID},
	})
	require.NoError(t, err)
}

func TestInventory_ReserveParts_OutOfStock(t *testing.T) {
	// Плазменный корпус (stock=0) — резервирование невозможно.
	_, err := inventoryClient.ReserveParts(context.Background(), &inventoryv1.ReservePartsRequest{
		Uuids: []string{HullOutOfStockUUID},
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.ResourceExhausted)
}

func TestInventory_ReserveParts_NotFound(t *testing.T) {
	_, err := inventoryClient.ReserveParts(context.Background(), &inventoryv1.ReservePartsRequest{
		Uuids: []string{uuid.New().String()},
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.NotFound)
}

func TestInventory_ReserveParts_SinglePart(t *testing.T) {
	// Резервируем одну деталь.
	_, err := inventoryClient.ReserveParts(context.Background(), &inventoryv1.ReservePartsRequest{
		Uuids: []string{ShieldEnergyUUID},
	})
	require.NoError(t, err)

	// Освобождаем обратно.
	_, err = inventoryClient.ReleaseParts(context.Background(), &inventoryv1.ReleasePartsRequest{
		Uuids: []string{ShieldEnergyUUID},
	})
	require.NoError(t, err)
}

// Тесты ReleaseParts (gRPC).

func TestInventory_ReleaseParts_Success(t *testing.T) {
	// Сначала резервируем, потом освобождаем — полный цикл.
	uuids := []string{HullTitaniumUUID, EngineIonBUUID}

	_, err := inventoryClient.ReserveParts(context.Background(), &inventoryv1.ReservePartsRequest{
		Uuids: uuids,
	})
	require.NoError(t, err)

	_, err = inventoryClient.ReleaseParts(context.Background(), &inventoryv1.ReleasePartsRequest{
		Uuids: uuids,
	})
	require.NoError(t, err)
}

func TestInventory_ReleaseParts_NothingToRelease(t *testing.T) {
	// Плазменный корпус (stock=0, reserved=0) — нечего освобождать.
	_, err := inventoryClient.ReleaseParts(context.Background(), &inventoryv1.ReleasePartsRequest{
		Uuids: []string{HullOutOfStockUUID},
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.FailedPrecondition)
}

func TestInventory_ReleaseParts_NotFound(t *testing.T) {
	_, err := inventoryClient.ReleaseParts(context.Background(), &inventoryv1.ReleasePartsRequest{
		Uuids: []string{uuid.New().String()},
	})
	require.Error(t, err)
	testutil.AssertGRPCStatus(t, err, codes.NotFound)
}

// Тесты Order Create с несовместимыми деталями (HTTP).

func TestOrder_Create_IncompatibleParts_WeakHullStrongEngine(t *testing.T) {
	// Алюминиевый корпус (strength=50) + Ионный двигатель B (required_strength=70).
	// Корпус не выдержит двигатель — ValidateCompatibility вернёт FailedPrecondition,
	// order-сервис преобразует в 409 Conflict.
	req := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonBUUID,
	}

	_, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusConflict)
}

func TestOrder_Create_IncompatibleParts_WithOptionalParts(t *testing.T) {
	// Алюминиевый корпус (strength=50) + Ионный двигатель B (required_strength=70) + Shield + Weapon.
	// Даже с опциональными деталями — несовместимость hull/engine блокирует создание.
	shieldUUID := ShieldEnergyUUID
	weaponUUID := WeaponLaserUUID
	req := &CreateOrderRequest{
		HullUUID:   HullAluminumUUID,
		EngineUUID: EngineIonBUUID,
		ShieldUUID: &shieldUUID,
		WeaponUUID: &weaponUUID,
	}

	_, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusConflict)
}

func TestOrder_Create_CompatibleParts_StrongHullStrongEngine(t *testing.T) {
	// Титановый корпус (strength=150) + Ионный двигатель B (required_strength=70) — совместимы.
	// Контрольный тест: при совместимых деталях заказ создаётся.
	req := &CreateOrderRequest{
		HullUUID:   HullTitaniumUUID,
		EngineUUID: EngineIonBUUID,
	}

	result, resp := createOrder(t, req)
	defer func() { _ = resp.Body.Close() }()

	testutil.AssertHTTPStatus(t, resp, http.StatusCreated)
	require.NotNil(t, result)
	assert.Equal(t, int64(HullTitaniumPrice+EngineIonBPrice), result.TotalPrice)
}

// Тесты Reserve/Release через полный жизненный цикл заказа.

func TestInventory_ReserveRelease_FullCycle(t *testing.T) {
	// Резервируем → освобождаем → снова резервируем — проверяем, что счётчики корректны.
	uuids := []string{WeaponLaserUUID}

	// Первый резерв.
	_, err := inventoryClient.ReserveParts(context.Background(), &inventoryv1.ReservePartsRequest{
		Uuids: uuids,
	})
	require.NoError(t, err)

	// Освобождаем.
	_, err = inventoryClient.ReleaseParts(context.Background(), &inventoryv1.ReleasePartsRequest{
		Uuids: uuids,
	})
	require.NoError(t, err)

	// Повторный резерв должен пройти (деталь снова доступна).
	_, err = inventoryClient.ReserveParts(context.Background(), &inventoryv1.ReservePartsRequest{
		Uuids: uuids,
	})
	require.NoError(t, err)

	// Финальное освобождение.
	_, err = inventoryClient.ReleaseParts(context.Background(), &inventoryv1.ReleasePartsRequest{
		Uuids: uuids,
	})
	require.NoError(t, err)
}
