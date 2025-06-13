-- +migrate Up
CREATE TABLE IF NOT EXISTS positions (
	id SERIAL PRIMARY KEY,
	created_at TIMESTAMPTZ,
	updated_at TIMESTAMPTZ,
	closed_at TIMESTAMPTZ,
	exchange TEXT,
	order_id TEXT UNIQUE NOT NULL,
	symbol TEXT,
	side TEXT,
	qty DOUBLE PRECISION,
	entry_price DOUBLE PRECISION,
	exit_price DOUBLE PRECISION,
	profit DOUBLE PRECISION,
	commission DOUBLE PRECISION,
	account_id TEXT NOT NULL,
	status TEXT DEFAULT 'open',
	reason TEXT
);

