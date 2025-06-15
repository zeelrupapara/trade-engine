package models

import "time"

type Account struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Balance   float64   `db:"balance"`
	Equity    float64   `db:"equity"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Deal struct {
	ID         string    `db:"id"`
	OrderID    string    `db:"order_id"`
	AccountID  string    `db:"account_id"`
	Symbol     string    `db:"symbol"`
	Side       string    `db:"side"`
	EntryPrice float64   `db:"entry_price"`
	ExitPrice  float64   `db:"exit_price"`
	Qty        float64   `db:"qty"`
	Profit     float64   `db:"profit"`
	Commission float64   `db:"commission"`
	Timestamp  time.Time `db:"timestamp"`
}
