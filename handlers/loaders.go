package handlers

import (
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"gitlab.com/zeelrupapara/trade-engine/models"
	"gitlab.com/zeelrupapara/trade-engine/pkg/utils"
	"go.uber.org/zap"
)

// ---------- ACCOUNT LOAD / CREATE ----------

func (ec *EngineCore) LoadOrCreateAccountByName(name string) error {
	now := time.Now()

	var acc models.Account
	found, err := ec.DB.From("accounts").Where(goqu.C("name").Eq(name)).ScanStruct(&acc)
	if err != nil {
		return fmt.Errorf("❌ failed to fetch account: %w", err)
	}

	if found {
		ec.Account = &acc
		ec.Logger.Info("✅ Account loaded", zap.String("id", acc.ID), zap.String("name", acc.Name), zap.Float64("balance", acc.Balance))
		return nil
	}

	// Create new account
	newAcc := &models.Account{
		ID:        utils.GenerateUUID(),
		Name:      name,
		Balance:   100000.0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = ec.DB.Insert("accounts").Rows(goqu.Record{
		"id":         newAcc.ID,
		"name":       newAcc.Name,
		"balance":    newAcc.Balance,
		"created_at": newAcc.CreatedAt,
		"updated_at": newAcc.UpdatedAt,
	}).Executor().Exec()
	if err != nil {
		return fmt.Errorf("❌ failed to insert new account: %w", err)
	}

	ec.Account = newAcc
	ec.Logger.Info("✅ New account created and loaded", zap.String("id", newAcc.ID), zap.String("name", name), zap.Float64("balance", newAcc.Balance))
	return nil
}

// ---------- LOAD OPEN POSITIONS INTO MEMORY ----------

func (ec *EngineCore) LoadOpenPositions() error {
	var positions []models.Position
	err := ec.DB.From("positions").Where(goqu.C("status").Eq("open")).ScanStructs(&positions)
	if err != nil {
		return fmt.Errorf("❌ failed to fetch open positions: %w", err)
	}

	for _, pos := range positions {
		// Add to in-memory map
		ec.Positions[pos.OrderID] = &pos
	}

	ec.Logger.Info("✅ Loaded open positions", zap.Int("count", len(positions)))
	return nil
}
