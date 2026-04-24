package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/FOCCms/go-microservices-course/order/pkg/app"
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
	// Создать gRPC соединение с InventoryService
	inventoryConn, err := grpc.NewClient(inventoryServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                inventoryServiceKeepaliveTime,
			Timeout:             inventoryServiceKeepaliveTimeout,
			PermitWithoutStream: true,
		}))
	if err != nil {
		slog.Error("не удалось подключиться к InventoryService", "error", err)
		return
	}
	defer func() {
		if closeErr := inventoryConn.Close(); closeErr != nil {
			slog.Error("не удалось закрыть соединение с InventoryService", "error", closeErr)
		}
	}()

	// Создать gRPC клиент PaymentService
	paymentConn, err := grpc.NewClient(paymentServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                paymentServiceKeepaliveTime,
			Timeout:             paymentServiceKeepaliveTimeout,
			PermitWithoutStream: true,
		}))
	if err != nil {
		slog.Error("не удалось подключиться к PaymentService", "error", err)
		return
	}
	defer func() {
		if closeErr := paymentConn.Close(); closeErr != nil {
			slog.Error("не удалось закрыть соединение с PaymentService", "error", closeErr)
		}
	}()

	// Создать OpenAPI сервер
	orderServer, err := app.NewHTTPHandler(inventoryv1.NewInventoryServiceClient(inventoryConn), paymentv1.NewPaymentServiceClient(paymentConn))
	if err != nil {
		slog.Error("ошибка создания сервера OpenAPI", "error", err)
		return
	}

	server := &http.Server{
		Addr:              net.JoinHostPort("localhost", httpPort),
		Handler:           orderServer,
		ReadHeaderTimeout: readHeaderTimeout, // Защита от Slowloris атаки
		ReadTimeout:       readTimeout,       // Лимит на чтение всего запроса
		WriteTimeout:      writeTimeout,      // Лимит на запись ответа
		IdleTimeout:       idleTimeout,       // Таймаут keep-alive соединений
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		slog.Info("🚀 HTTP-сервер запущен на порту", "port", httpPort)
		listenErr := server.ListenAndServe()
		if listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
			slog.Error("❌ ошибка запуска сервера", "error", listenErr)
			cancel() // будим main, чтобы не висеть бесконечно
		}
	}()

	// Ждём сигнал от ОС или падение сервера.
	<-ctx.Done()

	slog.Info("🛑 завершение работы сервера...")

	// Создаем контекст с таймаутом для остановки сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	err = server.Shutdown(shutdownCtx)
	if err != nil {
		slog.Error("❌ ошибка при остановке сервера", "error", err)
	}

	slog.Info("✅ сервер остановлен")
}

// const (
// 	inventoryServiceAddress          = "localhost:50051"
// 	inventoryServiceKeepaliveTime    = 30 * time.Second
// 	inventoryServiceKeepaliveTimeout = 3 * time.Second

// 	paymentServiceAddress          = "localhost:50052"
// 	paymentServiceKeepaliveTime    = 30 * time.Second
// 	paymentServiceKeepaliveTimeout = 3 * time.Second

// 	httpPort          = "8080"
// 	shutdownTimeout   = 10 * time.Second
// 	readHeaderTimeout = 5 * time.Second
// 	readTimeout       = 15 * time.Second
// 	writeTimeout      = 15 * time.Second
// 	idleTimeout       = 60 * time.Second
// )

// func main() {
// 	// Создать gRPC соединение с InventoryService
// 	inventoryConn, err := grpc.NewClient(inventoryServiceAddress,
// 		grpc.WithTransportCredentials(insecure.NewCredentials()),
// 		grpc.WithKeepaliveParams(keepalive.ClientParameters{
// 			Time:                inventoryServiceKeepaliveTime,
// 			Timeout:             inventoryServiceKeepaliveTimeout,
// 			PermitWithoutStream: true,
// 		}))
// 	if err != nil {
// 		slog.Error("не удалось подключиться к InventoryService", "error", err)
// 		return
// 	}
// 	defer func() {
// 		if closeErr := inventoryConn.Close(); closeErr != nil {
// 			slog.Error("не удалось закрыть соединение с InventoryService", "error", closeErr)
// 		}
// 	}()

// 	// Создать gRPC клиент PaymentService
// 	paymentConn, err := grpc.NewClient(paymentServiceAddress,
// 		grpc.WithTransportCredentials(insecure.NewCredentials()),
// 		grpc.WithKeepaliveParams(keepalive.ClientParameters{
// 			Time:                paymentServiceKeepaliveTime,
// 			Timeout:             paymentServiceKeepaliveTimeout,
// 			PermitWithoutStream: true,
// 		}))
// 	if err != nil {
// 		slog.Error("не удалось подключиться к PaymentService", "error", err)
// 		return
// 	}
// 	defer func() {
// 		if closeErr := paymentConn.Close(); closeErr != nil {
// 			slog.Error("не удалось закрыть соединение с PaymentService", "error", closeErr)
// 		}
// 	}()

// 	// Создаём хранилище и обработчик
// 	store := orderHandler.NewOrderStore()
// 	h := orderHandler.NewOrderHandler(
// 		inventoryv1.NewInventoryServiceClient(inventoryConn),
// 		paymentv1.NewPaymentServiceClient(paymentConn),
// 		store,
// 	)

// 	// Создать OpenAPI сервер
// 	orderServer, err := orderHandler.SetupServer(h)
// 	if err != nil {
// 		slog.Error("ошибка создания сервера OpenAPI", "error", err)
// 		return
// 	}

// 	server := &http.Server{
// 		Addr:              net.JoinHostPort("localhost", httpPort),
// 		Handler:           orderServer,
// 		ReadHeaderTimeout: readHeaderTimeout, // Защита от Slowloris атаки
// 		ReadTimeout:       readTimeout,       // Лимит на чтение всего запроса
// 		WriteTimeout:      writeTimeout,      // Лимит на запись ответа
// 		IdleTimeout:       idleTimeout,       // Таймаут keep-alive соединений
// 	}

// 	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
// 	defer cancel()

// 	go func() {
// 		slog.Info("🚀 HTTP-сервер запущен на порту", "port", httpPort)
// 		listenErr := server.ListenAndServe()
// 		if listenErr != nil && !errors.Is(listenErr, http.ErrServerClosed) {
// 			slog.Error("❌ ошибка запуска сервера", "error", listenErr)
// 			cancel() // будим main, чтобы не висеть бесконечно
// 		}
// 	}()

// 	// Ждём сигнал от ОС или падение сервера.
// 	<-ctx.Done()

// 	slog.Info("🛑 завершение работы сервера...")

// 	// Создаем контекст с таймаутом для остановки сервера
// 	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
// 	defer shutdownCancel()

// 	err = server.Shutdown(shutdownCtx)
// 	if err != nil {
// 		slog.Error("❌ ошибка при остановке сервера", "error", err)
// 	}

// 	slog.Info("✅ сервер остановлен")
// }
