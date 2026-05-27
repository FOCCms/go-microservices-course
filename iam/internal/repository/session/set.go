package session

import (
	"context"
	"time"

	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	repoConverter "github.com/FOCCms/go-microservices-course/iam/internal/repository/converter"
)

func (r *repository) Set(ctx context.Context, session model.Session, ttl time.Duration) error {
	cacheKey := r.getCacheKey(session.UUID.String())
	err := r.client.HSet(ctx, cacheKey, new(repoConverter.SessionToRedisView(session))).Err()
	if err != nil {
		return err
	}

	return r.client.Expire(ctx, cacheKey, ttl).Err()
}
