package handlers

import (
	"math"

	"gitlab.com/zeelrupapara/trade-engine/models"
	"go.uber.org/zap"
)

// scalp1m holds allocation and risk settings for the 1m ATR scalp strategy.
var scalp1m = struct {
	AllocationPercent float64 // % of total capital for this strategy
	RiskPercent       float64 // % of strategy capital risked per trade
	SLMultiplier      float64 // ATR × this for stop‐loss
	TPMultiplier      float64 // ATR × this for take‐profit
}{
	AllocationPercent: 0.25, // 25% of account
	RiskPercent:       0.01, // 1% of strategy capital
	SLMultiplier:      4.0,  // 1×ATR stop loss
	TPMultiplier:      2.0,  // 2×ATR take profit
}

// ComputeOrderVolume calculates the order size so that risk per trade
// does not exceed scalp1m.RiskPercent of scalp1m.AllocationPercent of account.
// It sizes by ATR‐based stop distance: riskAmount = stopDistance×qty×price.
func (ec *EngineCore) ComputeOrderVolume(price, atr float64) float64 {
	if price <= 0 || atr <= 0 {
		ec.Logger.Error("❌ Invalid inputs for order volume", zap.Float64("price", price), zap.Float64("atr", atr))
		return 0
	}

	// 1. Determine capital slice and per‐trade risk
	stratCapital := ec.Account.Balance * scalp1m.AllocationPercent // e.g. 25%
	riskAmount := stratCapital * scalp1m.RiskPercent               // e.g. 1% of that slice

	// 2. Compute stop‐loss price and stop distance
	sl := price - (atr * scalp1m.SLMultiplier)
	stopDistance := price - sl // absolute distance in price units

	// 3. Calculate quantity so that riskAmount = stopDistance * qty * price
	qty := riskAmount / (stopDistance * price)

	// 4. Round down to instrument precision (4 decimal places here)
	precision := 10000.0
	return math.Floor(qty*precision) / precision
}

// ComputeRiskLevels returns the stop‐loss and take‐profit levels based on ATR multipliers.
// ComputeRiskLevels calculates SL/TP based on direction and ATR
func (ec *EngineCore) ComputeRiskLevels(price, atr float64, direction string) (sl, tp float64) {
	switch direction {
	case "LONG":
		sl = price - (atr * scalp1m.SLMultiplier) // SL below entry
		tp = price + (atr * scalp1m.TPMultiplier) // TP above entry
	case "SHORT":
		sl = price + (atr * scalp1m.SLMultiplier) // SL above entry
		tp = price - (atr * scalp1m.TPMultiplier) // TP below entry
	default:
		ec.Logger.Error("❌ Unknown direction in ComputeRiskLevels", zap.String("direction", direction))
		sl, tp = price, price
	}
	return
}

// ComputeATR computes the classic ATR over provided Kline candles.
// If there are fewer than 2 candles, returns 0.
func (ec *EngineCore) ComputeATR(candles []models.Kline) float64 {
	if len(candles) < 2 {
		return 0
	}
	var sumTR float64
	for i := 1; i < len(candles); i++ {
		hi, lo, prev := candles[i].High, candles[i].Low, candles[i-1].Close
		tr := math.Max(hi-lo, math.Max(math.Abs(hi-prev), math.Abs(lo-prev)))
		sumTR += tr
	}
	return sumTR / float64(len(candles)-1)
}

// // Example usage in your trading loop:
// func (ec *EngineCore) OnNewCandle(candles []models.Kline) {
// 	atr := ec.ComputeATR(candles)
// 	price := candles[len(candles)-1].Close

// 	qty := ec.ComputeOrderVolume(price, atr)
// 	sl, tp := ec.ComputeRiskLevels(price, atr)

// 	ec.Logger.Info("Scalp1m trade parameters", zap.Float64("qty", qty), zap.Float64("sl", sl), zap.Float64("tp", tp), zap.Float64("atr", atr))

// 	// Then place your order with qty, sl, tp...
// 	_ = time.Now() // placeholder
// }
