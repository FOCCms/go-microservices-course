package part

import (
	"context"

	model "github.com/FOCCms/go-microservices-course/inventory/internal/model/entity"
)

func (r *repository) UpdateReservedBatch(ctx context.Context, parts []model.Part) error {
	const query = `
		UPDATE parts AS c
		SET reserved = batch.reserved
		FROM unnest($1::uuid[], $2::int[]) AS batch(uuid, reserved)
		WHERE c.uuid = batch.uuid
	`

	uuids := make([]string, len(parts))
	reservedVals := make([]int, len(parts))

	for i, p := range parts {
		uuids[i] = p.UUID()
		reservedVals[i] = p.Reserved()
	}

	_, err := r.pool.Exec(ctx, query, uuids, reservedVals)
	if err != nil {
		return err
	}
	return nil
}
