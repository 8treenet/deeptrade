package tradeflow

import (
	"deeptrade/binance"
	"log"
	"sync"
	"time"
)

var tradeFlow *TradeFlow
var once sync.Once
var fetchRecentTradeMutex sync.Mutex
var fetchRecentTradeLatestTime time.Time

// GetOnceTradeFlow .
func GetOnceTradeFlow() *TradeFlow {
	// 获取期货客户端
	once.Do(func() {
		tradeFlow = NewTradeflow()
	})
	return tradeFlow
}

func RunFetch(iswork func() bool) {
	fetchRecentTradeLatestTime = time.Now().AddDate(0, 0, -1)
	go func() {
		for {
			if !iswork() {
				GetOnceTradeFlow().Clear()
				time.Sleep(150 * time.Second)
				continue
			}
			err := FetchRecentTrade()
			if err != nil {
				log.Println("[系统] 拉取交易数据失败", err)
			}
			time.Sleep(150 * time.Second)
		}

	}()
	time.Sleep(500 * time.Second) //首次启动需要等待趋势数据
}

func FetchRecentTrade() (e error) {
	now := time.Now()
	fetchRecentTradeMutex.Lock()
	duration := now.Sub(fetchRecentTradeLatestTime)
	fetchRecentTradeMutex.Unlock()

	if duration <= 10*time.Second {
		return
	}

	list, e := binance.GetOnceFuturesClient().GetRecentTrades(binance.ETHUSDT_PERP, 1000)
	if e != nil {
		return
	}
	GetOnceTradeFlow().AddRecentTrade(list)
	fetchRecentTradeMutex.Lock()
	fetchRecentTradeLatestTime = time.Now()
	fetchRecentTradeMutex.Unlock()
	return
}
