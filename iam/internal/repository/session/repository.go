package session

import "fmt"

const (
	cacheKeyPrefix = "iam:session:"
)

type repository struct {
	client redisClient
}

func NewRepository(client redisClient) *repository {
	return &repository{
		client: client,
	}
}

func (r *repository) getCacheKey(uuid string) string {
	return fmt.Sprintf("%s%s", cacheKeyPrefix, uuid)
}
