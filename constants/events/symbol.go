package constants

const (
	// symbols worker names or listener name
	SYMBOLS_WORKER = "symbols_worker"
)

type SymbolEvent struct {
	Symbol string `json:"symbol"`
	Exchange string `json:"exchange"`
	
}