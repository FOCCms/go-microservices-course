package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	authV1API "github.com/FOCCms/go-microservices-course/iam/internal/api/auth/v1"
	userV1API "github.com/FOCCms/go-microservices-course/iam/internal/api/user/v1"
	"github.com/FOCCms/go-microservices-course/iam/internal/app"
	sessionRepository "github.com/FOCCms/go-microservices-course/iam/internal/repository/session"
	userRepository "github.com/FOCCms/go-microservices-course/iam/internal/repository/user"
	iamService "github.com/FOCCms/go-microservices-course/iam/internal/service/iam"
	"github.com/FOCCms/go-microservices-course/platform/pkg/grpc/health"
	authv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/auth/v1"
	userv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/user/v1"
)

const (
	grpcMaxConnectionIdle     = 15 * time.Minute
	grpcMaxConnectionAge      = 30 * time.Minute
	grpcMaxConnectionAgeGrace = 10 * time.Second
	grpcKeepaliveTime         = 5 * time.Minute
	grpcKeepaliveTimeout      = 5 * time.Second
	grpcMinPingInterval       = 5 * time.Minute
)

func NewGRPCServer(pool *pgxpool.Pool, rdb *redis.Client, ttl time.Duration, minBcryptCost int) *grpc.Server {
	userRepo := userRepository.NewRepository(pool)
	sessionRepo := sessionRepository.NewRepository(rdb)

	service := iamService.NewService(userRepo, sessionRepo, ttl, minBcryptCost)
	authApi := authV1API.NewAPI(service)
	userApi := userV1API.NewAPI(service)

	return initGRPCServer(authApi, userApi)
}

func initGRPCServer(authApi authv1.AuthServiceServer, userApi userv1.UserServiceServer) *grpc.Server {
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

	health.RegisterService(grpcServer)

	authv1.RegisterAuthServiceServer(grpcServer, authApi)
	userv1.RegisterUserServiceServer(grpcServer, userApi)

	return grpcServer
}
