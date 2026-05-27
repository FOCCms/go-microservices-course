package converter

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	errs "github.com/FOCCms/go-microservices-course/iam/internal/errors"
	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	repoModel "github.com/FOCCms/go-microservices-course/iam/internal/repository/redis_view"
)

func SessionToRedisView(s model.Session) repoModel.Session {
	view := repoModel.Session{
		UUID:        s.UUID.String(),
		UserUUID:    s.UserUUID.String(),
		Login:       s.Login,
		CreatedAtNs: s.CreatedAt.UnixNano(),
		ExpiresAtNs: s.ExpiresAt.UnixNano(),
	}

	if s.UpdatedAt != nil {
		view.UpdatedAtNs = new(s.UpdatedAt.UnixNano())
	}
	return view
}

func SessionFromRedisView(v repoModel.Session) (model.Session, error) {
	sessionUuid, err := uuid.Parse(v.UUID)
	if err != nil {
		return model.Session{}, fmt.Errorf("распарсить UUID сессии: %w", errs.ErrInvalidUUID)
	}
	userUuid, err := uuid.Parse(v.UUID)
	if err != nil {
		return model.Session{}, fmt.Errorf("распарсить UUID пользователя: %w", errs.ErrInvalidUUID)
	}
	session := model.Session{
		UUID:      sessionUuid,
		UserUUID:  userUuid,
		Login:     v.Login,
		CreatedAt: time.Unix(0, v.CreatedAtNs),
		UpdatedAt: nil,
		ExpiresAt: time.Unix(0, v.ExpiresAtNs),
	}

	if v.UpdatedAtNs != nil {
		session.UpdatedAt = new(time.Unix(0, *v.UpdatedAtNs))
	}

	return session, nil
}
