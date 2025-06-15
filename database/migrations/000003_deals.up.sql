-- +migrate Up
CREATE TABLE IF NOT EXISTS deals (
	id TEXT PRIMARY KEY,
	order_id TEXT NOT NULL,
	account_id TEXT NOT NULL,
	symbol TEXT,
	side TEXT,
	entry_price DOUBLE PRECISION,
	exit_price DOUBLE PRECISION,
	qty DOUBLE PRECISION,
	profit DOUBLE PRECISION,
	commission DOUBLE PRECISION,
	timestamp TIMESTAMPTZ
);