package part

import (
	"context"
	"fmt"
	"sort"

	"github.com/jackc/pgx/v5"

	errs "github.com/FOCCms/go-microservices-course/inventory/internal/errors"
	model "github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/converter"
	"github.com/FOCCms/go-microservices-course/inventory/internal/repository/record"
)

func (r *repository) ListForUpdate(ctx context.Context, filter record.PartFilter) ([]model.Part, error) {
	if len(filter.UUIDs) == 0 {
		return nil, fmt.Errorf("считать детали по списку uuid: %w", errs.ErrPartNotFound)
	}

	// сортировка перед for update
	sort.Slice(filter.UUIDs, func(i, j int) bool {
		return filter.UUIDs[i] < filter.UUIDs[j]
	})

	const query = `
				SELECT p.uuid, p.name, p.description, p.part_type, p.price, p.stock_quantity, p.properties, p.reserved, p.created_at
		    FROM parts p
		    JOIN unnest($1::uuid[]) WITH ORDINALITY AS input_uuids(uuid, order_num)
		    ON p.uuid = input_uuids.uuid
		    ORDER BY input_uuids.order_num
			FOR UPDATE
		`
	rows, err := r.getter.DefaultTrOrDB(ctx, r.pool).Query(ctx, query, filter.UUIDs)
	if err != nil {
		return nil, fmt.Errorf("считать детали: %w", err)
	}
	defer rows.Close()

	pts, err := pgx.CollectRows(rows, pgx.RowToStructByName[record.Part])
	if err != nil {
		return nil, fmt.Errorf("считать строки: %w", err)
	}

	if len(filter.UUIDs) > 0 && len(pts) != len(filter.UUIDs) {
		return nil, fmt.Errorf("считать детали по списку uuid: %w", errs.ErrPartNotFound)
	}

	parts := make([]model.Part, 0, len(pts))
	for _, p := range pts {
		part, err := converter.PartRecordToModel(p)
		if err != nil {
			return nil, fmt.Errorf("считать строки: %w", err)
		}
		parts = append(parts, part)
	}

	return parts, nil
}
