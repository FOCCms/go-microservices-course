package converter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	commonv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/common/v1"
)

func SessionModelToDto(s model.Session) *commonv1.Session {
	session := &commonv1.Session{
		Uuid:      s.UUID.String(),
		CreatedAt: timestamppb.New(s.CreatedAt),
		UpdatedAt: nil,
		ExpiresAt: timestamppb.New(s.ExpiresAt),
	}

	if s.UpdatedAt != nil {
		session.UpdatedAt = timestamppb.New(*s.UpdatedAt)
	}

	return session
}
