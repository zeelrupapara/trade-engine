package models

type SymbolData struct {
	Klines    []Kline
	AggTrades []AggTrade
	Ticker    Ticker
	Depth     Depth
	Streams   SymbolStreams
}
