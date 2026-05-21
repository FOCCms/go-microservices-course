package v1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func (a *api) CommitParts(ctx context.Context, req *inventoryv1.CommitPartsRequest) (*inventoryv1.CommitPartsResponse, error) {
	err := a.partService.Commit(ctx, req.Uuids)
	if err != nil {
		if errors.Is(err, errs.ErrCommitParts) {
			return nil, status.Errorf(codes.FailedPrecondition, "не все детали получилось списать")
		}
		if errors.Is(err, errs.ErrPartNotFound) {
			return nil, status.Errorf(codes.NotFound, "деталь не найдена")
		}
		return nil, status.Errorf(codes.Internal, "ошибка списания детали")
	}

	return &inventoryv1.CommitPartsResponse{}, nil
}
