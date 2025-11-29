package task

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"deeptrade/binance"
	tradeflow "deeptrade/task/trade_flow"
)

// GetMarketData 获取完整的市场数据
func GetMarketData() (*MarketData, error) {
	log.Println("[市场数据] 开始获取完整市场数据...")

	// 获取期货客户端
	client, err := binance.GetFuturesClient()
	if err != nil {
		log.Printf("[市场数据] 创建期货客户端失败: %v", err)
		return nil, err
	}

	symbol := binance.ETHUSDT_PERP

	// 并发获取所有数据
	var ticker *binance.FuturesTicker
	var klines3m []binance.Kline
	var orderBook *binance.Depth
	var positions []binance.Position
	var account *binance.FuturesAccountInfo
	var markPrice *binance.MarkPrice
	var fundingRate *binance.FundingRateHistory
	var openInterest *binance.OpenInterest
	var orderHistory []binance.Order
	var openOrders []binance.Order
	var fundingRateHistorys []binance.FundingRateHistory
	var bookTicker *binance.BookTicker

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	// 获取24小时价格统计
	wg.Add(1)
	go func() {
		defer wg.Done()
		t, err := client.Get24hrTicker(symbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("获取价格统计失败: %v", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		ticker = t
		mu.Unlock()
	}()

	// 获取3分钟K线数据（75条用于短期精准分析，覆盖约3.75小时，避免长期数据均值失真，提高对近期变化的敏感度）
	wg.Add(1)
	go func() {
		defer wg.Done()
		klines, err := client.GetKlines(symbol, binance.KlineInterval3m, 71)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("获取3分钟K线失败: %v", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		klines3m = klines
		mu.Unlock()
	}()

	// 获取订单簿深度
	wg.Add(1)
	go func() {
		defer wg.Done()
		depth, err := client.GetDepth(symbol, binance.DepthLevel20)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("获取订单簿失败: %v", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		orderBook = depth
		mu.Unlock()
	}()

	// 获取当前持仓
	wg.Add(1)
	go func() {
		defer wg.Done()
		pos, err := client.GetPositions(symbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("获取持仓信息失败: %v", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		positions = pos
		mu.Unlock()
	}()

	// 获取账户信息
	wg.Add(1)
	go func() {
		defer wg.Done()
		acc, err := client.GetAccountInfo()
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("获取账户信息失败: %v", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		account = acc
		mu.Unlock()
	}()

	// 获取标记价格
	wg.Add(1)
	go func() {
		defer wg.Done()
		mp, err := client.GetMarkPrice(symbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("获取标记价格失败: %v", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		markPrice = mp
		mu.Unlock()
	}()

	// 获取资金费率
	wg.Add(1)
	go func() {
		defer wg.Done()
		fr, err := client.GetLatestFundingRate(symbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("获取资金费率失败: %v", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		fundingRate = fr
		mu.Unlock()
	}()

	// 获取资金费率历史
	wg.Add(1)
	go func() {
		defer wg.Done()
		frs, err := client.GetFundingRateHistory(symbol, 6, 0, 0)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("获取资金费率历史失败: %v", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		fundingRateHistorys = frs
		mu.Unlock()
	}()

	// 获取持仓量
	wg.Add(1)
	go func() {
		defer wg.Done()
		oi, err := client.GetOpenInterest(symbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("获取持仓量失败: %v", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		openInterest = oi
		mu.Unlock()
	}()

	// 获取最近交易数据
	wg.Add(1)
	go func() {
		defer wg.Done()
		tradeflow.FetchRecentTrade()
	}()

	// 获取历史订单数据（最近15个订单）
	wg.Add(1)
	go func() {
		defer wg.Done()
		orders, err := client.GetOrderHistory(symbol, 15, 0, 0, 0) // 不限制时间范围，获取最新15个订单
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("获取历史订单失败: %v", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		orderHistory = orders
		mu.Unlock()
	}()

	// 获取当前最优挂单信息
	wg.Add(1)
	go func() {
		defer wg.Done()
		bt, err := client.GetBookTicker(symbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("获取最优挂单失败: %v", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		bookTicker = bt
		mu.Unlock()
	}()

	// 获取当前挂单信息
	wg.Add(1)
	go func() {
		defer wg.Done()
		orders, err := client.GetOpenOrders(symbol)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("获取当前挂单失败: %v", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		openOrders = orders
		mu.Unlock()
	}()

	wg.Wait()

	// 检查是否有错误
	if len(errs) > 0 {
		log.Printf("[市场数据] 获取市场数据时发生错误: %v", errs)
	}

	data := &MarketData{
		Ticker:              ticker,
		Klines3m:            klines3m,
		OrderBook:           orderBook,
		BookTicker:          bookTicker,
		Positions:           positions,
		Account:             account,
		MarkPrice:           markPrice.MarkPrice,
		FundingRate:         fundingRate,
		OpenInterest:        openInterest,
		OrderHistory:        orderHistory,
		OpenOrders:          openOrders,
		MarkPriceDetail:     markPrice,
		FundingRateHistorys: fundingRateHistorys,
		PositionInfo:        GetPositionInfo(positions),
	}

	log.Printf("[市场数据] 获取完成 - 当前价格: %s, 标记价格: %s", ticker.LastPrice, markPrice.MarkPrice)
	return data, nil
}

// AnalyzeFundingRateTrend 分析资金费率历史趋势
func AnalyzeFundingRateTrend(currentRate float64, fundingRateHistorys []binance.FundingRateHistory) string {
	if len(fundingRateHistorys) == 0 {
		return "暂无历史数据"
	}

	// 收集有效费率数据和时间戳
	var rates []float64
	var timestamps []int64
	for _, fr := range fundingRateHistorys {
		rate, _ := strconv.ParseFloat(fr.FundingRate, 64)
		rates = append(rates, rate)
		timestamps = append(timestamps, fr.FundingTime)
	}

	if len(rates) < 2 {
		return fmt.Sprintf("当前费率: %.6f (数据不足)", currentRate)
	}

	// 找出最高和最低费率
	maxRate := rates[0]
	minRate := rates[0]

	for _, rate := range rates {
		if rate > maxRate {
			maxRate = rate
		}
		if rate < minRate {
			minRate = rate
		}
	}

	// 计算实际时间范围 - 基于历史数据的最早和最晚时间
	var earliestTime, latestTime time.Time
	earliestTime = time.Unix(timestamps[0]/1000, 0)               // 第0条是最早的
	latestTime = time.Unix(timestamps[len(timestamps)-1]/1000, 0) // 最后一条是最新的
	timeRangeHours := latestTime.Sub(earliestTime).Hours()

	// 计算趋势：比较最近2次费率平均值与之前1次费率平均值
	trendDesc := "趋势平稳"
	if len(rates) >= 3 {
		recentAvg := (rates[len(rates)-1] + rates[len(rates)-2]) / 2 // 最近2次费率平均值（最新+倒数第二）
		earlierAvg := rates[len(rates)-3]                            // 之前1次费率（倒数第三）
		trend := recentAvg - earlierAvg

		if trend > 0.00001 {
			trendDesc = "趋势上升"
		} else if trend < -0.00001 {
			trendDesc = "趋势下降"
		}
	}

	return fmt.Sprintf("%.0f小时范围最高: %.6f, 最低: %.6f, 当前: %.6f, %s",
		timeRangeHours, maxRate, minRate, currentRate, trendDesc)
}

// FormatFundingAnalysis 格式化资金费率和持仓量分析
func FormatFundingAnalysis(marketData *MarketData) string {
	var analysis strings.Builder
	analysis.WriteString("资金费率:\n")

	if marketData.FundingRate != nil {
		rate, _ := strconv.ParseFloat(marketData.FundingRate.FundingRate, 64)
		// 修正显示：rate本身就是小数形式，不需要额外乘100
		analysis.WriteString(fmt.Sprintf("  当前资金费率: %.6f (每8小时结算)\n", rate))
		analysis.WriteString(fmt.Sprintf("  年化费率: %.2f%%\n", rate*3*365*100))

		// 市场情绪解读 - 使用新的分析函数
		trendInfo := AnalyzeFundingRateTrend(rate, marketData.FundingRateHistorys)
		analysis.WriteString(fmt.Sprintf("  费率趋势: %s\n", trendInfo))

		// 下次费率时间
		if marketData.FundingRate.FundingTime > 0 {
			nextFunding := time.Unix(marketData.FundingRate.FundingTime/1000, 0)

			// 如果当前时间已过结算时间，计算下一个结算周期
			for time.Until(nextFunding) <= 0 {
				nextFunding = nextFunding.Add(8 * time.Hour) // 资金费率每8小时结算一次
			}

			remaining := time.Until(nextFunding)
			analysis.WriteString(fmt.Sprintf("  下次结算: %s (剩余%v)\n",
				nextFunding.Format("15:04:05"), remaining.Round(time.Minute)))
		}
	} else {
		analysis.WriteString("  资金费率数据: 暂无\n")
	}

	// 持仓量分析
	if marketData.OpenInterest != nil {
		oiRaw := marketData.OpenInterest.OpenInterest
		oiFloat, err := strconv.ParseFloat(oiRaw, 64)
		if err != nil {
			analysis.WriteString(fmt.Sprintf("  未平仓合约: 数据解析错误 (原始值: '%s', 错误: %v)\n", oiRaw, err))
		} else if oiFloat > 0 && oiFloat < 1000000000 { // 调整为10亿张的合理性检查
			analysis.WriteString(fmt.Sprintf("  未平仓合约: %.0f 张\n", oiFloat))
		} else {
			analysis.WriteString(fmt.Sprintf("  未平仓合约: 数据异常 (原始值: '%s', 解析后: %.0f, 可能是API返回格式问题)\n", oiRaw, oiFloat))
		}
	} else {
		analysis.WriteString("  未平仓合约: 暂无数据\n")
	}

	return analysis.String()
}
