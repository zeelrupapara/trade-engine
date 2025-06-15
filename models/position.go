package models

import "time"

type Position struct {
	ID         string     `db:"id"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	ClosedAt   *time.Time `db:"closed_at"`
	Exchange   string     `db:"exchange"`
	OrderID    string     `db:"order_id"`
	Symbol     string     `db:"symbol"`
	Side       string     `db:"side"`
	Qty        float64    `db:"qty"`
	EntryPrice float64    `db:"entry_price"`
	ExitPrice  *float64   `db:"exit_price"`
	Profit     *float64   `db:"profit"`
	Commission *float64   `db:"commission"`
	AccountID  string     `db:"account_id"`
	Status     string     `db:"status"`
	Reason     string     `db:"reason"`
}
