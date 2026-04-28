package part

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	"github.com/FOCCms/go-microservices-course/inventory/internal/model"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/converter"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/record"
)

func (r *repository) List(ctx context.Context, filter record.PartFilter) ([]model.Part, error) {
	var rows pgx.Rows
	var err error

	if len(filter.UUIDs) > 0 {
		const query = `
				SELECT p.uuid, p.name, p.description, p.part_type, p.price, p.stock_quantity, p.created_at
		    FROM parts p
		    JOIN unnest($1::uuid[]) WITH ORDINALITY AS input_uuids(uuid, order_num)
		    ON p.uuid = input_uuids.uuid
		    ORDER BY input_uuids.order_num
		`
		rows, err = r.pool.Query(ctx, query, filter.UUIDs)
	} else {
		const query = `
			SELECT uuid, name, description, part_type, price, stock_quantity, created_at
			FROM parts
			WHERE $1 = 'UNSPECIFIED' OR part_type = $1
			ORDER BY name ASC
		`
		rows, err = r.pool.Query(ctx, query, filter.PartType)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pts, err := pgx.CollectRows(rows, pgx.RowToStructByName[record.Part])
	if err != nil {
		return nil, fmt.Errorf("считать строки: %w", err)
	}

	if len(filter.UUIDs) > 0 && len(pts) != len(filter.UUIDs) {
		return nil, errs.ErrPartNotFound
	}

	parts := make([]model.Part, 0, len(pts))
	for _, p := range pts {
		parts = append(parts, converter.PartToModel(p))
	}

	return parts, nil
}
