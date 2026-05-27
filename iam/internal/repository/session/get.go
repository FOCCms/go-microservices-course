package session

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	repoConverter "github.com/FOCCms/go-microservices-course/iam/internal/repository/converter"
	repoModel "github.com/FOCCms/go-microservices-course/iam/internal/repository/redis_view"
)

func (r *repository) Get(ctx context.Context, uuid string) (model.Session, error) {
	cacheKey := r.getCacheKey(uuid)

	var sessionRedisView repoModel.Session
	err := r.client.HGetAll(ctx, cacheKey).Scan(&sessionRedisView)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return model.Session{}, errs.ErrSessionNotFound
		}

		return model.Session{}, err
	}

	// HGetAll возвращает пустую map для несуществующего ключа, без ошибки
	if sessionRedisView.UUID == "" {
		return model.Session{}, errs.ErrSessionNotFound
	}

	return repoConverter.SessionFromRedisView(sessionRedisView)
}
