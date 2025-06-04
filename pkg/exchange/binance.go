package exchange

import (
	"context"

	"github.com/adshao/go-binance/v2"
	"gitlab.com/zeelrupapara/trade-engine/config"
)

type Binance struct {
	Client *binance.Client
}

func NewBinanceClient(cfg *config.AppConfig) *Binance {
	client := binance.NewClient(cfg.Binance.APIKey, cfg.Binance.APISecret)
	return &Binance{
		Client: client,
	}
}

func (b *Binance) GetSymbols() ([]string, error) {
	symbolMap := []string{}
	exchangeInfo, err := b.Client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		return symbolMap, err
	}

	for _, symbol := range exchangeInfo.Symbols {
		if symbol.Status == "TRADING" && symbol.QuoteAsset == "USDT" {
			symbolMap = append(symbolMap, symbol.Symbol)
		}
	}
	return symbolMap, nil
}
