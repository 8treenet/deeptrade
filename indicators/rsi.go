package indicators

import (
	"math"
)

// RSI 计算相对强弱指数 (Relative Strength Index)
// period: 计算周期，通常为14
// 返回值：RSI值数组
func RSI(closes []float64, period int) []float64 {
	if len(closes) < period+1 {
		return nil
	}

	rsi := make([]float64, len(closes))
	// 前period个值设为0
	for i := 0; i < period; i++ {
		rsi[i] = 0
	}

	// 计算初始平均涨幅和跌幅
	var gainSum, lossSum float64
	for i := 1; i <= period; i++ {
		change := closes[i] - closes[i-1]
		if change > 0 {
			gainSum += change
		} else {
			lossSum += -change
		}
	}

	avgGain := gainSum / float64(period)
	avgLoss := lossSum / float64(period)

	// 第一个RSI值（需要处理avgLoss为0的情况）
	if avgLoss == 0 {
		rsi[period] = 100 // 全部上涨，RSI为100
	} else {
		rs := avgGain / avgLoss
		rsi[period] = 100 - (100 / (1 + rs))
	}

	// 计算后续RSI值
	for i := period + 1; i < len(closes); i++ {
		change := closes[i] - closes[i-1]
		gain := math.Max(change, 0)
		loss := math.Max(-change, 0)

		// 平滑移动平均 (Wilder's smoothing)
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)

		// 计算RSI
		if avgLoss == 0 {
			rsi[i] = 100
		} else {
			rs := avgGain / avgLoss
			rsi[i] = 100 - (100 / (1 + rs))
		}
	}

	return rsi
}

// GetLatestRSI 获取最新RSI值
func GetLatestRSI(closes []float64, period int) float64 {
	rsi := RSI(closes, period)
	if len(rsi) == 0 {
		return 0
	}
	return rsi[len(rsi)-1]
}


// IsOverbought 判断是否超买 (RSI > 70)
func IsOverbought(rsi float64) bool {
	return rsi > 70
}

// IsOversold 判断是否超卖 (RSI < 30)
func IsOversold(rsi float64) bool {
	return rsi < 30
}
