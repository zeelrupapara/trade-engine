package handlers

import (
	"time"

	"gitlab.com/zeelrupapara/trade-engine/constants"
	"gitlab.com/zeelrupapara/trade-engine/pkg/exchange"
	"go.uber.org/zap"
)

func (ec *EngineCore) InitBotWorkflow(symbol, exchangeName string, interval int) {
	ec.Logger.Info("Starting Workflow", zap.String("symbol", symbol), zap.String("exchange", exchangeName), zap.Int("interval", interval))
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		var closeCh <-chan int
		switch exchangeName {
		case constants.BINANCE:
			closeCh = ec.Exchange[exchangeName].(*exchange.Binance).Symbols[symbol].Settings.WorkflowCloseCh
		default:
			closeCh = make(<-chan int)
		}

		for {
			select {
			case <-closeCh:
				ec.Logger.Info("Workflow Closed", zap.String("symbol", symbol), zap.String("exchange", exchangeName), zap.Int("interval", interval))
				return
			case <-ticker.C:
				ec.Logger.Info("Compute New Signal", zap.String("symbol", symbol), zap.String("exchange", exchangeName), zap.Int("interval", interval))
				ec.ComputeSignal(exchangeName, symbol, interval)
			}
		}
	}()
}
