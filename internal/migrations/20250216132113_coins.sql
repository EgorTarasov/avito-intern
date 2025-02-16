-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE users ADD COLUMN coin_balance INT;

CREATE TABLE IF NOT EXISTS merch(
    id BIGSERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    price INTEGER NOT NULL
);
INSERT INTO merch (name, price) VALUES
    ('t-shirt', 80),
    ('cup', 20),
    ('book', 50),
    ('pen', 10),
    ('powerbank', 200),
    ('hoody', 300),
    ('umbrella', 200),
    ('socks', 10),
    ('wallet', 50),
    ('pink-hoody', 500);

CREATE TABLE purchases (
    id SERIAL PRIMARY KEY,
    fk_user INTEGER NOT NULL REFERENCES users(id),
    fk_merch INTEGER NOT NULL REFERENCES merch(id),
    quantity INTEGER NOT NULL DEFAULT 1,
    purchased_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    fk_from_user INTEGER REFERENCES users(id),
    fk_to_user   INTEGER REFERENCES users(id),
    amount    INTEGER NOT NULL,
    type      TEXT NOT NULL,  -- 'transfer' or 'purchase'
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

DROP TABLE transactions;
DROP TABLE purchases;
DROP TABLE merch;

ALTER TABLE users
DROP COLUMN coin_balance;
-- +goose StatementEnd
