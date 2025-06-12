package handlers

import (
	"encoding/json"
	"fmt"
	"time"

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
			atr := ec.ComputeATR(symbolData.KlinesBuf.GetAll())
			quantity := ec.ComputeOrderVolume(symbolData.Ticker.BidPrice, atr)

			// 3. Apply risk management (SL/TP based on ATR)
			sl, tp := ec.ComputeRiskLevels(symbolData.Ticker.BidPrice, atr)

			// 4. Prepare order
			reason := signal.Reason
			order := models.Order{
				Exchange:   exchangeName,
				Symbol:     symbol,
				EntryPrice: symbolData.Ticker.BidPrice,
				Qty:        quantity,
				Side:       signal.Action,
				SL:         sl,
				TP:         tp,
				Reason:     signal.Reason,
			}

			ec.Logger.Sugar().Infof("Order: %+v | SL: %.2f | TP: %.2f | Reason: %s", order, sl, tp, reason)

			// 5. Insert order record into DB
			_, err := ec.DB.Insert("orders").Rows(goquRecordFromOrder(order, sl, tp, reason)).Executor().Exec()
			if err != nil {
				ec.Logger.Error("Failed to insert order", zap.Error(err))
				return
			}

			// 6. Publish order to NATS
			orderBytes, err := json.Marshal(order)
			if err != nil {
				ec.Logger.Error("Failed to marshal order", zap.Error(err))
				return
			}
			ec.Nats.Publish("signal."+symbol, orderBytes)
		}

		// Publish signal to the nats
		ec.Nats.Publish("signal."+symbol, []byte(fmt.Sprintf("%s", signal)))
	}
}

// goquRecordFromOrder converts models.Order to a map for DB insertion
func goquRecordFromOrder(order models.Order, sl, tp float64, reason string) map[string]interface{} {
	return map[string]interface{}{
		"exchange":  order.Exchange,
		"symbol":    order.Symbol,
		"price":     order.EntryPrice,
		"quantity":  order.Qty,
		"side":      order.Side,
		"type":      "MARKET",
		"sl":        sl,
		"tp":        tp,
		"reason":    reason,
		"timestamp": time.Now(),
	}
}
