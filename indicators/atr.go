package indicators

import (
	"math"
)

// ATR 计算平均真实波幅 (Average True Range)
// period: 计算周期，通常为14
func ATR(highs, lows, closes []float64, period int) []float64 {
	if len(highs) < period+1 || len(lows) < period+1 || len(closes) < period+1 {
		return nil
	}

	tr := make([]float64, len(highs))
	tr[0] = highs[0] - lows[0]

	// 计算真实波幅 (True Range)
	for i := 1; i < len(highs); i++ {
		hl := highs[i] - lows[i]
		hc := math.Abs(highs[i] - closes[i-1])
		lc := math.Abs(lows[i] - closes[i-1])

		// 取最大值
		tr[i] = math.Max(hl, math.Max(hc, lc))
	}

	// 计算ATR（Wilder's平滑）
	atr := make([]float64, len(tr))
	// 第一个ATR值为TR的简单移动平均
	var sum float64
	for i := 1; i <= period; i++ {
		sum += tr[i]
	}
	// Wilder首值：使用TR[1..period]的period个样本
	atr[period] = sum / float64(period)

	// 后续ATR值（Wilder平滑）
	for i := period + 1; i < len(tr); i++ {
		atr[i] = ((atr[i-1] * float64(period-1)) + tr[i]) / float64(period)
	}

	return atr
}

// GetLatestATR 获取最新ATR值
func GetLatestATR(highs, lows, closes []float64, period int) float64 {
	atr := ATR(highs, lows, closes, period)
	if len(atr) == 0 {
		return 0
	}
	return atr[len(atr)-1]
}

// CalculateStopLossPrice 计算动态止损价
// currentPrice: 当前价格
// atr: 平均真实波幅
// multiplier: 乘数（通常1.5-2.5）
// isLong: 是否做多
func CalculateStopLossPrice(currentPrice, atr float64, multiplier float64, isLong bool) float64 {
	stopDistance := atr * multiplier

	if isLong {
		return currentPrice - stopDistance
	} else {
		return currentPrice + stopDistance
	}
}

// CalculateTakeProfitPrice 计算动态止盈价
// currentPrice: 当前价格
// atr: 平均真实波幅
// multiplier: 乘数（通常1.5-2.5）
// isLong: 是否做多
func CalculateTakeProfitPrice(currentPrice, atr float64, multiplier float64, isLong bool) float64 {
	stopDistance := atr * multiplier

	if isLong {
		return currentPrice + stopDistance
	} else {
		return currentPrice - stopDistance
	}
}

// CalculateVolatilityPercent 计算波动率百分比
func CalculateVolatilityPercent(atr, price float64) float64 {
	if price == 0 {
		return 0
	}
	return (atr / price) * 100
}
