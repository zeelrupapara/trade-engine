package handlers

import (
	"fmt"

	yamlconf "gitlab.com/zeelrupapara/trade-engine/config/exchange"
	"gitlab.com/zeelrupapara/trade-engine/constants"
	"gitlab.com/zeelrupapara/trade-engine/pkg/exchange"
	"go.uber.org/zap"
)

func (ec *EngineCore) AddExchangeClient(exchangeName string) {
	switch exchangeName {
	case constants.BINANCE:
		client := exchange.NewBinanceClient(&ec.Config, ec.Logger)
		ec.Exchange[constants.BINANCE] = client
	}
}

func (ec *EngineCore) LoadExchanges(exchanges *yamlconf.Config) {
	for _, exchange := range exchanges.Exchanges {
		if exchange.Active {
			ec.Logger.Info(fmt.Sprintf("Loading %s Client", exchange.Name))
			ec.AddExchangeClient(exchange.Name)

			ec.Logger.Info(fmt.Sprintf("Loading %s Symbols", exchange.Name))
			if err := ec.LoadExSymbols(exchange.Name); err != nil {
				ec.Logger.Fatal("Failed to load symbols", zap.String("exchange", exchange.Name), zap.Error(err))
			}
		}
	}
}
