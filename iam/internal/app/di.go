package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	authV1API "github.com/FOCCms/go-microservices-course/iam/internal/api/auth/v1"
	userV1API "github.com/FOCCms/go-microservices-course/iam/internal/api/user/v1"
	"github.com/FOCCms/go-microservices-course/iam/internal/config"
	sessionRepository "github.com/FOCCms/go-microservices-course/iam/internal/repository/session"
	userRepository "github.com/FOCCms/go-microservices-course/iam/internal/repository/user"
	iamService "github.com/FOCCms/go-microservices-course/iam/internal/service/iam"
	"github.com/FOCCms/go-microservices-course/platform/pkg/closer"
	platformRedis "github.com/FOCCms/go-microservices-course/platform/pkg/redis"
	authv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/auth/v1"
	userv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/user/v1"
)

type diContainer struct {
	pgPool      *pgxpool.Pool
	redisClient *redis.Client

	userRepo    iamService.UserRepository
	sessionRepo iamService.SessionRepository

	iamService *iamService.Service

	authV1Handler authv1.AuthServiceServer
	userV1Handler userv1.UserServiceServer
}

func (d *diContainer) PGPool(ctx context.Context) (*pgxpool.Pool, error) {
	if d.pgPool == nil {
		// Подключаем БД
		pool, err := pgxpool.New(ctx, config.AppConfig().PG.DSN())
		if err != nil {
			return nil, fmt.Errorf("создать пул соединений: %w", err)
		}

		// Проверяем соединение
		err = pool.Ping(ctx)
		if err != nil {
			return nil, fmt.Errorf("проверить соединение с БД: %w", err)
		}

		closer.Add("PostgreSQL pool", func(_ context.Context) error {
			pool.Close()
			return nil
		})

		d.pgPool = pool
	}
	return d.pgPool, nil
}

func (d *diContainer) RedisClient(_ context.Context) (*redis.Client, error) {
	if d.redisClient == nil {
		rdb, err := platformRedis.NewClient(&redis.Options{
			Addr:            config.AppConfig().Redis.Address(),
			DialTimeout:     config.AppConfig().Redis.ConnectionTimeout,
			ReadTimeout:     config.AppConfig().Redis.ConnectionTimeout,
			WriteTimeout:    config.AppConfig().Redis.ConnectionTimeout,
			MaxIdleConns:    config.AppConfig().Redis.MaxIdle,
			ConnMaxIdleTime: config.AppConfig().Redis.IdleTimeout,
		}, slog.Default())
		if err != nil {
			return nil, fmt.Errorf("создать клиент Redis: %w", err)
		}

		closer.Add("Redis", func(_ context.Context) error {
			return rdb.Close()
		})

		d.redisClient = rdb
	}

	return d.redisClient, nil
}

func (d *diContainer) SessionRepo(ctx context.Context) (iamService.SessionRepository, error) {
	if d.sessionRepo == nil {
		redisClient, err := d.RedisClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать SessionRepo: %w", err)
		}
		d.sessionRepo = sessionRepository.NewRepository(redisClient)
	}

	return d.sessionRepo, nil
}

func (d *diContainer) UserRepo(ctx context.Context) (iamService.UserRepository, error) {
	if d.userRepo == nil {
		pool, err := d.PGPool(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать UserRepo: %w", err)
		}
		d.userRepo = userRepository.NewRepository(pool)
	}
	return d.userRepo, nil
}

func (d *diContainer) IAMService(ctx context.Context) (*iamService.Service, error) {
	if d.iamService == nil {
		userRepo, err := d.UserRepo(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать IAMService: %w", err)
		}
		sessionRepo, err := d.SessionRepo(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать IAMService: %w", err)
		}
		ttl, err := time.ParseDuration(config.AppConfig().Session.TTL)
		if err != nil {
			return nil, fmt.Errorf("инициализировать IAMService: %w", err)
		}

		d.iamService = iamService.NewService(userRepo, sessionRepo, ttl, bcrypt.DefaultCost)
	}

	return d.iamService, nil
}

func (d *diContainer) AuthV1API(ctx context.Context) (authv1.AuthServiceServer, error) {
	if d.authV1Handler == nil {
		service, err := d.IAMService(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать хендлер: %w", err)
		}
		d.authV1Handler = authV1API.NewAPI(service)
	}

	return d.authV1Handler, nil
}

func (d *diContainer) UserV1API(ctx context.Context) (userv1.UserServiceServer, error) {
	if d.userV1Handler == nil {
		service, err := d.IAMService(ctx)
		if err != nil {
			return nil, fmt.Errorf("инициализировать хендлер: %w", err)
		}
		d.userV1Handler = userV1API.NewAPI(service)
	}

	return d.userV1Handler, nil
}
