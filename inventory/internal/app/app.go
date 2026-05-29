package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/FOCCms/go-microservices-course/inventory/internal/config"
	"github.com/FOCCms/go-microservices-course/inventory/internal/interceptor"
	"github.com/FOCCms/go-microservices-course/platform/pkg/closer"
	"github.com/FOCCms/go-microservices-course/platform/pkg/grpc/health"
	"github.com/FOCCms/go-microservices-course/platform/pkg/logger"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

type App struct {
	diContainer *diContainer
	grpcServer  *grpc.Server
	listener    net.Listener
}

func New(ctx context.Context) (*App, error) {
	a := &App{}
	err := a.initDeps(ctx)
	if err != nil {
		closeAll()
	}
	return a, err
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		if err := a.runGRPCServer(); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			slog.Error("ошибка gRPC сервера", "error", err)
			cancel()
		}
	}()

	<-ctx.Done()
	closeAll()
	return nil
}

func (a *App) initDeps(ctx context.Context) error {
	a.initDI(ctx)
	a.initLogger(ctx)

	if err := a.initListener(ctx); err != nil {
		return err
	}
	if err := a.initGRPCServer(ctx); err != nil {
		return err
	}
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

func (a *App) initListener(ctx context.Context) error {
	lis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", config.AppConfig().GRPC.Address())
	if err != nil {
		return fmt.Errorf("создать listener: %w", err)
	}

	a.listener = lis
	return nil
}

func (a *App) initGRPCServer(ctx context.Context) error {
	iamClient, err := a.diContainer.IAMClient(ctx)
	if err != nil {
		return fmt.Errorf("инициализировать grpc сервер: %w", err)
	}
	authInterceptor := interceptor.AuthIncomingInterceptor(iamClient)

	a.grpcServer = grpc.NewServer(grpc.KeepaliveParams(
		keepalive.ServerParameters{
			MaxConnectionIdle:     config.AppConfig().GRPC.MaxConnectionIdle,
			MaxConnectionAge:      config.AppConfig().GRPC.MaxConnectionAge,
			MaxConnectionAgeGrace: config.AppConfig().GRPC.MaxConnectionAgeGrace,
			Time:                  config.AppConfig().GRPC.KeepaliveTime,
			Timeout:               config.AppConfig().GRPC.KeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             config.AppConfig().GRPC.MinPingInterval,
			PermitWithoutStream: true,
		}),
		grpc.ChainUnaryInterceptor(
			interceptor.UnaryErrorInterceptor,
			authInterceptor),
	)

	reflection.Register(a.grpcServer)
	health.RegisterService(a.grpcServer)

	api, err := a.diContainer.InventoryV1API(ctx)
	if err != nil {
		return fmt.Errorf("инициализировать grpc сервер: %w", err)
	}

	closer.Add("gRPC server", func(_ context.Context) error {
		a.grpcServer.GracefulStop()
		return nil
	})

	inventoryv1.RegisterInventoryServiceServer(a.grpcServer, api)
	return nil
}

func closeAll() {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), config.AppConfig().ShutdownConfig.ShutdownTimeout)
	defer shutdownCancel()

	if closeErr := closer.CloseAll(shutdownCtx); closeErr != nil {
		slog.Error("ошибка при завершении работы", "error", closeErr)
	}
}

func (a *App) runGRPCServer() error {
	slog.Info("gRPC-сервер запущен", "address", config.AppConfig().GRPC.Address())
	return a.grpcServer.Serve(a.listener)
}
