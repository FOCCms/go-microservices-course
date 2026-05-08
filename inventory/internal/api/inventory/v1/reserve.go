package v1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func (a *api) ReserveParts(ctx context.Context, req *inventoryv1.ReservePartsRequest) (*inventoryv1.ReservePartsResponse, error) {
	err := a.partService.Reserve(ctx, req.Uuids)
	if err != nil {
		if errors.Is(err, errs.ErrOutOfStock) {
			return nil, status.Errorf(codes.ResourceExhausted, "деталь закончилась")
		}
		if errors.Is(err, errs.ErrPartNotFound) {
			return nil, status.Errorf(codes.NotFound, "деталь не найдена")
		}
		return nil, status.Errorf(codes.Internal, "ошибка резервирования детали")
	}
	return &inventoryv1.ReservePartsResponse{}, nil
}
