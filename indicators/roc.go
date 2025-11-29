package indicators

// ROC 计算价格变化率 (Rate of Change)
// period: 计算周期，通常为12或20
// 返回值：ROC值数组（百分比形式）
func ROC(closes []float64, period int) []float64 {
	if len(closes) < period+1 {
		return nil
	}

	dataLen := len(closes)
	roc := make([]float64, dataLen)

	// 前period个值设为0
	for i := 0; i < period; i++ {
		roc[i] = 0
	}

	// 计算ROC值
	for i := period; i < dataLen; i++ {
		if closes[i-period] != 0 {
			roc[i] = ((closes[i] - closes[i-period]) / closes[i-period]) * 100
		} else {
			roc[i] = 0
		}
	}

	return roc
}

// GetLatestROC 获取最新ROC值
func GetLatestROC(closes []float64, period int) float64 {
	roc := ROC(closes, period)
	if len(roc) == 0 {
		return 0
	}
	return roc[len(roc)-1]
}

// IsROCBullishCross 判断ROC看涨交叉（从负值穿越到正值）
func IsROCBullishCross(roc []float64) bool {
	if len(roc) < 2 {
		return false
	}
	n := len(roc)
	return roc[n-1] > 0 && roc[n-2] <= 0
}

// IsROCBearishCross 判断ROC看跌交叉（从正值穿越到负值）
func IsROCBearishCross(roc []float64) bool {
	if len(roc) < 2 {
		return false
	}
	n := len(roc)
	return roc[n-1] < 0 && roc[n-2] >= 0
}

// IsROCOverbought 判断ROC是否超买（ROC值超过某个阈值，如10%）
func IsROCOverbought(roc float64, threshold float64) bool {
	return roc > threshold
}

// IsROCOversold 判断ROC是否超卖（ROC值低于某个阈值，如-10%）
func IsROCOversold(roc float64, threshold float64) bool {
	return roc < threshold
}
