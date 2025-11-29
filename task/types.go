package task

import (
	"deeptrade/binance"
	"time"
)

// TradingSignal LLM返回的交易信号
type TradingSignal struct {
	Action       string  `json:"action"`        // 操作类型: OPEN_LONG, OPEN_SHORT, CLOSE_LONG, CLOSE_SHORT, ADD_LONG, ADD_SHORT, HOLD
	Score        int     `json:"score"`         // 评分: -10到+10
	Confidence   float64 `json:"confidence"`    // 置信度: 0.0-1.0
	StopLoss     float64 `json:"stop_loss"`     // 止损价格
	TakeProfit   float64 `json:"take_profit"`   // 止盈价格
	PositionSize int     `json:"position_size"` // 仓位大小
	Reasoning    string  `json:"reasoning"`     // 分析原因
	Memory       string  `json:"memory"`        //记忆
}

// FuturesTicker 期货价格数据（别名）
type FuturesTicker = binance.FuturesTicker

// FuturesAccountInfo 期货账户信息（别名）
type FuturesAccountInfo = binance.FuturesAccountInfo

// MarketData 完整的市场数据
type MarketData struct {
	Ticker              *FuturesTicker               // 24小时价格统计
	Klines3m            []binance.Kline              // 3分钟K线数据
	Klines1m            []binance.Kline              // 1分钟K线数据
	OrderBook           *binance.Depth               // 订单簿深度
	BookTicker          *binance.BookTicker          // 当前最优挂单信息
	Positions           []binance.Position           // 当前持仓
	Account             *FuturesAccountInfo          // 账户信息
	MarkPrice           string                       // 标记价格
	FundingRate         *binance.FundingRateHistory  // 资金费率
	OpenInterest        *binance.OpenInterest        // 持仓量 (未平仓合约张数,非实际ETH)
	OrderHistory        []binance.Order              // 历史订单数据
	OpenOrders          []binance.Order              // 当前挂单数据
	MarkPriceDetail     *binance.MarkPrice           // 详细标记价格数据
	FundingRateHistorys []binance.FundingRateHistory // 资金费率历史
	PositionInfo        *PositionInfo                //持仓信息
}

// PositionInfo 持仓信息结构
type PositionInfo struct {
	HasLong          bool          // 是否有多头持仓
	HasShort         bool          // 是否有空头持仓
	LongAmt          float64       // 多头数量
	ShortAmt         float64       // 空头数量
	NetAmt           float64       // 净持仓（单向模式）
	IsDualSide       bool          // 是否为双向持仓模式
	Duration         time.Duration // 持仓时间
	UnRealizedProfit string        //未实现盈亏
}

// TechnicalAnalysisData 技术指标数据结构
type TechnicalAnalysisData struct {
	Price3m      []float64 // 3分钟收盘价
	High3m       []float64 // 3分钟最高价
	Low3m        []float64 // 3分钟最低价
	CurrentPrice float64   // 当前价格
	Has3mData    bool      // 是否有3分钟数据
}
