-- +goose Up
CREATE TABLE order_items (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_uuid UUID NOT NULL REFERENCES orders(uuid),
    part_uuid UUID NOT NULL,
    part_type VARCHAR(20) NOT NULL,
    price BIGINT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS order_items;
