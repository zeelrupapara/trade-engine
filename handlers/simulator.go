package handlers

import (
	"encoding/json"
	"math"
	"time"

	"github.com/nats-io/nats.go"
	"gitlab.com/zeelrupapara/trade-engine/constants"
	"gitlab.com/zeelrupapara/trade-engine/models"
	"gitlab.com/zeelrupapara/trade-engine/pkg/exchange"
	"go.uber.org/zap"
)

func (ec *EngineCore) SubscribeToOrders() {
	_, err := ec.Nats.Subscribe("orders.new", func(msg *nats.Msg) {
		var order models.Order
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			ec.Logger.Error("Invalid order format", zap.Error(err))
			return
		}
		ec.ProcessNewOrder(order)
	})
	if err != nil {
		ec.Logger.Fatal("❌ Failed to subscribe to orders.new", zap.Error(err))
	}
}

func (ec *EngineCore) ProcessNewOrder(order models.Order) {
	pos := &models.Position{
		OrderID:    order.OrderID,
		Symbol:     order.Symbol,
		Side:       order.Side,
		EntryPrice: order.EntryPrice,
		Qty:        order.Qty,
		Status:     "open",
		Reason:     order.Reason,
	}

	if err := ec.InsertPositionToDB(pos); err != nil {
		ec.Logger.Error("❌ Failed to save open position", zap.Error(err))
		return
	}

	ec.Positions[order.OrderID] = pos
	ec.Logger.Info("📥 Position opened", zap.String("order_id", order.OrderID), zap.String("symbol", order.Symbol))
}

func (ec *EngineCore) InsertPositionToDB(pos *models.Position) error {
	return ec.DB.Insert(pos).Error()
}

// ---------- SL/TP WATCHER ----------

func (ec *EngineCore) StartPriceWatcher() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ec.CheckAllPositions()
			case <-ec.StopCh:
				ec.Logger.Info("🛑 Watcher exiting")
				return
			}
		}
	}()
}

func (ec *EngineCore) CheckAllPositions() {
	for id, pos := range ec.Positions {
		price := 0.0
		switch pos.Exchange {
		case constants.BINANCE:
			switch pos.Side {
			case "buy":
				price = ec.Exchange[constants.BINANCE].(exchange.Binance).Symbols[pos.Symbol].Ticker.BidPrice
			case "sell":
				price = ec.Exchange[constants.BINANCE].(exchange.Binance).Symbols[pos.Symbol].Ticker.AskPrice
			}
		}
		if ShouldClose(pos, price) {
			ec.CloseAndRecordPosition(id, pos, price)
		}
	}
}

func ShouldClose(pos *models.Position, price float64) bool {
	switch pos.Side {
	case "buy":
		return (pos.EntryPrice > 0 && price <= pos.EntryPrice-0.01) || (price >= pos.EntryPrice+0.01)
	case "sell":
		return (pos.EntryPrice > 0 && price >= pos.EntryPrice+0.01) || (price <= pos.EntryPrice-0.01)
	default:
		return false
	}
}

// ---------- POSITION CLOSURE ----------

func (ec *EngineCore) CloseAndRecordPosition(id string, pos *models.Position, price float64) {
	now := time.Now()
	var profit float64

	if pos.Side == "buy" {
		profit = (price - pos.EntryPrice) * pos.Qty
	} else {
		profit = (pos.EntryPrice - price) * pos.Qty
	}

	commission := math.Abs(price * pos.Qty * 0.01)
	netProfit := profit - commission

	pos.ExitPrice = &price
	pos.Profit = &netProfit
	pos.Commission = &commission
	pos.Status = "closed"
	pos.ClosedAt = &now

	if err := ec.UpdateClosedPositionInDB(pos); err != nil {
		ec.Logger.Error("❌ Failed to update closed position", zap.Error(err))
		return
	}

	delete(ec.Positions, id)
	ec.Logger.Info("✅ Position closed",
		zap.String("order_id", pos.OrderID),
		zap.Float64("exit", price),
		zap.Float64("profit", netProfit),
		zap.Float64("commission", commission),
	)
}

func (ec *EngineCore) UpdateClosedPositionInDB(pos *models.Position) error {
	return ec.DB.Insert(pos).Error()
}
