package handlers

import (
	"fmt"

	"gitlab.com/zeelrupapara/trade-engine/constants"
	handlers "gitlab.com/zeelrupapara/trade-engine/handlers/strategy"
	"gitlab.com/zeelrupapara/trade-engine/models"
	"gitlab.com/zeelrupapara/trade-engine/pkg/exchange"
	"gitlab.com/zeelrupapara/trade-engine/pkg/utils"
	"go.uber.org/zap"
)

func (ec *EngineCore) ComputeSignal(exchangeName, symbol string, interval int) {
	switch exchangeName {
	case constants.BINANCE:
		symbolData := ec.Exchange[constants.BINANCE].(*exchange.Binance).Symbols[symbol]
		candles := symbolData.KlinesBuf.GetAll()
		// compute the candles close prices based on the interval
		prices := utils.GetPricesFromCandles(candles, interval)
		signal := models.StrategiesSignal{}
		switch symbolData.Settings.Strategy {
		case models.StrategyType(constants.EMR):
			ec.Logger.Info("Compute EMR Signal", zap.String("symbol", symbol), zap.String("exchange", exchangeName), zap.Int("interval", interval))
			// choose sterategy
			signal = handlers.ComputeEMRDecision(prices)
			ec.Logger.Sugar().Infof("Signal: %v", signal, zap.String("symbol", symbol), zap.String("exchange", exchangeName), zap.Int("interval", interval))
		}

		// Publish signal to the nats
		ec.Nats.Publish("signal."+symbol, []byte(fmt.Sprintf("%s", signal)))
	}
}
