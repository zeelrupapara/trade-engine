package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"gitlab.com/zeelrupapara/trade-engine/constants"
	"gitlab.com/zeelrupapara/trade-engine/models"
	"gitlab.com/zeelrupapara/trade-engine/pkg/exchange"
	"gitlab.com/zeelrupapara/trade-engine/pkg/utils"
	"go.uber.org/zap"
)

// ---------- NATS SUBSCRIPTION ----------

func (ec *EngineCore) SubscribeToOrders() {
	_, err := ec.Nats.Subscribe("orders.*", func(msg *nats.Msg) {
		var order models.Order
		if err := json.Unmarshal(msg.Data, &order); err != nil {
			ec.Logger.Error("Invalid order format", zap.Error(err))
			return
		}
		ec.Logger.Info("📥 New order received", zap.String("order_id", order.OrderID), zap.String("symbol", order.Symbol))
		ec.ProcessNewOrder(order)
	})
	if err != nil {
		ec.Logger.Fatal("❌ Failed to subscribe to orders.", zap.Error(err))
	}
}

// ---------- PROCESS NEW ORDER ----------

func (ec *EngineCore) OnlyAllowLimitedNumberOfPositions() bool {
	return !(len(ec.Positions) >= constants.MaxNumberOfPositions)
}

func (ec *EngineCore) ProcessNewOrder(order models.Order) {
	if ec.OnlyAllowLimitedNumberOfPositions() {
		pos := &models.Position{
			ID:         utils.GenerateUUID(),
			OrderID:    order.OrderID,
			AccountID:  order.AccountID,
			Exchange:   order.Exchange,
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
}

func (ec *EngineCore) InsertPositionToDB(pos *models.Position) error {
	tx, err := ec.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()
	pos.CreatedAt = now
	pos.UpdatedAt = now

	// Insert position
	positionRecord := goqu.Record{
		"id":          pos.ID,
		"created_at":  pos.CreatedAt,
		"updated_at":  pos.UpdatedAt,
		"exchange":    pos.Exchange,
		"order_id":    pos.OrderID,
		"symbol":      pos.Symbol,
		"side":        pos.Side,
		"qty":         pos.Qty,
		"entry_price": pos.EntryPrice,
		"account_id":  pos.AccountID,
		"status":      pos.Status,
		"reason":      pos.Reason,
	}
	if _, err := tx.Insert("positions").Rows(positionRecord).Executor().Exec(); err != nil {
		return err
	}

	// Entry deal
	entryDeal := goqu.Record{
		"id":          uuid.New().String(),
		"order_id":    pos.OrderID,
		"account_id":  pos.AccountID,
		"symbol":      pos.Symbol,
		"side":        pos.Side,
		"entry_price": pos.EntryPrice,
		"exit_price":  0, // Not applicable yet
		"qty":         pos.Qty,
		"profit":      0, // Not applicable yet
		"commission":  0, // Not applicable yet
		"timestamp":   now,
	}
	if _, err := tx.Insert("deals").Rows(entryDeal).Executor().Exec(); err != nil {
		return err
	}

	// Deduct cost
	cost := pos.EntryPrice * pos.Qty
	newBalance := ec.Account.Balance - cost
	if _, err := tx.Update("accounts").
		Set(goqu.Record{
			"balance":    newBalance,
			"updated_at": now,
		}).
		Where(goqu.C("id").Eq(pos.AccountID)).
		Executor().
		Exec(); err != nil {
		return err
	}

	// Commit transaction and update memory
	if err := tx.Commit(); err != nil {
		return err
	}
	ec.Account.Balance = newBalance
	ec.Account.UpdatedAt = now

	return nil
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

// ---------- POSITION CLOSURE + DEAL + ACCOUNT UPDATE ----------

func (ec *EngineCore) CloseAndRecordPosition(id string, pos *models.Position, exitPrice float64) {
	tx, err := ec.DB.Begin()
	if err != nil {
		ec.Logger.Error("❌ Begin TX failed", zap.Error(err))
		return
	}
	defer tx.Rollback()

	now := time.Now()
	var profit float64
	if pos.Side == "buy" {
		profit = (exitPrice - pos.EntryPrice) * pos.Qty
	} else {
		profit = (pos.EntryPrice - exitPrice) * pos.Qty
	}
	commission := math.Abs(exitPrice * pos.Qty * 0.01)
	netProfit := profit - commission

	// Update position
	pos.ExitPrice = &exitPrice
	pos.Profit = &netProfit
	pos.Commission = &commission
	pos.Status = "closed"
	pos.ClosedAt = &now

	if _, err := tx.Update("positions").
		Set(goqu.Record{
			"exit_price": pos.ExitPrice,
			"profit":     pos.Profit,
			"commission": pos.Commission,
			"status":     pos.Status,
			"closed_at":  pos.ClosedAt,
			"updated_at": now,
		}).
		Where(goqu.C("order_id").Eq(pos.OrderID)).
		Executor().
		Exec(); err != nil {
		ec.Logger.Error("❌ Failed to update position", zap.Error(err))
		return
	}

	// Insert exit deal
	exitDeal := goqu.Record{
		"id":          uuid.New().String(),
		"order_id":    pos.OrderID,
		"account_id":  pos.AccountID,
		"symbol":      pos.Symbol,
		"side":        pos.Side,
		"entry_price": pos.EntryPrice,
		"exit_price":  exitPrice,
		"qty":         pos.Qty,
		"profit":      netProfit,
		"commission":  commission,
		"timestamp":   now,
	}
	if _, err := tx.Insert("deals").Rows(exitDeal).Executor().Exec(); err != nil {
		ec.Logger.Error("❌ Failed to insert deal", zap.Error(err))
		return
	}

	// Update account
	newBalance := ec.Account.Balance + netProfit
	if _, err := tx.Update("accounts").
		Set(goqu.Record{
			"balance":    newBalance,
			"updated_at": now,
		}).
		Where(goqu.C("id").Eq(pos.AccountID)).
		Executor().
		Exec(); err != nil {
		ec.Logger.Error("❌ Failed to update account", zap.Error(err))
		return
	}

	if err := tx.Commit(); err != nil {
		ec.Logger.Error("❌ TX Commit Failed", zap.Error(err))
		return
	}

	// Update memory and cleanup
	ec.Account.Balance = newBalance
	ec.Account.UpdatedAt = now
	delete(ec.Positions, id)

	ec.Logger.Info("✅ Position closed",
		zap.String("order_id", pos.OrderID),
		zap.Float64("exit", exitPrice),
		zap.Float64("profit", netProfit),
		zap.Float64("commission", commission),
	)
}

func (ec *EngineCore) UpdateClosedPositionInDB(pos *models.Position) error {
	_, err := ec.DB.Update("positions").
		Set(goqu.Record{
			"exit_price": pos.ExitPrice,
			"profit":     pos.Profit,
			"commission": pos.Commission,
			"status":     pos.Status,
			"closed_at":  pos.ClosedAt,
		}).
		Where(goqu.C("order_id").Eq(pos.OrderID)).
		Executor().
		Exec()
	return err
}

func (ec *EngineCore) ApplyProfitToAccount(accountID string, profit float64) error {
	if ec.Account == nil || ec.Account.ID != accountID {
		return fmt.Errorf("❌ account not loaded in memory or mismatched")
	}

	ec.Account.Balance += profit
	ec.Account.UpdatedAt = time.Now()

	_, err := ec.DB.Update("accounts").
		Set(goqu.Record{
			"balance":    ec.Account.Balance,
			"updated_at": ec.Account.UpdatedAt,
		}).
		Where(goqu.C("id").Eq(accountID)).
		Executor().
		Exec()
	return err
}
