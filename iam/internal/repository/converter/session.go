package converter

import (
	"time"

	"github.com/google/uuid"

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

func SessionFromRedisView(v repoModel.Session) model.Session {
	session := model.Session{
		UUID:      uuid.MustParse(v.UUID),
		UserUUID:  uuid.MustParse(v.UserUUID),
		Login:     v.Login,
		CreatedAt: time.Unix(0, v.CreatedAtNs),
		UpdatedAt: nil,
		ExpiresAt: time.Unix(0, v.ExpiresAtNs),
	}

	if v.UpdatedAtNs != nil {
		session.UpdatedAt = new(time.Unix(0, *v.UpdatedAtNs))
	}

	return session
}
