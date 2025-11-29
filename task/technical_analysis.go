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

	// —— 3分钟K线块（75条数据优化格式；缺项跳过）——
	if technicalData.Has3mData && len(technicalData.Price3m) > 0 {
		dataCount3m := len(technicalData.Price3m)
		hours3m := float64(dataCount3m) * 3.0 / 60.0 // 3分钟K线：75条 = 3.75小时
		curr3m := technicalData.Price3m[len(technicalData.Price3m)-1]

		// 计算3分钟技术指标（75条短期精准数据）
		ta3m := indicators.AnalyzeAll(technicalData.High3m, technicalData.Low3m, technicalData.Price3m, curr3m)

		// 数据基础
		analysis.WriteString("  [3分钟K线分析]\n")
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
