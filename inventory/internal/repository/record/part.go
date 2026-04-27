package record

import (
	"time"

	"github.com/google/uuid"
)

// Part представляет деталь космического корабля.
type Part struct {
	UUID          string
	Name          string
	Description   string
	Price         int64
	PartType      string
	StockQuantity int64
	CreatedAt     time.Time
}

type PartFilter struct {
	// UUIDs — если не пустой, возвращаются только эти детали (приоритет)
	UUIDs []uuid.UUID
	// PartType — фильтр по типу (игнорируется если UUIDs заполнен)
	PartType string
}
