package handlers

import (
	"math"

	"gitlab.com/zeelrupapara/trade-engine/models"
)

// ComputeOrderVolume - calculate quantity based on fixed risk per trade
func (ec *EngineCore) ComputeOrderVolume(price, atr float64) float64 {
	capital := 1000.0
	riskPercent := 0.01
	riskAmount := capital * riskPercent

	if atr == 0 || price == 0 {
		return 0
	}
	units := riskAmount / atr
	return math.Floor(units*100) / 100
}

// ComputeRiskLevels - basic SL/TP calculation using ATR
func (ec *EngineCore) ComputeRiskLevels(price, atr float64) (sl, tp float64) {
	sl = price - atr
	tp = price + (2 * atr)
	return
}

// ComputeATR - Average True Range based on historical candles
func (ec *EngineCore) ComputeATR(candles []models.Kline) float64 {
	atr := 0.0
	for i := 1; i < len(candles); i++ {
		tr := math.Max(candles[i].High-candles[i].Low, math.Abs(candles[i].High-candles[i-1].Close))
		tr = math.Max(tr, math.Abs(candles[i].Low-candles[i-1].Close))
		atr += tr
	}
	if len(candles) > 1 {
		atr = atr / float64(len(candles)-1)
	}
	return atr
}
