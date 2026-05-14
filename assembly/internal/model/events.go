package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderPaidEvent struct {
	EventUUID uuid.UUID
	OrderUUID uuid.UUID
	UserUUID  uuid.UUID
}

type ShipAssembledEvent struct {
	EventUUID   uuid.UUID
	OrderUUID   uuid.UUID
	UserUUID    uuid.UUID
	BuildTime   time.Duration
	AssembledAd time.Time
}
