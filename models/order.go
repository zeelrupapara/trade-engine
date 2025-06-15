package models

type Order struct {
	OrderID    string  `json:"order_id" db:"order_id"`
	Exchange   string  `json:"exchange" db:"exchange"`
	AccountID  string  `json:"account_id" db:"account_id"`
	Symbol     string  `json:"symbol" db:"symbol"`
	Side       string  `json:"side" db:"side"`
	EntryPrice float64 `json:"entry_price" db:"entry_price"`
	SL         float64 `json:"sl,omitempty" db:"sl"`
	TP         float64 `json:"tp,omitempty" db:"tp"`
	Qty        float64 `json:"qty" db:"qty"`
	Reason     string  `json:"reason" db:"reason"`
}
