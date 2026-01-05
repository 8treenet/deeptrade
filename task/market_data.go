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
	var klines15m []binance.Kline
	var orderBook *binance.Depth
	var positions []binance.Position
	var account *binance.FuturesAccountInfo
	var markPrice *binance.MarkPrice
	var fundingRate *binance.FundingRateHistory
	var openInterest *binance.OpenInterest
	var orderHistory []binance.Order
	var openOrders []binance.Order
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

	// 获取15分钟K线数据（100条用于趋势方向判断，覆盖约25小时）
	wg.Add(1)
	go func() {
		defer wg.Done()
		klines, err := client.GetKlines(symbol, binance.KlineInterval15m, 100)
		if err != nil {
			mu.Lock()
			errs = append(errs, fmt.Errorf("获取15分钟K线失败: %v", err))
			mu.Unlock()
			return
		}
		mu.Lock()
		klines15m = klines
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
		Ticker:          ticker,
		Klines3m:        klines3m,
		Klines15m:       klines15m,
		OrderBook:       orderBook,
		BookTicker:      bookTicker,
		Positions:       positions,
		Account:         account,
		MarkPrice:       markPrice.MarkPrice,
		FundingRate:     fundingRate,
		OpenInterest:    openInterest,
		OrderHistory:    orderHistory,
		OpenOrders:      openOrders,
		MarkPriceDetail: markPrice,
		PositionInfo:    GetPositionInfo(positions),
	}

	log.Printf("[市场数据] 获取完成 - 当前价格: %s, 标记价格: %s", ticker.LastPrice, markPrice.MarkPrice)
	return data, nil
}

// FormatFundingAnalysis 格式化资金费率和持仓量分析
func FormatFundingAnalysis(marketData *MarketData) string {
	var analysis strings.Builder
	analysis.WriteString("资金费率:\n")

	if marketData.FundingRate != nil && marketData.FundingRate.FundingRate != "" {
		rate, _ := strconv.ParseFloat(marketData.FundingRate.FundingRate, 64)
		// 修正显示：rate本身就是小数形式，不需要额外乘100
		analysis.WriteString(fmt.Sprintf("  当前资金费率: %.6f (每8小时结算)\n", rate))
		analysis.WriteString(fmt.Sprintf("  年化费率: %.2f%%\n", rate*3*365*100))

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
		if marketData.FundingRate == nil {
			analysis.WriteString("  资金费率数据: 暂无 (API返回nil)\n")
		} else {
			analysis.WriteString(fmt.Sprintf("  资金费率数据: 暂无 (FundingRate字段为空, Symbol=%s, Time=%d)\n",
				marketData.FundingRate.Symbol, marketData.FundingRate.FundingTime))
		}
	}

	// 持仓量分析
	if marketData.OpenInterest != nil {
		oiRaw := marketData.OpenInterest.OpenInterest
		oiFloat, err := strconv.ParseFloat(oiRaw, 64)
		if err != nil {
			analysis.WriteString(fmt.Sprintf("  未平仓合约: 数据解析错误 (原始值: '%s', 错误: %v)\n", oiRaw, err))
		} else if oiFloat > 0 && oiFloat < 100000000000 { // 调整为1000亿张的合理性检查(ETHUSDT合约每张0.001ETH，实际持仓量可达数十亿张)
			analysis.WriteString(fmt.Sprintf("  未平仓合约: %.0f 张 (约 %.2f ETH)\n", oiFloat, oiFloat*0.001))
		} else {
			analysis.WriteString(fmt.Sprintf("  未平仓合约: 数据异常 (原始值: '%s', 解析后: %.0f, 可能是API返回格式问题)\n", oiRaw, oiFloat))
		}
	} else {
		analysis.WriteString("  未平仓合约: 暂无数据\n")
	}

	return analysis.String()
}
