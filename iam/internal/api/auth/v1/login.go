package v1

import (
	"context"
	"fmt"

	"github.com/FOCCms/go-microservices-course/iam/internal/api/converter"
	authv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/auth/v1"
)

func (a *api) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	uuid, err := a.iamService.Login(ctx, converter.LoginRequestToInput(req))
	if err != nil {
		return nil, fmt.Errorf("залогинить пользователя: %w", err)
	}

	return &authv1.LoginResponse{
		SessionUuid: uuid.String(),
	}, nil
}
