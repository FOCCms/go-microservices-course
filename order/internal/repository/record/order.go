package record

import "time"

type Order struct {
	UUID            string    `db:"uuid"`
	HullUUID        string    `db:"hull_uuid"`
	EngineUUID      string    `db:"engine_uuid"`
	ShieldUUID      *string   `db:"shield_uuid"`
	WeaponUUID      *string   `db:"weapon_uuid"`
	TotalPrice      int64     `db:"total_price"`
	TransactionUUID *string   `db:"transaction_uuid"`
	PaymentMethod   *string   `db:"payment_method"`
	Status          string    `db:"status"`
	CreatedAt       time.Time `db:"created_at"`
}
