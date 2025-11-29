package indicators

import (
	"math"
)

// BollingerResult 布林带计算结果
type BollingerResult struct {
	UpperBand []float64 // 上轨
	LowerBand []float64 // 下轨
	MA        []float64 // 中轨 (移动平均线)
	Width     []float64 // 带宽
}

// BollingerBands 计算布林带
// period: 周期，通常为20
// stdDev: 标准差倍数，通常为2
func BollingerBands(closes []float64, period int, stdDev float64) *BollingerResult {
	if len(closes) < period {
		return nil
	}

	result := &BollingerResult{
		UpperBand: make([]float64, len(closes)),
		LowerBand: make([]float64, len(closes)),
		MA:        make([]float64, len(closes)),
		Width:     make([]float64, len(closes)),
	}

	for i := period - 1; i < len(closes); i++ {
		// 计算简单移动平均 (SMA)
		var sum float64
		for j := i - period + 1; j <= i; j++ {
			sum += closes[j]
		}
		ma := sum / float64(period)
		result.MA[i] = ma

		// 计算标准差
		var variance float64
		for j := i - period + 1; j <= i; j++ {
			diff := closes[j] - ma
			variance += diff * diff
		}
		std := math.Sqrt(variance / float64(period))

		// 计算上下轨
		result.UpperBand[i] = ma + (std * stdDev)
		result.LowerBand[i] = ma - (std * stdDev)

		// 计算带宽（防止除零）
		band := result.UpperBand[i] - result.LowerBand[i]
		if ma != 0 {
			result.Width[i] = band / ma
		} else {
			result.Width[i] = 0
		}
	}

	return result
}

// GetLatestBollingerBands 获取最新布林带值
func GetLatestBollingerBands(closes []float64, period int, stdDev float64) *BollingerResult {
	return BollingerBands(closes, period, stdDev)
}

// IsPriceAboveUpperBand 判断价格是否突破上轨
func IsPriceAboveUpperBand(price, upperBand float64) bool {
	return price > upperBand
}

// IsPriceBelowLowerBand 判断价格是否跌破下轨
func IsPriceBelowLowerBand(price, lowerBand float64) bool {
	return price < lowerBand
}
