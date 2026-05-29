package app

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/FOCCms/go-microservices-course/assembly/internal/config"
	"github.com/FOCCms/go-microservices-course/platform/pkg/closer"
	"github.com/FOCCms/go-microservices-course/platform/pkg/logger"
)

type App struct {
	diContainer *diContainer
}

func New(ctx context.Context) *App {
	a := &App{}
	a.initDeps(ctx)
	return a
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

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

func (a *App) initDeps(ctx context.Context) {
	a.initDI(ctx)
	a.initLogger(ctx)
}

func (a *App) initDI(_ context.Context) {
	a.diContainer = &diContainer{}
}

func (a *App) initLogger(_ context.Context) {
	logger.Init(logger.Config{
		Level:             config.AppConfig().Logger.Level,
		ServiceName:       config.AppConfig().OtelConfig.ServiceName,
		Environment:       config.AppConfig().Stage,
		EnableOTLP:        true,
		CollectorEndpoint: config.AppConfig().OtelConfig.Endpoint,
	})
	closer.Add("logger", func(ctx context.Context) error {
		return logger.Close()
	})
}

func closeAll() {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.AppConfig().ShutdownConfig.ShutdownTimeout)
	defer shutdownCancel()

	if closeErr := closer.CloseAll(shutdownCtx); closeErr != nil {
		slog.Error("ошибка при завершении работы", "error", closeErr)
	}
}

func (a *App) runConsumer(ctx context.Context) error {
	slog.Info("🚀 Kafka-потребитель OrderPaid запущен")
	srv, err := a.diContainer.OrderConsumerService()
	if err != nil {
		return fmt.Errorf("запустить потребитель OrderPaid: %w", err)
	}
	return srv.RunConsumer(ctx)
}
