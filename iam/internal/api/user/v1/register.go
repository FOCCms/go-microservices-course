package v1

import (
	"context"
	"fmt"

	"github.com/FOCCms/go-microservices-course/iam/internal/api/converter"
	userv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/user/v1"
)

func (a *api) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
	input := converter.RegisterRequestToInput(req)
	uuid, err := a.iamService.Register(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("зарегистрировать пользователя: %w", err)
	}

	return &userv1.RegisterResponse{
		UserUuid: uuid.String(),
	}, nil
}
