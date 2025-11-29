package indicators

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	binance "deeptrade/binance"
	tradeflow "deeptrade/task/trade_flow"
)

// TradeFlowAnalysis 交易流分析结果
type TradeFlowAnalysis struct {
	// 基础统计
	TotalTrades   int     `json:"total_trades"`
	TotalVolume   float64 `json:"total_volume"`
	TotalValue    float64 `json:"total_value"`
	AvgTradeValue float64 `json:"avg_trade_value"`
	AvgPrice      float64 `json:"avg_price"` // 总体成交均价(USDT/ETH)
	MaxTradeValue float64 `json:"max_trade_value"`

	// 买卖方向统计
	BuyVolume     float64 `json:"buy_volume"`
	SellVolume    float64 `json:"sell_volume"`
	BuyRatio      float64 `json:"buy_ratio"`
	SellRatio     float64 `json:"sell_ratio"`
	NetFlowVolume float64 `json:"net_flow_volume"`
	NetFlowValue  float64 `json:"net_flow_value"`

	// 分层交易统计（基于交易规模）
	SmallTrades  TradeSizeStats `json:"small_trades"`  // < $20k
	MediumTrades TradeSizeStats `json:"medium_trades"` // $20k - $70k
	LargeTrades  TradeSizeStats `json:"large_trades"`  // $70k - $175k
	WhaleTrades  TradeSizeStats `json:"whale_trades"`  // > $175k

	// 大单方向分析（关键新增指标）
	LargeBuyRatio  float64 `json:"large_buy_ratio"`  // 大单买入占比
	LargeSellRatio float64 `json:"large_sell_ratio"` // 大单卖出占比
	WhaleBuyRatio  float64 `json:"whale_buy_ratio"`  // 巨鲸买入占比
	WhaleSellRatio float64 `json:"whale_sell_ratio"` // 巨鲸卖出占比

	// 机构行为指标
	AggressiveRatio   float64 `json:"aggressive_ratio"`   // 主动交易占比
	InstitutionalFlow float64 `json:"institutional_flow"` // 机构资金流向（-1到+1）

	// 时间序列指标
	TradeIntensity float64 `json:"trade_intensity"` // 交易强度（笔/秒）
	AvgInterval    float64 `json:"avg_interval"`    // 平均交易间隔（毫秒）
	TradeFrequency string  `json:"trade_frequency"` // 交易频率描述

	// 价格冲击分析
	PriceImpact         float64 `json:"price_impact"`          // 价格冲击百分比
	VolumeWeightedPrice float64 `json:"volume_weighted_price"` // 成交量加权平均价

	// 市场情绪综合评分
	SentimentScore    float64 `json:"sentiment_score"`    // -10到+10
	MomentumIndicator string  `json:"momentum_indicator"` // 动量指标描述

	// 时间分层分析（新增）
	Recent5Min  *TimeLayerAnalysis `json:"recent_5min"`  // 最近5分钟分析
	Recent20Min *TimeLayerAnalysis `json:"recent_20min"` // 最近20分钟分析
}

// TradeSizeStats 分层交易统计
type TradeSizeStats struct {
	Count      int     `json:"count"`
	Volume     float64 `json:"volume"`
	Value      float64 `json:"value"`
	AvgPrice   float64 `json:"avg_price"` // 成交均价(USDT/ETH)
	BuyVolume  float64 `json:"buy_volume"`
	SellVolume float64 `json:"sell_volume"`
	BuyRatio   float64 `json:"buy_ratio"`
}

// TimeLayerAnalysis 时间分层分析结果
type TimeLayerAnalysis struct {
	// 基础统计
	TotalTrades   int     `json:"total_trades"`
	TotalVolume   float64 `json:"total_volume"`
	TotalValue    float64 `json:"total_value"`
	AvgTradeValue float64 `json:"avg_trade_value"`
	AvgPrice      float64 `json:"avg_price"` // 成交均价(USDT/ETH)

	// 买卖方向统计
	BuyVolume     float64 `json:"buy_volume"`
	SellVolume    float64 `json:"sell_volume"`
	BuyRatio      float64 `json:"buy_ratio"`
	SellRatio     float64 `json:"sell_ratio"`
	NetFlowVolume float64 `json:"net_flow_volume"`
	NetFlowValue  float64 `json:"net_flow_value"`

	// 分层交易统计（基于交易规模）
	SmallTrades  TradeSizeStats `json:"small_trades"`  // < $20k
	MediumTrades TradeSizeStats `json:"medium_trades"` // $20k - $70k
	LargeTrades  TradeSizeStats `json:"large_trades"`  // $70k - $175k
	WhaleTrades  TradeSizeStats `json:"whale_trades"`  // > $175k

	// 大单方向分析
	LargeBuyRatio  float64 `json:"large_buy_ratio"`  // 大单买入占比
	LargeSellRatio float64 `json:"large_sell_ratio"` // 大单卖出占比
	WhaleBuyRatio  float64 `json:"whale_buy_ratio"`  // 巨鲸买入占比
	WhaleSellRatio float64 `json:"whale_sell_ratio"` // 巨鲸卖出占比

	// 时间统计
	TimeWindow     string  `json:"time_window"`     // 时间窗口描述
	TradeIntensity float64 `json:"trade_intensity"` // 交易强度（笔/秒）
	TradeFrequency string  `json:"trade_frequency"` // 交易频率描述

	// 市场情绪评分
	SentimentScore    float64 `json:"sentiment_score"`    // -10到+10
	MomentumIndicator string  `json:"momentum_indicator"` // 动量指标描述

	// 关键信号
	Signals []string `json:"signals"` // 关键交易信号
}

// TradeFlowConfig 交易流分析配置
type TradeFlowConfig struct {
	SmallThreshold  float64 `json:"small_threshold"`  // 小单阈值（美元）
	MediumThreshold float64 `json:"medium_threshold"` // 中单阈值
	LargeThreshold  float64 `json:"large_threshold"`  // 大单阈值
	WhaleThreshold  float64 `json:"whale_threshold"`  // 巨鲸阈值
}

// DefaultTradeFlowConfig 默认配置
func DefaultTradeFlowConfig() *TradeFlowConfig {
	return &TradeFlowConfig{
		SmallThreshold:  20000.0,  // $20k
		MediumThreshold: 70000.0,  // $70k
		LargeThreshold:  175000.0, // $175k
		WhaleThreshold:  350000.0, // $350k
	}
}

// AnalyzeTradeFlow 分析交易流（增强版）
func AnalyzeTradeFlow(currentPrice float64, config *TradeFlowConfig) *TradeFlowAnalysis {
	if config == nil {
		config = DefaultTradeFlowConfig()
	}

	trades := tradeflow.GetOnceTradeFlow().GetRecentTradesLast10Minutes()
	if len(trades) == 0 {
		return &TradeFlowAnalysis{}
	}

	analysis := &TradeFlowAnalysis{
		TotalTrades: len(trades),
	}

	// 动态调整阈值（基于当前价格）
	adjustThresholds(config, currentPrice)

	// 初始化统计变量
	var timeIntervals []int64
	var totalBuyValue, totalSellValue float64
	var priceValues []float64
	var volumes []float64

	// 分析每笔交易
	for i, trade := range trades {
		qty, _ := strconv.ParseFloat(trade.Qty, 64)
		price, _ := strconv.ParseFloat(trade.Price, 64)
		tradeValue := qty * price

		// 基础统计
		analysis.TotalVolume += qty
		analysis.TotalValue += tradeValue
		priceValues = append(priceValues, price)
		volumes = append(volumes, qty)

		// 最大交易
		if tradeValue > analysis.MaxTradeValue {
			analysis.MaxTradeValue = tradeValue
		}

		// 时间间隔分析
		if i > 0 {
			interval := trade.Time - trades[i-1].Time
			timeIntervals = append(timeIntervals, interval)
		}

		// 买卖方向统计
		if !trade.IsBuyerMaker { // 主动买入
			analysis.BuyVolume += qty
			totalBuyValue += tradeValue
		} else { // 主动卖出
			analysis.SellVolume += qty
			totalSellValue += tradeValue
		}

		// 分层统计（使用动态阈值）
		categorizeTrade(analysis, trade, qty, tradeValue, config)
	}

	// 计算衍生指标
	calculateDerivedMetrics(analysis, totalBuyValue, totalSellValue, timeIntervals, priceValues, volumes)

	// 添加时间分层分析（最近2分钟）
	analysis.Recent5Min = AnalyzeTimeLayerTrades(tradeflow.GetOnceTradeFlow().GetRecentTradesLast5Minutes(), 2*60*1000, currentPrice, config)
	analysis.Recent20Min = AnalyzeTimeLayerTrades(tradeflow.GetOnceTradeFlow().GetRecentTradesLast20Minutes(), 15*60*1000, currentPrice, config)

	return analysis
}

// adjustThresholds 动态调整阈值
func adjustThresholds(config *TradeFlowConfig, currentPrice float64) {
	if currentPrice <= 0 {
		return
	}

	// 基于ETH数量调整阈值
	baseETH := 5.0                                        // 基础5 ETH
	config.SmallThreshold = baseETH * currentPrice        // 5 ETH
	config.MediumThreshold = baseETH * 3.5 * currentPrice // 17.5 ETH
	config.LargeThreshold = baseETH * 10 * currentPrice   // 50 ETH
	config.WhaleThreshold = baseETH * 20 * currentPrice   // 100 ETH
}

// categorizeTrade 交易分类统计
func categorizeTrade(analysis *TradeFlowAnalysis, trade binance.RecentTrade, qty, tradeValue float64, config *TradeFlowConfig) {
	var stats *TradeSizeStats

	// 根据交易价值分类（使用动态阈值）
	switch {
	case tradeValue >= config.WhaleThreshold:
		stats = &analysis.WhaleTrades
	case tradeValue >= config.LargeThreshold:
		stats = &analysis.LargeTrades
	case tradeValue >= config.MediumThreshold:
		stats = &analysis.MediumTrades
	default:
		stats = &analysis.SmallTrades
	}

	// 更新统计
	stats.Count++
	stats.Volume += qty
	stats.Value += tradeValue

	if !trade.IsBuyerMaker { // 主动买入
		stats.BuyVolume += qty
	} else { // 主动卖出
		stats.SellVolume += qty
	}
}

// calculateDerivedMetrics 计算衍生指标
func calculateDerivedMetrics(analysis *TradeFlowAnalysis, totalBuyValue, totalSellValue float64,
	timeIntervals []int64, priceValues []float64, volumes []float64) {

	// 平均交易值
	if analysis.TotalTrades > 0 {
		analysis.AvgTradeValue = analysis.TotalValue / float64(analysis.TotalTrades)
	}

	// 总体成交均价
	if analysis.TotalVolume > 0 {
		analysis.AvgPrice = analysis.TotalValue / analysis.TotalVolume
	}

	// 买卖比例
	if analysis.TotalVolume > 0 {
		analysis.BuyRatio = (analysis.BuyVolume / analysis.TotalVolume) * 100
		analysis.SellRatio = (analysis.SellVolume / analysis.TotalVolume) * 100
		analysis.NetFlowVolume = analysis.BuyVolume - analysis.SellVolume
		analysis.NetFlowValue = totalBuyValue - totalSellValue
	}

	// 大单方向分析（关键指标）
	calculateLargeTradeDirection(analysis)

	// 计算各分类平均交易金额
	calculateAvgValuePerTrade(analysis)

	// 机构行为指标
	calculateInstitutionalMetrics(analysis)

	// 时间序列指标
	calculateTimeMetrics(analysis, timeIntervals)

	// 价格冲击分析
	calculatePriceImpact(analysis, priceValues, volumes)

	// 综合评分
	calculateSentimentScore(analysis)
}

// calculateLargeTradeDirection 计算大单方向分析（关键新增）
func calculateLargeTradeDirection(analysis *TradeFlowAnalysis) {
	// 大单方向分析
	if analysis.LargeTrades.Volume > 0 {
		analysis.LargeBuyRatio = (analysis.LargeTrades.BuyVolume / analysis.LargeTrades.Volume) * 100
		analysis.LargeSellRatio = (analysis.LargeTrades.SellVolume / analysis.LargeTrades.Volume) * 100
	}

	// 巨鲸方向分析
	if analysis.WhaleTrades.Volume > 0 {
		analysis.WhaleBuyRatio = (analysis.WhaleTrades.BuyVolume / analysis.WhaleTrades.Volume) * 100
		analysis.WhaleSellRatio = (analysis.WhaleTrades.SellVolume / analysis.WhaleTrades.Volume) * 100
	}
}

// calculateAvgValuePerTrade 计算各分类平均每笔交易金额和成交均价
func calculateAvgValuePerTrade(analysis *TradeFlowAnalysis) {
	// 小单平均交易金额和成交均价
	if analysis.SmallTrades.Count > 0 {
		if analysis.SmallTrades.Volume > 0 {
			analysis.SmallTrades.AvgPrice = analysis.SmallTrades.Value / analysis.SmallTrades.Volume
		}
	}

	// 中单平均交易金额和成交均价
	if analysis.MediumTrades.Count > 0 {
		if analysis.MediumTrades.Volume > 0 {
			analysis.MediumTrades.AvgPrice = analysis.MediumTrades.Value / analysis.MediumTrades.Volume
		}
	}

	// 大单平均交易金额和成交均价
	if analysis.LargeTrades.Count > 0 {
		if analysis.LargeTrades.Volume > 0 {
			analysis.LargeTrades.AvgPrice = analysis.LargeTrades.Value / analysis.LargeTrades.Volume
		}
	}

	// 巨鲸平均交易金额和成交均价
	if analysis.WhaleTrades.Count > 0 {
		if analysis.WhaleTrades.Volume > 0 {
			analysis.WhaleTrades.AvgPrice = analysis.WhaleTrades.Value / analysis.WhaleTrades.Volume
		}
	}
}

// calculateInstitutionalMetrics 计算机构行为指标
func calculateInstitutionalMetrics(analysis *TradeFlowAnalysis) {
	// 主动交易比例（衡量市场急迫性）
	totalLargeVolume := analysis.LargeTrades.Volume + analysis.WhaleTrades.Volume
	if analysis.TotalVolume > 0 {
		analysis.AggressiveRatio = (totalLargeVolume / analysis.TotalVolume) * 100
	}

	// 机构资金流向（-1到+1，正值为买入偏向）
	// 需要处理Volume为0的情况，避免除零错误
	largeBuyValue := 0.0
	if analysis.LargeTrades.Volume > 0 {
		largeBuyValue = analysis.LargeTrades.BuyVolume * (analysis.LargeTrades.Value / analysis.LargeTrades.Volume)
	}

	whaleBuyValue := 0.0
	if analysis.WhaleTrades.Volume > 0 {
		whaleBuyValue = analysis.WhaleTrades.BuyVolume * (analysis.WhaleTrades.Value / analysis.WhaleTrades.Volume)
	}

	largeSellValue := 0.0
	if analysis.LargeTrades.Volume > 0 {
		largeSellValue = analysis.LargeTrades.SellVolume * (analysis.LargeTrades.Value / analysis.LargeTrades.Volume)
	}

	whaleSellValue := 0.0
	if analysis.WhaleTrades.Volume > 0 {
		whaleSellValue = analysis.WhaleTrades.SellVolume * (analysis.WhaleTrades.Value / analysis.WhaleTrades.Volume)
	}

	totalInstitutionalValue := largeBuyValue + whaleBuyValue + largeSellValue + whaleSellValue
	if totalInstitutionalValue > 0 {
		analysis.InstitutionalFlow = (largeBuyValue + whaleBuyValue - largeSellValue - whaleSellValue) / totalInstitutionalValue
	}
}

// calculateTimeMetrics 计算时间维度指标
func calculateTimeMetrics(analysis *TradeFlowAnalysis, timeIntervals []int64) {
	if len(timeIntervals) > 0 {
		// 平均交易间隔
		var totalInterval int64
		for _, interval := range timeIntervals {
			totalInterval += interval
		}
		analysis.AvgInterval = float64(totalInterval) / float64(len(timeIntervals)) // 毫秒

		// 交易强度（笔/秒）
		if analysis.AvgInterval > 0 {
			analysis.TradeIntensity = 1000.0 / analysis.AvgInterval
		}

		// 交易频率描述
		analysis.TradeFrequency = describeTradeFrequency(analysis.TradeIntensity)
	}
}

// calculatePriceImpact 计算价格冲击分析
func calculatePriceImpact(analysis *TradeFlowAnalysis, priceValues []float64, volumes []float64) {
	if len(priceValues) > 1 && len(volumes) > 0 {
		// 成交量加权平均价 (VWAP)
		var totalValue, totalVolume float64
		for i := 0; i < len(priceValues) && i < len(volumes); i++ {
			totalValue += priceValues[i] * volumes[i]
			totalVolume += volumes[i]
		}
		if totalVolume > 0 {
			analysis.VolumeWeightedPrice = totalValue / totalVolume
		}

		// 价格冲击（首尾价格变化相对于VWAP）
		firstPrice := priceValues[0]
		lastPrice := priceValues[len(priceValues)-1]
		if analysis.VolumeWeightedPrice > 0 {
			analysis.PriceImpact = ((lastPrice - firstPrice) / analysis.VolumeWeightedPrice) * 100
		}
	}
}

// calculateSentimentScore 计算综合情绪评分（基于1000笔交易的动态权重）
func calculateSentimentScore(analysis *TradeFlowAnalysis) {
	score := 0.0
	totalTrades := float64(analysis.TotalTrades)
	if totalTrades == 0 {
		totalTrades = 1000.0 // 默认值
	}

	// 1. 基础买卖不平衡 (基础权重: 30%, 始终计入)
	buySellImbalance := (analysis.BuyRatio - analysis.SellRatio) / 100.0 // -1到+1
	score += buySellImbalance * 3.0

	// 2. 大单方向偏向 (最大权重: 30%, 根据样本量动态调整)
	if analysis.LargeTrades.Count > 0 {
		largeDirection := (analysis.LargeBuyRatio - analysis.LargeSellRatio) / 100.0
		// 样本量置信度：30笔=100%, 15笔=50%, <10笔=20%
		largeConfidence := 1.0
		if analysis.LargeTrades.Count < 30 {
			largeConfidence = math.Max(0.2, float64(analysis.LargeTrades.Count)/30.0)
		}
		score += largeDirection * 3.0 * largeConfidence
	}

	// 3. 巨鲸方向 (最大权重: 25%, 根据样本量动态调整)
	if analysis.WhaleTrades.Count > 0 {
		whaleDirection := (analysis.WhaleBuyRatio - analysis.WhaleSellRatio) / 100.0
		// 样本量置信度：10笔=100%, 5笔=50%, <3笔=20%
		whaleConfidence := 1.0
		if analysis.WhaleTrades.Count < 10 {
			whaleConfidence = math.Max(0.2, float64(analysis.WhaleTrades.Count)/10.0)
		}
		score += whaleDirection * 2.5 * whaleConfidence
	}

	// 4. 机构流向 (最大权重: 10%, 根据样本量动态调整)
	totalInstitutional := float64(analysis.LargeTrades.Count + analysis.WhaleTrades.Count)
	if totalInstitutional > 0 {
		// 样本量置信度：40笔=100%, 20笔=50%, <10笔=20%
		institutionalConfidence := 1.0
		if totalInstitutional < 40 {
			institutionalConfidence = math.Max(0.2, totalInstitutional/40.0)
		}
		score += analysis.InstitutionalFlow * 1.0 * institutionalConfidence
	}

	// 5. 交易活跃度 (权重: 5%, 作为辅助指标)
	institutionalRatio := totalInstitutional / totalTrades
	if institutionalRatio > 0.04 { // 机构级交易占比>4%时才考虑活跃度
		activityBonus := math.Min(institutionalRatio, 0.15) * 0.5
		if buySellImbalance > 0 { // 买盘时活跃度加分
			score += activityBonus
		} else { // 卖盘时活跃度减分
			score -= activityBonus
		}
	}

	// 转换为-10到+10的评分
	analysis.SentimentScore = math.Max(-10, math.Min(10, score))

	// 动量指标描述
	analysis.MomentumIndicator = describeMomentum(analysis)
}

// describeTradeFrequency 描述交易频率
func describeTradeFrequency(intensity float64) string {
	switch {
	case intensity >= 2.0:
		return "极高频"
	case intensity >= 1.0:
		return "高频"
	case intensity >= 0.5:
		return "中频"
	case intensity >= 0.2:
		return "低频"
	default:
		return "极低频"
	}
}

// describeMomentum 描述动量指标
func describeMomentum(analysis *TradeFlowAnalysis) string {
	if math.Abs(analysis.SentimentScore) >= 7 {
		if analysis.SentimentScore > 0 {
			return "强势买入动量"
		}
		return "强势卖出动量"
	} else if math.Abs(analysis.SentimentScore) >= 4 {
		if analysis.SentimentScore > 0 {
			return "明显买入倾向"
		}
		return "明显卖出倾向"
	} else if math.Abs(analysis.SentimentScore) >= 2 {
		if analysis.SentimentScore > 0 {
			return "温和买盘"
		}
		return "温和卖盘"
	}
	return "相对平衡"
}

// GenerateTradeFlowReport 生成交易流分析报告（LLM友好格式）
func GenerateTradeFlowReport(analysis *TradeFlowAnalysis) string {
	if analysis.TotalTrades == 0 {
		return "交易流分析: 无近期交易数据"
	}

	var report strings.Builder
	// 时间分层分析（最近3分钟）- 放在最前面
	if analysis.Recent5Min != nil {
		report.WriteString(fmt.Sprintf("📅 微观信号窗口：最近5分钟 %d笔交易\n", analysis.Recent5Min.TotalTrades))

		if analysis.Recent5Min.TotalTrades > 0 {
			report.WriteString(fmt.Sprintf("    成交: %d笔, %.2f ETH, 均价%.0f USDT\n",
				analysis.Recent5Min.TotalTrades, analysis.Recent5Min.TotalVolume, analysis.Recent5Min.AvgPrice))
			report.WriteString(fmt.Sprintf("    买卖比: %.1f%% vs %.1f%%, 资金净流向: %s %.2f ETH\n",
				analysis.Recent5Min.BuyRatio, analysis.Recent5Min.SellRatio,
				getNetFlowDirection(analysis.Recent5Min.NetFlowVolume), math.Abs(analysis.Recent5Min.NetFlowVolume)))
			report.WriteString(fmt.Sprintf("    情绪评分: %.1f/10 (%s), 交易频率: %s (%.1f笔/秒)\n",
				analysis.Recent5Min.SentimentScore, analysis.Recent5Min.MomentumIndicator,
				analysis.Recent5Min.TradeFrequency, analysis.Recent5Min.TradeIntensity))

			// 关键信号
			if len(analysis.Recent5Min.Signals) > 0 {
				report.WriteString("    🔔 关键信号: " + strings.Join(analysis.Recent5Min.Signals, ", ") + "\n")
			}
		}
	}

	// 时间分层分析（最近15分钟)
	if analysis.Recent5Min != nil {
		report.WriteString(fmt.Sprintf("\n📅 趋势确认窗口：最近20分钟 %d笔交易\n", analysis.Recent20Min.TotalTrades))

		if analysis.Recent20Min.TotalTrades > 0 {
			report.WriteString(fmt.Sprintf("    成交: %d笔, %.2f ETH, 均价%.0f USDT\n",
				analysis.Recent20Min.TotalTrades, analysis.Recent20Min.TotalVolume, analysis.Recent20Min.AvgPrice))
			report.WriteString(fmt.Sprintf("    买卖比: %.1f%% vs %.1f%%, 资金净流向: %s %.2f ETH\n",
				analysis.Recent20Min.BuyRatio, analysis.Recent20Min.SellRatio,
				getNetFlowDirection(analysis.Recent20Min.NetFlowVolume), math.Abs(analysis.Recent20Min.NetFlowVolume)))
			report.WriteString(fmt.Sprintf("    情绪评分: %.1f/10 (%s), 交易频率: %s (%.1f笔/秒)\n",
				analysis.Recent20Min.SentimentScore, analysis.Recent20Min.MomentumIndicator,
				analysis.Recent20Min.TradeFrequency, analysis.Recent20Min.TradeIntensity))

			// 关键信号
			if len(analysis.Recent20Min.Signals) > 0 {
				report.WriteString("    🔔 关键信号: " + strings.Join(analysis.Recent20Min.Signals, ", ") + "\n")
			}
		}
	}

	// 添加主要时间窗口信息
	report.WriteString(fmt.Sprintf("\n📅 主要决策窗口: 最近10分钟，%d笔交易\n", analysis.TotalTrades))

	// 核心情绪指标（优先展示）
	report.WriteString("🎯 核心情绪指标:\n")
	report.WriteString(fmt.Sprintf("  综合评分: %.1f/10 (%s)\n", analysis.SentimentScore, analysis.MomentumIndicator))
	report.WriteString(fmt.Sprintf("  买卖比例: %.1f%% 买 vs %.1f%% 卖\n", analysis.BuyRatio, analysis.SellRatio))
	direction := "净流入"
	if analysis.NetFlowVolume < 0 {
		direction = "净流出"
	}
	report.WriteString(fmt.Sprintf("  资金净流向: %s %.0f ETH\n", direction, math.Abs(analysis.NetFlowVolume)))

	// 大单方向分析（关键新增）
	if analysis.LargeTrades.Count > 0 {
		report.WriteString(fmt.Sprintf("  大单倾向: %.1f%% 买 vs %.1f%% 卖\n", analysis.LargeBuyRatio, analysis.LargeSellRatio))
	}
	if analysis.WhaleTrades.Count > 0 {
		report.WriteString(fmt.Sprintf("  巨鲸倾向: %.1f%% 买 vs %.1f%% 卖\n", analysis.WhaleBuyRatio, analysis.WhaleSellRatio))
	}

	// 机构行为分析
	report.WriteString(fmt.Sprintf("  机构活跃度: %.1f%%, 资金流向: %s\n",
		analysis.AggressiveRatio, getInstitutionalFlowDirection(analysis.InstitutionalFlow)))

	// 分层交易统计
	report.WriteString("\n📊 交易规模分析:\n")
	report.WriteString(fmt.Sprintf("  总成交: %d笔, %.2f ETH, 平均约%.0f USDT/笔, 成交均价%.0f USDT\n",
		analysis.TotalTrades, analysis.TotalVolume, analysis.AvgTradeValue, analysis.AvgPrice))

	printTradeSizeStats(&report, "小单(<$20k)", &analysis.SmallTrades, analysis.TotalTrades)
	printTradeSizeStats(&report, "中单($20k-70k)", &analysis.MediumTrades, analysis.TotalTrades)
	printTradeSizeStats(&report, "大单($70k-175k)", &analysis.LargeTrades, analysis.TotalTrades)
	printTradeSizeStats(&report, "巨鲸(>$175k)", &analysis.WhaleTrades, analysis.TotalTrades)

	// 时间维度分析
	report.WriteString("\n⏱️ 交易活跃度:\n")
	report.WriteString(fmt.Sprintf("  交易频率: %s (%.2f笔/秒)\n", analysis.TradeFrequency, analysis.TradeIntensity))
	report.WriteString(fmt.Sprintf("  平均间隔: %.0f毫秒\n", analysis.AvgInterval))

	// 价格冲击分析（更客观的描述）
	if analysis.PriceImpact != 0 {
		report.WriteString("\n💥 价格冲击分析:\n")
		report.WriteString(fmt.Sprintf("  价格冲击: %.3f%% (%s)\n",
			analysis.PriceImpact, getPriceImpactDirection(analysis.PriceImpact)))
		report.WriteString(fmt.Sprintf("  成交量VWAP: %.2f USDT\n", analysis.VolumeWeightedPrice))
	}

	// 专业交易信号分析（基于1000笔交易的样本量阈值）
	report.WriteString("\n🎯 交易信号分析:\n")
	hasSignal := false

	// 计算各类交易占比（基于1000笔总量）
	totalTrades := float64(analysis.TotalTrades)
	largeTradeRatio := float64(analysis.LargeTrades.Count) / totalTrades * 100
	whaleTradeRatio := float64(analysis.WhaleTrades.Count) / totalTrades * 100

	// 大单方向建议（需要至少30笔，约3%的占比）
	if analysis.LargeTrades.Count >= 30 {
		if analysis.LargeBuyRatio > 70 {
			report.WriteString(fmt.Sprintf("  ✓ 大单明显偏向买入 (%.1f%%, %d笔/%.1f%%) → 做多信号\n",
				analysis.LargeBuyRatio, analysis.LargeTrades.Count, largeTradeRatio))
			hasSignal = true
		} else if analysis.LargeBuyRatio < 30 {
			report.WriteString(fmt.Sprintf("  ✓ 大单明显偏向卖出 (%.1f%%, %d笔/%.1f%%) → 做空信号\n",
				100-analysis.LargeBuyRatio, analysis.LargeTrades.Count, largeTradeRatio))
			hasSignal = true
		} else if analysis.LargeTrades.Count >= 50 {
			// 样本量充足但方向不明显（至少5%的大单）
			report.WriteString(fmt.Sprintf("  ⚠ 大单方向不明确 (买入%.1f%% vs 卖出%.1f%%, %d笔/%.1f%%) → 观望\n",
				analysis.LargeBuyRatio, 100-analysis.LargeBuyRatio, analysis.LargeTrades.Count, largeTradeRatio))
			hasSignal = true
		}
	} else if analysis.LargeTrades.Count >= 15 {
		// 样本量较小（1.5%-3%），降低置信度
		if analysis.LargeBuyRatio > 75 || analysis.LargeBuyRatio < 25 {
			direction := "买入"
			ratio := analysis.LargeBuyRatio
			if analysis.LargeBuyRatio < 50 {
				direction = "卖出"
				ratio = 100 - analysis.LargeBuyRatio
			}
			report.WriteString(fmt.Sprintf("  ⚠ 大单偏向%s (%.1f%%, %d笔/%.1f%%) → 弱信号，样本量偏小\n",
				direction, ratio, analysis.LargeTrades.Count, largeTradeRatio))
			hasSignal = true
		}
	} else if analysis.LargeTrades.Count > 0 {
		report.WriteString(fmt.Sprintf("  ⚠ 大单样本量不足 (%d笔/%.1f%%) → 信号不可靠\n",
			analysis.LargeTrades.Count, largeTradeRatio))
		hasSignal = true
	}

	// 巨鲸方向建议（需要至少10笔，约1%的占比）
	if analysis.WhaleTrades.Count >= 10 {
		if analysis.WhaleBuyRatio > 75 {
			report.WriteString(fmt.Sprintf("  ✓ 巨鲸大额买入 (%.1f%%, %d笔/%.1f%%) → 强烈做多信号\n",
				analysis.WhaleBuyRatio, analysis.WhaleTrades.Count, whaleTradeRatio))
			hasSignal = true
		} else if analysis.WhaleBuyRatio < 25 {
			report.WriteString(fmt.Sprintf("  ✓ 巨鲸大额卖出 (%.1f%%, %d笔/%.1f%%) → 强烈做空信号\n",
				100-analysis.WhaleBuyRatio, analysis.WhaleTrades.Count, whaleTradeRatio))
			hasSignal = true
		} else if analysis.WhaleTrades.Count >= 20 {
			// 巨鲸单充足但方向分散（至少2%）
			report.WriteString(fmt.Sprintf("  ⚠ 巨鲸方向分散 (买入%.1f%% vs 卖出%.1f%%, %d笔/%.1f%%) → 市场博弈激烈\n",
				analysis.WhaleBuyRatio, 100-analysis.WhaleBuyRatio, analysis.WhaleTrades.Count, whaleTradeRatio))
			hasSignal = true
		}
	} else if analysis.WhaleTrades.Count >= 5 {
		// 巨鲸单较少（0.5%-1%），极度谨慎
		if analysis.WhaleBuyRatio > 80 || analysis.WhaleBuyRatio < 20 {
			direction := "买入"
			ratio := analysis.WhaleBuyRatio
			if analysis.WhaleBuyRatio < 50 {
				direction = "卖出"
				ratio = 100 - analysis.WhaleBuyRatio
			}
			report.WriteString(fmt.Sprintf("  ⚠ 少量巨鲸%s (%d笔/%.1f%%, %.1f%%) → 信号可信度较低\n",
				direction, analysis.WhaleTrades.Count, whaleTradeRatio, ratio))
			hasSignal = true
		}
	} else if analysis.WhaleTrades.Count > 0 {
		report.WriteString(fmt.Sprintf("  ⚠ 巨鲸单极少 (%d笔/%.1f%%) → 不构成有效信号\n",
			analysis.WhaleTrades.Count, whaleTradeRatio))
		hasSignal = true
	}

	// 机构流向建议（需要足够的大单+巨鲸交易，至少40笔，约4%）
	totalInstitutionalTrades := analysis.LargeTrades.Count + analysis.WhaleTrades.Count
	institutionalRatio := float64(totalInstitutionalTrades) / totalTrades * 100

	if totalInstitutionalTrades >= 40 && math.Abs(analysis.InstitutionalFlow) > 0.35 {
		if analysis.InstitutionalFlow > 0 {
			report.WriteString(fmt.Sprintf("  ✓ 机构资金净流入 (流向指数%.2f, %d笔/%.1f%%) → 跟随买入\n",
				analysis.InstitutionalFlow, totalInstitutionalTrades, institutionalRatio))
		} else {
			report.WriteString(fmt.Sprintf("  ✓ 机构资金净流出 (流向指数%.2f, %d笔/%.1f%%) → 跟随卖出\n",
				analysis.InstitutionalFlow, totalInstitutionalTrades, institutionalRatio))
		}
		hasSignal = true
	} else if totalInstitutionalTrades >= 20 && math.Abs(analysis.InstitutionalFlow) > 0.5 {
		// 样本量较小但流向极其明显，给出弱信号
		direction := "流入"
		if analysis.InstitutionalFlow < 0 {
			direction = "流出"
		}
		report.WriteString(fmt.Sprintf("  ⚠ 机构资金强势%s (流向指数%.2f, %d笔/%.1f%%) → 弱信号，样本偏小但方向明确\n",
			direction, analysis.InstitutionalFlow, totalInstitutionalTrades, institutionalRatio))
		hasSignal = true
	} else if totalInstitutionalTrades > 0 {
		report.WriteString(fmt.Sprintf("  ⚠ 机构级交易不足 (%d笔/%.1f%%) → 流向判断不可靠\n",
			totalInstitutionalTrades, institutionalRatio))
		hasSignal = true
	}

	// 活跃度建议（基于大单+巨鲸占比和情绪评分）
	if institutionalRatio > 8 && math.Abs(analysis.SentimentScore) > 4 {
		if analysis.SentimentScore > 0 {
			report.WriteString(fmt.Sprintf("  ✓ 高活跃度看多 (机构占比%.1f%%, 情绪%.1f) → 积极做多氛围\n",
				institutionalRatio, analysis.SentimentScore))
			hasSignal = true
		} else {
			report.WriteString(fmt.Sprintf("  ✓ 高活跃度看空 (机构占比%.1f%%, 情绪%.1f) → 积极做空氛围\n",
				institutionalRatio, analysis.SentimentScore))
			hasSignal = true
		}
	}

	// 如果没有明确信号
	if !hasSignal {
		report.WriteString("  ℹ️  当前交易数据不足以产生明确信号，建议继续观察\n")
		report.WriteString(fmt.Sprintf("  当前总交易量: %d笔 (大单%d笔, 巨鲸%d笔)\n",
			analysis.TotalTrades, analysis.LargeTrades.Count, analysis.WhaleTrades.Count))
	}

	return report.String()
}

// 辅助函数
func printTradeSizeStats(report *strings.Builder, label string, stats *TradeSizeStats, totalTrades int) {
	if stats.Count > 0 {
		if stats.Volume > 0 {
			stats.BuyRatio = (stats.BuyVolume / stats.Volume) * 100
		}
		percentage := 0.0
		if totalTrades > 0 {
			percentage = float64(stats.Count) / float64(totalTrades) * 100
		}
		report.WriteString(fmt.Sprintf("  %s: %d笔 (%.1f%%), 买卖比 %.1f:%.1f, 成交均价%.0f USDT\n",
			label, stats.Count, percentage, stats.BuyRatio, 100-stats.BuyRatio, stats.AvgPrice))
	}
}

func getInstitutionalFlowDirection(flow float64) string {
	if flow > 0.2 {
		return "净流入"
	} else if flow < -0.2 {
		return "净流出"
	}
	return "相对平衡"
}

func getPriceImpactDirection(impact float64) string {
	if impact > 0.1 {
		return "价格上涨压力"
	} else if impact < -0.1 {
		return "价格下跌压力"
	}
	return "价格稳定"
}

// AnalyzeTimeLayerTrades 分析特定时间窗口内的交易流
func AnalyzeTimeLayerTrades(trades []binance.RecentTrade, timeWindowMs int64, currentPrice float64, config *TradeFlowConfig) *TimeLayerAnalysis {
	if config == nil {
		config = DefaultTradeFlowConfig()
	}

	if len(trades) == 0 {
		return &TimeLayerAnalysis{}
	}

	// 获取交易数据中的最新时间戳作为参考点
	var latestTime int64
	if len(trades) > 0 {
		latestTime = trades[len(trades)-1].Time // 最新交易在最后
	} else {
		latestTime = time.Now().UnixMilli() // 备用：当前系统时间
	}
	cutoffTime := latestTime - timeWindowMs

	// 筛选时间窗口内的交易
	var filteredTrades []binance.RecentTrade
	for _, trade := range trades {
		if trade.Time >= cutoffTime {
			filteredTrades = append(filteredTrades, trade)
		}
	}

	if len(filteredTrades) == 0 {
		return &TimeLayerAnalysis{
			TimeWindow: fmt.Sprintf("最近%d分钟内无交易", timeWindowMs/60000),
		}
	}

	analysis := &TimeLayerAnalysis{
		TotalTrades: len(filteredTrades),
		TimeWindow:  fmt.Sprintf("最近%.1f分钟", float64(timeWindowMs)/60000.0),
	}

	// 动态调整阈值（基于当前价格）
	adjustThresholds(config, currentPrice)

	// 初始化统计变量
	var timeIntervals []int64
	var totalBuyValue, totalSellValue float64
	var priceValues []float64
	var volumes []float64

	// 分析每笔交易
	for i, trade := range filteredTrades {
		qty, _ := strconv.ParseFloat(trade.Qty, 64)
		price, _ := strconv.ParseFloat(trade.Price, 64)
		tradeValue := qty * price

		// 基础统计
		analysis.TotalVolume += qty
		analysis.TotalValue += tradeValue
		priceValues = append(priceValues, price)
		volumes = append(volumes, qty)

		// 时间间隔分析
		if i > 0 {
			interval := trade.Time - filteredTrades[i-1].Time
			timeIntervals = append(timeIntervals, interval)
		}

		// 买卖方向统计
		if !trade.IsBuyerMaker { // 主动买入
			analysis.BuyVolume += qty
			totalBuyValue += tradeValue
		} else { // 主动卖出
			analysis.SellVolume += qty
			totalSellValue += tradeValue
		}

		// 分层统计（使用动态阈值）
		categorizeTimeLayerTrade(analysis, trade, qty, tradeValue, config)
	}

	// 计算衍生指标
	calculateTimeLayerDerivedMetrics(analysis, totalBuyValue, totalSellValue, timeIntervals, priceValues, volumes)

	return analysis
}

// categorizeTimeLayerTrade 时间分层交易分类统计
func categorizeTimeLayerTrade(analysis *TimeLayerAnalysis, trade binance.RecentTrade, qty, tradeValue float64, config *TradeFlowConfig) {
	var stats *TradeSizeStats

	// 根据交易价值分类（使用动态阈值）
	switch {
	case tradeValue >= config.WhaleThreshold:
		stats = &analysis.WhaleTrades
	case tradeValue >= config.LargeThreshold:
		stats = &analysis.LargeTrades
	case tradeValue >= config.MediumThreshold:
		stats = &analysis.MediumTrades
	default:
		stats = &analysis.SmallTrades
	}

	// 更新统计
	stats.Count++
	stats.Volume += qty
	stats.Value += tradeValue

	if !trade.IsBuyerMaker { // 主动买入
		stats.BuyVolume += qty
	} else { // 主动卖出
		stats.SellVolume += qty
	}
}

// calculateTimeLayerDerivedMetrics 计算时间分层衍生指标
func calculateTimeLayerDerivedMetrics(analysis *TimeLayerAnalysis, totalBuyValue, totalSellValue float64,
	timeIntervals []int64, priceValues []float64, volumes []float64) {

	// 平均交易值
	if analysis.TotalTrades > 0 {
		analysis.AvgTradeValue = analysis.TotalValue / float64(analysis.TotalTrades)
	}

	// 总体成交均价
	if analysis.TotalVolume > 0 {
		analysis.AvgPrice = analysis.TotalValue / analysis.TotalVolume
	}

	// 买卖比例
	if analysis.TotalVolume > 0 {
		analysis.BuyRatio = (analysis.BuyVolume / analysis.TotalVolume) * 100
		analysis.SellRatio = (analysis.SellVolume / analysis.TotalVolume) * 100
		analysis.NetFlowVolume = analysis.BuyVolume - analysis.SellVolume
		analysis.NetFlowValue = totalBuyValue - totalSellValue
	}

	// 大单方向分析
	if analysis.LargeTrades.Volume > 0 {
		analysis.LargeBuyRatio = (analysis.LargeTrades.BuyVolume / analysis.LargeTrades.Volume) * 100
		analysis.LargeSellRatio = (analysis.LargeTrades.SellVolume / analysis.LargeTrades.Volume) * 100
	}

	// 巨鲸方向分析
	if analysis.WhaleTrades.Volume > 0 {
		analysis.WhaleBuyRatio = (analysis.WhaleTrades.BuyVolume / analysis.WhaleTrades.Volume) * 100
		analysis.WhaleSellRatio = (analysis.WhaleTrades.SellVolume / analysis.WhaleTrades.Volume) * 100
	}

	// 计算各分类平均成交均价
	calculateTimeLayerAvgPrice(analysis)

	// 时间序列指标
	if len(timeIntervals) > 0 {
		// 平均交易间隔
		var totalInterval int64
		for _, interval := range timeIntervals {
			totalInterval += interval
		}
		avgInterval := float64(totalInterval) / float64(len(timeIntervals)) // 毫秒

		// 交易强度（笔/秒）
		if avgInterval > 0 {
			analysis.TradeIntensity = 1000.0 / avgInterval
		}

		// 交易频率描述
		analysis.TradeFrequency = describeTradeFrequency(analysis.TradeIntensity)
	}

	// 综合评分（简化版）
	calculateTimeLayerSentimentScore(analysis)

	// 生成关键信号
	generateTimeLayerSignals(analysis)
}

// calculateTimeLayerAvgPrice 计算时间分层各分类成交均价
func calculateTimeLayerAvgPrice(analysis *TimeLayerAnalysis) {
	// 小单成交均价
	if analysis.SmallTrades.Count > 0 && analysis.SmallTrades.Volume > 0 {
		analysis.SmallTrades.AvgPrice = analysis.SmallTrades.Value / analysis.SmallTrades.Volume
	}

	// 中单成交均价
	if analysis.MediumTrades.Count > 0 && analysis.MediumTrades.Volume > 0 {
		analysis.MediumTrades.AvgPrice = analysis.MediumTrades.Value / analysis.MediumTrades.Volume
	}

	// 大单成交均价
	if analysis.LargeTrades.Count > 0 && analysis.LargeTrades.Volume > 0 {
		analysis.LargeTrades.AvgPrice = analysis.LargeTrades.Value / analysis.LargeTrades.Volume
	}

	// 巨鲸成交均价
	if analysis.WhaleTrades.Count > 0 && analysis.WhaleTrades.Volume > 0 {
		analysis.WhaleTrades.AvgPrice = analysis.WhaleTrades.Value / analysis.WhaleTrades.Volume
	}
}

// calculateTimeLayerSentimentScore 计算时间分层情绪评分
func calculateTimeLayerSentimentScore(analysis *TimeLayerAnalysis) {
	score := 0.0

	// 1. 基础买卖不平衡 (权重: 40%)
	buySellImbalance := (analysis.BuyRatio - analysis.SellRatio) / 100.0 // -1到+1
	score += buySellImbalance * 4.0

	// 2. 大单方向偏向 (权重: 35%)
	if analysis.LargeTrades.Count > 0 {
		largeDirection := (analysis.LargeBuyRatio - analysis.LargeSellRatio) / 100.0
		// 样本量置信度：15笔=100%, 8笔=50%, <5笔=20%
		largeConfidence := 1.0
		if analysis.LargeTrades.Count < 15 {
			largeConfidence = math.Max(0.2, float64(analysis.LargeTrades.Count)/15.0)
		}
		score += largeDirection * 3.5 * largeConfidence
	}

	// 3. 巨鲸方向 (权重: 25%)
	if analysis.WhaleTrades.Count > 0 {
		whaleDirection := (analysis.WhaleBuyRatio - analysis.WhaleSellRatio) / 100.0
		// 样本量置信度：5笔=100%, 3笔=60%, <2笔=20%
		whaleConfidence := 1.0
		if analysis.WhaleTrades.Count < 5 {
			whaleConfidence = math.Max(0.2, float64(analysis.WhaleTrades.Count)/5.0)
		}
		score += whaleDirection * 2.5 * whaleConfidence
	}

	// 转换为-10到+10的评分
	analysis.SentimentScore = math.Max(-10, math.Min(10, score))

	// 动量指标描述
	analysis.MomentumIndicator = describeMomentum(&TradeFlowAnalysis{
		SentimentScore:    analysis.SentimentScore,
		MomentumIndicator: analysis.MomentumIndicator,
	})
}

// generateTimeLayerSignals 生成时间分层关键信号
func generateTimeLayerSignals(analysis *TimeLayerAnalysis) {
	var signals []string
	totalTrades := float64(analysis.TotalTrades)

	// 大单方向信号
	if analysis.LargeTrades.Count >= 8 { // 至少8笔大单
		largeTradeRatio := float64(analysis.LargeTrades.Count) / totalTrades * 100
		if analysis.LargeBuyRatio > 75 {
			signals = append(signals, fmt.Sprintf("大单明显买入(%d笔,%.1f%%)", analysis.LargeTrades.Count, largeTradeRatio))
		} else if analysis.LargeBuyRatio < 25 {
			signals = append(signals, fmt.Sprintf("大单明显卖出(%d笔,%.1f%%)", analysis.LargeTrades.Count, largeTradeRatio))
		}
	}

	// 巨鲸信号
	if analysis.WhaleTrades.Count >= 3 { // 至少3笔巨鲸单
		whaleTradeRatio := float64(analysis.WhaleTrades.Count) / totalTrades * 100
		if analysis.WhaleBuyRatio > 80 {
			signals = append(signals, fmt.Sprintf("巨鲸大额买入(%d笔,%.1f%%)", analysis.WhaleTrades.Count, whaleTradeRatio))
		} else if analysis.WhaleBuyRatio < 20 {
			signals = append(signals, fmt.Sprintf("巨鲸大额卖出(%d笔,%.1f%%)", analysis.WhaleTrades.Count, whaleTradeRatio))
		}
	}

	// 交易活跃度信号
	if analysis.TradeIntensity > 1.5 && math.Abs(analysis.SentimentScore) > 3 {
		if analysis.SentimentScore > 0 {
			signals = append(signals, fmt.Sprintf("高频买盘活跃(%.1f笔/秒)", analysis.TradeIntensity))
		} else {
			signals = append(signals, fmt.Sprintf("高频卖盘活跃(%.1f笔/秒)", analysis.TradeIntensity))
		}
	}

	// 资金流向信号
	if analysis.NetFlowVolume > 0 && math.Abs(analysis.NetFlowVolume) > 0.5 {
		signals = append(signals, fmt.Sprintf("资金净流入%.2fETH", analysis.NetFlowVolume))
	} else if analysis.NetFlowVolume < 0 && math.Abs(analysis.NetFlowVolume) > 0.5 {
		signals = append(signals, fmt.Sprintf("资金净流出%.2fETH", math.Abs(analysis.NetFlowVolume)))
	}

	analysis.Signals = signals
}

// getNetFlowDirection 获取资金流向描述
func getNetFlowDirection(netFlow float64) string {
	if netFlow > 0 {
		return "净流入"
	} else if netFlow < 0 {
		return "净流出"
	}
	return "相对平衡"
}
