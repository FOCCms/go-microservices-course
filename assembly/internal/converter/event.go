package converter

import (
	"time"

	"github.com/FOCCms/go-microservices-course/assembly/internal/model"
)

func ToShipAssembledEvent(e model.OrderPaidEvent, d time.Duration, t time.Time) model.ShipAssembledEvent {
	return model.ShipAssembledEvent{
		EventUUID:   e.EventUUID,
		OrderUUID:   e.OrderUUID,
		UserUUID:    e.UserUUID,
		BuildTime:   d,
		AssembledAd: t,
	}
}
