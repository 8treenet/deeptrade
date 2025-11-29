package tradeflow

import (
	"deeptrade/binance"
	"log"
	"sort"
	"sync"
	"time"
)

func NewTradeflow() *TradeFlow {
	return &TradeFlow{dataMap: make(map[int64]binance.RecentTrade)}
}

type TradeFlow struct {
	mutex   sync.Mutex
	dataMap map[int64]binance.RecentTrade //id=>obj
}

func (tf *TradeFlow) Clear() {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()
	tf.dataMap = make(map[int64]binance.RecentTrade)
}

// AddRecentTrade 添加记录
func (tf *TradeFlow) AddRecentTrade(data []binance.RecentTrade) {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()
	size := 10000

	for _, v := range data {
		tf.dataMap[v.ID] = v
	}
	if len(tf.dataMap) < size {
		log.Printf("[系统] 添加交易数据，当前已有%d数据\n", len(tf.dataMap))
		return
	}

	// 将map转换为切片以便排序
	trades := make([]binance.RecentTrade, 0, len(tf.dataMap))
	for _, trade := range tf.dataMap {
		trades = append(trades, trade)
	}

	// 使用内置排序函数，根据时间戳降序排序（最新的在前）
	sort.Slice(trades, func(i, j int) bool {
		return trades[i].Time > trades[j].Time
	})

	// 只保留最近10000条数据
	newdata := make(map[int64]binance.RecentTrade, size)
	for i := 0; i < size && i < len(trades); i++ {
		newdata[trades[i].ID] = trades[i]
	}

	// 更新数据映射
	tf.dataMap = newdata
	log.Printf("[系统] 添加交易数据，当前已有%d数据\n", len(tf.dataMap))
}

// GetRecentTradesLast5Minutes 获取最近5分钟的交易数据
func (tf *TradeFlow) GetRecentTradesLast5Minutes() []binance.RecentTrade {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()

	result := tf.getRecentTradesByDuration(5 * time.Minute)
	// 根据时间戳升序排序（最新的在后）
	sort.Slice(result, func(i, j int) bool {
		return result[i].Time < result[j].Time
	})
	return result
}

// GetRecentTradesLast10Minutes 获取最近10分钟的交易数据
func (tf *TradeFlow) GetRecentTradesLast10Minutes() []binance.RecentTrade {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()

	result := tf.getRecentTradesByDuration(10 * time.Minute)
	// 根据时间戳升序排序（最新的在后）
	sort.Slice(result, func(i, j int) bool {
		return result[i].Time < result[j].Time
	})
	return result
}

// GetRecentTradesLast20Minutes 获取最近20分钟的交易数据
func (tf *TradeFlow) GetRecentTradesLast20Minutes() []binance.RecentTrade {
	tf.mutex.Lock()
	defer tf.mutex.Unlock()

	result := tf.getRecentTradesByDuration(20 * time.Minute)
	// 根据时间戳升序排序（最新的在后）
	sort.Slice(result, func(i, j int) bool {
		return result[i].Time < result[j].Time
	})
	return result
}

// getRecentTradesByDuration 根据指定时间范围获取最近的交易数据
func (tf *TradeFlow) getRecentTradesByDuration(duration time.Duration) []binance.RecentTrade {
	// 计算时间阈值（当前时间减去指定时间范围）
	thresholdTime := time.Now().Add(-duration).UnixNano() / int64(time.Millisecond)

	// 将map转换为切片以便排序和过滤
	trades := make([]binance.RecentTrade, 0, len(tf.dataMap))
	for _, trade := range tf.dataMap {
		// 只保留在指定时间范围内的交易
		if trade.Time >= thresholdTime {
			trades = append(trades, trade)
		}
	}

	// 根据时间戳降序排序（最新的在前）
	sort.Slice(trades, func(i, j int) bool {
		return trades[i].Time > trades[j].Time
	})

	return trades
}
