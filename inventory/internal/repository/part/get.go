package part

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model"
	repoConverter "github.com/FOCCms/go-microservices-course/inventory/internal/repository/converter"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/record"
)

func (r *repository) Get(ctx context.Context, id string) (model.Part, error) {
	const query = `SELECT uuid, name, part_type, price, stock_quantity FROM parts WHERE uuid = $1`

	var p record.Part

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.UUID, &p.Name, &p.PartType, &p.Price, &p.StockQuantity,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Part{}, errs.ErrPartNotFound
		}
		return model.Part{}, err
	}

	return repoConverter.PartToModel(p), nil
}
