package indicators

import "math"

// CCI 计算商品通道指数 (Commodity Channel Index)
// period: 计算周期，通常为20
// 返回值：CCI值数组
func CCI(highs, lows, closes []float64, period int) []float64 {
	if len(highs) < period || len(lows) < period || len(closes) < period {
		return nil
	}

	dataLen := len(closes)
	cci := make([]float64, dataLen)

	// 计算每个时间点的典型价格
	typicalPrices := make([]float64, dataLen)
	for i := 0; i < dataLen; i++ {
		typicalPrices[i] = (highs[i] + lows[i] + closes[i]) / 3.0
	}

	// 计算CCI值
	for i := period - 1; i < dataLen; i++ {
		// 计算最近period个典型价格的简单移动平均
		var sma float64
		for j := i - period + 1; j <= i; j++ {
			sma += typicalPrices[j]
		}
		sma /= float64(period)

		// 计算平均绝对偏差
		var mad float64
		for j := i - period + 1; j <= i; j++ {
			mad += math.Abs(typicalPrices[j] - sma)
		}
		mad /= float64(period)

		// 计算CCI
		if mad != 0 {
			cci[i] = (typicalPrices[i] - sma) / (0.015 * mad)
		} else {
			cci[i] = 0
		}
	}

	return cci
}

// GetLatestCCI 获取最新CCI值
func GetLatestCCI(highs, lows, closes []float64, period int) float64 {
	cci := CCI(highs, lows, closes, period)
	if len(cci) == 0 {
		return 0
	}
	return cci[len(cci)-1]
}

// IsCCIOverbought 判断CCI是否超买（CCI > 100）
func IsCCIOverbought(cci float64) bool {
	return cci > 100
}

// IsCCIOversold 判断CCI是否超卖（CCI < -100）
func IsCCIOversold(cci float64) bool {
	return cci < -100
}
