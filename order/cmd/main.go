package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/FOCCms/go-microservices-course/order/pkg/app"
	orderv1 "github.com/FOCCms/go-microservices-course/shared/pkg/openapi/order/v1"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

const (
	inventoryServiceAddress          = "localhost:50051"
	inventoryServiceKeepaliveTime    = 30 * time.Second
	inventoryServiceKeepaliveTimeout = 3 * time.Second

	paymentServiceAddress          = "localhost:50052"
	paymentServiceKeepaliveTime    = 30 * time.Second
	paymentServiceKeepaliveTimeout = 3 * time.Second

	httpPort          = "8080"
	shutdownTimeout   = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
)

func main() {
	if err := godotenv.Load("order.env"); err != nil {
		slog.Error("не удалось загрузить .env файл", "error", err)
		return
	}
	dsn := os.Getenv("DB_URI")

	// Создать gRPC соединение с InventoryService.
	inventoryConn, err := initInventoryConn()
	if err != nil {
		slog.Error("не удалось подключиться к InventoryService", "error", err)
		return
	}
	defer func() {
		if closeErr := inventoryConn.Close(); closeErr != nil {
			slog.Error("не удалось закрыть соединение с InventoryService", "error", closeErr)
		}
	}()

	// Создать gRPC клиент PaymentService.
	paymentConn, err := initPaymentConn()
	if err != nil {
		slog.Error("не удалось подключиться к PaymentService", "error", err)
		return
	}
	defer func() {
		if closeErr := paymentConn.Close(); closeErr != nil {
			slog.Error("не удалось закрыть соединение с PaymentService", "error", closeErr)
		}
	}()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Подключаем БД
	pool, txManager, err := initDBConnection(ctx, dsn)
	if err != nil {
		return
	}

	// Создать OpenAPI сервер.
	orderServer, err := app.NewHTTPHandler(pool, txManager, inventoryv1.NewInventoryServiceClient(inventoryConn), paymentv1.NewPaymentServiceClient(paymentConn))
	if err != nil {
		slog.Error("ошибка создания сервера OpenAPI", "error", err)
		return
	}

	server := initServer(orderServer)

	go func() {
		slog.Info("🚀 HTTP-сервер запущен на порту", "port", httpPort)
		listenErr := server.ListenAndServe()
		if listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			slog.Error("❌ ошибка запуска сервера", "error", listenErr)
			cancel() // будим main, чтобы не висеть бесконечно.
		}
	}()

	// Ждём сигнал от ОС или падение сервера.
	<-ctx.Done()
	slog.Info("🛑 завершение работы сервера...")
	// Создаем контекст с таймаутом для остановки сервера.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("❌ ошибка при остановке сервера", "error", err)
	}

	slog.Info("✅ сервер остановлен")
}

func initDBConnection(ctx context.Context, dsn string) (*pgxpool.Pool, *manager.Manager, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("создание пула соединений", "error", err)
		return nil, nil, err
	}
	defer pool.Close()

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		slog.Error("проверка соединения с БД", "error", err)
		return nil, nil, err
	}
	slog.Info("подключение к PostgreSQL установлено")

	// Создаём Transaction Manager для pgx
	txManager, err := manager.New(trmpgx.NewDefaultFactory(pool))
	if err != nil {
		slog.Error("создание transaction manager", "error", err)
		return nil, nil, err
	}
	return pool, txManager, nil
}

func initPaymentConn() (*grpc.ClientConn, error) {
	paymentConn, err := grpc.NewClient(paymentServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                paymentServiceKeepaliveTime,
			Timeout:             paymentServiceKeepaliveTimeout,
			PermitWithoutStream: true,
		}))
	return paymentConn, err
}

func initInventoryConn() (*grpc.ClientConn, error) {
	inventoryConn, err := grpc.NewClient(inventoryServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                inventoryServiceKeepaliveTime,
			Timeout:             inventoryServiceKeepaliveTimeout,
			PermitWithoutStream: true,
		}))
	return inventoryConn, err
}

func initServer(orderServer *orderv1.Server) *http.Server {
	return &http.Server{
		Addr:              net.JoinHostPort("localhost", httpPort),
		Handler:           orderServer,
		ReadHeaderTimeout: readHeaderTimeout, // Защита от Slowloris атаки
		ReadTimeout:       readTimeout,       // Лимит на чтение всего запроса
		WriteTimeout:      writeTimeout,      // Лимит на запись ответа
		IdleTimeout:       idleTimeout,       // Таймаут keep-alive соединений
	}
}
