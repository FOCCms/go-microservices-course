package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderItem struct {
	UUID      uuid.UUID
	OrderUUID uuid.UUID
	PartUUID  uuid.UUID
	PartType  PartType
	Price     int64
	CreatedAt time.Time
}
