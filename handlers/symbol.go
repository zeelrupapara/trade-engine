package handlers

import (
	"gitlab.com/zeelrupapara/trade-engine/constants"
	"gitlab.com/zeelrupapara/trade-engine/pkg/exchange"
	"go.uber.org/zap"
)

func (ec *EngineCore) LoadExSymbols(exchangeName string) error {
	switch exchangeName {
	case constants.BINANCE:
		// Remove the old symbols and update
		ec.Exchange[constants.BINANCE].(*exchange.Binance).Symbols = nil
		if err := ec.Exchange[constants.BINANCE].(*exchange.Binance).MapExchangeSymbols(); err != nil {
			ec.Logger.Error(err.Error(), zap.Any("Config", "LoadExSymbols"))
			return err
		}
	}
	return nil
}

func (ec *EngineCore) SubcribeSymbol(exchangeName, symbol string) error {
	switch exchangeName {
	case constants.BINANCE:
		ec.Exchange[constants.BINANCE].(*exchange.Binance).SubscribeSymbol(symbol)
	}
	return nil
}

func (ec *EngineCore) UnsubcribeSymbol(exchangeName, symbol string) error {
	switch exchangeName {
	case constants.BINANCE:
		ec.Exchange[constants.BINANCE].(*exchange.Binance).UnsubscribeSymbol(symbol)
	}
	return nil
}