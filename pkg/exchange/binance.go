package exchange

import (
	"context"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2"
	"gitlab.com/zeelrupapara/trade-engine/config"
	"gitlab.com/zeelrupapara/trade-engine/models"
	"gitlab.com/zeelrupapara/trade-engine/pkg/utils"
	"go.uber.org/zap"
)

type Binance struct {
	Client  *binance.Client
	Symbols map[string]*models.SymbolData
	mu      sync.Mutex
	Logger  *zap.Logger
}

func NewBinanceClient(cfg *config.AppConfig, logger *zap.Logger) *Binance {
	client := binance.NewClient(cfg.Binance.APIKey, cfg.Binance.APISecret)
	logger.Info("Binance Client Created")
	return &Binance{
		Client:  client,
		Logger:  logger,
		Symbols: make(map[string]*models.SymbolData),
	}
}

func (b *Binance) MapExchangeSymbols() error {
	exchangeInfo, err := b.Client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		b.Logger.Error(err.Error(), zap.Any("Config", "LoadExSymbols"))
		return err
	}

	// Get Symbols and map in the core memory
	for _, symbol := range exchangeInfo.Symbols {
		if symbol.Status == "TRADING" && symbol.QuoteAsset == "USDT" {
			b.Symbols[symbol.Symbol] = &models.SymbolData{
				Klines:    []models.Kline{},
				AggTrades: []models.AggTrade{},
				Ticker:    models.Ticker{},
				Depth:     models.Depth{},
				Streams:   models.SymbolStreams{},
			}
		}
	}

	return nil
}

func (b *Binance) SubscribeSymbol(symbol string) {
	b.mu.Lock()
	if _, exists := b.Symbols[symbol]; exists {
		b.mu.Unlock()
		b.Logger.Info("Symbol already subscribed", zap.String("symbol", symbol))
		return
	}
	data := &models.SymbolData{
		Streams: models.SymbolStreams{
			KlineStop:  make(chan struct{}),
			AggStop:    make(chan struct{}),
			TickerStop: make(chan struct{}),
			DepthStop:  make(chan struct{}),
		},
	}
	b.Symbols[symbol] = data
	b.mu.Unlock()

	go b.startKline(symbol, data.Streams.KlineStop)
	go b.startAggTrade(symbol, data.Streams.AggStop)
	go b.startTicker(symbol, data.Streams.TickerStop)
	go b.startDepth(symbol, data.Streams.DepthStop)
}

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

func (b *Binance) startKline(symbol string, stopChan chan struct{}) {
	doneC, _, err := binance.WsKlineServe(symbol, "1m", func(e *binance.WsKlineEvent) {
		b.mu.Lock()
		if data, ok := b.Symbols[symbol]; ok {
			data.Klines = append(data.Klines, models.Kline{
				Symbol:    e.Symbol,
				StartTime: time.UnixMilli(e.Kline.StartTime),
				EndTime:   time.UnixMilli(e.Kline.EndTime),
				Open:      utils.ParseFloat(e.Kline.Open),
				High:      utils.ParseFloat(e.Kline.High),
				Low:       utils.ParseFloat(e.Kline.Low),
				Close:     utils.ParseFloat(e.Kline.Close),
				Volume:    utils.ParseFloat(e.Kline.Volume),
			})
		}
		b.mu.Unlock()
	}, b.errorHandler(symbol, "kline"))
	if err != nil {
		b.Logger.Error("Failed to start kline stream", zap.String("symbol", symbol), zap.Error(err))
		return
	}
	<-stopChan
	close(doneC)
	b.Logger.Info("Stopped kline stream", zap.String("symbol", symbol))
}

func (b *Binance) startAggTrade(symbol string, stopChan chan struct{}) {
	doneC, _, err := binance.WsAggTradeServe(symbol, func(e *binance.WsAggTradeEvent) {
		b.mu.Lock()
		if data, ok := b.Symbols[symbol]; ok {
			data.AggTrades = append(data.AggTrades, models.AggTrade{
				Symbol:    e.Symbol,
				Price:     utils.ParseFloat(e.Price),
				Quantity:  utils.ParseFloat(e.Quantity),
				Timestamp: time.UnixMilli(e.Time),
			})
		}
		b.mu.Unlock()
	}, b.errorHandler(symbol, "aggTrade"))
	if err != nil {
		b.Logger.Error("Failed to start aggTrade stream", zap.String("symbol", symbol), zap.Error(err))
		return
	}
	<-stopChan
	close(doneC)
	b.Logger.Info("Stopped aggTrade stream", zap.String("symbol", symbol))
}

func (b *Binance) startTicker(symbol string, stopChan chan struct{}) {
	doneC, _, err := binance.WsBookTickerServe(symbol, func(e *binance.WsBookTickerEvent) {
		b.mu.Lock()
		if data, ok := b.Symbols[symbol]; ok {
			data.Ticker = models.Ticker{
				Symbol:    e.Symbol,
				BidPrice:  utils.ParseFloat(e.BestBidPrice),
				AskPrice:  utils.ParseFloat(e.BestAskPrice),
				Timestamp: time.Now(),
			}
		}
		b.mu.Unlock()
	}, b.errorHandler(symbol, "ticker"))
	if err != nil {
		b.Logger.Error("Failed to start ticker stream", zap.String("symbol", symbol), zap.Error(err))
		return
	}
	<-stopChan
	close(doneC)
	b.Logger.Info("Stopped ticker stream", zap.String("symbol", symbol))
}

func (b *Binance) startDepth(symbol string, stopChan chan struct{}) {
	doneC, _, err := binance.WsDepthServe(symbol, func(e *binance.WsDepthEvent) {
		b.mu.Lock()
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
		b.mu.Unlock()
	}, b.errorHandler(symbol, "depth"))
	if err != nil {
		b.Logger.Error("Failed to start depth stream", zap.String("symbol", symbol), zap.Error(err))
		return
	}
	<-stopChan
	close(doneC)
	b.Logger.Info("Stopped depth stream", zap.String("symbol", symbol))
}

func (b *Binance) errorHandler(symbol, stream string) func(error) {
	return func(err error) {
		b.Logger.Error("WebSocket error", zap.String("symbol", symbol), zap.String("stream", stream), zap.Error(err))
	}
}
