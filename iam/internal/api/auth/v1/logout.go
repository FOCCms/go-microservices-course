package v1

import (
	"context"
	"fmt"

	authv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/auth/v1"
)

func (a *api) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	err := a.iamService.Logout(ctx, req.SessionUuid)
	if err != nil {
		return nil, fmt.Errorf("разлогинить пользователя: %w", err)
	}

	return &authv1.LogoutResponse{}, nil
}
