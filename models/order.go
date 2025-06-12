package models

type Order struct {
	OrderID    string  `json:"order_id"`
	Exchange   string  `json:"exchange"`
	Symbol     string  `json:"symbol"`
	Side       string  `json:"side"`
	EntryPrice float64 `json:"entry_price"`
	SL         float64 `json:"sl,omitempty"`
	TP         float64 `json:"tp,omitempty"`
	Qty        float64 `json:"qty"`
	Reason     string  `json:"reason"`
}
