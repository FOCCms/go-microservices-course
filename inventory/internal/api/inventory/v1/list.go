package v1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/FOCCms/go-microservices-course/inventory/internal/converter"
	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

// ListParts возвращает список деталей с опциональной фильтрацией по типу.
func (a *api) ListParts(
	ctx context.Context,
	req *inventoryv1.ListPartsRequest,
) (*inventoryv1.ListPartsResponse, error) {
	parts, err := a.partService.List(ctx, valueobject.PartFilter{
		UUIDs:    req.GetUuids(),
		PartType: converter.ProtoPartTypeToPartType(req.PartType),
	})
	if err != nil {
		if errors.Is(err, errs.ErrInvalidUUID) {
			return nil, status.Errorf(codes.InvalidArgument, "неверный формат uuid")
		}
		if errors.Is(err, errs.ErrPartNotFound) {
			return nil, status.Errorf(codes.NotFound, "деталь не найдена")
		}
		return nil, status.Errorf(codes.Internal, "ошибка получения деталей: %v", err)
	}

	return &inventoryv1.ListPartsResponse{Parts: converter.PartsToProto(parts)}, nil
}
