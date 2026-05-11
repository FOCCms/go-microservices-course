package v1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func (a *api) ReleaseParts(ctx context.Context, req *inventoryv1.ReleasePartsRequest) (*inventoryv1.ReleasePartsResponse, error) {
	err := a.partService.Release(ctx, req.Uuids)
	if err != nil {
		if errors.Is(err, errs.ErrNothingToRelease) {
			return nil, status.Errorf(codes.FailedPrecondition, "нечего освобождать")
		}
		if errors.Is(err, errs.ErrPartNotFound) {
			return nil, status.Errorf(codes.NotFound, "деталь не найдена")
		}
		return nil, status.Errorf(codes.Internal, "ошибка освобождения детали")
	}

	return &inventoryv1.ReleasePartsResponse{}, nil
}
