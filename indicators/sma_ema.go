package indicators

// SMA 计算简单移动平均线 (Simple Moving Average)
// period: 周期
func SMA(closes []float64, period int) []float64 {
	if len(closes) < period {
		return nil
	}

	sma := make([]float64, len(closes))
	for i := period - 1; i < len(closes); i++ {
		var sum float64
		for j := i - period + 1; j <= i; j++ {
			sum += closes[j]
		}
		sma[i] = sum / float64(period)
	}

	return sma
}

// GetLatestSMA 获取最新SMA值
func GetLatestSMA(closes []float64, period int) float64 {
	sma := SMA(closes, period)
	if len(sma) == 0 {
		return 0
	}
	return sma[len(sma)-1]
}

// GetLatestEMA 获取最新EMA值
func GetLatestEMA(closes []float64, period int) float64 {
	ema := EMA(closes, period)
	if len(ema) == 0 {
		return 0
	}
	return ema[len(ema)-1]
}
