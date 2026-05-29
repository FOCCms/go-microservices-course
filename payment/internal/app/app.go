package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/FOCCms/go-microservices-course/payment/internal/config"
	"github.com/FOCCms/go-microservices-course/payment/internal/interceptor"
	"github.com/FOCCms/go-microservices-course/platform/pkg/closer"
	"github.com/FOCCms/go-microservices-course/platform/pkg/grpc/health"
	"github.com/FOCCms/go-microservices-course/platform/pkg/logger"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

const (
	grpcMaxConnectionIdle     = 15 * time.Minute
	grpcMaxConnectionAge      = 30 * time.Minute
	grpcMaxConnectionAgeGrace = 10 * time.Second
	grpcKeepaliveTime         = 5 * time.Minute
	grpcKeepaliveTimeout      = 5 * time.Second
	grpcMinPingInterval       = 5 * time.Second
	shutdownTimeout           = 5 * time.Second
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

	a.startGracefulShutdown(ctx, cancel)

	return a.runGRPCServer()
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
	opts := make([]grpc.ServerOption, 0, 2+len(Interceptors()))
	opts = append(opts,
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				MaxConnectionIdle:     grpcMaxConnectionIdle,
				MaxConnectionAge:      grpcMaxConnectionAge,
				MaxConnectionAgeGrace: grpcMaxConnectionAgeGrace,
				Time:                  grpcKeepaliveTime,
				Timeout:               grpcKeepaliveTimeout,
			}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             grpcMinPingInterval,
			PermitWithoutStream: true,
		}))
	opts = append(opts, Interceptors()...)

	a.grpcServer = grpc.NewServer(opts...)
	closer.Add("gRPC server", func(_ context.Context) error {
		a.grpcServer.GracefulStop()
		return nil
	})

	reflection.Register(a.grpcServer)
	health.RegisterService(a.grpcServer)

	api, err := a.diContainer.PaymentV1API(ctx)
	if err != nil {
		return fmt.Errorf("инициализировать grpc сервер: %w", err)
	}

	paymentv1.RegisterPaymentServiceServer(a.grpcServer, api)
	return nil
}

func (a *App) startGracefulShutdown(ctx context.Context, cancel context.CancelFunc) {
	go func() {
		// Ждём сигнал от ОС или падение сервера.
		<-ctx.Done()
		cancel()

		closeAll()
	}()
}

func closeAll() {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	if closeErr := closer.CloseAll(shutdownCtx); closeErr != nil {
		slog.Error("ошибка при завершении работы", "error", closeErr)
	}
}

func (a *App) runGRPCServer() error {
	slog.Info("gRPC-сервер запущен", "address", config.AppConfig().GRPC.Address())
	return a.grpcServer.Serve(a.listener)
}

func Interceptors() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.UnaryErrorInterceptor),
	}
}
