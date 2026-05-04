-- +goose Up
CREATE TABLE orders (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    total_price BIGINT NOT NULL DEFAULT 0,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING_PAYMENT',
    transaction_uuid UUID,
    payment_method VARCHAR(50),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS orders;
