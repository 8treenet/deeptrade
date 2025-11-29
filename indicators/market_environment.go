package indicators

import (
	"math"
)

// MarketEnvironment 市场环境类型
type MarketEnvironment int

const (
	MarketEnvironmentUnknown  MarketEnvironment = iota
	MarketEnvironmentBullish                    // 牛市/上涨趋势
	MarketEnvironmentBearish                    // 熊市/下跌趋势
	MarketEnvironmentSideways                   // 横盘/震荡
)

// MarketCondition 市场条件
type MarketCondition struct {
	Environment   MarketEnvironment
	Volatility    float64 // 波动率 (0-1)
	TrendStrength float64 // 趋势强度 (0-1)
}

// AnalyzeMarketEnvironment 分析市场环境
// 使用多个指标综合判断：价格相对位置 + ATR波动率 + 趋势线
func AnalyzeMarketEnvironment(prices []float64, sma20, sma50 float64, atr float64) *MarketCondition {
	if len(prices) < 2 {
		return &MarketCondition{
			Environment:   MarketEnvironmentUnknown,
			Volatility:    0,
			TrendStrength: 0,
		}
	}

	latestPrice := prices[len(prices)-1]
	prevPrice := prices[len(prices)-2]
	priceChange := (latestPrice - prevPrice) / prevPrice

	// 计算波动率 (ATR / 价格)
	volatility := 0.0
	if latestPrice > 0 {
		volatility = atr / latestPrice
	}

	// 分析趋势
	var env MarketEnvironment
	trendStrength := 0.0

	// 基于价格和移动平均线判断
	if sma20 > 0 && sma50 > 0 {
		if latestPrice > sma20 && sma20 > sma50 && priceChange > 0 {
			env = MarketEnvironmentBullish
			trendStrength = math.Min(1.0, math.Abs(priceChange)*10+0.5) // 价格上涨 + MA排列
		} else if latestPrice < sma20 && sma20 < sma50 && priceChange < 0 {
			env = MarketEnvironmentBearish
			trendStrength = math.Min(1.0, math.Abs(priceChange)*10+0.5) // 价格下跌 + MA排列
		} else {
			env = MarketEnvironmentSideways
			trendStrength = 0.3 // 横盘趋势较弱
		}
	} else {
		// 简单判断：基于价格变化
		if priceChange > 0.01 { // 涨幅超过1%
			env = MarketEnvironmentBullish
			trendStrength = math.Min(1.0, math.Abs(priceChange)*5)
		} else if priceChange < -0.01 { // 跌幅超过1%
			env = MarketEnvironmentBearish
			trendStrength = math.Min(1.0, math.Abs(priceChange)*5)
		} else {
			env = MarketEnvironmentSideways
			trendStrength = 0.2
		}
	}

	// 波动率调整
	if volatility > 0.05 { // 5%以上为高波动
		trendStrength *= 0.8 // 高波动时降低趋势强度
	}

	return &MarketCondition{
		Environment:   env,
		Volatility:    volatility,
		TrendStrength: trendStrength,
	}
}

