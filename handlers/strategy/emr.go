package handlers

import (
	"fmt"

	"gitlab.com/zeelrupapara/trade-engine/models"
	"gitlab.com/zeelrupapara/trade-engine/pkg/indicators"
)

func ComputeEMRDecision(prices []float64) models.StrategiesSignal {
	if len(prices) < 30 {
		return models.StrategiesSignal{
			Action: "NONE",
			Reason: "Not enough data",
		}
	}

	ema9 := indicators.EMA(prices, 9)
	ema21 := indicators.EMA(prices, 21)
	rsi := indicators.RSI(prices, 14)
	_, _, macdHist := indicators.MACD(prices)

	if ema9 > ema21 && macdHist > 0 && rsi < 70 {
		return models.StrategiesSignal{
			Action: "LONG",
			Reason: fmt.Sprintf("EMA9=%.2f>EMA21=%.2f, MACD=%.2f>0, RSI=%.2f<70", ema9, ema21, macdHist, rsi),
		}
	}

	if ema9 < ema21 && macdHist < 0 && rsi > 30 {
		return models.StrategiesSignal{
			Action: "SHORT",
			Reason: fmt.Sprintf("EMA9=%.2f<EMA21=%.2f, MACD=%.2f<0, RSI=%.2f>30", ema9, ema21, macdHist, rsi),
		}
	}

	return models.StrategiesSignal{
		Action: "NONE",
		Reason: "Conditions not met",
	}
}
