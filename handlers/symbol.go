package handlers

import (
	"errors"
	"fmt"

	"gitlab.com/zeelrupapara/trade-engine/constants"
	"gitlab.com/zeelrupapara/trade-engine/pkg/exchange"
	"go.uber.org/zap"
)

func (ec *EngineCore) LoadExSymbols(exchangeName string, active bool) error {
	if !active {
		return errors.New(fmt.Sprintf("Symbol Load Skipped: Exchange %s is not active", exchangeName))
	}

	switch exchangeName {
	case constants.BINANCE:
		// Remove the old symbols and update
		ec.Exchange.Symbols[constants.BINANCE] = []string{}
		binanceSymbols, err := ec.Exchange.Clients[constants.BINANCE].(*exchange.Binance).GetSymbols()
		if err != nil {
			ec.Logger.Debug(err.Error(), zap.Any("Symbol", "UpdatesExSymbols()"))
		}
		ec.Exchange.Symbols[constants.BINANCE] = binanceSymbols
	}
	return nil
}
