-- +goose Up
UPDATE orders SET user_uuid = '00000000-0000-0000-0000-000000000000'::uuid WHERE user_uuid IS NULL;

-- +goose Down
UPDATE orders SET user_uuid = null WHERE user_uuid IS NOT NULL;
