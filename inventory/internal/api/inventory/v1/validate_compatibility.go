package v1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

func (a *api) ValidateCompatibility(ctx context.Context, req *inventoryv1.ValidateCompatibilityRequest) (*inventoryv1.ValidateCompatibilityResponse, error) {
	err := a.partService.ValidateCompatibility(ctx, model.ShipSlots{
		Hull:   req.GetHullUuid(),
		Engine: req.GetEngineUuid(),
		Shield: req.GetShieldUuid(),
		Weapon: req.GetWeaponUuid(),
	})
	if err != nil {
		if errors.Is(err, errs.ErrInvalidUUID) {
			return nil, status.Errorf(codes.InvalidArgument, "неверный формат uuid")
		}
		if errors.Is(err, errs.ErrPartNotFound) {
			return nil, status.Errorf(codes.NotFound, "деталь не найдена")
		}
		if errors.Is(err, errs.ErrIncompatibleParts) {
			return nil, status.Errorf(codes.FailedPrecondition, "детали несовместимы")
		}
		if errors.Is(err, errs.ErrPartTypeMismatch) {
			return nil, status.Errorf(codes.InvalidArgument, "неверный слот для детали")
		}
		if errors.Is(err, errs.ErrPartIsNonUnique) {
			return nil, status.Errorf(codes.InvalidArgument, "переданы повторяющиеся UUID")
		}

		return nil, status.Errorf(codes.Internal, "ошибка проверки совместимости")
	}
	return &inventoryv1.ValidateCompatibilityResponse{}, nil
}
