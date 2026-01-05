package task

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	binance "deeptrade/binance"
	"deeptrade/indicators"
	"deeptrade/utils"
)

// PrepareTechnicalData 准备技术分析数据
func PrepareTechnicalData(marketData *MarketData) *TechnicalAnalysisData {
	data := &TechnicalAnalysisData{}

	// 解析3分钟K线数据（75条短期精准指标，避免长期数据均值失真）
	if len(marketData.Klines3m) > 0 {
		data.High3m = make([]float64, 0, len(marketData.Klines3m))
		data.Low3m = make([]float64, 0, len(marketData.Klines3m))
		data.Price3m = make([]float64, 0, len(marketData.Klines3m))

		for _, k := range marketData.Klines3m {
			if h, err := strconv.ParseFloat(k.High, 64); err == nil {
				data.High3m = append(data.High3m, h)
			}
			if l, err := strconv.ParseFloat(k.Low, 64); err == nil {
				data.Low3m = append(data.Low3m, l)
			}
			if c, err := strconv.ParseFloat(k.Close, 64); err == nil {
				data.Price3m = append(data.Price3m, c)
			}
		}

		if len(data.Price3m) > 0 && data.CurrentPrice == 0 {
			data.CurrentPrice = data.Price3m[len(data.Price3m)-1]
		}
		data.Has3mData = true
	}

	// 解析15分钟K线数据（100条用于趋势方向判断，覆盖约25小时）
	if len(marketData.Klines15m) > 0 {
		data.High15m = make([]float64, 0, len(marketData.Klines15m))
		data.Low15m = make([]float64, 0, len(marketData.Klines15m))
		data.Price15m = make([]float64, 0, len(marketData.Klines15m))

		for _, k := range marketData.Klines15m {
			if h, err := strconv.ParseFloat(k.High, 64); err == nil {
				data.High15m = append(data.High15m, h)
			}
			if l, err := strconv.ParseFloat(k.Low, 64); err == nil {
				data.Low15m = append(data.Low15m, l)
			}
			if c, err := strconv.ParseFloat(k.Close, 64); err == nil {
				data.Price15m = append(data.Price15m, c)
			}
		}
		data.Has15mData = true
	}

	// 如果K线数据都没有，使用ticker价格
	if data.CurrentPrice == 0 && marketData.Ticker != nil {
		if price, err := strconv.ParseFloat(marketData.Ticker.LastPrice, 64); err == nil {
			data.CurrentPrice = price
		}
	}

	return data
}

// getPositionIn24RangeDesc 获取24小时区间位置描述
func getPositionIn24RangeDesc(positionPercent float64) string {
	switch {
	case positionPercent < 20:
		return "下沿(偏弱)"
	case positionPercent < 40:
		return "下半部"
	case positionPercent < 60:
		return "中部"
	case positionPercent < 80:
		return "上半部"
	default:
		return "上沿(偏强)"
	}
}

// FormatTechnicalIndicators 格式化技术指标分析
func FormatTechnicalIndicators(technicalData *TechnicalAnalysisData, marketData *MarketData) string {
	var analysis strings.Builder
	analysis.WriteString("技术指标分析:\n")

	// 添加24小时统计数据
	if marketData != nil && marketData.Ticker != nil {
		ticker := marketData.Ticker
		analysis.WriteString("  [24小时统计数据]\n")

		// 当前价格和变化
		lastPrice := utils.ParseFloatSafe(ticker.LastPrice, 0)
		if lastPrice > 0 {
			analysis.WriteString(fmt.Sprintf("  当前价格:     %.2f\n", lastPrice))

			// 价格变化
			priceChange := utils.ParseFloatSafe(ticker.PriceChange, 0)
			if priceChange != 0 || ticker.PriceChange != "" {
				if priceChange >= 0 {
					analysis.WriteString(fmt.Sprintf("  24h变化:     +%.2f (+%s%%)\n", priceChange, ticker.PriceChangePercent))
				} else {
					analysis.WriteString(fmt.Sprintf("  24h变化:     %.2f (%s%%)\n", priceChange, ticker.PriceChangePercent))
				}
			}

			// 24小时价格区间
			highPrice := utils.ParseFloatSafe(ticker.HighPrice, 0)
			lowPrice := utils.ParseFloatSafe(ticker.LowPrice, 0)
			if highPrice > 0 && lowPrice > 0 {
				rangeWidth := highPrice - lowPrice

				analysis.WriteString(fmt.Sprintf("  24h区间:     %.2f - %.2f (区间宽度: %.2f)\n",
					lowPrice, highPrice, rangeWidth))

				if rangeWidth > 0 {
					positionInRange := ((lastPrice - lowPrice) / rangeWidth) * 100
					analysis.WriteString(fmt.Sprintf("  当前位置:     24h区间%s位置 (距离最低价%.1f%%)\n",
						getPositionIn24RangeDesc(positionInRange), positionInRange))
				} else {
					analysis.WriteString("  当前位置:     24h区间异常/无波动\n")
				}
			}

			// 24小时最高价
			if highPrice > 0 {
				analysis.WriteString(fmt.Sprintf("  24h最高价: %.2f\n", highPrice))
			}

			// 24小时最低价
			if lowPrice > 0 {
				analysis.WriteString(fmt.Sprintf("  24h最低价: %.2f\n", lowPrice))
			}

			// 24小时加权平均价
			weightedAvgPrice := utils.ParseFloatSafe(ticker.WeightedAvgPrice, 0)
			if weightedAvgPrice > 0 {
				analysis.WriteString(fmt.Sprintf("  24h加权均价: %.2f\n", weightedAvgPrice))
			}

			// 成交量数据
			volume := utils.ParseFloatSafe(ticker.Volume, 0)
			quoteVolume := utils.ParseFloatSafe(ticker.QuoteVolume, 0)
			if volume > 0 && quoteVolume > 0 {
				avgPrice := quoteVolume / volume
				analysis.WriteString(fmt.Sprintf("  24h成交量:   %.0f ETH, 成交额: %.0f USDT, 均价: %.2f\n",
					volume, quoteVolume, avgPrice))
			}

			// 交易次数
			if ticker.Count > 0 {
				analysis.WriteString(fmt.Sprintf("  24h交易次数: %d次\n", ticker.Count))
			}

			analysis.WriteString("\n")
		}
	}

	// ===== 15分钟K线趋势分析（优先显示，作为背景判断） =====
	if technicalData.Has15mData && len(technicalData.Price15m) > 0 {
		dataCount15m := len(technicalData.Price15m)
		hours15m := float64(dataCount15m) * 15.0 / 60.0 // 15分钟K线：100条 = 25小时
		curr15m := technicalData.Price15m[len(technicalData.Price15m)-1]

		analysis.WriteString("  [15分钟K线趋势分析] (主趋势方向)\n")
		analysis.WriteString(fmt.Sprintf("  基于%d条15分钟K线数据，覆盖时间范围:%.1f小时\n", dataCount15m, hours15m))

		// 计算EMA判断趋势
		ema10_15m := indicators.GetLatestEMA(technicalData.Price15m, 10)
		ema20_15m := indicators.GetLatestEMA(technicalData.Price15m, 20)
		ema30_15m := indicators.GetLatestEMA(technicalData.Price15m, 30)
		ema50_15m := indicators.GetLatestEMA(technicalData.Price15m, 50)

		if ema10_15m > 0 && ema20_15m > 0 && ema30_15m > 0 && ema50_15m > 0 {
			analysis.WriteString(fmt.Sprintf("  EMA队列:     EMA10=%.2f, EMA20=%.2f, EMA30=%.2f, EMA50=%.2f\n", ema10_15m, ema20_15m, ema30_15m, ema50_15m))

			// EMA排列判断趋势
			bullishAlignment := ema10_15m > ema20_15m && ema20_15m > ema30_15m && ema30_15m > ema50_15m
			bearishAlignment := ema10_15m < ema20_15m && ema20_15m < ema30_15m && ema30_15m < ema50_15m

			if bullishAlignment && curr15m > ema10_15m {
				analysis.WriteString("  趋势判断:    ★ 多头趋势 (EMA整齐排列向上，价格站稳均线上方)\n")
			} else if bearishAlignment && curr15m < ema10_15m {
				analysis.WriteString("  趋势判断:    ★ 空头趋势 (EMA整齐排列向下，价格站稳均线下方)\n")
			} else {
				analysis.WriteString("  趋势判断:    震荡行情 (EMA混乱交织，价格穿越均线)\n")
			}
		}

		// ATR波动率
		atr15m := indicators.GetLatestATR(technicalData.High15m, technicalData.Low15m, technicalData.Price15m, 14)
		if atr15m > 0 {
			volatility15m := indicators.CalculateVolatilityPercent(atr15m, curr15m)
			analysis.WriteString(fmt.Sprintf("  ATR波动率:   %.4f (%.1f%%)\n", atr15m, volatility15m))
		}

		analysis.WriteString("\n")
	}

	// ===== 3分钟K线块（75条数据优化格式；缺项跳过） =====
	if technicalData.Has3mData && len(technicalData.Price3m) > 0 {
		dataCount3m := len(technicalData.Price3m)
		hours3m := float64(dataCount3m) * 3.0 / 60.0 // 3分钟K线：75条 = 3.75小时
		curr3m := technicalData.Price3m[len(technicalData.Price3m)-1]

		// 计算3分钟技术指标（75条短期精准数据）
		ta3m := indicators.AnalyzeAll(technicalData.High3m, technicalData.Low3m, technicalData.Price3m, curr3m)

		// 数据基础
		analysis.WriteString("  [3分钟K线分析] (入场信号)\n")
		analysis.WriteString(fmt.Sprintf("  基于%d条3分钟K线数据，覆盖时间范围:%.1f小时\n", dataCount3m, hours3m))

		// RSI
		if ta3m.RSI > 0 {
			analysis.WriteString(fmt.Sprintf("  RSI(14):     %.2f (超卖<30, 超买>70)\n", ta3m.RSI))
		}

		// MACD三值 + MACD状态
		if ta3m.MACD != 0 || ta3m.MACDSignal != 0 || ta3m.MACDHistogram != 0 {
			analysis.WriteString(fmt.Sprintf("  MACD:        线=%.4f, 信号=%.4f, 柱=%.4f\n", ta3m.MACD, ta3m.MACDSignal, ta3m.MACDHistogram))
		}
		var macdState3m string
		// 使用ta3m中的MACD值，避免重复计算
		if macd3m := indicators.GetLatestMACD(technicalData.Price3m, 12, 26, 9); macd3m != nil {
			if indicators.IsBullishCross(macd3m.MACDLine, macd3m.SignalLine) {
				macdState3m = "金叉"
			} else if indicators.IsBearishCross(macd3m.MACDLine, macd3m.SignalLine) {
				macdState3m = "死叉"
			} else if len(macd3m.MACDLine) > 0 && len(macd3m.SignalLine) > 0 {
				lm := macd3m.MACDLine[len(macd3m.MACDLine)-1]
				ls := macd3m.SignalLine[len(macd3m.SignalLine)-1]
				if lm > ls {
					macdState3m = "多头排列"
				} else {
					macdState3m = "空头排列"
				}
			}
		}
		if macdState3m != "" {
			analysis.WriteString(fmt.Sprintf("  MACD状态:     %s\n", macdState3m))
		}

		// 布林带 + 位置
		if ta3m.BBUpper > 0 && ta3m.BBMiddle > 0 && ta3m.BBLower > 0 {
			analysis.WriteString(fmt.Sprintf("  布林带:       上轨=%.2f, 中轨=%.2f, 下轨=%.2f\n", ta3m.BBUpper, ta3m.BBMiddle, ta3m.BBLower))
			bandWidth3m := ta3m.BBUpper - ta3m.BBLower
			if bandWidth3m > 0 {
				upDist := (ta3m.BBUpper - curr3m) / bandWidth3m
				lowDist := (curr3m - ta3m.BBLower) / bandWidth3m
				switch {
				case upDist < 0:
					analysis.WriteString("  布林带位置:   突破上轨 (强烈买入)\n")
				case lowDist < 0:
					analysis.WriteString("  布林带位置:   跌破下轨 (强烈卖出)\n")
				case upDist < 0.02:
					analysis.WriteString("  布林带位置:   接近上轨\n")
				case lowDist < 0.02:
					analysis.WriteString("  布林带位置:   接近下轨\n")
				default:
					analysis.WriteString("  布林带位置:   在中轨附近\n")
				}
			}
		}

		// ATR + 波动率
		if ta3m.ATR > 0 {
			analysis.WriteString(fmt.Sprintf("  ATR:          %.4f (当前波动率: %.1f%%)\n", ta3m.ATR, ta3m.Volatility))
		}

		// EMA/SMA 队列（缺项则整行略过）
		ema10_3m := indicators.GetLatestEMA(technicalData.Price3m, 10)
		ema30_3m := indicators.GetLatestEMA(technicalData.Price3m, 30)
		sma10_3m := indicators.GetLatestSMA(technicalData.Price3m, 10)
		sma30_3m := indicators.GetLatestSMA(technicalData.Price3m, 30)
		sma60_3m := indicators.GetLatestSMA(technicalData.Price3m, 60)

		if ema10_3m > 0 && ta3m.EMA20 > 0 && ema30_3m > 0 && ta3m.EMA50 > 0 {
			analysis.WriteString(fmt.Sprintf("  EMA队列:     EMA10=%.2f, EMA20=%.2f, EMA30=%.2f, EMA50=%.2f\n", ema10_3m, ta3m.EMA20, ema30_3m, ta3m.EMA50))
		}
		if sma10_3m > 0 && sma30_3m > 0 && sma60_3m > 0 {
			analysis.WriteString(fmt.Sprintf("  SMA队列:     SMA10=%.2f, SMA30=%.2f, SMA60=%.2f\n", sma10_3m, sma30_3m, sma60_3m))
		}

		// EMA排列
		if ema10_3m > 0 && ta3m.EMA20 > 0 && ema30_3m > 0 && ta3m.EMA50 > 0 {
			if ema10_3m > ta3m.EMA20 && ta3m.EMA20 > ema30_3m && ema30_3m > ta3m.EMA50 {
				analysis.WriteString("  EMA排列:      强势多头排列 (10>20>30>50)\n")
			} else if ema10_3m < ta3m.EMA20 && ta3m.EMA20 < ema30_3m && ema30_3m < ta3m.EMA50 {
				analysis.WriteString("  EMA排列:      强势空头排列 (10<20<30<50)\n")
			} else {
				analysis.WriteString("  EMA排列:      混乱无序 (震荡行情)\n")
			}
		}

		// 价格位置（相对短期均线）- 修复逻辑一致性
		if ema10_3m > 0 && sma10_3m > 0 {
			// 使用统一的容差判断，避免微小差异导致的矛盾
			tolerance := 0.01 // 容差范围
			aboveEMA10 := curr3m > ema10_3m+tolerance
			aboveSMA10 := curr3m > sma10_3m+tolerance

			if aboveEMA10 && aboveSMA10 {
				analysis.WriteString("  价格位置:     站上所有短期均线\n")
			} else if aboveEMA10 && !aboveSMA10 {
				analysis.WriteString("  价格位置:     在EMA10上方，SMA10下方\n")
			} else if !aboveEMA10 && aboveSMA10 {
				analysis.WriteString("  价格位置:     在SMA10上方，EMA10下方\n")
			} else {
				analysis.WriteString("  价格位置:     跌破所有短期均线\n")
			}
		}

		// Stochastic %K/%D (14,3,3)
		if ta3m.StochK > 0 && ta3m.StochD > 0 {
			analysis.WriteString(fmt.Sprintf("  Stochastic:   %%K=%.2f, %%D=%.2f (超卖<20, 超买>80)\n", ta3m.StochK, ta3m.StochD))
		}

		// CCI (20周期)
		if ta3m.CCI != 0 {
			analysis.WriteString(fmt.Sprintf("  CCI(20):      %.2f (超卖<-100, 超买>100)\n", ta3m.CCI))
		}

		// Williams %R (14周期)
		if ta3m.WilliamsR != 0 {
			analysis.WriteString(fmt.Sprintf("  Williams %%R:  %.2f (超卖<-80, 超买>-20)\n", ta3m.WilliamsR))
		}

		// ROC (12周期)
		if ta3m.ROC != 0 {
			analysis.WriteString(fmt.Sprintf("  ROC(12):      %.2f%% (负值看跌, 正值看涨)\n", ta3m.ROC))
		}

		// 市场环境 + 技术信号
		if ta3m.MarketEnv != "" {
			analysis.WriteString(fmt.Sprintf("  市场环境:     %s (趋势强度: %.2f/10)\n", ta3m.MarketEnv, ta3m.TrendStrength*10))
		}
		if len(ta3m.Signals) > 0 {
			analysis.WriteString(fmt.Sprintf("  技术信号:     %s\n", strings.Join(ta3m.Signals, ", ")))
		}
	}

	return analysis.String()
}

// CalculateVolatilityForPosition 计算用于仓位调整的波动率
func CalculateVolatilityForPosition(technicalData *TechnicalAnalysisData) float64 {
	// 获取ATR计算波动率
	if len(technicalData.High3m) > 0 && len(technicalData.Low3m) > 0 && len(technicalData.Price3m) > 0 {
		atr := indicators.GetLatestATR(technicalData.High3m, technicalData.Low3m, technicalData.Price3m, 14)
		return indicators.CalculateVolatilityPercent(atr, technicalData.CurrentPrice)
	}
	return 0
}

// FormatVolumeAnalysis 格式化成交量分析
func FormatVolumeAnalysis(klines []binance.Kline, currentPrice float64) string {
	if len(klines) == 0 {
		return "成交量分析: 无数据"
	}

	// 使用分层时间成交量分析
	config := indicators.DefaultVolumeAnalysisConfig()
	analysis := indicators.AnalyzeVolumeLayers(klines, config)

	return indicators.FormatVolumeAnalysisForLLM(analysis)
}

// extractRecentVolumes 提取成交量数据
func extractRecentVolumes(klines []binance.Kline, count int) []float64 {
	if len(klines) < count {
		count = len(klines)
	}

	volumes := make([]float64, count)
	start := len(klines) - count

	for i := 0; i < count; i++ {
		if vol, err := strconv.ParseFloat(klines[start+i].Volume, 64); err == nil {
			volumes[i] = vol
		}
	}

	return volumes
}

// calculateVolumeTrend 计算成交量趋势
func calculateVolumeTrend(recent, earlier []float64) string {
	if len(recent) == 0 || len(earlier) == 0 {
		return "数据不足"
	}

	recentAvg := average(recent)
	earlierAvg := average(earlier)

	if earlierAvg == 0 {
		return "无法计算"
	}

	ratio := recentAvg / earlierAvg

	switch {
	case ratio > 1.2:
		return "显著上升"
	case ratio > 1.05:
		return "温和上升"
	case ratio < 0.8:
		return "显著下降"
	case ratio < 0.95:
		return "温和下降"
	default:
		return "相对稳定"
	}
}

// calculateBuyRatio 计算主动买入比例
func calculateBuyRatio(klines []binance.Kline) float64 {
	if len(klines) == 0 {
		return 50.0
	}

	var totalVolume, buyVolume float64

	for _, kline := range klines {
		if vol, err := strconv.ParseFloat(kline.Volume, 64); err == nil {
			totalVolume += vol
		}
		if buyVol, err := strconv.ParseFloat(kline.TakerBuyBaseAssetVolume, 64); err == nil {
			buyVolume += buyVol
		}
	}

	if totalVolume > 0 {
		return (buyVolume / totalVolume) * 100
	}
	return 50.0
}

// analyzePriceVolumeRelationship 分析量价关系
func analyzePriceVolumeRelationship(klines []binance.Kline) string {
	if len(klines) < 10 {
		return "数据不足"
	}

	var priceChanges []float64
	var volumeChanges []float64

	for i := 1; i < len(klines); i++ {
		// 计算价格变化
		prevPrice, _ := strconv.ParseFloat(klines[i-1].Close, 64)
		currPrice, _ := strconv.ParseFloat(klines[i].Close, 64)
		if prevPrice > 0 {
			priceChanges = append(priceChanges, (currPrice-prevPrice)/prevPrice)
		}

		// 计算成交量变化
		prevVol, _ := strconv.ParseFloat(klines[i-1].Volume, 64)
		currVol, _ := strconv.ParseFloat(klines[i].Volume, 64)
		if prevVol > 0 {
			volumeChanges = append(volumeChanges, (currVol-prevVol)/prevVol)
		}
	}

	if len(priceChanges) != len(volumeChanges) || len(priceChanges) == 0 {
		return "数据不足"
	}

	correlation := calculateCorrelation(priceChanges, volumeChanges)

	switch {
	case correlation > 0.3:
		return "量价同步上涨 (健康上涨)"
	case correlation < -0.3:
		return "量价背离 (警惕反转)"
	default:
		return "量价关系不明显"
	}
}

// calculateCorrelation 计算相关性
func calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}

	n := float64(len(x))
	var sumX, sumY, sumXY, sumX2, sumY2 float64

	for i := 0; i < len(x); i++ {
		sumX += x[i]
		sumY += y[i]
		sumXY += x[i] * y[i]
		sumX2 += x[i] * x[i]
		sumY2 += y[i] * y[i]
	}

	numerator := n*sumXY - sumX*sumY
	denominator := math.Sqrt((n*sumX2 - sumX*sumX) * (n*sumY2 - sumY*sumY))

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

// average 计算平均值
func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// detectUnusualVolume 检测异常成交量
func detectUnusualVolume(klines []binance.Kline) string {
	if len(klines) < 20 {
		return ""
	}

	// 获取最新成交量
	currentVol, _ := strconv.ParseFloat(klines[len(klines)-1].Volume, 64)

	// 计算前20期平均成交量
	var sumVol float64
	for i := len(klines) - 21; i < len(klines)-1; i++ {
		if vol, err := strconv.ParseFloat(klines[i].Volume, 64); err == nil {
			sumVol += vol
		}
	}
	avgVol := sumVol / 20

	if avgVol == 0 {
		return ""
	}

	ratio := currentVol / avgVol

	if ratio > 3.0 {
		return fmt.Sprintf("异常放量(%.1f倍)", ratio)
	} else if ratio > 2.0 {
		return fmt.Sprintf("明显放量(%.1f倍)", ratio)
	} else if ratio < 0.3 {
		return fmt.Sprintf("异常缩量(%.1f倍)", ratio)
	} else if ratio < 0.5 {
		return fmt.Sprintf("明显缩量(%.1f倍)", ratio)
	}

	return ""
}

// AnalyzeTrendBy15m 基于15分钟K线分析趋势方向
func AnalyzeTrendBy15m(technicalData *TechnicalAnalysisData) *TrendAnalysis {
	result := &TrendAnalysis{
		Direction:   "SIDEWAYS",
		Strength:    5.0,
		IsTrending:  false,
		Description: "数据不足",
	}

	if !technicalData.Has15mData || len(technicalData.Price15m) < 50 {
		return result
	}

	prices := technicalData.Price15m
	highs := technicalData.High15m
	lows := technicalData.Low15m
	currentPrice := technicalData.CurrentPrice

	// 计算15分钟EMA
	ema10 := indicators.GetLatestEMA(prices, 10)
	ema20 := indicators.GetLatestEMA(prices, 20)
	ema30 := indicators.GetLatestEMA(prices, 30)
	ema50 := indicators.GetLatestEMA(prices, 50)

	if ema10 == 0 || ema20 == 0 || ema30 == 0 || ema50 == 0 {
		return result
	}

	// 计算ATR评估波动性
	atr := indicators.GetLatestATR(highs, lows, prices, 14)
	volatility := indicators.CalculateVolatilityPercent(atr, currentPrice)

	// 判断EMA排列
	bullishAlignment := ema10 > ema20 && ema20 > ema30 && ema30 > ema50
	bearishAlignment := ema10 < ema20 && ema20 < ema30 && ema30 < ema50

	// 计算EMA斜率（最近5根K线的变化趋势）
	ema10Series := indicators.EMA(prices, 10)
	ema20Series := indicators.EMA(prices, 20)
	var ema10Slope, ema20Slope float64
	if len(ema10Series) >= 6 && len(ema20Series) >= 6 {
		ema10Slope = (ema10Series[len(ema10Series)-1] - ema10Series[len(ema10Series)-6]) / 5.0
		ema20Slope = (ema20Series[len(ema20Series)-1] - ema20Series[len(ema20Series)-6]) / 5.0
	}

	// 价格相对于EMA的位置
	aboveAllEMA := currentPrice > ema10 && currentPrice > ema20 && currentPrice > ema30 && currentPrice > ema50
	belowAllEMA := currentPrice < ema10 && currentPrice < ema20 && currentPrice < ema30 && currentPrice < ema50

	// 综合判断趋势
	if bullishAlignment && aboveAllEMA {
		result.Direction = "UP"
		result.IsTrending = true

		// 计算趋势强度
		strength := 5.0
		if ema10Slope > 0 {
			strength += 1.5
		}
		if ema20Slope > 0 {
			strength += 1.0
		}
		if volatility > 1.5 {
			strength += 1.0 // 波动扩张
		}
		if volatility < 0.8 {
			strength -= 1.0 // 波动收缩
		}

		result.Strength = math.Min(10.0, math.Max(0.0, strength))
		result.Description = fmt.Sprintf("多头趋势 EMA排列整齐(%.2f>%.2f>%.2f>%.2f) 价格站稳均线上方 ATR波动率%.1f%%",
			ema10, ema20, ema30, ema50, volatility)
	} else if bearishAlignment && belowAllEMA {
		result.Direction = "DOWN"
		result.IsTrending = true

		// 计算趋势强度
		strength := 5.0
		if ema10Slope < 0 {
			strength += 1.5
		}
		if ema20Slope < 0 {
			strength += 1.0
		}
		if volatility > 1.5 {
			strength += 1.0
		}
		if volatility < 0.8 {
			strength -= 1.0
		}

		result.Strength = math.Min(10.0, math.Max(0.0, strength))
		result.Description = fmt.Sprintf("空头趋势 EMA排列整齐(%.2f<%.2f<%.2f<%.2f) 价格站稳均线下方 ATR波动率%.1f%%",
			ema10, ema20, ema30, ema50, volatility)
	} else {
		result.Direction = "SIDEWAYS"
		result.IsTrending = false

		// 判断震荡强度
		strength := 5.0
		emaSpread := math.Abs(ema10-ema50) / ema30 * 100 // EMA离散程度
		if emaSpread < 0.5 {
			strength -= 2.0 // 非常拥挤
		} else if emaSpread > 2.0 {
			strength += 1.0 // 有一定波动
		}

		result.Strength = math.Min(10.0, math.Max(0.0, strength))
		result.Description = fmt.Sprintf("震荡行情 EMA混乱交织 价格频繁穿越均线 EMA离散度%.2f%% ATR波动率%.1f%%",
			emaSpread, volatility)
	}

	return result
}

// IdentifySupportResistance 识别支撑阻力位
func IdentifySupportResistance(technicalData *TechnicalAnalysisData) []SupportResistanceLevel {
	if !technicalData.Has15mData || len(technicalData.Price15m) < 50 {
		return nil
	}

	var levels []SupportResistanceLevel
	currentPrice := technicalData.CurrentPrice
	highs := technicalData.High15m
	lows := technicalData.Low15m

	// 识别局部高低点（Fractal方法）
	var swingHighs, swingLows []float64

	for i := 2; i < len(highs)-2; i++ {
		// 局部高点：比左右各2根K线的最高价都高
		if highs[i] > highs[i-1] && highs[i] > highs[i-2] &&
			highs[i] > highs[i+1] && highs[i] > highs[i+2] {
			swingHighs = append(swingHighs, highs[i])
		}
		// 局部低点：比左右各2根K线的最低价都低
		if lows[i] < lows[i-1] && lows[i] < lows[i-2] &&
			lows[i] < lows[i+1] && lows[i] < lows[i+2] {
			swingLows = append(swingLows, lows[i])
		}
	}

	// 聚类相近的价格点（容差0.5%）
	tolerance := 0.005

	// 聚类阻力位
	resistanceClusters := clusterPriceLevels(swingHighs, tolerance)
	for _, cluster := range resistanceClusters {
		avgPrice := average(cluster)
		distance := (avgPrice - currentPrice) / currentPrice * 100

		// 只保留距离当前价格±3%以内的价位
		if distance > 0 && distance < 3.0 {
			strength := calculateLevelStrength(cluster, len(highs))
			levels = append(levels, SupportResistanceLevel{
				Price:      avgPrice,
				Type:       "RESISTANCE",
				Strength:   strength,
				TouchCount: len(cluster),
				Distance:   distance,
			})
		}
	}

	// 聚类支撑位
	supportClusters := clusterPriceLevels(swingLows, tolerance)
	for _, cluster := range supportClusters {
		avgPrice := average(cluster)
		distance := (currentPrice - avgPrice) / currentPrice * 100

		// 只保留距离当前价格±3%以内的价位
		if distance > 0 && distance < 3.0 {
			strength := calculateLevelStrength(cluster, len(lows))
			levels = append(levels, SupportResistanceLevel{
				Price:      avgPrice,
				Type:       "SUPPORT",
				Strength:   strength,
				TouchCount: len(cluster),
				Distance:   distance,
			})
		}
	}

	// 按距离排序（距离近的在前）
	sortLevelsByDistance(levels)

	// 最多返回5个最相关的价位
	if len(levels) > 5 {
		levels = levels[:5]
	}

	return levels
}

// clusterPriceLevels 聚类相近的价格点
func clusterPriceLevels(prices []float64, tolerance float64) [][]float64 {
	if len(prices) == 0 {
		return nil
	}

	var clusters [][]float64
	used := make([]bool, len(prices))

	for i := 0; i < len(prices); i++ {
		if used[i] {
			continue
		}

		cluster := []float64{prices[i]}
		used[i] = true

		for j := i + 1; j < len(prices); j++ {
			if used[j] {
				continue
			}
			// 检查是否在容差范围内
			if math.Abs(prices[j]-prices[i])/prices[i] < tolerance {
				cluster = append(cluster, prices[j])
				used[j] = true
			}
		}

		if len(cluster) >= 2 { // 至少触及2次才有效
			clusters = append(clusters, cluster)
		}
	}

	return clusters
}

// calculateLevelStrength 计算价格位强度
func calculateLevelStrength(cluster []float64, totalPeriods int) int {
	// 基于触及次数和总周期数计算强度
	touchCount := len(cluster)

	switch {
	case touchCount >= 5:
		return 5 // 非常强
	case touchCount >= 4:
		return 4 // 强
	case touchCount >= 3:
		return 3 // 中等
	case touchCount >= 2:
		return 2 // 较弱
	default:
		return 1 // 弱
	}
}

// sortLevelsByDistance 按距离排序支撑阻力位
func sortLevelsByDistance(levels []SupportResistanceLevel) {
	for i := 0; i < len(levels)-1; i++ {
		for j := i + 1; j < len(levels); j++ {
			if levels[i].Distance > levels[j].Distance {
				levels[i], levels[j] = levels[j], levels[i]
			}
		}
	}
}

// FormatTrendAnalysis 格式化趋势分析结果
func FormatTrendAnalysis(trend *TrendAnalysis) string {
	var sb strings.Builder

	sb.WriteString("【多周期趋势分析】\n")
	sb.WriteString(fmt.Sprintf("  趋势方向: %s\n", trend.Direction))
	sb.WriteString(fmt.Sprintf("  趋势强度: %.1f/10\n", trend.Strength))
	sb.WriteString(fmt.Sprintf("  市场状态: %s\n", map[bool]string{true: "趋势市", false: "震荡市"}[trend.IsTrending]))
	sb.WriteString(fmt.Sprintf("  分析详情: %s\n", trend.Description))

	return sb.String()
}

// FormatSupportResistance 格式化支撑阻力位
func FormatSupportResistance(levels []SupportResistanceLevel, currentPrice float64) string {
	if len(levels) == 0 {
		return "【支撑阻力位】暂无有效价位"
	}

	var sb strings.Builder
	sb.WriteString("【支撑阻力位分析】\n")

	// 分别处理支撑和阻力
	var supports, resistances []SupportResistanceLevel
	for _, level := range levels {
		if level.Type == "SUPPORT" {
			supports = append(supports, level)
		} else {
			resistances = append(resistances, level)
		}
	}

	// 输出阻力位
	if len(resistances) > 0 {
		sb.WriteString("  阻力位:\n")
		for i, r := range resistances {
			strengthStr := strings.Repeat("★", r.Strength)
			sb.WriteString(fmt.Sprintf("    R%d: %.2f (强度:%s 触及%d次 距离+%.2f%%)\n",
				i+1, r.Price, strengthStr, r.TouchCount, r.Distance))
		}
	}

	// 输出支撑位
	if len(supports) > 0 {
		sb.WriteString("  支撑位:\n")
		for i, s := range supports {
			strengthStr := strings.Repeat("★", s.Strength)
			sb.WriteString(fmt.Sprintf("    S%d: %.2f (强度:%s 触及%d次 距离-%.2f%%)\n",
				i+1, s.Price, strengthStr, s.TouchCount, s.Distance))
		}
	}

	// 交易建议
	sb.WriteString("  交易建议: ")
	if len(supports) > 0 && len(resistances) > 0 {
		nearestSupport := supports[0]
		nearestResistance := resistances[0]
		space := nearestResistance.Price - nearestSupport.Price
		spacePercent := space / currentPrice * 100

		if spacePercent < 1.0 {
			sb.WriteString("支撑阻力位空间过小(<1%)，建议观望\n")
		} else {
			sb.WriteString(fmt.Sprintf("支撑阻力空间%.2f%%，可在支撑位附近做多，阻力位附近做空\n", spacePercent))
		}
	} else if len(supports) > 0 {
		sb.WriteString(fmt.Sprintf("价格接近支撑位%.2f，可考虑做多\n", supports[0].Price))
	} else if len(resistances) > 0 {
		sb.WriteString(fmt.Sprintf("价格接近阻力位%.2f，可考虑做空\n", resistances[0].Price))
	}

	return sb.String()
}
