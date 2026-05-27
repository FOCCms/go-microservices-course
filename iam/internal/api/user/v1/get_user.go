package v1

import (
	"context"
	"fmt"

	"github.com/FOCCms/go-microservices-course/iam/internal/api/converter"
	userv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/user/v1"
)

func (a *api) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	user, err := a.iamService.GetUser(ctx, req.GetUserUuid())
	if err != nil {
		return nil, fmt.Errorf("получить пользователя: %w", err)
	}

	return &userv1.GetUserResponse{
		User: converter.UserModelToDto(user),
	}, nil
}
