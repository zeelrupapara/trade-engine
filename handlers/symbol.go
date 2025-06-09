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

func (ec *EngineCore) SubcribeSymbol(exchangeName, symbol string) {
	switch exchangeName {
	case constants.BINANCE:
		ec.Exchange[constants.BINANCE].(*exchange.Binance).SubscribeSymbol(symbol)
		// take small day for data collection time.sleep(10 sec)
		ec.Logger.Info("Init Bot Workflow", zap.String("symbol", symbol), zap.String("exchange", exchangeName), zap.Int("interval", ec.Exchange[constants.BINANCE].(*exchange.Binance).Symbols[symbol].Settings.Interval))
		ec.InitBotWorkflow(symbol, exchangeName, ec.Exchange[constants.BINANCE].(*exchange.Binance).Symbols[symbol].Settings.Interval)
	}
}

func (ec *EngineCore) UnsubcribeSymbol(exchangeName, symbol string) {
	switch exchangeName {
	case constants.BINANCE:
		ec.Exchange[constants.BINANCE].(*exchange.Binance).UnsubscribeSymbol(symbol)
		ec.Exchange[constants.BINANCE].(*exchange.Binance).Symbols[symbol].Settings.WorkflowCloseCh <- 1
	}
}
