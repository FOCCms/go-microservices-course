package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	"github.com/FOCCms/go-microservices-course/iam/internal/repository/converter"
)

const duplicateKeyValueViolatesUniqueConstraint = "23505"

func (r *repository) Create(ctx context.Context, user model.User) error {
	u := converter.UserModelToRecord(user)

	const query = `INSERT INTO users (uuid, login, password_hash, created_at) VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, query, u.UUID, u.Login, u.PasswordHash, u.CreatedAt)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == duplicateKeyValueViolatesUniqueConstraint {
			return errs.ErrUserAlreadyExists
		}
		return fmt.Errorf("создать пользователя: %w", err)
	}
	return nil
}
