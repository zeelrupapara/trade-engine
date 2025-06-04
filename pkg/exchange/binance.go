package exchange

import (
	"context"
	"fmt"

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

	// Loop and print symbols (example: filter for USDT pairs)
	fmt.Println("List of USDT trading pairs:")
	for _, symbol := range exchangeInfo.Symbols {
		symbolMap = append(symbolMap, symbol.Symbol)
	}
	return symbolMap, nil
}
