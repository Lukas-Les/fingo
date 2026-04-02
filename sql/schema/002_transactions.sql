-- +goose Up
CREATE TYPE transaction_type_enum AS ENUM ('income', 'expense');
CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount NUMERIC(10, 2) NOT NULL,
    transaction_type transaction_type_enum NOT NULL,
    category TEXT,
    description TEXT,
    party TEXT,
    transaction_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE transactions;
DROP TYPE transaction_type;
