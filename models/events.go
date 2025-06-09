package models

type SymbolEventType string

type StrategyType string

const (
	SymbolEventType_AddSymbol    SymbolEventType = "add"
	SymbolEventType_RemoveSymbol SymbolEventType = "remove"
)

type SymbolEvent struct {
	Type     SymbolEventType `json:"type"`
	Symbol   string          `json:"symbol"`
	Exchange string          `json:"exchange"`
}

// Subject

var (
	GroupSymbol string = "engine.symbol_group"
)

var (
	SubjectEngineSymbol string = "engine.symbol"
)
