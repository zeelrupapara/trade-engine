package models

import "time"

type SymbolStreams struct {
	KlineStop  chan struct{}
	AggStop    chan struct{}
	TickerStop chan struct{}
	DepthStop  chan struct{}
}

type Kline struct {
	Symbol    string
	StartTime time.Time
	EndTime   time.Time
	Open      float64
	Close     float64
	High      float64
	Low       float64
	Volume    float64
}

type AggTrade struct {
	Symbol    string
	Price     float64
	Quantity  float64
	Timestamp time.Time
}

type Ticker struct {
	Symbol    string
	BidPrice  float64
	AskPrice  float64
	Timestamp time.Time
}

type Depth struct {
	Symbol    string
	Bids      [][2]float64 // [Price, Quantity]
	Asks      [][2]float64
	Timestamp time.Time
}
