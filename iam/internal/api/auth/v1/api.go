package v1

import (
	authv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/auth/v1"
)

type api struct {
	authv1.UnimplementedAuthServiceServer

	iamService IAMService
}

func NewAPI(iamService IAMService) *api {
	return &api{
		iamService: iamService,
	}
}
