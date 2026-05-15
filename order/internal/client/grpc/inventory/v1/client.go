package v1

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/FOCCms/go-microservices-course/order/internal/client/grpc/inventory/v1/converter"
	errs "github.com/FOCCms/go-microservices-course/order/internal/errors"
	"github.com/FOCCms/go-microservices-course/order/internal/model"
	inventoryv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/inventory/v1"
)

type client struct {
	inventoryClient inventoryv1.InventoryServiceClient
}

func New(c inventoryv1.InventoryServiceClient) *client {
	return &client{inventoryClient: c}
}

func (c *client) ListParts(ctx context.Context, uuids []uuid.UUID) ([]model.Part, error) {
	uuidsStr := uuidsToStr(uuids)

	resp, err := c.inventoryClient.ListParts(ctx, &inventoryv1.ListPartsRequest{
		Uuids: uuidsStr,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return nil, errs.ErrPartNotFound
		}
		return nil, fmt.Errorf("получить список деталей: %w", err)
	}

	return converter.PartsToModel(resp.GetParts()), nil
}

func (c *client) ValidateCompatibility(ctx context.Context, slots model.ShipSlots) error {
	_, err := c.inventoryClient.ValidateCompatibility(ctx, &inventoryv1.ValidateCompatibilityRequest{
		HullUuid:   slots.Hull,
		EngineUuid: slots.Engine,
		ShieldUuid: slots.Shield,
		WeaponUuid: slots.Weapon,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return err
		}
		if st.Code() == codes.InvalidArgument {
			return errs.ErrPartTypeMismatch
		}
		if st.Code() == codes.FailedPrecondition {
			return errs.ErrIncompatibleParts
		}
		return fmt.Errorf("проверить совместимость деталей: %w", err)
	}

	return nil
}

func (c *client) ReserveParts(ctx context.Context, uuids []uuid.UUID) error {
	uuidsStr := uuidsToStr(uuids)

	_, err := c.inventoryClient.ReserveParts(ctx, &inventoryv1.ReservePartsRequest{
		Uuids: uuidsStr,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return err
		}
		if st.Code() == codes.ResourceExhausted {
			return errs.ErrOutOfStock
		}
		return fmt.Errorf("зарезервировать детали: %w", err)
	}
	return nil
}

func (c *client) ReleaseParts(ctx context.Context, uuids []uuid.UUID) error {
	uuidsStr := uuidsToStr(uuids)

	_, err := c.inventoryClient.ReleaseParts(ctx, &inventoryv1.ReleasePartsRequest{
		Uuids: uuidsStr,
	})
	if err != nil {
		return fmt.Errorf("освободить детали: %w", err)
	}
	return nil
}

func (c *client) CommitParts(ctx context.Context, uuids []uuid.UUID) error {
	uuidsStr := uuidsToStr(uuids)

	_, err := c.inventoryClient.CommitParts(ctx, &inventoryv1.CommitPartsRequest{
		Uuids: uuidsStr,
	})
	if err != nil {
		return fmt.Errorf("списать детали: %w", err)
	}
	return nil
}

func uuidsToStr(uuids []uuid.UUID) []string {
	uuidsStr := make([]string, len(uuids))
	for i, u := range uuids {
		uuidsStr[i] = u.String()
	}
	return uuidsStr
}
