# 成交量分析设计方案

## 概述

本文档详细描述了如何将K线中的成交量相关指标整合到深度交易系统中，以便LLM能够更好地感知成交量趋势和做出更准确的交易决策。

## 1. 分层时间设计

### 1.1 时间窗口分层

基于3分钟K线数据，我们设计以下时间窗口分层：

| 时间窗口 | K线数量 | 覆盖时间 | 用途 |
|---------|---------|---------|------|
| 微观窗口 | 5条 | 15分钟 | 短期情绪和即时信号 |
| 短期窗口 | 15条 | 45分钟 | 短期趋势确认 |
| 中期窗口 | 30条 | 90分钟 | 中期趋势分析 |
| 长期窗口 | 70条 | 210分钟(3.5小时) | 长期趋势背景 |

### 1.2 分层数据结构

```go
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

type VolumeWindowAnalysis struct {
    TimeWindow     string  `json:"time_window"`     // 时间窗口描述
    KlineCount     int     `json:"kline_count"`     // K线数量
    
    // 基础成交量统计
    AvgVolume      float64 `json:"avg_volume"`       // 平均成交量
    MaxVolume      float64 `json:"max_volume"`       // 最大成交量
    MinVolume      float64 `json:"min_volume"`       // 最小成交量
    VolumeStdDev   float64 `json:"volume_std_dev"`   // 成交量标准差
    VolumeRange    float64 `json:"volume_range"`      // 成交量区间
    
    // 成交量趋势
    VolumeTrend    string  `json:"volume_trend"`     // 成交量趋势
    VolumeMomentum float64 `json:"volume_momentum"`  // 成交量动量
    
    // 买卖压力分析
    BuyRatio       float64 `json:"buy_ratio"`         // 主动买入比例
    SellRatio      float64 `json:"sell_ratio"`        // 主动卖出比例
    NetFlowVolume  float64 `json:"net_flow_volume"`   // 净流入成交量
    
    // 量价关系
    PriceVolumeCorrelation float64 `json:"price_volume_correlation"` // 量价相关性
    VolumePriceSync       string  `json:"volume_price_sync"`        // 量价同步性描述
    
    // 关键信号
    Signals []VolumeSignal `json:"signals"` // 检测到的信号
}

type CrossWindowAnalysis struct {
    VolumeAcceleration  float64 `json:"volume_acceleration"`  // 成交量加速度
    TrendConsistency   float64 `json:"trend_consistency"`    // 趋势一致性
    MultiTimeframeSignal string `json:"multi_timeframe_signal"` // 多时间框架信号
}
```

## 2. 成交量信号设计

### 2.1 单根K线信号

#### 2.1.1 巨量信号

```go
type GiantVolumeSignal struct {
    SignalType    string  `json:"signal_type"`     // "GIANT_VOLUME"
    KlineIndex    int     `json:"kline_index"`     // K线索引
    Volume        float64 `json:"volume"`          // 成交量
    VolumeRatio   float64 `json:"volume_ratio"`    // 与平均成交量的比率
    PriceChange   float64 `json:"price_change"`    // 价格变化百分比
    BuyRatio      float64 `json:"buy_ratio"`       // 主动买入比例
    Significance  string  `json:"significance"`    // 重要性级别
    Interpretation string  `json:"interpretation"`  // 信号解读
}
```

**检测条件：**
- 当前成交量 > 20期平均成交量的2.5倍
- 或者成交量 > 20期最大成交量的1.8倍
- 或者成交量在最近70条中排名前5%

#### 2.1.2 地量信号

```go
type LowVolumeSignal struct {
    SignalType    string  `json:"signal_type"`     // "LOW_VOLUME"
    KlineIndex    int     `json:"kline_index"`     // K线索引
    Volume        float64 `json:"volume"`          // 成交量
    VolumeRatio   float64 `json:"volume_ratio"`    // 与平均成交量的比率
    PriceRange    float64 `json:"price_range"`     // 价格波动幅度
    Significance  string  `json:"significance"`    // 重要性级别
    Interpretation string  `json:"interpretation"`  // 信号解读
}
```

**检测条件：**
- 当前成交量 < 20期平均成交量的0.3倍
- 或者成交量在最近70条中排名后10%

#### 2.1.3 量价背离信号

```go
type VolumePriceDivergenceSignal struct {
    SignalType       string  `json:"signal_type"`        // "VOLUME_PRICE_DIVERGENCE"
    KlineIndex       int     `json:"kline_index"`        // K线索引
    PriceDirection   string  `json:"price_direction"`    // "UP"或"DOWN"
    VolumeDirection  string  `json:"volume_direction"`   // "UP"或"DOWN"
    DivergenceType   string  `json:"divergence_type"`    // "BULLISH"或"BEARISH"
    Strength        float64 `json:"strength"`           // 背离强度
    Interpretation  string  `json:"interpretation"`     // 信号解读
}
```

**检测条件：**
- 价格创新高但成交量萎缩（看跌背离）
- 价格创新低但成交量放大（看涨背离）

### 2.2 多根K线信号

#### 2.2.1 连续放量信号

```go
type ContinuousVolumeSignal struct {
    SignalType       string  `json:"signal_type"`         // "CONTINUOUS_VOLUME"
    StartIndex       int     `json:"start_index"`         // 起始索引
    EndIndex         int     `json:"end_index"`           // 结束索引
    Duration        int     `json:"duration"`            // 持续时间（根K线）
    VolumeTrend     string  `json:"volume_trend"`        // "EXPANDING"或"CONTRACTING"
    AvgVolumeRatio  float64 `json:"avg_volume_ratio"`     // 平均成交量比率
    PriceTrend      string  `json:"price_trend"`         // 价格趋势
    Accumulation    float64 `json:"accumulation"`        // 累积成交量
    Interpretation   string  `json:"interpretation"`      // 信号解读
}
```

**检测条件：**
- 连续3根以上K线成交量递增
- 或者连续3根以上K线成交量递减

#### 2.2.2 成交量堆信号

```go
type VolumeStackSignal struct {
    SignalType       string  `json:"signal_type"`         // "VOLUME_STACK"
    StartIndex       int     `json:"start_index"`         // 起始索引
    EndIndex         int     `json:"end_index"`           // 结束索引
    Duration        int     `json:"duration"`            // 持续时间
    StackVolume     float64 `json:"stack_volume"`        // 堆积成交量
    StackRatio      float64 `json:"stack_ratio"`         // 占总成交量比例
    PriceAction     string  `json:"price_action"`        // 价格行为
    Distribution    string  `json:"distribution"`        // 成交量分布
    Interpretation   string  `json:"interpretation"`      // 信号解读
}
```

**检测条件：**
- 短时间内成交量异常集中
- 5根K线内成交量占总成交量的30%以上

#### 2.2.3 吸筹派发信号

```go
type AccumulationDistributionSignal struct {
    SignalType         string  `json:"signal_type"`           // "ACCUMULATION_DISTRIBUTION"
    StartIndex         int     `json:"start_index"`           // 起始索引
    EndIndex           int     `json:"end_index"`             // 结束索引
    Duration          int     `json:"duration"`              // 持续时间
    Pattern           string  `json:"pattern"`               // "ACCUMULATION"或"DISTRIBUTION"
    VolumeProfile     float64 `json:"volume_profile"`        // 成交量分布特征
    PriceRange        float64 `json:"price_range"`           // 价格区间
    BuySellImbalance  float64 `json:"buy_sell_imbalance"`    // 买卖不平衡度
    Interpretation     string  `json:"interpretation"`        // 信号解读
}
```

**检测条件：**
- 价格区间内成交量放大但价格波动小（吸筹）
- 价格下跌时成交量放大，上涨时成交量萎缩（派发）

### 2.3 时间框架信号

#### 2.3.1 多时间框架共振信号

```go
type MultiTimeframeSignal struct {
    SignalType       string                 `json:"signal_type"`        // "MULTI_TIMEFRAME"
    Consensus        string                 `json:"consensus"`         // 共识方向
    Confidence       float64                `json:"confidence"`        // 置信度
    TimeframeSignals map[string]string      `json:"timeframe_signals"` // 各时间框架信号
    Strength         float64                `json:"strength"`          // 信号强度
    Interpretation   string                 `json:"interpretation"`    // 信号解读
}
```

#### 2.3.2 时间框架背离信号

```go
type TimeframeDivergenceSignal struct {
    SignalType         string                 `json:"signal_type"`          // "TIMEFRAME_DIVERGENCE"
    ShortTimeframe     string                 `json:"short_timeframe"`      // 短时间框架
    LongTimeframe      string                 `json:"long_timeframe"`       // 长时间框架
    ShortSignal        string                 `json:"short_signal"`         // 短期信号
    LongSignal         string                 `json:"long_signal"`          // 长期信号
    DivergenceType     string                 `json:"divergence_type"`      // 背离类型
    Significance       float64                `json:"significance"`         // 重要性
    Interpretation     string                 `json:"interpretation"`       // 信号解读
}
```

## 3. 信号检测算法

### 3.1 巨量检测算法

```go
func DetectGiantVolume(klines []binance.Kline, index int) *GiantVolumeSignal {
    if index < 20 || index >= len(klines) {
        return nil
    }
    
    currentVol, _ := strconv.ParseFloat(klines[index].Volume, 64)
    
    // 计算前20期平均成交量
    var sumVol float64
    var maxVol float64
    for i := index - 20; i < index; i++ {
        vol, _ := strconv.ParseFloat(klines[i].Volume, 64)
        sumVol += vol
        if vol > maxVol {
            maxVol = vol
        }
    }
    avgVol := sumVol / 20
    
    // 计算最近70条中的排名
    allVolumes := make([]float64, 70)
    for i := 0; i < 70 && index-i >= 0; i++ {
        vol, _ := strconv.ParseFloat(klines[index-i].Volume, 64)
        allVolumes[i] = vol
    }
    
    rank := calculatePercentile(currentVol, allVolumes)
    
    // 检测条件
    ratio := currentVol / avgVol
    if ratio > 2.5 || currentVol > maxVol*1.8 || rank > 95 {
        // 计算价格变化和买入比例
        prevClose, _ := strconv.ParseFloat(klines[index-1].Close, 64)
        currClose, _ := strconv.ParseFloat(klines[index].Close, 64)
        priceChange := (currClose - prevClose) / prevClose * 100
        
        buyVol, _ := strconv.ParseFloat(klines[index].TakerBuyBaseAssetVolume, 64)
        buyRatio := buyVol / currentVol * 100
        
        significance := "中等"
        if ratio > 4 || rank > 98 {
            significance = "极高"
        } else if ratio > 3 || rank > 96 {
            significance = "高"
        }
        
        return &GiantVolumeSignal{
            SignalType:    "GIANT_VOLUME",
            KlineIndex:    index,
            Volume:        currentVol,
            VolumeRatio:   ratio,
            PriceChange:   priceChange,
            BuyRatio:      buyRatio,
            Significance:  significance,
            Interpretation: generateGiantVolumeInterpretation(priceChange, buyRatio, significance),
        }
    }
    
    return nil
}
```

### 3.2 量价背离检测算法

```go
func DetectVolumePriceDivergence(klines []binance.Kline, index int) *VolumePriceDivergenceSignal {
    if index < 10 || index >= len(klines) {
        return nil
    }
    
    // 获取最近10根K线的价格和成交量
    prices := make([]float64, 10)
    volumes := make([]float64, 10)
    
    for i := 0; i < 10; i++ {
        price, _ := strconv.ParseFloat(klines[index-i].Close, 64)
        volume, _ := strconv.ParseFloat(klines[index-i].Volume, 64)
        prices[i] = price
        volumes[i] = volume
    }
    
    // 检测价格新高但成交量萎缩
    if prices[0] > prices[1] && prices[1] > prices[2] {
        if volumes[0] < volumes[1]*0.8 && volumes[1] < volumes[2]*0.8 {
            return &VolumePriceDivergenceSignal{
                SignalType:      "VOLUME_PRICE_DIVERGENCE",
                KlineIndex:      index,
                PriceDirection:  "UP",
                VolumeDirection: "DOWN",
                DivergenceType:  "BEARISH",
                Strength:       calculateDivergenceStrength(prices, volumes),
                Interpretation:  "价格创新高但成交量萎缩，看跌背离，可能预示反转",
            }
        }
    }
    
    // 检测价格新低但成交量放大
    if prices[0] < prices[1] && prices[1] < prices[2] {
        if volumes[0] > volumes[1]*1.2 && volumes[1] > volumes[2]*1.2 {
            return &VolumePriceDivergenceSignal{
                SignalType:      "VOLUME_PRICE_DIVERGENCE",
                KlineIndex:      index,
                PriceDirection:  "DOWN",
                VolumeDirection: "UP",
                DivergenceType:  "BULLISH",
                Strength:       calculateDivergenceStrength(prices, volumes),
                Interpretation:  "价格创新低但成交量放大，看涨背离，可能预示反弹",
            }
        }
    }
    
    return nil
}
```

## 4. LLM集成方案

### 4.1 成交量报告格式

```go
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
```

### 4.2 信号优先级排序

```go
type SignalPriority struct {
    Signal      interface{}
    Priority    int
    Timeframe   string
    Significance float64
}

func PrioritizeSignals(analysis *TimeLayerVolumeAnalysis) []SignalPriority {
    var priorities []SignalPriority
    
    // 按时间框架和信号类型分配优先级
    addSignals := func(window *VolumeWindowAnalysis, timeframe string, basePriority int) {
        for _, signal := range window.Signals {
            var priority int
            var significance float64
            
            switch s := signal.(type) {
            case *GiantVolumeSignal:
                priority = basePriority + 10 // 巨量信号高优先级
                significance = s.VolumeRatio
            case *VolumePriceDivergenceSignal:
                priority = basePriority + 8 // 背离信号高优先级
                significance = s.Strength
            case *ContinuousVolumeSignal:
                priority = basePriority + 6 // 连续放量中等优先级
                significance = s.AvgVolumeRatio
            default:
                priority = basePriority
                significance = 1.0
            }
            
            priorities = append(priorities, SignalPriority{
                Signal:      signal,
                Priority:    priority,
                Timeframe:   timeframe,
                Significance: significance,
            })
        }
    }
    
    addSignals(analysis.MicroWindow, "微观", 30)
    addSignals(analysis.ShortWindow, "短期", 20)
    addSignals(analysis.MediumWindow, "中期", 10)
    addSignals(analysis.LongWindow, "长期", 5)
    
    // 按优先级排序
    sort.Slice(priorities, func(i, j int) bool {
        if priorities[i].Priority != priorities[j].Priority {
            return priorities[i].Priority > priorities[j].Priority
        }
        return priorities[i].Significance > priorities[j].Significance
    })
    
    return priorities
}
```

## 5. 实施步骤

### 5.1 第一阶段：基础架构
1. 创建成交量分析数据结构
2. 实现基础的时间窗口分析
3. 集成到现有的技术分析流程

### 5.2 第二阶段：信号检测
1. 实现单根K线信号检测
2. 实现多根K线信号检测
3. 添加信号优先级排序

### 5.3 第三阶段：LLM集成
1. 格式化成交量报告
2. 优化LLM提示词
3. 测试和验证效果

### 5.4 第四阶段：优化完善
1. 根据实际交易效果调整参数
2. 添加更多信号类型
3. 优化性能和稳定性

## 6. 配置参数

```go
type VolumeAnalysisConfig struct {
    // 时间窗口配置
    MicroWindowKlines   int     `json:"micro_window_klines"`    // 微观窗口K线数量
    ShortWindowKlines   int     `json:"short_window_klines"`    // 短期窗口K线数量
    MediumWindowKlines  int     `json:"medium_window_klines"`   // 中期窗口K线数量
    LongWindowKlines    int     `json:"long_window_klines"`     // 长期窗口K线数量
    
    // 巨量检测参数
    GiantVolumeRatio    float64 `json:"giant_volume_ratio"`     // 巨量比率阈值
    GiantVolumePercentile float64 `json:"giant_volume_percentile"` // 巨量百分位阈值
    
    // 地量检测参数
    LowVolumeRatio      float64 `json:"low_volume_ratio"`       // 地量比率阈值
    LowVolumePercentile float64 `json:"low_volume_percentile"`  // 地量百分位阈值
    
    // 背离检测参数
    DivergenceLookback  int     `json:"divergence_lookback"`    // 背离检测回看期
    DivergenceThreshold float64 `json:"divergence_threshold"`   // 背离检测阈值
    
    // 连续放量参数
    ContinuousMinLength int     `json:"continuous_min_length"`  // 连续最短长度
    ContinuousThreshold float64 `json:"continuous_threshold"`    // 连续变化阈值
}
```

## 7. 总结

本设计方案通过多层次的时间窗口分析和丰富的信号检测，将原始的K线成交量数据转换为LLM易于理解的结构化信息。这种方法不仅保留了原始数据的完整性，还提供了有价值的趋势分析和交易信号，有助于提高LLM的交易决策准确性。

通过分层时间设计，系统能够在不同时间尺度上捕捉市场情绪和资金流向的变化；通过多样化的信号设计，系统能够识别出各种重要的市场行为模式。这种综合性的分析方法将为深度交易系统提供更强大的市场洞察力。