package v1

import (
	userv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/user/v1"
)

type api struct {
	userv1.UnimplementedUserServiceServer

	iamService IAMService
}

func NewAPI(iamService IAMService) *api {
	return &api{
		iamService: iamService,
	}
}
