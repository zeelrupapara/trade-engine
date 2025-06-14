package handlers

import (
	"math"

	"gitlab.com/zeelrupapara/trade-engine/models"
)

// ---------- STATIC STRATEGY (scalp_1m) CONFIG ----------

var scalp1m = struct {
	AllocationPercent float64 // % of total capital for this strategy
	RiskPercent       float64 // % of strategy capital risked per trade
	SLMultiplier      float64 // ATR × this for stop‐loss
	TPMultiplier      float64 // ATR × this for take‐profit
}{
	AllocationPercent: 0.25, // 25%
	RiskPercent:       0.01, // 1%
	SLMultiplier:      1.0,
	TPMultiplier:      2.0,
}

// ComputeOrderVolume calculates units based on ATR stop‐distance
// but now limits the capital to scalp1m.AllocationPercent of your account
// and risks only scalp1m.RiskPercent of that slice.
func (ec *EngineCore) ComputeOrderVolume(price, atr float64) float64 {
	if atr <= 0 || price <= 0 {
		return 0
	}

	// Total capital for this strategy
	stratCapital := ec.Account.Balance * scalp1m.AllocationPercent
	// Amount we are willing to lose on this trade
	riskAmount := stratCapital * scalp1m.RiskPercent
	// Stop‐loss distance in price units
	stopDist := atr * scalp1m.SLMultiplier

	units := riskAmount / stopDist
	// Round down to two decimals (adjust to your instrument precision)
	return math.Floor(units*100) / 100
}

// ComputeRiskLevels now uses the strategy’s multipliers on ATR
// to produce a tighter SL or a larger TP.
func (ec *EngineCore) ComputeRiskLevels(price, atr float64) (sl, tp float64) {
	sl = price - (atr * scalp1m.SLMultiplier)
	tp = price + (atr * scalp1m.TPMultiplier)
	return
}

// ComputeATR remains a rolling ATR but now explicitly handles
// “not enough candles” and timestamps the last computation.
func (ec *EngineCore) ComputeATR(candles []models.Kline) float64 {
	if len(candles) < 2 {
		return 0
	}
	var sumTR float64
	for i := 1; i < len(candles); i++ {
		hi, lo := candles[i].High, candles[i].Low
		prevC := candles[i-1].Close
		tr := math.Max(hi-lo, math.Max(math.Abs(hi-prevC), math.Abs(lo-prevC)))
		sumTR += tr
	}
	// Classic ATR = average of the TRs
	return sumTR / float64(len(candles)-1)
}
