package main

import (
	"context"
	"log/slog"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	"github.com/FOCCms/go-microservices-course/inventory/internal/config"
	"github.com/FOCCms/go-microservices-course/inventory/pkg/app"
)

const (
	grpcMaxConnectionIdle     = 15 * time.Minute
	grpcMaxConnectionAge      = 30 * time.Minute
	grpcMaxConnectionAgeGrace = 10 * time.Second
	grpcKeepaliveTime         = 5 * time.Minute
	grpcKeepaliveTimeout      = 5 * time.Second
	grpcMinPingInterval       = 5 * time.Minute
)

func main() {
	configPath := config.ResolveConfigPath()

	// .env опционален — ошибка загрузки допустима.
	err := godotenv.Load("inventory.env")
	if err != nil {
		slog.Warn("не удалось загрузить .env конфигурацию", "error", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("не удалось загрузить конфигурацию", "error", err)
		return
	}

	lis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", cfg.GRPC.Address())
	if err != nil {
		slog.Error("не удалось создать listener", "error", err)
		return
	}
	opts := make([]grpc.ServerOption, 0, 2+len(app.Interceptors()))
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
	opts = append(opts, app.Interceptors()...)

	grpcServer := grpc.NewServer(opts...)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Подключаем БД
	pool, err := pgxpool.New(ctx, cfg.PG.DSN())
	if err != nil {
		slog.Error("создание пула соединений", "error", err)
		return
	}
	defer pool.Close()

	// Проверяем соединение
	err = pool.Ping(ctx)
	if err != nil {
		slog.Error("проверка соединения с БД", "error", err)
		return
	}

	slog.Info("подключение к PostgreSQL установлено")

	// Регистрируем сервисы
	app.RegisterServices(grpcServer, pool)

	// Включаем reflection для postman/grpcurl
	reflection.Register(grpcServer)

	slog.Info("запуск InventoryService", "адрес", cfg.GRPC.Address())

	go func() {
		slog.Info("🚀 gRPC сервер запущен", "address", cfg.GRPC.Address())
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			slog.Error("ошибка запуска сервера", "error", serveErr)
			cancel() // будим main, чтобы не висеть бесконечно
		}
	}()

	// Ждём сигнал от ОС или падение сервера.
	<-ctx.Done()
	slog.Info("🛑 остановка gRPC сервера")
	grpcServer.GracefulStop()
	slog.Info("✅ сервер остановлен")
}
