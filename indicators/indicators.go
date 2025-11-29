package indicators

// TechnicalAnalysis 技术分析结果汇总
type TechnicalAnalysis struct {
	RSI           float64  // RSI值
	MACD          float64  // MACD值
	MACDSignal    float64  // MACD信号线
	MACDHistogram float64  // MACD柱状图
	BBUpper       float64  // 布林带上轨
	BBLower       float64  // 布林带下轨
	BBMiddle      float64  // 布林带中轨
	BBWidth       float64  // 布林带宽度
	ATR           float64  // ATR值
	EMA20         float64  // 20周期EMA
	EMA50         float64  // 50周期EMA
	StochK        float64  // Stochastic %K值
	StochD        float64  // Stochastic %D值
	CCI           float64  // CCI值
	WilliamsR     float64  // Williams %R值
	ROC           float64  // ROC值
	CurrentPrice  float64  // 当前价格
	MarketEnv     string   // 市场环境
	Volatility    float64  // 波动率
	TrendStrength float64  // 趋势强度
	Signals       []string // 技术信号列表
}

// AnalyzeAll 综合技术分析
func AnalyzeAll(highs, lows, closes []float64, currentPrice float64) *TechnicalAnalysis {
	analysis := &TechnicalAnalysis{
		CurrentPrice: currentPrice,
		Signals:      make([]string, 0),
	}

	// RSI - 使用70条数据确保稳定性
	if rsi := GetLatestRSI(closes, 14); rsi > 0 {
		analysis.RSI = rsi
		if IsOverbought(rsi) {
			analysis.Signals = append(analysis.Signals, "RSI超买")
		} else if IsOversold(rsi) {
			analysis.Signals = append(analysis.Signals, "RSI超卖")
		}
	}

	// MACD
	if macd := GetLatestMACD(closes, 12, 26, 9); macd != nil {
		analysis.MACD = macd.MACDLine[len(macd.MACDLine)-1]
		analysis.MACDSignal = macd.SignalLine[len(macd.SignalLine)-1]
		analysis.MACDHistogram = macd.Histogram[len(macd.Histogram)-1]

		if IsBullishCross(macd.MACDLine, macd.SignalLine) {
			analysis.Signals = append(analysis.Signals, "MACD金叉")
		} else if IsBearishCross(macd.MACDLine, macd.SignalLine) {
			analysis.Signals = append(analysis.Signals, "MACD死叉")
		}
	}

	// 布林带
	if bb := GetLatestBollingerBands(closes, 20, 2); bb != nil {
		n := len(bb.UpperBand)
		if n > 0 {
			analysis.BBUpper = bb.UpperBand[n-1]
			analysis.BBLower = bb.LowerBand[n-1]
			analysis.BBMiddle = bb.MA[n-1]
			analysis.BBWidth = bb.Width[n-1]

			if IsPriceAboveUpperBand(currentPrice, analysis.BBUpper) {
				analysis.Signals = append(analysis.Signals, "价格突破布林上轨")
			} else if IsPriceBelowLowerBand(currentPrice, analysis.BBLower) {
				analysis.Signals = append(analysis.Signals, "价格跌破布林下轨")
			}
		}
	}

	// ATR
	if atr := GetLatestATR(highs, lows, closes, 14); atr > 0 {
		analysis.ATR = atr
		analysis.Volatility = CalculateVolatilityPercent(atr, currentPrice)
	}

	// EMA - 修复价格位置判断
	if ema20 := GetLatestEMA(closes, 20); ema20 > 0 {
		analysis.EMA20 = ema20
		// 增加容错范围，避免微小差异导致的错误判断
		if currentPrice > ema20+0.01 { // 价格必须显著高于EMA20
			analysis.Signals = append(analysis.Signals, "价格在EMA20上方")
		} else if currentPrice < ema20-0.01 { // 价格必须显著低于EMA20
			analysis.Signals = append(analysis.Signals, "价格在EMA20下方")
		} else {
			// 价格接近EMA20，使用中性描述
			analysis.Signals = append(analysis.Signals, "价格接近EMA20")
		}
	}

	if ema50 := GetLatestEMA(closes, 50); ema50 > 0 {
		analysis.EMA50 = ema50
	}

	// Stochastic %K/%D (14,3,3)
	if stochK, stochD := GetLatestStochastic(highs, lows, closes, 14, 3, 3); stochK > 0 && stochD > 0 {
		analysis.StochK = stochK
		analysis.StochD = stochD

		if IsStochasticOverbought(stochK, stochD) {
			analysis.Signals = append(analysis.Signals, "Stochastic超买")
		} else if IsStochasticOversold(stochK, stochD) {
			analysis.Signals = append(analysis.Signals, "Stochastic超卖")
		}

		stochResult := Stochastic(highs, lows, closes, 14, 3, 3)
		if stochResult != nil && IsStochasticBullishCross(stochResult.K, stochResult.D) {
			analysis.Signals = append(analysis.Signals, "Stochastic金叉")
		} else if stochResult != nil && IsStochasticBearishCross(stochResult.K, stochResult.D) {
			analysis.Signals = append(analysis.Signals, "Stochastic死叉")
		}
	}

	// CCI (20周期)
	if cci := GetLatestCCI(highs, lows, closes, 20); cci != 0 {
		analysis.CCI = cci

		if IsCCIOverbought(cci) {
			analysis.Signals = append(analysis.Signals, "CCI超买")
		} else if IsCCIOversold(cci) {
			analysis.Signals = append(analysis.Signals, "CCI超卖")
		}
	}

	// Williams %R (14周期)
	if wr := GetLatestWilliamsR(highs, lows, closes, 14); wr != 0 {
		analysis.WilliamsR = wr

		if IsWilliamsROverbought(wr) {
			analysis.Signals = append(analysis.Signals, "Williams %R超买")
		} else if IsWilliamsROversold(wr) {
			analysis.Signals = append(analysis.Signals, "Williams %R超卖")
		}

		wrValues := WilliamsR(highs, lows, closes, 14)
		if wrValues != nil && IsWilliamsRBullishCross(wrValues) {
			analysis.Signals = append(analysis.Signals, "Williams %R看涨交叉")
		} else if wrValues != nil && IsWilliamsRBearishCross(wrValues) {
			analysis.Signals = append(analysis.Signals, "Williams %R看跌交叉")
		}
	}

	// ROC (12周期)
	if roc := GetLatestROC(closes, 12); roc != 0 {
		analysis.ROC = roc

		if IsROCOverbought(roc, 10) {
			analysis.Signals = append(analysis.Signals, "ROC超买")
		} else if IsROCOversold(roc, -10) {
			analysis.Signals = append(analysis.Signals, "ROC超卖")
		}

		rocValues := ROC(closes, 12)
		if rocValues != nil && IsROCBullishCross(rocValues) {
			analysis.Signals = append(analysis.Signals, "ROC看涨交叉")
		} else if rocValues != nil && IsROCBearishCross(rocValues) {
			analysis.Signals = append(analysis.Signals, "ROC看跌交叉")
		}
	}

	// 市场环境
	cond := AnalyzeMarketEnvironment(closes, analysis.EMA20, analysis.EMA50, analysis.ATR)
	switch cond.Environment {
	case MarketEnvironmentBullish:
		analysis.MarketEnv = "上涨趋势"
	case MarketEnvironmentBearish:
		analysis.MarketEnv = "下跌趋势"
	case MarketEnvironmentSideways:
		analysis.MarketEnv = "震荡市场"
	default:
		analysis.MarketEnv = "未知"
	}
	analysis.TrendStrength = cond.TrendStrength

	return analysis
}
