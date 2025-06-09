package models

type StrategiesSignal struct {
	Action string `json:"action"` // "LONG", "SHORT", or "NONE"
	Reason string `json:"reason"` // e.g. "EMA9>EMA21, MACD>0, RSI<70"
}
