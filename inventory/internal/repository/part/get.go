package part

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	repoConverter "github.com/FOCCms/go-microservices-course/inventory/internal/repository/converter"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/record"
)

func (r *repository) Get(ctx context.Context, id string) (model.Part, error) {
	const query = `SELECT uuid, name, description, part_type, price, stock_quantity, properties, reserved, created_at FROM parts WHERE uuid = $1`

	var p record.Part

	err := r.getter.DefaultTrOrDB(ctx, r.pool).QueryRow(ctx, query, id).Scan(
		&p.UUID, &p.Name, &p.Description, &p.PartType, &p.Price, &p.StockQuantity, &p.Properties, &p.Reserved, &p.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Part{}, fmt.Errorf("считать деталь: %w", errs.ErrPartNotFound)
		}
		return model.Part{}, fmt.Errorf("считать деталь: %w", err)
	}

	return repoConverter.PartRecordToModel(p)
}
