package models

import "gitlab.com/zeelrupapara/trade-engine/pkg/ringbuf"

type SymbolData struct {
	// Instead of slices, we now use ring buffers to keep a fixed number of recent items.
	KlinesBuf    *ringbuf.RingBuffer[Kline]
	AggTradesBuf *ringbuf.RingBuffer[AggTrade]

	// We still overwrite the latest Ticker and Depth on each update.
	Ticker       Ticker
	Depth        Depth
	
	Streams      SymbolStreams
}
