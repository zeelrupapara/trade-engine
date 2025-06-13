package handlers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
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
				OrderID:    utils.GenerateUUID(),
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
			if err := ec.InsertOrderToDB(order); err != nil {
				ec.Logger.Error("Failed to insert order to DB", zap.Error(err))
				return
			}

			// 6. Publish order to NATS
			orderBytes, err := json.Marshal(order)
			if err != nil {
				ec.Logger.Error("Failed to marshal order", zap.Error(err))
				return
			}
			ec.Nats.Publish("orders."+symbol, orderBytes)
		}

		// Publish signal to the nats
		ec.Nats.Publish("signal."+symbol, []byte(fmt.Sprintf("%s", signal)))
	}
}

func (ec *EngineCore) InsertOrderToDB(order models.Order) error {
	record := map[string]interface{}{
		"order_id":    order.OrderID,
		"exchange":    order.Exchange,
		"account_id":  order.AccountID,
		"symbol":      order.Symbol,
		"side":        order.Side,
		"entry_price": order.EntryPrice,
		"sl":          order.SL,
		"tp":          order.TP,
		"qty":         order.Qty,
		"reason":      order.Reason,
		"status":      "new",
		"timestamp":   time.Now(),
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	}

	_, err := ec.DB.Insert(goqu.T("orders")).Rows(record).Executor().Exec()
	if err != nil {
		ec.Logger.Error("❌ Failed to insert order", zap.Error(err))
	}
	return err
}
