package models

type Exchange struct {
	// Client Mapping for connection like map["binance"] = *binance.Client
	Clients map[string]interface{}
	// List of Symbols [exchange] = List of Symbols
	Symbols map[string][]string
	// Active or Disabled
	Active bool
}
