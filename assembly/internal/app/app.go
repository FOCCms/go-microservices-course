package app

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/FOCCms/go-microservices-course/platform/pkg/closer"
)

const (
	shutdownTimeout = 10 * time.Second
)

type App struct {
	diContainer *diContainer
}

func New(ctx context.Context) *App {
	a := &App{}
	err := a.initDeps(ctx)
	if err != nil {
		closeAll()
	}
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

func (a *App) initDeps(ctx context.Context) error {
	a.initDI(ctx)
	return nil
}

func (a *App) initDI(_ context.Context) {
	a.diContainer = &diContainer{}
}

func closeAll() {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
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
