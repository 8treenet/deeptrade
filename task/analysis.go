package task

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"deeptrade/utils"

	"github.com/cloudwego/eino/schema"
)

// AnalyzeWithLLM 使用LLM分析市场数据
func AnalyzeWithLLM(marketData *MarketData) (*TradingSignal, error) {
	log.Println("[LLM分析] 开始调用LLM分析...")

	// 准备技术分析数据
	technicalData := PrepareTechnicalData(marketData)
	balanceInfo := GetAccountBalanceInfo(marketData)

	// 构建各种分析
	technicalAnalysis := FormatTechnicalIndicators(technicalData, marketData)
	volumeAnalysis := FormatVolumeAnalysis(marketData.Klines3m, technicalData.CurrentPrice)
	fundingAnalysis := FormatFundingAnalysis(marketData)
	tradeFlowAnalysis := FormatTradeFlowAnalysis(marketData, technicalData.CurrentPrice)
	bookTickerAnalysis := FormatBookTickerData(marketData)

	positionAnalysis := FormatPositionWithSLTP(marketData.Positions, marketData.OpenOrders)

	// 构建用户消息
	currentPrice := marketData.Ticker.LastPrice
	priceChange := marketData.Ticker.PriceChangePercent

	// 重新解析价格为浮点数，以便格式化（前面声明的可能基于不同的数据源）
	var currentPriceFloat float64
	if cp, err := strconv.ParseFloat(currentPrice, 64); err == nil {
		currentPriceFloat = cp
	}

	// 解析价格变化为浮点数，以便格式化
	var priceChangeFloat float64
	if pc, err := strconv.ParseFloat(priceChange, 64); err == nil {
		priceChangeFloat = pc
	}

	// 添加当前时间信息
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// 直接使用MarketData中已有的历史订单数据，避免重复API调用
	tradeRecords := GetTradeRecordsFromMarketData(marketData, 6)
	tradeRecordsAnalysis := FormatTradeRecords(tradeRecords)
	log.Printf("最近订单记录: \n%v\n", tradeRecordsAnalysis)

	// 构建用户消息 - 专注于数据呈现和决策触发
	userMsg := fmt.Sprintf(`[当前时间: %s]
📊 完整市场数据
## 基础信息
价格: %.2f (%.2f%%) | 持仓: %s
💰 账户余额详情:
	 • 钱包余额: %.2f USDT
	 • 可用余额: %.2f USDT (可用于开仓)
	 • 保证金余额: %.2f USDT

## 技术指标
%s

## 成交量趋势分析
%s

## 专业交易流分析
%s

## memory
%s

## 资金状况
%s

## 最优挂单信息
%s

## 原始订单簿数据
%s
`,
		currentTime,
		currentPriceFloat, priceChangeFloat, positionAnalysis,
		balanceInfo.WalletBalance, balanceInfo.AvailableBalance,
		balanceInfo.MarginBalance,
		technicalAnalysis,
		volumeAnalysis,
		tradeFlowAnalysis,
		GetMemory(),
		fundingAnalysis,
		bookTickerAnalysis,
		FormatRawOrderBookData(marketData),
	)

	// 创建消息
	message := &schema.Message{
		Role:    schema.User,
		Content: userMsg,
	}
	log.Println("userMsg ", userMsg)
	// 调用LLM
	response, err := utils.Run(marketData.PositionInfo.HasLong || marketData.PositionInfo.HasShort, message)
	if err != nil {
		log.Printf("[LLM分析] LLM调用失败: %v", err)
		return nil, err
	}

	// 解析JSON响应
	log.Printf("[LLM分析] LLM原始响应: %s", response)
	signal, err := ParseLLMResponse(response)
	if err != nil {
		log.Printf("[LLM分析] 解析LLM响应失败: %v", err)
		return nil, err
	}
	log.Printf("[LLM分析] %s (评分: %d, 置信度: %.2f%%)", signal.Action, signal.Score, signal.Confidence*100)
	return signal, nil
}

// ParseLLMResponse 解析LLM响应（需要修改llm.go使用）
func ParseLLMResponse(response string) (*TradingSignal, error) {
	strings.ReplaceAll(response, "0.,", "0.0,")
	var signal TradingSignal
	lines := strings.Split(response, "\n")
	var jsonStr string
	inJson := false

	for _, line := range lines {
		if strings.Contains(line, "{") {
			inJson = true
		}
		if inJson {
			jsonStr += line + "\n"
		}
		if strings.Contains(line, "}") {
			break
		}
	}

	if err := json.Unmarshal([]byte(jsonStr), &signal); err != nil {
		return nil, fmt.Errorf("解析LLM响应失败 response: %v  err: %v", response, err)
	}

	return &signal, nil
}
