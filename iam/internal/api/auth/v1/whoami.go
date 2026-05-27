package v1

import (
	"context"
	"fmt"

	"github.com/FOCCms/go-microservices-course/iam/internal/api/converter"
	authv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/auth/v1"
)

func (a *api) Whoami(ctx context.Context, req *authv1.WhoamiRequest) (*authv1.WhoamiResponse, error) {
	session, user, err := a.iamService.Whoami(ctx, req.SessionUuid)
	if err != nil {
		return nil, fmt.Errorf("получить сессию и пользователя: %w", err)
	}

	return &authv1.WhoamiResponse{
		Session: converter.SessionModelToDto(session),
		User:    converter.UserModelToDto(user),
	}, nil
}
