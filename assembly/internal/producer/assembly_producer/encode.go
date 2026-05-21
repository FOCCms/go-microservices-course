package assembly_producer

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/FOCCms/go-microservices-course/assembly/internal/model"
	eventsv1 "github.com/FOCCms/go-microservices-course/shared/pkg/proto/events/v1"
)

func encodeShipAssembled(event model.ShipAssembledEvent) ([]byte, error) {
	msg := &eventsv1.ShipAssembled{
		EventUuid:    event.EventUUID.String(),
		OrderUuid:    event.OrderUUID.String(),
		UserUuid:     event.UserUUID.String(),
		BuildTimeSec: int64(event.BuildTime.Seconds()),
		AssembledAt:  timestamppb.New(event.AssembledAt),
	}

	payload, err := proto.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("сериализовать ShipAssembled: %w", err)
	}
	return payload, nil
}
