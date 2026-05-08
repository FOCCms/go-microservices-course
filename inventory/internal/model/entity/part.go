package model

import (
	"time"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/valueobject"
)

type Part struct {
	uuid          string
	name          string
	description   string
	price         int64
	partType      valueobject.PartType
	reserved      int
	stockQuantity int
	createdAt     time.Time
	properties    valueobject.PartProperties
}

func RestorePart(uuid, name, description string, partType valueobject.PartType, price int64,
	stockQuantity, reserved int, properties valueobject.PartProperties, createdAt time.Time,
) Part {
	return Part{
		uuid:          uuid,
		name:          name,
		description:   description,
		partType:      partType,
		price:         price,
		stockQuantity: stockQuantity,
		reserved:      reserved,
		properties:    properties,
		createdAt:     createdAt,
	}
}

func (p *Part) Reserve() error {
	if p.Available() <= 0 {
		return errs.ErrOutOfStock
	}
	p.reserved += 1
	return nil
}

func (p *Part) Release() error {
	if p.reserved <= 0 {
		return errs.ErrNothingToRelease
	}
	p.reserved -= 1
	return nil
}

func (p *Part) Available() int {
	return p.stockQuantity - p.reserved
}

func (p *Part) UUID() string {
	return p.uuid
}

func (p *Part) Name() string {
	return p.name
}

func (p *Part) Description() string {
	return p.description
}

func (p *Part) PartType() valueobject.PartType {
	return p.partType
}

func (p *Part) Price() int64 {
	return p.price
}

func (p *Part) StockQuantity() int {
	return p.stockQuantity
}

func (p *Part) Reserved() int {
	return p.reserved
}

func (p *Part) Properties() valueobject.PartProperties {
	return p.properties
}

func (p *Part) CreatedAt() time.Time {
	return p.createdAt
}
