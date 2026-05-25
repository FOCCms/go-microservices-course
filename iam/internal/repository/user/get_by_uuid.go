package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	repoConverter "github.com/FOCCms/go-microservices-course/iam/internal/repository/converter"
	"github.com/FOCCms/go-microservices-course/iam/internal/repository/record"
)

func (r *repository) GetByUUID(ctx context.Context, uuid uuid.UUID) (model.User, error) {
	const query = `SELECT uuid, login, password_hash, created_at, updated_at FROM users WHERE uuid = $1`

	var u record.User

	err := r.pool.QueryRow(ctx, query, uuid).Scan(
		&u.UUID, &u.Login, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, fmt.Errorf("считать пользователя: %w", errs.ErrUserNotFound)
		}
		return model.User{}, fmt.Errorf("считать пользователя: %w", err)
	}
	return repoConverter.UserRecordToModel(u), nil
}
