package v1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/FOCCms/go-microservices-course/inventory/internal/converter"
	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

// GetPart возвращает деталь по UUID.
func (a *api) GetPart(ctx context.Context, req *inventoryv1.GetPartRequest) (*inventoryv1.GetPartResponse, error) {
	if req.GetUuid() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "uuid не указан")
	}

	part, err := a.partService.Get(ctx, req.GetUuid())
	if err != nil {
		if errors.Is(err, errs.ErrInvalidUUID) {
			return nil, status.Errorf(codes.InvalidArgument, "неверный формат uuid: %s", req.GetUuid())
		}
		if errors.Is(err, errs.ErrPartNotFound) {
			return nil, status.Errorf(codes.NotFound, "деталь с UUID %s не найдена", req.GetUuid())
		}
		return nil, status.Errorf(codes.Internal, "ошибка получения детали: %v", err)
	}

	return &inventoryv1.GetPartResponse{
		Part: converter.PartToProto(part),
	}, nil
}
