package v1

import (
	"context"
	"fmt"

	authv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/auth/v1"
	commonv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/common/v1"
)

type Client struct {
	iamClient authv1.AuthServiceClient
}

func New(c authv1.AuthServiceClient) *Client {
	return &Client{
		iamClient: c,
	}
}

func (c *Client) Whoami(ctx context.Context, sessionUuid string) (*commonv1.Session, *commonv1.User, error) {
	resp, err := c.iamClient.Whoami(ctx, &authv1.WhoamiRequest{SessionUuid: sessionUuid})
	if err != nil {
		return nil, nil, fmt.Errorf("получить сессию и пользователя: %w", err)
	}

	return resp.GetSession(), resp.GetUser(), nil
}
