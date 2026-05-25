package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/FOCCms/go-microservices-course/order/internal/config"
	"github.com/FOCCms/go-microservices-course/order/internal/middleware"
	"github.com/FOCCms/go-microservices-course/platform/pkg/closer"
	"github.com/FOCCms/go-microservices-course/platform/pkg/logger"
)

const (
	inventoryServiceAddress          = "localhost:50051"
	inventoryServiceKeepaliveTime    = 30 * time.Second
	inventoryServiceKeepaliveTimeout = 3 * time.Second

	paymentServiceAddress          = "localhost:50052"
	paymentServiceKeepaliveTime    = 30 * time.Second
	paymentServiceKeepaliveTimeout = 3 * time.Second

	iamServiceAddress          = "localhost:50053"
	iamServiceKeepaliveTime    = 60 * time.Second
	iamServiceKeepaliveTimeout = 20 * time.Second

	shutdownTimeout   = 10 * time.Second
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
)

type App struct {
	diContainer *diContainer
	httpServer  *http.Server
}

func New(ctx context.Context) *App {
	a := &App{}
	err := a.initDeps(ctx)
	if err != nil {
		closeAll()
	}
	return a
}

func (a *App) initDeps(ctx context.Context) error {
	a.initDI(ctx)
	a.initLogger(ctx)
	if err := a.initHTTPServer(ctx); err != nil {
		return err
	}
	return nil
}

func (a *App) initDI(_ context.Context) {
	a.diContainer = &diContainer{}
}

func (a *App) initLogger(_ context.Context) {
	logger.Init(config.AppConfig().Logger.Level)
}

func (a *App) initHTTPServer(ctx context.Context) error {
	orderServer, err := a.diContainer.OrderV1API(ctx)
	if err != nil {
		return fmt.Errorf("инициализировать http сервер: %w", err)
	}
	iamClient, err := a.diContainer.IAMClient(ctx)
	if err != nil {
		return fmt.Errorf("инициализировать http сервер: %w", err)
	}

	authMiddleware := middleware.AuthMiddleware(iamClient)

	handler := middleware.ErrorsMiddleware(authMiddleware(orderServer))

	a.httpServer = &http.Server{
		Addr:              config.AppConfig().HTTP.Addr(),
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout, // Защита от Slowloris атаки
		ReadTimeout:       readTimeout,       // Лимит на чтение всего запроса
		WriteTimeout:      writeTimeout,      // Лимит на запись ответа
		IdleTimeout:       idleTimeout,       // Таймаут keep-alive соединений
	}
	closer.Add("http server", func(ctx context.Context) error {
		return a.httpServer.Shutdown(ctx)
	})

	return nil
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		slog.Info("🚀 HTTP-сервер запущен", "address", config.AppConfig().HTTP.Addr())
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("ошибка запуска сервера", "error", err)
			cancel()
		}
	}()

	go func() {
		if err := a.runConsumer(ctx); err != nil {
			slog.Error("ошибка запуска потребителя", "error", err)
			cancel()
		}
	}()

	<-ctx.Done()
	closeAll()
	return nil
}

func closeAll() {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if closeErr := closer.CloseAll(shutdownCtx); closeErr != nil {
		slog.Error("ошибка при завершении работы", "error", closeErr)
	}
}

func (a *App) runConsumer(ctx context.Context) error {
	slog.Info("🚀 Kafka-потребитель ShipAssembled запущен")
	srv, err := a.diContainer.AssemblyConsumerService(ctx)
	if err != nil {
		return fmt.Errorf("запустить потребитель OrderPaid: %w", err)
	}
	return srv.RunConsumer(ctx)
}
