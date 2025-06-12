-- +migrate Up
CREATE TABLE IF NOT EXISTS positions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMP,

    order_id VARCHAR(255) NOT NULL UNIQUE,
    exchange VARCHAR(20) NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL,
    qty DOUBLE PRECISION NOT NULL,
    entry_price DOUBLE PRECISION NOT NULL,
    exit_price DOUBLE PRECISION,
    profit DOUBLE PRECISION,
    commission DOUBLE PRECISION,
    status VARCHAR(10) NOT NULL DEFAULT 'open',
    reason TEXT,

    -- Indexes
    INDEX idx_closed_at (closed_at)
);

