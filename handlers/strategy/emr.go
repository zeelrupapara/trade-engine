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
			Reason: "Not enough data to compute EMR",
		}
	}

	ema9 := indicators.EMA(prices, 9)
	ema21 := indicators.EMA(prices, 21)
	rsi := indicators.RSI(prices, 14)
	_, _, macdHist := indicators.MACD(prices)

	logReason := fmt.Sprintf("EMA9=%.2f, EMA21=%.2f, MACD=%.2f, RSI=%.2f", ema9, ema21, macdHist, rsi)

	if ema9 > ema21 && macdHist > 0 && rsi < 70 {
		return models.StrategiesSignal{
			Action: "LONG",
			Reason: fmt.Sprintf("EMR long signal: %s", logReason),
		}
	}

	if ema9 < ema21 && macdHist < 0 && rsi > 30 {
		return models.StrategiesSignal{
			Action: "SHORT",
			Reason: fmt.Sprintf("EMR short signal: %s", logReason),
		}
	}

	return models.StrategiesSignal{
		Action: "NONE",
		Reason: fmt.Sprintf("EMR no signal: %s", logReason),
	}
}
