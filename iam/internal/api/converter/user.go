package converter

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/FOCCms/go-microservices-course/iam/internal/model"
	"github.com/FOCCms/go-microservices-course/iam/internal/service/input"
	authv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/auth/v1"
	commonv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/common/v1"
	userv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/user/v1"
)

func RegisterRequestToInput(req *userv1.RegisterRequest) input.RegisterInput {
	return input.RegisterInput{
		Login:    req.GetInfo().GetInfo().GetLogin(),
		Password: req.GetInfo().GetPassword(),
	}
}

func LoginRequestToInput(req *authv1.LoginRequest) input.LoginInput {
	return input.LoginInput{
		Login:    req.GetLogin(),
		Password: req.GetPassword(),
	}
}

func UserModelToDto(user model.User) *commonv1.User {
	dto := &commonv1.User{
		Uuid: user.UUID.String(),
		Info: &commonv1.UserInfo{
			Login: user.Login,
		},
		CreatedAt: timestamppb.New(user.CreatedAt),
	}
	if user.UpdatedAt != nil {
		dto.UpdatedAt = timestamppb.New(*user.UpdatedAt)
	}

	return dto
}
