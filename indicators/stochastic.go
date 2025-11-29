package indicators

// StochasticResult Stochastic计算结果
type StochasticResult struct {
	K []float64 // %K线
	D []float64 // %D线
}

// Stochastic 计算随机指标 (Stochastic Oscillator)
// kPeriod: %K周期，通常为14
// dPeriod: %D周期，通常为3
// slowingPeriod: 减速周期，通常为3
func Stochastic(highs, lows, closes []float64, kPeriod, dPeriod, slowingPeriod int) *StochasticResult {
	if len(highs) < kPeriod || len(lows) < kPeriod || len(closes) < kPeriod {
		return nil
	}

	dataLen := len(closes)
	k := make([]float64, dataLen)

	// 计算%K值
	for i := kPeriod - 1; i < dataLen; i++ {
		// 找出最近k周期的最高价和最低价
		highestHigh := highs[i]
		lowestLow := lows[i]

		for j := i - kPeriod + 1; j <= i; j++ {
			if highs[j] > highestHigh {
				highestHigh = highs[j]
			}
			if lows[j] < lowestLow {
				lowestLow = lows[j]
			}
		}

		// 计算%K
		if highestHigh != lowestLow {
			k[i] = ((closes[i] - lowestLow) / (highestHigh - lowestLow)) * 100
		} else {
			k[i] = 50 // 如果没有变化，设为50
		}
	}

	// 对%K进行平滑处理（减速）
	if slowingPeriod > 1 {
		smoothedK := make([]float64, dataLen)
		for i := kPeriod - 2 + slowingPeriod; i < dataLen; i++ {
			var sum float64
			for j := 0; j < slowingPeriod; j++ {
				sum += k[i-j]
			}
			smoothedK[i] = sum / float64(slowingPeriod)
		}
		k = smoothedK
	}

	// 计算%D（%K的移动平均）
	d := make([]float64, dataLen)
	if len(k) >= dPeriod {
		for i := kPeriod - 2 + slowingPeriod + dPeriod - 1; i < dataLen; i++ {
			var sum float64
			for j := 0; j < dPeriod; j++ {
				sum += k[i-j]
			}
			d[i] = sum / float64(dPeriod)
		}
	}

	return &StochasticResult{
		K: k,
		D: d,
	}
}

// GetLatestStochastic 获取最新Stochastic值
func GetLatestStochastic(highs, lows, closes []float64, kPeriod, dPeriod, slowingPeriod int) (float64, float64) {
	result := Stochastic(highs, lows, closes, kPeriod, dPeriod, slowingPeriod)
	if result == nil || len(result.K) == 0 || len(result.D) == 0 {
		return 0, 0
	}

	latestK := result.K[len(result.K)-1]
	latestD := result.D[len(result.D)-1]

	return latestK, latestD
}

// IsStochasticOverbought 判断Stochastic是否超买（%K和%D都>80）
func IsStochasticOverbought(k, d float64) bool {
	return k > 80 && d > 80
}

// IsStochasticOversold 判断Stochastic是否超卖（%K和%D都<20）
func IsStochasticOversold(k, d float64) bool {
	return k < 20 && d < 20
}

// IsStochasticBullishCross 判断Stochastic金叉（%K从下向上穿越%D）
func IsStochasticBullishCross(k, d []float64) bool {
	if len(k) < 2 || len(d) < 2 {
		return false
	}
	n := len(k)
	return k[n-1] > d[n-1] && k[n-2] <= d[n-2]
}

// IsStochasticBearishCross 判断Stochastic死叉（%K从上向下穿越%D）
func IsStochasticBearishCross(k, d []float64) bool {
	if len(k) < 2 || len(d) < 2 {
		return false
	}
	n := len(k)
	return k[n-1] < d[n-1] && k[n-2] >= d[n-2]
}
