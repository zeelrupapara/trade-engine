package handlers

import "math"

// EMA computes the Exponential Moving Average for the given data and period.
// EMA = (price - prevEMA) * (2/(period+1)) + prevEMA:contentReference[oaicite:0]{index=0}.
func EMA(data []float64, period int) float64 {
	if len(data) < period {
		return 0.0
	}
	// Calculate initial SMA for first EMA period
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	ema := sum / float64(period)
	multiplier := 2.0 / float64(period+1)
	// Apply EMA formula to remaining data points
	for i := period; i < len(data); i++ {
		ema = (data[i]-ema)*multiplier + ema
	}
	return ema
}

// RSI calculates the Relative Strength Index for the given data and period.
// RSI = 100 - 100/(1+RS), where RS = avg gain / avg loss:contentReference[oaicite:1]{index=1}.
func RSI(data []float64, period int) float64 {
	if len(data) < period+1 {
		return 0.0
	}
	gains, losses := 0.0, 0.0
	// Initial average gain and loss (simple)
	for i := 1; i <= period; i++ {
		delta := data[i] - data[i-1]
		if delta > 0 {
			gains += delta
		} else {
			losses -= delta
		}
	}
	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)
	// Wilder's smoothing for subsequent values (if needed)
	for i := period + 1; i < len(data); i++ {
		delta := data[i] - data[i-1]
		if delta > 0 {
			avgGain = (avgGain*float64(period-1) + delta) / float64(period)
			avgLoss = (avgLoss * (float64(period - 1))) / float64(period)
		} else {
			avgGain = (avgGain * (float64(period - 1))) / float64(period)
			avgLoss = (avgLoss*(float64(period-1)) - delta) / float64(period)
		}
	}
	if avgLoss == 0 {
		return 100.0
	}
	rs := avgGain / avgLoss
	return 100.0 - 100.0/(1.0+rs)
}

// MACD computes the MACD line, signal line, and histogram from closing prices.
// MACD line = EMA12 - EMA26; Signal = EMA9 of MACD line:contentReference[oaicite:2]{index=2}.
func MACD(data []float64) (macdLine float64, signalLine float64, histo float64) {
	if len(data) < 26 {
		return 0, 0, 0
	}
	ema12 := EMA(data, 12)
	ema26 := EMA(data, 26)
	macdLine = ema12 - ema26
	// Compute MACD series to get signal line
	macdSeries := make([]float64, 0, len(data)-26)
	for i := 26; i < len(data); i++ {
		partEma12 := EMA(data[:i+1], 12)
		partEma26 := EMA(data[:i+1], 26)
		macdSeries = append(macdSeries, partEma12-partEma26)
	}
	if len(macdSeries) >= 9 {
		signalLine = EMA(macdSeries, 9)
	}
	histo = macdLine - signalLine
	return
}

// ATR computes the latest Average True Range over the given period.
// high[i], low[i], close[i] must be aligned time‑series slices of equal length.
func ATR(high, low, close []float64, period int) float64 {
	// need at least period+1 points to compute the first TR
	if len(high) != len(low) || len(low) != len(close) || len(close) < period+1 {
		return 0.0
	}

	// build True Range (TR) series
	tr := make([]float64, 0, len(close)-1)
	for i := 1; i < len(close); i++ {
		hl := high[i] - low[i]
		hc := math.Abs(high[i] - close[i-1])
		lc := math.Abs(low[i] - close[i-1])
		tr = append(tr, math.Max(hl, math.Max(hc, lc)))
	}

	// first ATR = simple average of first 'period' TRs
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += tr[i]
	}
	atr := sum / float64(period)

	// Wilder smoothing
	for i := period; i < len(tr); i++ {
		atr = (atr*float64(period-1) + tr[i]) / float64(period)
	}
	return atr
}
