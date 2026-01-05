package task

import (
	"deeptrade/binance"
	"deeptrade/conf"
	"deeptrade/indicators"
	"deeptrade/utils"
	"log"
	"strconv"
	"sync"
	"time"
)

var (
	offSystem      bool //可关闭交易系统，用于固定时间关闭交易，但需要判断持仓
	sleepSec       int  //动态定时器睡眠的秒数
	offSystemMutex sync.Mutex
)

func setOffSystem(ps *PositionInfo) {
	offSystemMutex.Lock()
	defer offSystemMutex.Unlock()

	if ps.HasLong || ps.HasShort {
		offSystem = false
		return
	}
	offSystem = true
}

func GetOffSystem() (vaule bool) {
	offSystemMutex.Lock()
	defer offSystemMutex.Unlock()

	vaule = offSystem
	return
}

func InitOffSystem() {
	client := binance.GetOnceFuturesClient()
	pos, err := client.GetPositions(binance.ETHUSDT_PERP)
	if err != nil {
		log.Println(err)
	}
	posinfo := GetPositionInfo(pos)
	setOffSystem(posinfo)
}

func GetSleepSec() (vaule int) {
	vaule = sleepSec
	if vaule <= 0 {
		vaule = conf.Get().Trading.TriggerTime * 60
	}
	return
}

var (
	workHourList = []int{9, 10, 11, 12, 18, 19, 20, 21, 22, 23}
)

// IsWork 判断是否执行
func IsWork() bool {
	return true
	now := time.Now()
	h := now.Hour()
	weekday := now.Weekday()
	if utils.InSlice(workHourList, h) && weekday > 0 && weekday < 6 {
		//周一至周五 10-23点
		return true
	}
	if utils.InSlice([]int{0, 1, 2, 3}, h) && weekday > 1 && weekday < 7 {
		//周二至周六 0-2点
		return true
	}

	if !GetOffSystem() {
		return true //如果持仓中就继续
	}
	return false
}

// RunQuantitativeTrading 运行量化交易主流程
func RunQuantitativeTrading() error {
	log.Println("========================================")
	log.Println("[量化交易] 启动ETH期货量化交易系统")
	log.Println("========================================")

	// 1. 获取市场数据
	marketData, err := GetMarketData()
	if err != nil {
		log.Printf("[量化交易] 错误: 无法获取市场数据 - %v", err)
		return err
	}
	if marketData.PositionInfo.HasLong && marketData.PositionInfo.HasShort {
		//当双向持仓的情况下出现，平仓后关闭系统
		clSignal := &TradingSignal{Action: "CLOSE_LONG", PositionSize: 100}
		ExecuteTrade(clSignal, marketData)
		csSignal := &TradingSignal{Action: "CLOSE_SHORT", PositionSize: 100}
		ExecuteTrade(csSignal, marketData)
		log.Printf("[系统] 系统无法处理双向持仓的情况，被迫关闭")
		panic("异常")
	}

	// 2. 创建市场状态快照（在LLM分析前）
	marketState := CreateMarketStateSnapshot(marketData)

	// 3. LLM分析
	signal, err := AnalyzeWithLLM(marketData)
	if err != nil {
		log.Printf("[量化交易] 错误: LLM分析失败 - %v", err)
		return err
	}
	log.Printf("[量化交易] 分析结果: %s (评分: %d, 置信度: %.2f%%)", signal.Action, signal.Score, signal.Confidence*100)
	log.Printf("[量化交易] 记忆: %s", signal.Memory)
	log.Printf("[量化交易] 动作: %s，仓位: %v", signal.Action, signal.PositionSize)

	// 4. 先记录决策到Memory（获取decisionID）
	decisionID := SetMemoryFromSignal(signal, marketState)
	log.Printf("[量化交易] 决策已记录 ID: %s", decisionID)

	// 5. 执行交易（传入decisionID用于更新入场价格）
	err = ExecuteTradeWithMemory(signal, marketData, decisionID)
	if err != nil {
		log.Printf("[量化交易] 错误: 交易执行失败 - %v", err)
		return err
	}

	refreshTimer()
	return err
}

// CreateMarketStateSnapshot 创建市场状态快照
func CreateMarketStateSnapshot(marketData *MarketData) MarketStateSnapshot {
	snapshot := MarketStateSnapshot{}

	// 获取当前价格
	if marketData.Ticker != nil {
		snapshot.Price, _ = strconv.ParseFloat(marketData.Ticker.LastPrice, 64)
	}

	// 准备技术分析数据
	technicalData := PrepareTechnicalData(marketData)

	// 获取趋势分析
	trendAnalysis := AnalyzeTrendBy15m(technicalData)
	snapshot.Trend = trendAnalysis.Direction
	snapshot.TrendStrength = trendAnalysis.Strength

	// 获取波动率
	if len(technicalData.High3m) > 0 && len(technicalData.Low3m) > 0 && len(technicalData.Price3m) > 0 {
		atr := indicators.GetLatestATR(technicalData.High3m, technicalData.Low3m, technicalData.Price3m, 14)
		snapshot.Volatility = indicators.CalculateVolatilityPercent(atr, technicalData.CurrentPrice)
	}

	// 获取RSI
	if len(technicalData.Price3m) > 0 {
		ta := indicators.AnalyzeAll(technicalData.High3m, technicalData.Low3m, technicalData.Price3m, technicalData.CurrentPrice)
		snapshot.RSI = ta.RSI

		// MACD信号
		macd := indicators.GetLatestMACD(technicalData.Price3m, 12, 26, 9)
		if macd != nil && len(macd.MACDLine) > 0 && len(macd.SignalLine) > 0 {
			lastMACD := macd.MACDLine[len(macd.MACDLine)-1]
			lastSignal := macd.SignalLine[len(macd.SignalLine)-1]
			if lastMACD > lastSignal {
				snapshot.MACDSignal = "BULLISH"
			} else if lastMACD < lastSignal {
				snapshot.MACDSignal = "BEARISH"
			} else {
				snapshot.MACDSignal = "NEUTRAL"
			}
		}
	}

	// 获取支撑阻力位
	srLevels := IdentifySupportResistance(technicalData)
	for _, level := range srLevels {
		if level.Type == "SUPPORT" && (snapshot.SupportLevel == 0 || level.Distance < (snapshot.Price-snapshot.SupportLevel)/snapshot.Price*100) {
			snapshot.SupportLevel = level.Price
		}
		if level.Type == "RESISTANCE" && (snapshot.ResistanceLevel == 0 || level.Distance < (snapshot.ResistanceLevel-snapshot.Price)/snapshot.Price*100) {
			snapshot.ResistanceLevel = level.Price
		}
	}

	// 获取资金费率
	if marketData.FundingRate != nil {
		snapshot.FundingRate, _ = strconv.ParseFloat(marketData.FundingRate.FundingRate, 64)
	}

	return snapshot
}

func refreshTimer() {
	log.Println("========================================")
	log.Println("[量化交易] 本轮交易流程完成")
	log.Println("========================================")

	time.Sleep(30 * time.Second)
	client, err := binance.GetFuturesClient()
	positions, err := client.GetPositions(binance.ETHUSDT_PERP)
	if err != nil {
		return
	}
	ps := GetPositionInfo(positions)
	setOffSystem(ps)
	sleepSec = conf.Get().Trading.TriggerTime*60 - 30
	if !ps.HasLong && !ps.HasShort {
		CloseFetchPosition()
		return
	}

	//sleepSec = 3*60 - 30
	// if ps.Duration >= time.Minute*15 {
	// 	now := time.Now()
	// 	switch {
	// 	case now.Hour() >= 21 && now.Hour() <= 23:
	// 		sleepSec = 4*60 - 30 //晚高峰阶段
	// 	case now.Hour() >= 0 && now.Hour() <= 2:
	// 		sleepSec = 4*60 - 30 //晚高峰阶段
	// 	default:
	// 		//如果持仓时间大等于15分钟，等待6分钟
	// 		sleepSec = 6*60 - 30
	// 	}
	// }
	StartFetchPosition()
}
