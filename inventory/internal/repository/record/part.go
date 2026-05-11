package record

import (
	"time"
)

// Part представляет деталь космического корабля.
type Part struct {
	UUID          string    `db:"uuid"`
	Name          string    `db:"name"`
	Description   string    `db:"description"`
	Price         int64     `db:"price"`
	PartType      string    `db:"part_type"`
	StockQuantity int       `db:"stock_quantity"`
	CreatedAt     time.Time `db:"created_at"`
	Reserved      int       `db:"reserved"`
	Properties    []byte    `db:"properties"`
}

type PartFilter struct {
	// UUIDs — если не пустой, возвращаются только эти детали (приоритет)
	UUIDs []string
	// PartType — фильтр по типу (игнорируется если UUIDs заполнен)
	PartType string
}
