-- +goose Up
ALTER TABLE transactions ADD COLUMN deleted_at TIMESTAMP;

-- +goose Down
ALTER TABLE transactions DROP COLUMN deleted_at;
