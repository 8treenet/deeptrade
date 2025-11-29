package indicators

// MACDResult MACD计算结果
type MACDResult struct {
	MACDLine   []float64 // MACD线
	SignalLine []float64 // 信号线
	Histogram  []float64 // 柱状图
}

// MACD 计算指数平滑异同移动平均线
// fastPeriod: 快线周期，通常为12
// slowPeriod: 慢线周期，通常为26
// signalPeriod: 信号线周期，通常为9
func MACD(closes []float64, fastPeriod, slowPeriod, signalPeriod int) *MACDResult {
	if len(closes) < slowPeriod+signalPeriod {
		return nil
	}

	// 计算EMA
	emaFast := EMA(closes, fastPeriod)
	emaSlow := EMA(closes, slowPeriod)

	// 计算MACD线
	macdLine := make([]float64, len(closes))
	for i := range closes {
		if i < slowPeriod-1 {
			macdLine[i] = 0
		} else {
			macdLine[i] = emaFast[i] - emaSlow[i]
		}
	}

	// 计算信号线（MACD线的EMA），对齐至有效MACD起点以减少暖启动偏差
	signalLine := make([]float64, len(closes))
	startIdx := slowPeriod - 1
	if len(macdLine)-startIdx >= signalPeriod {
		emaPart := EMA(macdLine[startIdx:], signalPeriod)
		for i := 0; i < len(emaPart); i++ {
			signalLine[startIdx+i] = emaPart[i]
		}
	}

	// 计算柱状图
	histogram := make([]float64, len(closes))
	for i := range macdLine {
		if i < slowPeriod+signalPeriod-2 {
			histogram[i] = 0
		} else {
			histogram[i] = macdLine[i] - signalLine[i]
		}
	}

	return &MACDResult{
		MACDLine:   macdLine,
		SignalLine: signalLine,
		Histogram:  histogram,
	}
}

// GetLatestMACD 获取最新MACD值
func GetLatestMACD(closes []float64, fastPeriod, slowPeriod, signalPeriod int) *MACDResult {
	return MACD(closes, fastPeriod, slowPeriod, signalPeriod)
}

// IsBullishCross 判断金叉（MACD线从下向上穿越信号线）
func IsBullishCross(macdLine, signalLine []float64) bool {
	if len(macdLine) < 2 || len(signalLine) < 2 {
		return false
	}
	n := len(macdLine)
	return macdLine[n-1] > signalLine[n-1] && macdLine[n-2] <= signalLine[n-2]
}

// IsBearishCross 判断死叉（MACD线从上向下穿越信号线）
func IsBearishCross(macdLine, signalLine []float64) bool {
	if len(macdLine) < 2 || len(signalLine) < 2 {
		return false
	}
	n := len(macdLine)
	return macdLine[n-1] < signalLine[n-1] && macdLine[n-2] >= signalLine[n-2]
}

// EMA 计算指数移动平均线
// period: 周期
func EMA(closes []float64, period int) []float64 {
	if len(closes) < period {
		return nil
	}

	ema := make([]float64, len(closes))
	multiplier := 2.0 / (float64(period) + 1.0)

	// 第一个值使用SMA
	var sum float64
	for i := 0; i < period; i++ {
		sum += closes[i]
	}
	ema[period-1] = sum / float64(period)

	// 后续值使用EMA公式
	for i := period; i < len(closes); i++ {
		ema[i] = (closes[i] * multiplier) + (ema[i-1] * (1 - multiplier))
	}

	return ema
}
