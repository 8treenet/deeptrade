package indicators

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	binance "deeptrade/binance"
)

// VolumeAnalysisConfig 成交量分析配置
type VolumeAnalysisConfig struct {
	// 时间窗口配置
	MicroWindowKlines  int `json:"micro_window_klines"`  // 微观窗口K线数量
	ShortWindowKlines  int `json:"short_window_klines"`  // 短期窗口K线数量
	MediumWindowKlines int `json:"medium_window_klines"` // 中期窗口K线数量
	LongWindowKlines   int `json:"long_window_klines"`   // 长期窗口K线数量

	// 巨量检测参数
	GiantVolumeRatio      float64 `json:"giant_volume_ratio"`      // 巨量比率阈值
	GiantVolumePercentile float64 `json:"giant_volume_percentile"` // 巨量百分位阈值

	// 地量检测参数
	LowVolumeRatio      float64 `json:"low_volume_ratio"`      // 地量比率阈值
	LowVolumePercentile float64 `json:"low_volume_percentile"` // 地量百分位阈值

	// 背离检测参数
	DivergenceLookback  int     `json:"divergence_lookback"`  // 背离检测回看期
	DivergenceThreshold float64 `json:"divergence_threshold"` // 背离检测阈值

	// 连续放量参数
	ContinuousMinLength int     `json:"continuous_min_length"` // 连续最短长度
	ContinuousThreshold float64 `json:"continuous_threshold"`  // 连续变化阈值
}

// DefaultVolumeAnalysisConfig 默认成交量分析配置
func DefaultVolumeAnalysisConfig() *VolumeAnalysisConfig {
	return &VolumeAnalysisConfig{
		MicroWindowKlines:  5,  // 微观窗口5条K线（15分钟）
		ShortWindowKlines:  15, // 短期窗口15条K线（45分钟）
		MediumWindowKlines: 30, // 中期窗口30条K线（90分钟）
		LongWindowKlines:   70, // 长期窗口70条K线（3.5小时）

		GiantVolumeRatio:      2.5,  // 巨量比率阈值
		GiantVolumePercentile: 95.0, // 巨量百分位阈值

		LowVolumeRatio:      0.3,  // 地量比率阈值
		LowVolumePercentile: 10.0, // 地量百分位阈值

		DivergenceLookback:  10,  // 背离检测回看期
		DivergenceThreshold: 0.2, // 背离检测阈值

		ContinuousMinLength: 3,   // 连续最短长度
		ContinuousThreshold: 1.1, // 连续变化阈值
	}
}

// VolumeSignal 成交量信号接口
type VolumeSignal interface {
	GetSignalType() string
	GetInterpretation() string
	GetSignificance() float64
}

// GiantVolumeSignal 巨量信号
type GiantVolumeSignal struct {
	SignalType     string  `json:"signal_type"`    // "GIANT_VOLUME"
	KlineIndex     int     `json:"kline_index"`    // K线索引
	Volume         float64 `json:"volume"`         // 成交量
	VolumeRatio    float64 `json:"volume_ratio"`   // 与平均成交量的比率
	PriceChange    float64 `json:"price_change"`   // 价格变化百分比
	BuyRatio       float64 `json:"buy_ratio"`      // 主动买入比例
	Significance   string  `json:"significance"`   // 重要性级别
	Interpretation string  `json:"interpretation"` // 信号解读
}

func (s *GiantVolumeSignal) GetSignalType() string     { return s.SignalType }
func (s *GiantVolumeSignal) GetInterpretation() string { return s.Interpretation }
func (s *GiantVolumeSignal) GetSignificance() float64 {
	switch s.Significance {
	case "极高":
		return 10.0
	case "高":
		return 8.0
	case "中等":
		return 6.0
	default:
		return 4.0
	}
}

// LowVolumeSignal 地量信号
type LowVolumeSignal struct {
	SignalType     string  `json:"signal_type"`    // "LOW_VOLUME"
	KlineIndex     int     `json:"kline_index"`    // K线索引
	Volume         float64 `json:"volume"`         // 成交量
	VolumeRatio    float64 `json:"volume_ratio"`   // 与平均成交量的比率
	PriceRange     float64 `json:"price_range"`    // 价格波动幅度
	Significance   string  `json:"significance"`   // 重要性级别
	Interpretation string  `json:"interpretation"` // 信号解读
}

func (s *LowVolumeSignal) GetSignalType() string     { return s.SignalType }
func (s *LowVolumeSignal) GetInterpretation() string { return s.Interpretation }
func (s *LowVolumeSignal) GetSignificance() float64 {
	switch s.Significance {
	case "极高":
		return 9.0
	case "高":
		return 7.0
	case "中等":
		return 5.0
	default:
		return 3.0
	}
}

// VolumePriceDivergenceSignal 量价背离信号
type VolumePriceDivergenceSignal struct {
	SignalType      string  `json:"signal_type"`      // "VOLUME_PRICE_DIVERGENCE"
	KlineIndex      int     `json:"kline_index"`      // K线索引
	PriceDirection  string  `json:"price_direction"`  // "UP"或"DOWN"
	VolumeDirection string  `json:"volume_direction"` // "UP"或"DOWN"
	DivergenceType  string  `json:"divergence_type"`  // "BULLISH"或"BEARISH"
	Strength        float64 `json:"strength"`         // 背离强度
	Interpretation  string  `json:"interpretation"`   // 信号解读
}

func (s *VolumePriceDivergenceSignal) GetSignalType() string     { return s.SignalType }
func (s *VolumePriceDivergenceSignal) GetInterpretation() string { return s.Interpretation }
func (s *VolumePriceDivergenceSignal) GetSignificance() float64  { return s.Strength * 10 }

// ContinuousVolumeSignal 连续放量信号
type ContinuousVolumeSignal struct {
	SignalType     string  `json:"signal_type"`      // "CONTINUOUS_VOLUME"
	StartIndex     int     `json:"start_index"`      // 起始索引
	EndIndex       int     `json:"end_index"`        // 结束索引
	Duration       int     `json:"duration"`         // 持续时间（根K线）
	VolumeTrend    string  `json:"volume_trend"`     // "EXPANDING"或"CONTRACTING"
	AvgVolumeRatio float64 `json:"avg_volume_ratio"` // 平均成交量比率
	PriceTrend     string  `json:"price_trend"`      // 价格趋势
	Accumulation   float64 `json:"accumulation"`     // 累积成交量
	Interpretation string  `json:"interpretation"`   // 信号解读
}

func (s *ContinuousVolumeSignal) GetSignalType() string     { return s.SignalType }
func (s *ContinuousVolumeSignal) GetInterpretation() string { return s.Interpretation }
func (s *ContinuousVolumeSignal) GetSignificance() float64 {
	if s.Duration >= 5 {
		return 7.0
	}
	return 5.0
}

// VolumeStackSignal 成交量堆信号
type VolumeStackSignal struct {
	SignalType     string  `json:"signal_type"`    // "VOLUME_STACK"
	StartIndex     int     `json:"start_index"`    // 起始索引
	EndIndex       int     `json:"end_index"`      // 结束索引
	Duration       int     `json:"duration"`       // 持续时间
	StackVolume    float64 `json:"stack_volume"`   // 堆积成交量
	StackRatio     float64 `json:"stack_ratio"`    // 占总成交量比例
	PriceAction    string  `json:"price_action"`   // 价格行为
	Distribution   string  `json:"distribution"`   // 成交量分布
	Interpretation string  `json:"interpretation"` // 信号解读
}

func (s *VolumeStackSignal) GetSignalType() string     { return s.SignalType }
func (s *VolumeStackSignal) GetInterpretation() string { return s.Interpretation }
func (s *VolumeStackSignal) GetSignificance() float64 {
	if s.StackRatio > 0.4 {
		return 8.0
	} else if s.StackRatio > 0.3 {
		return 6.0
	}
	return 4.0
}

// VolumeWindowAnalysis 时间窗口成交量分析
type VolumeWindowAnalysis struct {
	TimeWindow string `json:"time_window"` // 时间窗口描述
	KlineCount int    `json:"kline_count"` // K线数量

	// 基础成交量统计
	AvgVolume    float64 `json:"avg_volume"`     // 平均成交量
	MaxVolume    float64 `json:"max_volume"`     // 最大成交量
	MinVolume    float64 `json:"min_volume"`     // 最小成交量
	VolumeStdDev float64 `json:"volume_std_dev"` // 成交量标准差
	VolumeRange  float64 `json:"volume_range"`   // 成交量区间

	// 成交量趋势
	VolumeTrend    string  `json:"volume_trend"`    // 成交量趋势
	VolumeMomentum float64 `json:"volume_momentum"` // 成交量动量

	// 买卖压力分析
	BuyRatio      float64 `json:"buy_ratio"`       // 主动买入比例
	SellRatio     float64 `json:"sell_ratio"`      // 主动卖出比例
	NetFlowVolume float64 `json:"net_flow_volume"` // 净流入成交量

	// 量价关系
	PriceVolumeCorrelation float64 `json:"price_volume_correlation"` // 量价相关性
	VolumePriceSync        string  `json:"volume_price_sync"`        // 量价同步性描述

	// 关键信号
	Signals []VolumeSignal `json:"signals"` // 检测到的信号
}

// TimeLayerVolumeAnalysis 分层时间成交量分析
type TimeLayerVolumeAnalysis struct {
	// 微观窗口（最近5条K线）
	MicroWindow *VolumeWindowAnalysis `json:"micro_window"`

	// 短期窗口（最近15条K线）
	ShortWindow *VolumeWindowAnalysis `json:"short_window"`

	// 中期窗口（最近30条K线）
	MediumWindow *VolumeWindowAnalysis `json:"medium_window"`

	// 长期窗口（全部70条K线）
	LongWindow *VolumeWindowAnalysis `json:"long_window"`

	// 跨窗口比较
	CrossWindowComparison *CrossWindowAnalysis `json:"cross_window_comparison"`
}

// CrossWindowAnalysis 跨窗口比较分析
type CrossWindowAnalysis struct {
	VolumeAcceleration   float64 `json:"volume_acceleration"`    // 成交量加速度
	TrendConsistency     float64 `json:"trend_consistency"`      // 趋势一致性
	MultiTimeframeSignal string  `json:"multi_timeframe_signal"` // 多时间框架信号
}

// AnalyzeVolumeLayers 分析分层时间成交量
func AnalyzeVolumeLayers(klines []binance.Kline, config *VolumeAnalysisConfig) *TimeLayerVolumeAnalysis {
	if config == nil {
		config = DefaultVolumeAnalysisConfig()
	}

	if len(klines) == 0 {
		return &TimeLayerVolumeAnalysis{}
	}

	analysis := &TimeLayerVolumeAnalysis{}

	// 分析各个时间窗口
	if len(klines) >= config.MicroWindowKlines {
		start := len(klines) - config.MicroWindowKlines
		microKlines := klines[start:]
		analysis.MicroWindow = analyzeVolumeWindow(microKlines, "微观窗口（15分钟）", config)
	}

	if len(klines) >= config.ShortWindowKlines {
		start := len(klines) - config.ShortWindowKlines
		shortKlines := klines[start:]
		analysis.ShortWindow = analyzeVolumeWindow(shortKlines, "短期窗口（45分钟）", config)
	}

	if len(klines) >= config.MediumWindowKlines {
		start := len(klines) - config.MediumWindowKlines
		mediumKlines := klines[start:]
		analysis.MediumWindow = analyzeVolumeWindow(mediumKlines, "中期窗口（90分钟）", config)
	}

	if len(klines) >= config.LongWindowKlines {
		start := len(klines) - config.LongWindowKlines
		longKlines := klines[start:]
		analysis.LongWindow = analyzeVolumeWindow(longKlines, "长期窗口（3.5小时）", config)
	}

	// 跨窗口比较分析
	analysis.CrossWindowComparison = analyzeCrossWindowComparison(analysis)

	return analysis
}

// analyzeVolumeWindow 分析单个时间窗口的成交量
func analyzeVolumeWindow(klines []binance.Kline, timeWindow string, config *VolumeAnalysisConfig) *VolumeWindowAnalysis {
	if len(klines) == 0 {
		return &VolumeWindowAnalysis{TimeWindow: timeWindow}
	}

	analysis := &VolumeWindowAnalysis{
		TimeWindow: timeWindow,
		KlineCount: len(klines),
	}

	// 提取基础数据
	volumes := make([]float64, len(klines))
	prices := make([]float64, len(klines))
	buyVolumes := make([]float64, len(klines))

	var totalVolume, totalBuyVolume float64

	for i, kline := range klines {
		vol, _ := strconv.ParseFloat(kline.Volume, 64)
		price, _ := strconv.ParseFloat(kline.Close, 64)
		buyVol, _ := strconv.ParseFloat(kline.TakerBuyBaseAssetVolume, 64)

		volumes[i] = vol
		prices[i] = price
		buyVolumes[i] = buyVol

		totalVolume += vol
		totalBuyVolume += buyVol
	}

	// 基础成交量统计
	analysis.AvgVolume = totalVolume / float64(len(klines))
	analysis.MaxVolume = maxFloat64(volumes)
	analysis.MinVolume = minFloat64(volumes)
	analysis.VolumeRange = analysis.MaxVolume - analysis.MinVolume
	analysis.VolumeStdDev = calculateStdDev(volumes, analysis.AvgVolume)

	// 买卖压力分析
	if totalVolume > 0 {
		analysis.BuyRatio = (totalBuyVolume / totalVolume) * 100
		analysis.SellRatio = 100 - analysis.BuyRatio
		analysis.NetFlowVolume = totalBuyVolume - (totalVolume - totalBuyVolume)
	}

	// 成交量趋势
	analysis.VolumeTrend = calculateVolumeTrend(volumes)
	analysis.VolumeMomentum = calculateVolumeMomentum(volumes)

	// 量价关系
	analysis.PriceVolumeCorrelation = calculateCorrelation(prices, volumes)
	analysis.VolumePriceSync = describeVolumePriceSync(analysis.PriceVolumeCorrelation)

	// 检测信号
	analysis.Signals = detectVolumeSignals(klines, volumes, prices, buyVolumes, config)

	return analysis
}

// detectVolumeSignals 检测成交量信号
func detectVolumeSignals(klines []binance.Kline, volumes, prices, buyVolumes []float64, config *VolumeAnalysisConfig) []VolumeSignal {
	var signals []VolumeSignal

	// 检测巨量信号
	if giantSignal := detectGiantVolume(klines, volumes, prices, buyVolumes, config); giantSignal != nil {
		signals = append(signals, giantSignal)
	}

	// 检测地量信号
	if lowSignal := detectLowVolume(klines, volumes, prices, config); lowSignal != nil {
		signals = append(signals, lowSignal)
	}

	// 检测量价背离信号
	if divergenceSignal := detectVolumePriceDivergence(klines, volumes, prices, config); divergenceSignal != nil {
		signals = append(signals, divergenceSignal)
	}

	// 检测连续放量信号
	if continuousSignal := detectContinuousVolume(klines, volumes, prices, config); continuousSignal != nil {
		signals = append(signals, continuousSignal)
	}

	// 检测成交量堆信号
	if stackSignal := detectVolumeStack(klines, volumes, prices, config); stackSignal != nil {
		signals = append(signals, stackSignal)
	}

	return signals
}

// detectGiantVolume 检测巨量信号
func detectGiantVolume(klines []binance.Kline, volumes, prices, buyVolumes []float64, config *VolumeAnalysisConfig) *GiantVolumeSignal {
	if len(klines) < 20 || len(volumes) < 20 {
		return nil
	}

	// 获取最新K线数据
	currentIndex := len(klines) - 1
	currentVol := volumes[currentIndex]
	currentPrice := prices[currentIndex]
	currentBuyVol := buyVolumes[currentIndex]

	// 计算前20期平均成交量和最大成交量
	var sumVol, maxVol float64
	for i := currentIndex - 20; i < currentIndex; i++ {
		sumVol += volumes[i]
		if volumes[i] > maxVol {
			maxVol = volumes[i]
		}
	}
	avgVol := sumVol / 20

	// 计算最近70条中的排名（如果有足够数据）
	rank := 0.0
	if len(volumes) >= 70 {
		allVolumes := make([]float64, 70)
		copy(allVolumes, volumes[len(volumes)-70:])
		sort.Float64s(allVolumes)
		rank = float64(sort.SearchFloat64s(allVolumes, currentVol)) / 70.0 * 100
	}

	// 检测条件
	ratio := currentVol / avgVol
	if ratio > config.GiantVolumeRatio || currentVol > maxVol*1.8 || rank > config.GiantVolumePercentile {
		fmt.Printf("DEBUG: 巨量检测触发 - ratio=%.2f, currentVol=%.2f, avgVol=%.2f, maxVol=%.2f, rank=%.2f\n",
			ratio, currentVol, avgVol, maxVol, rank)
		// 计算价格变化
		priceChange := 0.0
		if currentIndex > 0 && prices[currentIndex-1] > 0 {
			priceChange = (currentPrice - prices[currentIndex-1]) / prices[currentIndex-1] * 100
		}

		// 计算买入比例
		buyRatio := 0.0
		if currentVol > 0 {
			buyRatio = currentBuyVol / currentVol * 100
		}

		// 确定重要性级别
		significance := "中等"
		if ratio > 4 || rank > 98 {
			significance = "极高"
		} else if ratio > 3 || rank > 96 {
			significance = "高"
		}

		return &GiantVolumeSignal{
			SignalType:     "GIANT_VOLUME",
			KlineIndex:     currentIndex,
			Volume:         currentVol,
			VolumeRatio:    ratio,
			PriceChange:    priceChange,
			BuyRatio:       buyRatio,
			Significance:   significance,
			Interpretation: generateGiantVolumeInterpretation(priceChange, buyRatio, significance),
		}
	}

	return nil
}

// detectLowVolume 检测地量信号
func detectLowVolume(klines []binance.Kline, volumes, prices []float64, config *VolumeAnalysisConfig) *LowVolumeSignal {
	if len(klines) < 20 || len(volumes) < 20 {
		return nil
	}

	// 获取最新K线数据
	currentIndex := len(klines) - 1
	currentVol := volumes[currentIndex]
	currentPrice := prices[currentIndex]

	// 计算前20期平均成交量
	var sumVol float64
	for i := currentIndex - 20; i < currentIndex; i++ {
		sumVol += volumes[i]
	}
	avgVol := sumVol / 20

	// 计算价格波动幅度
	priceRange := 0.0
	if currentIndex > 0 {
		high, _ := strconv.ParseFloat(klines[currentIndex].High, 64)
		low, _ := strconv.ParseFloat(klines[currentIndex].Low, 64)
		priceRange = (high - low) / currentPrice * 100
	}

	// 检测条件
	ratio := currentVol / avgVol
	if ratio < config.LowVolumeRatio {
		// 确定重要性级别
		significance := "中等"
		if ratio < 0.15 {
			significance = "极高"
		} else if ratio < 0.25 {
			significance = "高"
		}

		return &LowVolumeSignal{
			SignalType:     "LOW_VOLUME",
			KlineIndex:     currentIndex,
			Volume:         currentVol,
			VolumeRatio:    ratio,
			PriceRange:     priceRange,
			Significance:   significance,
			Interpretation: generateLowVolumeInterpretation(fmt.Sprintf("%.2f%%", priceRange), significance),
		}
	}

	return nil
}

// detectVolumePriceDivergence 检测量价背离信号
func detectVolumePriceDivergence(klines []binance.Kline, volumes, prices []float64, config *VolumeAnalysisConfig) *VolumePriceDivergenceSignal {
	if len(klines) < config.DivergenceLookback || len(volumes) < config.DivergenceLookback {
		return nil
	}

	// 获取最近10根K线的价格和成交量
	lookback := config.DivergenceLookback
	recentPrices := prices[len(prices)-lookback:]
	recentVolumes := volumes[len(volumes)-lookback:]

	// 检测价格新高但成交量萎缩（看跌背离）
	if isPriceNewHigh(recentPrices) && isVolumeContracting(recentVolumes) {
		strength := calculateDivergenceStrength(recentPrices, recentVolumes)
		if strength > config.DivergenceThreshold {
			return &VolumePriceDivergenceSignal{
				SignalType:      "VOLUME_PRICE_DIVERGENCE",
				KlineIndex:      len(klines) - 1,
				PriceDirection:  "UP",
				VolumeDirection: "DOWN",
				DivergenceType:  "BEARISH",
				Strength:        strength,
				Interpretation:  "价格创新高但成交量萎缩，看跌背离，可能预示反转",
			}
		}
	}

	// 检测价格新低但成交量放大（看涨背离）
	if isPriceNewLow(recentPrices) && isVolumeExpanding(recentVolumes) {
		strength := calculateDivergenceStrength(recentPrices, recentVolumes)
		if strength > config.DivergenceThreshold {
			return &VolumePriceDivergenceSignal{
				SignalType:      "VOLUME_PRICE_DIVERGENCE",
				KlineIndex:      len(klines) - 1,
				PriceDirection:  "DOWN",
				VolumeDirection: "UP",
				DivergenceType:  "BULLISH",
				Strength:        strength,
				Interpretation:  "价格创新低但成交量放大，看涨背离，可能预示反弹",
			}
		}
	}

	return nil
}

// detectContinuousVolume 检测连续放量信号
func detectContinuousVolume(klines []binance.Kline, volumes, prices []float64, config *VolumeAnalysisConfig) *ContinuousVolumeSignal {
	if len(volumes) < config.ContinuousMinLength+1 {
		return nil
	}

	minLength := config.ContinuousMinLength
	threshold := config.ContinuousThreshold

	// 检测连续放量
	for i := len(volumes) - minLength; i >= 0; i-- {
		if isContinuousExpanding(volumes[i:i+minLength], threshold) {
			// 计算价格趋势
			priceTrend := calculatePriceTrend(prices[i : i+minLength])

			// 计算平均成交量比率
			avgRatio := calculateAverageVolumeRatio(volumes[i : i+minLength])

			// 计算累积成交量
			var accumulation float64
			for j := i; j < i+minLength; j++ {
				accumulation += volumes[j]
			}

			return &ContinuousVolumeSignal{
				SignalType:     "CONTINUOUS_VOLUME",
				StartIndex:     i,
				EndIndex:       i + minLength - 1,
				Duration:       minLength,
				VolumeTrend:    "EXPANDING",
				AvgVolumeRatio: avgRatio,
				PriceTrend:     priceTrend,
				Accumulation:   accumulation,
				Interpretation: fmt.Sprintf("连续%d根K线成交量递增，%s，可能预示趋势延续", minLength, priceTrend),
			}
		}

		if isContinuousContracting(volumes[i:i+minLength], threshold) {
			// 计算价格趋势
			priceTrend := calculatePriceTrend(prices[i : i+minLength])

			// 计算平均成交量比率
			avgRatio := calculateAverageVolumeRatio(volumes[i : i+minLength])

			// 计算累积成交量
			var accumulation float64
			for j := i; j < i+minLength; j++ {
				accumulation += volumes[j]
			}

			return &ContinuousVolumeSignal{
				SignalType:     "CONTINUOUS_VOLUME",
				StartIndex:     i,
				EndIndex:       i + minLength - 1,
				Duration:       minLength,
				VolumeTrend:    "CONTRACTING",
				AvgVolumeRatio: avgRatio,
				PriceTrend:     priceTrend,
				Accumulation:   accumulation,
				Interpretation: fmt.Sprintf("连续%d根K线成交量递减，%s，可能预示趋势衰竭", minLength, priceTrend),
			}
		}
	}

	return nil
}

// detectVolumeStack 检测成交量堆信号
func detectVolumeStack(klines []binance.Kline, volumes, prices []float64, config *VolumeAnalysisConfig) *VolumeStackSignal {
	if len(volumes) < 5 {
		return nil
	}

	// 检查最近5根K线的成交量集中度
	recent5 := volumes[len(volumes)-5:]
	var stackVolume, totalVolume float64

	for _, vol := range recent5 {
		stackVolume += vol
	}

	// 计算总成交量（最近30根）
	totalStart := len(volumes) - 30
	if totalStart < 0 {
		totalStart = 0
	}
	for i := totalStart; i < len(volumes); i++ {
		totalVolume += volumes[i]
	}

	stackRatio := stackVolume / totalVolume

	// 检测条件：5根K线内成交量占总成交量的30%以上
	if stackRatio > 0.3 {
		// 分析价格行为
		priceAction := analyzePriceAction(prices[len(prices)-5:])

		// 分析成交量分布
		distribution := analyzeVolumeDistribution(recent5)

		return &VolumeStackSignal{
			SignalType:   "VOLUME_STACK",
			StartIndex:   len(volumes) - 5,
			EndIndex:     len(volumes) - 1,
			Duration:     5,
			StackVolume:  stackVolume,
			StackRatio:   stackRatio,
			PriceAction:  priceAction,
			Distribution: distribution,
			Interpretation: fmt.Sprintf("成交量堆信号：5根K线内成交量集中度%.1f%%，%s，%s",
				stackRatio*100, priceAction, distribution),
		}
	}

	return nil
}

// 辅助函数

// generateGiantVolumeInterpretation 生成巨量信号解读
func generateGiantVolumeInterpretation(priceChange, buyRatio float64, significance string) string {
	var action, direction string

	if priceChange > 2 {
		action = "大幅上涨"
	} else if priceChange > 0.5 {
		action = "温和上涨"
	} else if priceChange < -2 {
		action = "大幅下跌"
	} else if priceChange < -0.5 {
		action = "温和下跌"
	} else {
		action = "横盘整理"
	}

	if buyRatio > 60 {
		direction = "主动买入主导"
	} else if buyRatio < 40 {
		direction = "主动卖出主导"
	} else {
		direction = "买卖相对平衡"
	}

	return fmt.Sprintf("%s巨量，%s，%s，%s信号", significance, action, direction,
		getSignalDirection(priceChange, buyRatio))
}

// generateLowVolumeInterpretation 生成地量信号解读
func generateLowVolumeInterpretation(priceRangeStr, significance string) string {
	var marketState string

	// 尝试解析价格范围字符串
	priceRange := 0.0
	if _, err := fmt.Sscanf(priceRangeStr, "%f", &priceRange); err == nil {
		if priceRange < 0.5 {
			marketState = "极度冷清"
		} else if priceRange < 1.0 {
			marketState = "相对冷清"
		} else {
			marketState = "波动中成交萎缩"
		}
	} else {
		marketState = "波动中成交萎缩"
	}

	return fmt.Sprintf("%s地量，%s，可能预示变盘临近", significance, marketState)
}

// getSignalDirection 获取信号方向
func getSignalDirection(priceChange, buyRatio float64) string {
	if priceChange > 1 && buyRatio > 60 {
		return "强烈看多"
	} else if priceChange > 0.5 && buyRatio > 55 {
		return "偏多"
	} else if priceChange < -1 && buyRatio < 40 {
		return "强烈看空"
	} else if priceChange < -0.5 && buyRatio < 45 {
		return "偏空"
	}
	return "中性"
}

// 数学计算辅助函数

func calculateStdDev(values []float64, mean float64) float64 {
	var sumSquaredDiff float64
	for _, v := range values {
		diff := v - mean
		sumSquaredDiff += diff * diff
	}
	return math.Sqrt(sumSquaredDiff / float64(len(values)))
}

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

func maxFloat64(values []float64) float64 {
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func minFloat64(values []float64) float64 {
	min := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// 更多辅助函数实现...
func calculateVolumeTrend(volumes []float64) string {
	if len(volumes) < 3 {
		return "数据不足"
	}

	// 简单的线性回归计算趋势
	n := float64(len(volumes))
	var sumX, sumY, sumXY float64

	for i, v := range volumes {
		x := float64(i)
		sumX += x
		sumY += v
		sumXY += x * v
	}

	slope := (n*sumXY - sumX*sumY) / (n*n - sumX*sumX)

	if slope > 0.1 {
		return "上升趋势"
	} else if slope < -0.1 {
		return "下降趋势"
	}
	return "相对稳定"
}

func calculateVolumeMomentum(volumes []float64) float64 {
	if len(volumes) < 2 {
		return 0
	}

	// 计算最近的成交量变化率
	recent := volumes[len(volumes)-1]
	previous := volumes[len(volumes)-2]

	if previous > 0 {
		return (recent - previous) / previous
	}
	return 0
}

func describeVolumePriceSync(correlation float64) string {
	if correlation > 0.5 {
		return "量价同步上涨"
	} else if correlation < -0.5 {
		return "量价背离"
	} else {
		return "量价关系不明显"
	}
}

func calculateDivergenceStrength(prices, volumes []float64) float64 {
	// 简化的背离强度计算
	priceTrend := calculateSimpleTrend(prices)
	volumeTrend := calculateSimpleTrend(volumes)

	return math.Abs(priceTrend - volumeTrend)
}

func calculateSimpleTrend(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	first := values[0]
	last := values[len(values)-1]

	if first > 0 {
		return (last - first) / first
	}
	return 0
}

func isPriceNewHigh(prices []float64) bool {
	if len(prices) < 3 {
		return false
	}

	current := prices[len(prices)-1]
	for i := 0; i < len(prices)-1; i++ {
		if prices[i] >= current {
			return false
		}
	}
	return true
}

func isPriceNewLow(prices []float64) bool {
	if len(prices) < 3 {
		return false
	}

	current := prices[len(prices)-1]
	for i := 0; i < len(prices)-1; i++ {
		if prices[i] <= current {
			return false
		}
	}
	return true
}

func isVolumeContracting(volumes []float64) bool {
	if len(volumes) < 3 {
		return false
	}

	// 检查成交量是否递减
	for i := 1; i < len(volumes); i++ {
		if volumes[i] >= volumes[i-1]*0.9 {
			return false
		}
	}
	return true
}

func isVolumeExpanding(volumes []float64) bool {
	if len(volumes) < 3 {
		return false
	}

	// 检查成交量是否递增
	for i := 1; i < len(volumes); i++ {
		if volumes[i] <= volumes[i-1]*1.1 {
			return false
		}
	}
	return true
}

func isContinuousExpanding(volumes []float64, threshold float64) bool {
	for i := 1; i < len(volumes); i++ {
		if volumes[i] <= volumes[i-1]*threshold {
			return false
		}
	}
	return true
}

func isContinuousContracting(volumes []float64, threshold float64) bool {
	for i := 1; i < len(volumes); i++ {
		if volumes[i] >= volumes[i-1]/threshold {
			return false
		}
	}
	return true
}

func calculatePriceTrend(prices []float64) string {
	if len(prices) < 2 {
		return "未知"
	}

	first := prices[0]
	last := prices[len(prices)-1]
	change := (last - first) / first * 100

	if change > 1 {
		return "上涨趋势"
	} else if change < -1 {
		return "下跌趋势"
	}
	return "横盘整理"
}

func calculateAverageVolumeRatio(volumes []float64) float64 {
	if len(volumes) < 2 {
		return 1.0
	}

	var sumRatio float64
	for i := 1; i < len(volumes); i++ {
		if volumes[i-1] > 0 {
			sumRatio += volumes[i] / volumes[i-1]
		}
	}

	return sumRatio / float64(len(volumes)-1)
}

func analyzePriceAction(prices []float64) string {
	if len(prices) < 2 {
		return "数据不足"
	}

	first := prices[0]
	last := prices[len(prices)-1]
	change := (last - first) / first * 100

	if change > 2 {
		return "大幅上涨"
	} else if change > 0.5 {
		return "温和上涨"
	} else if change < -2 {
		return "大幅下跌"
	} else if change < -0.5 {
		return "温和下跌"
	}
	return "横盘整理"
}

func analyzeVolumeDistribution(volumes []float64) string {
	if len(volumes) < 3 {
		return "数据不足"
	}

	// 检查成交量分布是否均匀
	maxVol := maxFloat64(volumes)
	minVol := minFloat64(volumes)

	if maxVol/minVol > 3 {
		return "分布不均"
	} else if maxVol/minVol > 2 {
		return "分布较散"
	}
	return "分布均匀"
}

func analyzeCrossWindowComparison(analysis *TimeLayerVolumeAnalysis) *CrossWindowAnalysis {
	comparison := &CrossWindowAnalysis{
		VolumeAcceleration:   0,
		TrendConsistency:     0,
		MultiTimeframeSignal: "无明确信号",
	}

	// 计算成交量加速度（微观vs短期）
	if analysis.MicroWindow != nil && analysis.ShortWindow != nil {
		comparison.VolumeAcceleration = analysis.MicroWindow.AvgVolume / analysis.ShortWindow.AvgVolume
	}

	// 计算趋势一致性
	trends := []string{}
	if analysis.MicroWindow != nil {
		trends = append(trends, analysis.MicroWindow.VolumeTrend)
	}
	if analysis.ShortWindow != nil {
		trends = append(trends, analysis.ShortWindow.VolumeTrend)
	}
	if analysis.MediumWindow != nil {
		trends = append(trends, analysis.MediumWindow.VolumeTrend)
	}

	// 计算一致性
	consistent := 0
	for i := 1; i < len(trends); i++ {
		if trends[i] == trends[0] {
			consistent++
		}
	}

	if len(trends) > 1 {
		comparison.TrendConsistency = float64(consistent) / float64(len(trends)-1)
	}

	// 生成多时间框架信号
	if comparison.TrendConsistency > 0.7 {
		if trends[0] == "上升趋势" {
			comparison.MultiTimeframeSignal = "多时间框架共振看多"
		} else if trends[0] == "下降趋势" {
			comparison.MultiTimeframeSignal = "多时间框架共振看空"
		}
	}

	return comparison
}

// FormatVolumeAnalysisForLLM 格式化成交量分析为LLM友好的报告
func FormatVolumeAnalysisForLLM(analysis *TimeLayerVolumeAnalysis) string {
	var report strings.Builder

	report.WriteString("📊 多时间框架成交量分析:\n\n")

	// 微观窗口分析
	if analysis.MicroWindow != nil {
		report.WriteString("🔍 微观窗口（15分钟）:\n")
		report.WriteString(formatWindowAnalysis(analysis.MicroWindow))
		report.WriteString(formatSignals(analysis.MicroWindow.Signals))
	}

	// 短期窗口分析
	if analysis.ShortWindow != nil {
		report.WriteString("\n📈 短期窗口（45分钟）:\n")
		report.WriteString(formatWindowAnalysis(analysis.ShortWindow))
		report.WriteString(formatSignals(analysis.ShortWindow.Signals))
	}

	// 中期窗口分析
	if analysis.MediumWindow != nil {
		report.WriteString("\n📊 中期窗口（90分钟）:\n")
		report.WriteString(formatWindowAnalysis(analysis.MediumWindow))
		report.WriteString(formatSignals(analysis.MediumWindow.Signals))
	}

	// 长期窗口分析
	if analysis.LongWindow != nil {
		report.WriteString("\n🌐 长期窗口（3.5小时）:\n")
		report.WriteString(formatWindowAnalysis(analysis.LongWindow))
		report.WriteString(formatSignals(analysis.LongWindow.Signals))
	}

	// 跨窗口比较
	if analysis.CrossWindowComparison != nil {
		report.WriteString("\n🔄 跨窗口比较:\n")
		report.WriteString(fmt.Sprintf("  成交量加速度: %.2f\n", analysis.CrossWindowComparison.VolumeAcceleration))
		report.WriteString(fmt.Sprintf("  趋势一致性: %.1f%%\n", analysis.CrossWindowComparison.TrendConsistency*100))
		report.WriteString(fmt.Sprintf("  多时间框架信号: %s\n", analysis.CrossWindowComparison.MultiTimeframeSignal))
	}

	return report.String()
}

func formatWindowAnalysis(window *VolumeWindowAnalysis) string {
	var report strings.Builder

	report.WriteString(fmt.Sprintf("  基础统计: 平均%.2f ETH, 区间%.2f-%.2f ETH, 标准差%.2f\n",
		window.AvgVolume, window.MinVolume, window.MaxVolume, window.VolumeStdDev))
	report.WriteString(fmt.Sprintf("  成交量趋势: %s, 动量: %.3f\n", window.VolumeTrend, window.VolumeMomentum))
	report.WriteString(fmt.Sprintf("  买卖压力: 买入%.1f%% vs 卖出%.1f%%, 净流向: %.2f ETH\n",
		window.BuyRatio, window.SellRatio, window.NetFlowVolume))
	report.WriteString(fmt.Sprintf("  量价关系: 相关性%.3f, %s\n",
		window.PriceVolumeCorrelation, window.VolumePriceSync))

	return report.String()
}

func formatSignals(signals []VolumeSignal) string {
	if len(signals) == 0 {
		return "  📢 信号: 无明显信号\n"
	}

	var report strings.Builder
	report.WriteString("  📢 关键信号:\n")

	for _, signal := range signals {
		report.WriteString(fmt.Sprintf("    • %s: %s\n", signal.GetSignalType(), signal.GetInterpretation()))
	}

	return report.String()
}
