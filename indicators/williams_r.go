package indicators

// WilliamsR 计算威廉指标 (Williams %R)
// period: 计算周期，通常为14
// 返回值：Williams %R值数组
func WilliamsR(highs, lows, closes []float64, period int) []float64 {
	if len(highs) < period || len(lows) < period || len(closes) < period {
		return nil
	}

	dataLen := len(closes)
	williamsR := make([]float64, dataLen)

	// 计算Williams %R值
	for i := period - 1; i < dataLen; i++ {
		// 找出最近period的最高价和最低价
		highestHigh := highs[i]
		lowestLow := lows[i]

		for j := i - period + 1; j <= i; j++ {
			if highs[j] > highestHigh {
				highestHigh = highs[j]
			}
			if lows[j] < lowestLow {
				lowestLow = lows[j]
			}
		}

		// 计算Williams %R
		if highestHigh != lowestLow {
			williamsR[i] = ((highestHigh - closes[i]) / (highestHigh - lowestLow)) * -100
		} else {
			williamsR[i] = -50 // 如果没有变化，设为-50
		}
	}

	return williamsR
}

// GetLatestWilliamsR 获取最新Williams %R值
func GetLatestWilliamsR(highs, lows, closes []float64, period int) float64 {
	wr := WilliamsR(highs, lows, closes, period)
	if len(wr) == 0 {
		return 0
	}
	return wr[len(wr)-1]
}

// IsWilliamsROverbought 判断Williams %R是否超买（Williams %R > -20）
func IsWilliamsROverbought(wr float64) bool {
	return wr > -20
}

// IsWilliamsROversold 判断Williams %R是否超卖（Williams %R < -80）
func IsWilliamsROversold(wr float64) bool {
	return wr < -80
}

// IsWilliamsRBullishCross 判断Williams %R看涨交叉（从超卖区域向上穿越-80）
func IsWilliamsRBullishCross(wr []float64) bool {
	if len(wr) < 2 {
		return false
	}
	n := len(wr)
	return wr[n-1] > -80 && wr[n-2] <= -80
}

// IsWilliamsRBearishCross 判断Williams %R看跌交叉（从超买区域向下穿越-20）
func IsWilliamsRBearishCross(wr []float64) bool {
	if len(wr) < 2 {
		return false
	}
	n := len(wr)
	return wr[n-1] < -20 && wr[n-2] >= -20
}
