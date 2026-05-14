package assembly_consumer

import (
	"fmt"
	"time"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
	eventsv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/events/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
)

func decodeShipAssembled(data []byte) (model.ShipAssembledEvent, error) {
	var pb eventsv1.ShipAssembled
	if err := proto.Unmarshal(data, &pb); err != nil {
		return model.ShipAssembledEvent{}, fmt.Errorf("десериализовать protobuf: %w", err)
	}

	eventUUID, err := uuid.Parse(pb.EventUuid)
	if err != nil {
		return model.ShipAssembledEvent{}, fmt.Errorf("распарсить event uuid: %w", err)
	}

	orderUUID, err := uuid.Parse(pb.OrderUuid)
	if err != nil {
		return model.ShipAssembledEvent{}, fmt.Errorf("распарсить order uuid: %w", err)
	}

	userUUID, err := uuid.Parse(pb.UserUuid)
	if err != nil {
		return model.ShipAssembledEvent{}, fmt.Errorf("распарсить user uuid: %w", err)
	}

	buildTime := time.Duration(pb.BuildTimeSec) * time.Second

	var assembledAd time.Time
	if pb.AssembledAt != nil {
		assembledAd = pb.AssembledAt.AsTime()
	}

	return model.ShipAssembledEvent{
		EventUUID:   eventUUID,
		OrderUUID:   orderUUID,
		UserUUID:    userUUID,
		BuildTime:   buildTime,
		AssembledAd: assembledAd,
	}, nil

}

func decodeOrderPaid(data []byte) (model.OrderPaidEvent, error) {
	var pb eventsv1.OrderPaid
	if err := proto.Unmarshal(data, &pb); err != nil {
		return model.OrderPaidEvent{}, fmt.Errorf("десериализовать protobuf: %w", err)
	}

	eventUUID, err := uuid.Parse(pb.EventUuid)
	if err != nil {
		return model.OrderPaidEvent{}, fmt.Errorf("распарсить event uuid: %w", err)
	}

	orderUUID, err := uuid.Parse(pb.OrderUuid)
	if err != nil {
		return model.OrderPaidEvent{}, fmt.Errorf("распарсить order uuid: %w", err)
	}

	userUUID, err := uuid.Parse(pb.UserUuid)
	if err != nil {
		return model.OrderPaidEvent{}, fmt.Errorf("распарсить user uuid: %w", err)
	}

	return model.OrderPaidEvent{
		EventUUID: eventUUID,
		OrderUUID: orderUUID,
		UserUUID:  userUUID,
	}, nil
}
