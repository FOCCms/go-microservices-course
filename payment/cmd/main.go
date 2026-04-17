package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	svc "github.com/FOCCms/go-microservices-course/payment/pkg/service"
	paymentv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/payment/v1"
)

const (
	grpcAddress               = "127.0.0.1:50052"
	grpcMaxConnectionIdle     = 15 * time.Minute
	grpcMaxConnectionAge      = 30 * time.Minute
	grpcMaxConnectionAgeGrace = 10 * time.Second
	grpcKeepaliveTime         = 5 * time.Minute
	grpcKeepaliveTimeout      = 5 * time.Second
	grpcMinPingInterval       = 5 * time.Second
)

func main() {
	lc := net.ListenConfig{}
	lis, err := lc.Listen(context.Background(), "tcp", grpcAddress)
	if err != nil {
		slog.Error("не удалось создать listener", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer(
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
		}),
	)
	paymentv1.RegisterPaymentServiceServer(grpcServer, &svc.PaymentServer{})

	// Включаем reflection для postman/grpcurl
	reflection.Register(grpcServer)

	slog.Info("запуск PaymentService", "адрес", grpcAddress)

	go func() {
		slog.Info("🚀 gRPC сервер запущен", "address", grpcAddress)
		if serveErr := grpcServer.Serve(lis); serveErr != nil {
			slog.Error("ошибка запуска сервера", "error", serveErr)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("🛑 остановка gRPC сервера...")
	grpcServer.GracefulStop()
	slog.Info("✅ сервер остановлен")
}
