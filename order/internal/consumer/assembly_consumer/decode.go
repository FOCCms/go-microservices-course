package assembly_consumer

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	"github.com/FOCCms/go-microservices-course/order/internal/model"
	eventsv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/events/v1"
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

	var assembledAt time.Time
	if pb.AssembledAt != nil {
		assembledAt = pb.AssembledAt.AsTime()
	}

	return model.ShipAssembledEvent{
		EventUUID:   eventUUID,
		OrderUUID:   orderUUID,
		UserUUID:    userUUID,
		BuildTime:   buildTime,
		AssembledAt: assembledAt,
	}, nil
}
