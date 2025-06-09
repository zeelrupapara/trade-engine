package utils

import "gitlab.com/zeelrupapara/trade-engine/models"

func GetPricesFromCandles(candles []models.Kline, interval int) []float64 {
	baseInterval := 5 // base candle interval in minutes

	if interval < baseInterval {
		// invalid interval, return empty slice or handle error as you want
		return nil
	}

	groupSize := interval / baseInterval
	if groupSize <= 0 {
		return nil
	}

	var result []float64
	for i := 0; i < len(candles); i += groupSize {
		if i+groupSize > len(candles) {
			break
		}
		lastCandle := candles[i+groupSize-1]
		result = append(result, lastCandle.Close)
	}
	return result
}
