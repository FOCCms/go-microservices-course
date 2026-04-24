package order

import (
	"sync"

	"github.com/google/uuid"

	"github.com/FOCCms/go-microservices-course/order/internal/repository/record"
)

type repository struct {
	mu     sync.RWMutex
	orders map[uuid.UUID]record.Order
}

func NewRepository() *repository {
	return &repository{
		orders: make(map[uuid.UUID]record.Order),
	}
}
