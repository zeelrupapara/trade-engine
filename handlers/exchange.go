package handlers

import (
	"gitlab.com/zeelrupapara/trade-engine/constants"
	"gitlab.com/zeelrupapara/trade-engine/pkg/exchange"
)

func (ec *EngineCore) AddExchangeClient(exchangeName string) {
	switch exchangeName {
	case constants.BINANCE:
		client := exchange.NewBinanceClient(&ec.Config, ec.Logger)
		ec.Exchange[constants.BINANCE] = client
	}
}
