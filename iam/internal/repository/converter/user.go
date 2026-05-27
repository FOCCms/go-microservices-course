package converter

import (
	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	"github.com/FOCCms/go-microservices-course/iam/internal/repository/record"
)

func UserRecordToModel(u record.User) model.User {
	user := model.User{
		UUID:         uuid.MustParse(u.UUID),
		Login:        u.Login,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}
	if u.UpdatedAt != nil {
		user.UpdatedAt = u.UpdatedAt
	}
	return user
}

func UserModelToRecord(u model.User) record.User {
	user := record.User{
		UUID:         u.UUID.String(),
		Login:        u.Login,
		PasswordHash: u.PasswordHash,
		CreatedAt:    u.CreatedAt,
	}
	if u.UpdatedAt != nil {
		user.UpdatedAt = u.UpdatedAt
	}
	return user
}
