package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/FOCCms/go-microservices-course/order/internal/config"
	"github.com/FOCCms/go-microservices-course/order/internal/middleware"
	"github.com/FOCCms/go-microservices-course/platform/pkg/closer"
	"github.com/FOCCms/go-microservices-course/platform/pkg/logger"
	"github.com/FOCCms/go-microservices-course/platform/pkg/metrics"
	"github.com/FOCCms/go-microservices-course/platform/pkg/tracing"
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
	a.initLogger(ctx)
	if err := a.initTracing(ctx); err != nil {
		return err
	}
	a.initMetrics(ctx)
	a.initDI(ctx)
	if err := a.initHTTPServer(ctx); err != nil {
		return err
	}
	return nil
}

func (a *App) initMetrics(_ context.Context) {
	metrics.Init(config.AppConfig().OtelConfig.ServiceName)
}

func (a *App) initTracing(ctx context.Context) error {
	shutdown, err := tracing.InitTracer(ctx, tracing.Config{
		CollectorEndpoint: config.AppConfig().OtelConfig.Endpoint,
		ServiceName:       config.AppConfig().OtelConfig.ServiceName,
		Environment:       config.AppConfig().Stage,
		ServiceVersion:    config.AppConfig().ServiceVersion,
		SamplingRatio:     config.AppConfig().TracingConfig.SamplingRatio,
	})
	if err != nil {
		return fmt.Errorf("инициализировать tracing: %w", err)
	}
	closer.Add("tracing", func(ctx context.Context) error {
		return shutdown(ctx)
	})

	return nil
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

	handler := authMiddleware(orderServer)
	handler = middleware.ErrorsMiddleware(handler)
	handler = otelhttp.NewHandler(handler, config.AppConfig().OtelConfig.ServiceName)

	a.httpServer = &http.Server{
		Addr:              config.AppConfig().HTTP.Addr(),
		Handler:           handler,
		ReadHeaderTimeout: config.AppConfig().HTTP.ReadHeaderTimeout, // Защита от Slowloris атаки
		ReadTimeout:       config.AppConfig().HTTP.ReadTimeout,       // Лимит на чтение всего запроса
		WriteTimeout:      config.AppConfig().HTTP.WriteTimeout,      // Лимит на запись ответа
		IdleTimeout:       config.AppConfig().HTTP.IdleTimeout,       // Таймаут keep-alive соединений
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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.AppConfig().ShutdownConfig.ShutdownTimeout)
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
