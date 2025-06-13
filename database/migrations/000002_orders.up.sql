-- +migrate Up
CREATE TABLE IF NOT EXISTS orders (
	order_id TEXT PRIMARY KEY,
	exchange TEXT,
	account_id TEXT NOT NULL,
	symbol TEXT,
	side TEXT,
	entry_price DOUBLE PRECISION,
	sl DOUBLE PRECISION,
	tp DOUBLE PRECISION,
	qty DOUBLE PRECISION,
	reason TEXT,
	status TEXT DEFAULT 'new',
	timestamp TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ
);
