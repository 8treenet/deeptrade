package task

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"deeptrade/indicators"
)

// FormatTradeFlowAnalysis 格式化交易流分析
func FormatTradeFlowAnalysis(marketData *MarketData, currentPriceFloat float64) string {
	var analysis strings.Builder

	// 使用新的专业交易流分析系统
	tradeFlowConfig := indicators.DefaultTradeFlowConfig()
	tradeFlowAnalysis := indicators.AnalyzeTradeFlow(currentPriceFloat, tradeFlowConfig)
	// 增强版交易流分析
	if tradeFlowAnalysis.TotalTrades > 0 {
		// 生成专业报告（格式化为LLM友好的简洁版本）
		analysis.WriteString(indicators.GenerateTradeFlowReport(tradeFlowAnalysis))
	} else {
		analysis.WriteString("交易流分析: 无近期交易数据\n")
	}

	return analysis.String()
}

// FormatRawOrderBookData 格式化原始订单簿数据供LLM分析
func FormatRawOrderBookData(marketData *MarketData) string {
	var analysis strings.Builder

	if marketData.OrderBook != nil && len(marketData.OrderBook.Asks) > 0 && len(marketData.OrderBook.Bids) > 0 {
		analysis.WriteString("### 原始订单簿数据 (前10档)\n")

		analysis.WriteString("**买单 (Bids)**\n")
		for i := 0; i < len(marketData.OrderBook.Bids) && i < 10; i++ {
			price, _ := strconv.ParseFloat(marketData.OrderBook.Bids[i].Price, 64)
			quantity, _ := strconv.ParseFloat(marketData.OrderBook.Bids[i].Quantity, 64)
			analysis.WriteString(fmt.Sprintf("  买%d: $%.2f (%.2f ETH)\n", i+1, price, quantity))
		}

		analysis.WriteString("\n**卖单 (Asks)**\n")
		for i := 0; i < len(marketData.OrderBook.Asks) && i < 10; i++ {
			price, _ := strconv.ParseFloat(marketData.OrderBook.Asks[i].Price, 64)
			quantity, _ := strconv.ParseFloat(marketData.OrderBook.Asks[i].Quantity, 64)
			analysis.WriteString(fmt.Sprintf("  卖%d: $%.2f (%.2f ETH)\n", i+1, price, quantity))
		}

		// 基础统计
		var totalBidVolume, totalAskVolume float64
		for i := 0; i < len(marketData.OrderBook.Bids) && i < 20; i++ {
			quantity, _ := strconv.ParseFloat(marketData.OrderBook.Bids[i].Quantity, 64)
			totalBidVolume += quantity
		}
		for i := 0; i < len(marketData.OrderBook.Asks) && i < 20; i++ {
			quantity, _ := strconv.ParseFloat(marketData.OrderBook.Asks[i].Quantity, 64)
			totalAskVolume += quantity
		}

		bestBid, _ := strconv.ParseFloat(marketData.OrderBook.Bids[0].Price, 64)
		bestAsk, _ := strconv.ParseFloat(marketData.OrderBook.Asks[0].Price, 64)
		spread := bestAsk - bestBid
		spreadPercent := spread / bestAsk * 100

		analysis.WriteString(fmt.Sprintf("\n**基础统计**\n"))
		analysis.WriteString(fmt.Sprintf("  买一价: $%.2f | 卖一价: $%.2f\n", bestBid, bestAsk))
		analysis.WriteString(fmt.Sprintf("  价差: $%.2f (%.3f%%)\n", spread, spreadPercent))
		analysis.WriteString(fmt.Sprintf("  买量(前20): %.2f ETH | 卖量(前20): %.2f ETH\n", totalBidVolume, totalAskVolume))
	} else {
		analysis.WriteString("订单簿数据: 暂无\n")
	}

	return analysis.String()
}

// FormatBookTickerData 格式化最优挂单数据
func FormatBookTickerData(marketData *MarketData) string {
	var analysis strings.Builder

	if marketData.BookTicker != nil {
		analysis.WriteString("### 最优挂单信息\n")

		// 解析价格和数量
		bidPrice, _ := strconv.ParseFloat(marketData.BookTicker.BidPrice, 64)
		bidQty, _ := strconv.ParseFloat(marketData.BookTicker.BidQty, 64)
		askPrice, _ := strconv.ParseFloat(marketData.BookTicker.AskPrice, 64)
		askQty, _ := strconv.ParseFloat(marketData.BookTicker.AskQty, 64)

		// 计算价差和价差百分比
		spread := askPrice - bidPrice
		spreadPercent := spread / askPrice * 100

		// 计算总挂单量
		totalVolume := bidQty + askQty
		bidRatio := bidQty / totalVolume * 100
		askRatio := askQty / totalVolume * 100

		analysis.WriteString(fmt.Sprintf("  最优买价: $%.2f (挂单量: %.2f ETH, 占比: %.1f%%)\n", bidPrice, bidQty, bidRatio))
		analysis.WriteString(fmt.Sprintf("  最优卖价: $%.2f (挂单量: %.2f ETH, 占比: %.1f%%)\n", askPrice, askQty, askRatio))
		analysis.WriteString(fmt.Sprintf("  价差: $%.4f (%.3f%%)\n", spread, spreadPercent))

		// 添加市场深度分析
		if bidQty > 0 && askQty > 0 {
			bidAskRatio := bidQty / askQty
			analysis.WriteString(fmt.Sprintf("  买卖量比: %.2f (买方%s)\n", bidAskRatio,
				func() string {
					if bidAskRatio > 1.2 {
						return "占优"
					} else if bidAskRatio < 0.8 {
						return "弱势"
					} else {
						return "平衡"
					}
				}()))
		}

		// 添加时间戳信息
		if marketData.BookTicker.Time > 0 {
			// 将时间戳转换为可读格式
			updateTime := time.Unix(0, marketData.BookTicker.Time*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000")
			analysis.WriteString(fmt.Sprintf("  更新时间: %s\n", updateTime))
		}

		// 添加流动性评估
		liquidityScore := totalVolume / spread // 简单的流动性评分：总量/价差
		var liquidityLevel string
		switch {
		case liquidityScore > 1000:
			liquidityLevel = "极高"
		case liquidityScore > 500:
			liquidityLevel = "高"
		case liquidityScore > 200:
			liquidityLevel = "中等"
		case liquidityScore > 50:
			liquidityLevel = "低"
		default:
			liquidityLevel = "极低"
		}
		analysis.WriteString(fmt.Sprintf("  流动性评估: %s (评分: %.1f)\n", liquidityLevel, liquidityScore))

	} else {
		analysis.WriteString("最优挂单信息: 暂无数据\n")
	}

	return analysis.String()
}
