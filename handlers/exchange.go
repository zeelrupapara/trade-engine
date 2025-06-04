package handlers

import (
	"gitlab.com/zeelrupapara/trade-engine/constants"
	"gitlab.com/zeelrupapara/trade-engine/pkg/exchange"
)

func (ec *EngineCore) AddExchangeClient(exchangeName string, active bool) {
	if !active {
		return
	}
	switch exchangeName {
	case constants.BINANCE:
		client := exchange.NewBinanceClient(&ec.Config)
		ec.Exchange.Clients[constants.BINANCE] = client
	}
}
