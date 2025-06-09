// exchange/binance.go
package exchange

import (
	"context"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"go.uber.org/zap"

	"gitlab.com/zeelrupapara/trade-engine/config"
	"gitlab.com/zeelrupapara/trade-engine/constants"
	"gitlab.com/zeelrupapara/trade-engine/models"
	"gitlab.com/zeelrupapara/trade-engine/pkg/ringbuf"
	"gitlab.com/zeelrupapara/trade-engine/pkg/utils"
)

// Constants to set buffer sizes. Adjust as needed.
const (
	MaxKlines    = 1000
	MaxAggTrades = 1000
)

// Binance wraps go-binance client + per-symbol buffers.
type Binance struct {
	Client  *binance.Client
	Symbols map[string]*models.SymbolData
	mu      sync.Mutex
	Logger  *zap.Logger
}

// NewBinanceClient sets up the Binance client and the Symbols map.
func NewBinanceClient(cfg *config.AppConfig, logger *zap.Logger) *Binance {
	client := binance.NewClient(cfg.Binance.APIKey, cfg.Binance.APISecret)
	b := &Binance{
		Client:  client,
		Logger:  logger,
		Symbols: make(map[string]*models.SymbolData),
	}
	return b
}

// MapExchangeSymbols fetches all USDT-quoted trading symbols and initializes their SymbolData.
func (b *Binance) MapExchangeSymbols() error {
	if b.Symbols == nil {
		b.Logger.Warn("Symbols map was nil, initializing")
		b.Symbols = make(map[string]*models.SymbolData)
	}

	exchangeInfo, err := b.Client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		b.Logger.Error(err.Error(), zap.Any("Config", "LoadExSymbols"))
		return err
	}

	for _, symbol := range exchangeInfo.Symbols {
		if symbol.Status == "TRADING" && symbol.QuoteAsset == "USDT" {
			b.Logger.Debug("Symbol Ready to Map", zap.String("symbol", symbol.Symbol))

			// Initialize SymbolData with two ring buffers of fixed size.
			b.Symbols[symbol.Symbol] = &models.SymbolData{
				KlinesBuf:    ringbuf.New[models.Kline](MaxKlines),
				AggTradesBuf: ringbuf.New[models.AggTrade](MaxAggTrades),
				Ticker:       models.Ticker{},
				Depth:        models.Depth{},
				Streams: models.SymbolStreams{
					KlineStop:  make(chan struct{}),
					AggStop:    make(chan struct{}),
					TickerStop: make(chan struct{}),
					DepthStop:  make(chan struct{}),
				},
				Settings: &models.SymbolSettings{
					TradeCount:      0,
					Interval:        5,
					Strategy:        models.StrategyType(constants.EMR),
					WorkflowCloseCh: make(chan int, 1),
				},
			}
		}
	}
	return nil
}

// SubscribeSymbol starts WebSocket streams for a given symbol, using ring buffers.
func (b *Binance) SubscribeSymbol(symbol string) {
	b.mu.Lock()
	data := &models.SymbolData{
		// Create new ring buffers when subscribing a single symbol at runtime:
		KlinesBuf:    ringbuf.New[models.Kline](MaxKlines),
		AggTradesBuf: ringbuf.New[models.AggTrade](MaxAggTrades),

		// Initialize ticker & depth so there's a zero value until the first update arrives:
		Ticker: models.Ticker{},
		Depth:  models.Depth{},

		Streams: models.SymbolStreams{
			KlineStop:  make(chan struct{}),
			AggStop:    make(chan struct{}),
			TickerStop: make(chan struct{}),
			DepthStop:  make(chan struct{}),
		},
		Settings: &models.SymbolSettings{
			TradeCount:      0,
			Interval:        5,
			Strategy:        models.StrategyType(constants.EMR),
			WorkflowCloseCh: make(chan int, 1),
		},
	}
	b.Symbols[symbol] = data
	b.mu.Unlock()

	// Launch each streaming goroutine:
	go b.startKline(symbol, data.Streams.KlineStop)
	go b.startAggTrade(symbol, data.Streams.AggStop)
	go b.startTicker(symbol, data.Streams.TickerStop)
	go b.startDepth(symbol, data.Streams.DepthStop)

	b.Logger.Info("Successfully Subscribed symbol", zap.String("symbol", symbol))
}

// UnsubscribeSymbol closes all stop channels and removes the SymbolData entry.
func (b *Binance) UnsubscribeSymbol(symbol string) {
	b.mu.Lock()
	data, ok := b.Symbols[symbol]
	if !ok {
		b.mu.Unlock()
		b.Logger.Warn("Symbol not subscribed", zap.String("symbol", symbol))
		return
	}
	close(data.Streams.KlineStop)
	close(data.Streams.AggStop)
	close(data.Streams.TickerStop)
	close(data.Streams.DepthStop)
	delete(b.Symbols, symbol)
	b.mu.Unlock()
	b.Logger.Info("Unsubscribed symbol", zap.String("symbol", symbol))
}

// startKline listens to 1m candlestick events and stores them in the ring buffer.
func (b *Binance) startKline(symbol string, stopChan chan struct{}) {
	doneC, _, err := binance.WsKlineServe(symbol, "5m",
		func(e *binance.WsKlineEvent) {
			b.mu.Lock()
			defer b.mu.Unlock()

			if data, ok := b.Symbols[symbol]; ok {
				k := models.Kline{
					Symbol:    e.Symbol,
					StartTime: time.UnixMilli(e.Kline.StartTime),
					EndTime:   time.UnixMilli(e.Kline.EndTime),
					Open:      utils.ParseFloat(e.Kline.Open),
					High:      utils.ParseFloat(e.Kline.High),
					Low:       utils.ParseFloat(e.Kline.Low),
					Close:     utils.ParseFloat(e.Kline.Close),
					Volume:    utils.ParseFloat(e.Kline.Volume),
				}
				// Instead of append, use ring buffer:
				data.KlinesBuf.Add(k)
			}
		},
		b.errorHandler(symbol, "kline"),
	)
	if err != nil {
		b.Logger.Error("Failed to start kline stream", zap.String("symbol", symbol), zap.Error(err))
		return
	}

	// Block until unsubscribe
	<-stopChan
	close(doneC)
	b.Logger.Info("Stopped kline stream", zap.String("symbol", symbol))
}

// startAggTrade listens to aggregated trade events and stores them in the ring buffer.
func (b *Binance) startAggTrade(symbol string, stopChan chan struct{}) {
	doneC, _, err := binance.WsAggTradeServe(symbol,
		func(e *binance.WsAggTradeEvent) {
			b.mu.Lock()
			defer b.mu.Unlock()

			if data, ok := b.Symbols[symbol]; ok {
				t := models.AggTrade{
					Symbol:    e.Symbol,
					Price:     utils.ParseFloat(e.Price),
					Quantity:  utils.ParseFloat(e.Quantity),
					Timestamp: time.UnixMilli(e.Time),
				}
				// Use ring buffer instead of slice append:
				data.AggTradesBuf.Add(t)
			}
		},
		b.errorHandler(symbol, "aggTrade"),
	)
	if err != nil {
		b.Logger.Error("Failed to start aggTrade stream", zap.String("symbol", symbol), zap.Error(err))
		return
	}

	// Block until unsubscribe
	<-stopChan
	close(doneC)
	b.Logger.Info("Stopped aggTrade stream", zap.String("symbol", symbol))
}

// startTicker replaces the in-memory ticker on each update (no buffer needed).
func (b *Binance) startTicker(symbol string, stopChan chan struct{}) {
	doneC, _, err := binance.WsBookTickerServe(symbol,
		func(e *binance.WsBookTickerEvent) {
			b.mu.Lock()
			defer b.mu.Unlock()

			if data, ok := b.Symbols[symbol]; ok {
				data.Ticker = models.Ticker{
					Symbol:    e.Symbol,
					BidPrice:  utils.ParseFloat(e.BestBidPrice),
					AskPrice:  utils.ParseFloat(e.BestAskPrice),
					Timestamp: time.Now(),
				}
			}
		},
		b.errorHandler(symbol, "ticker"),
	)
	if err != nil {
		b.Logger.Error("Failed to start ticker stream", zap.String("symbol", symbol), zap.Error(err))
		return
	}

	// Block until unsubscribe
	<-stopChan
	close(doneC)
	b.Logger.Info("Stopped ticker stream", zap.String("symbol", symbol))
}

// startDepth replaces the in-memory depth snapshot on each update (no buffer needed).
func (b *Binance) startDepth(symbol string, stopChan chan struct{}) {
	doneC, _, err := binance.WsDepthServe(symbol,
		func(e *binance.WsDepthEvent) {
			b.mu.Lock()
			defer b.mu.Unlock()

			if data, ok := b.Symbols[symbol]; ok {
				bids := make([][2]float64, len(e.Bids))
				asks := make([][2]float64, len(e.Asks))
				for i, v := range e.Bids {
					bids[i][0] = utils.ParseFloat(v.Price)
					bids[i][1] = utils.ParseFloat(v.Quantity)
				}
				for i, v := range e.Asks {
					asks[i][0] = utils.ParseFloat(v.Price)
					asks[i][1] = utils.ParseFloat(v.Quantity)
				}
				data.Depth = models.Depth{
					Symbol:    symbol,
					Bids:      bids,
					Asks:      asks,
					Timestamp: time.Now(),
				}
			}
		},
		b.errorHandler(symbol, "depth"),
	)
	if err != nil {
		b.Logger.Error("Failed to start depth stream", zap.String("symbol", symbol), zap.Error(err))
		return
	}

	// Block until unsubscribe
	<-stopChan
	close(doneC)
	b.Logger.Info("Stopped depth stream", zap.String("symbol", symbol))
}

// errorHandler returns a callback that logs WebSocket errors for a symbol.
func (b *Binance) errorHandler(symbol, stream string) func(error) {
	return func(err error) {
		b.Logger.Error("WebSocket error", zap.String("symbol", symbol), zap.String("stream", stream), zap.Error(err))
	}
}
