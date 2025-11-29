package binance

import (
	"time"
)

// Symbol 表示交易对符号
type Symbol string

const (
	// ETHUSDT 现货交易对
	ETHUSDT Symbol = "ETHUSDT"
	// ETHUSDT 合约交易对
	ETHUSDT_PERP Symbol = "ETHUSDT"
	// BTCUSDT 现货交易对
	BTCUSDT Symbol = "BTCUSDT"
	// BTCUSDT_PERP 合约交易对
	BTCUSDT_PERP Symbol = "BTCUSDT"
	// ADAUSDT 现货交易对
	ADAUSDT Symbol = "ADAUSDT"
	// BNBUSDT 现货交易对
	BNBUSDT Symbol = "BNBUSDT"
)

// OrderSide 订单方向
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderType 订单类型
type OrderType string

const (
	// 通用/现货
	OrderTypeMarket      OrderType = "MARKET"
	OrderTypeLimit       OrderType = "LIMIT"
	OrderTypeStopLoss    OrderType = "STOP_LOSS"
	OrderTypeStopLossLMT OrderType = "STOP_LOSS_LIMIT"
	OrderTypeTakeProfit  OrderType = "TAKE_PROFIT"
	// 期货专用
	OrderTypeStop             OrderType = "STOP"
	OrderTypeStopMarket       OrderType = "STOP_MARKET"
	OrderTypeTakeProfitMarket OrderType = "TAKE_PROFIT_MARKET"
)

// TimeInForce 订单有效期
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC" // Good Till Cancel
	TimeInForceIOC TimeInForce = "IOC" // Immediate or Cancel
	TimeInForceFOK TimeInForce = "FOK" // Fill or Kill
)

// OrderStatus 订单状态
type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCanceled        OrderStatus = "CANCELED"
	OrderStatusPendingCancel   OrderStatus = "PENDING_CANCEL"
	OrderStatusRejected        OrderStatus = "REJECTED"
	OrderStatusExpired         OrderStatus = "EXPIRED"
)

// KlineInterval K线间隔
type KlineInterval string

const (
	KlineInterval1m  KlineInterval = "1m"
	KlineInterval3m  KlineInterval = "3m"
	KlineInterval5m  KlineInterval = "5m"
	KlineInterval15m KlineInterval = "15m"
	KlineInterval30m KlineInterval = "30m"
	KlineInterval1h  KlineInterval = "1h"
	KlineInterval2h  KlineInterval = "2h"
	KlineInterval4h  KlineInterval = "4h"
	KlineInterval6h  KlineInterval = "6h"
	KlineInterval8h  KlineInterval = "8h"
	KlineInterval12h KlineInterval = "12h"
	KlineInterval1d  KlineInterval = "1d"
	KlineInterval3d  KlineInterval = "3d"
	KlineInterval1w  KlineInterval = "1w"
	KlineInterval1M  KlineInterval = "1M"
)

// Ticker 24小时价格统计
type Ticker struct {
	Symbol             string `json:"symbol"`             // 交易对
	PriceChange        string `json:"priceChange"`        // 24小时价格变化
	PriceChangePercent string `json:"priceChangePercent"` // 24小时价格变化百分比
	WeightedAvgPrice   string `json:"weightedAvgPrice"`   // 加权平均价
	PrevClosePrice     string `json:"prevClosePrice"`     // 前收盘价
	LastPrice          string `json:"lastPrice"`          // 最新价格
	LastQty            string `json:"lastQty"`            // 最新成交量
	BidPrice           string `json:"bidPrice"`           // 当前最高买价
	BidQty             string `json:"bidQty"`             // 当前最高买价的数量
	AskPrice           string `json:"askPrice"`           // 当前最低卖价
	AskQty             string `json:"askQty"`             // 当前最低卖价的数量
	OpenPrice          string `json:"openPrice"`          // 开盘价
	HighPrice          string `json:"highPrice"`          // 24小时最高价
	LowPrice           string `json:"lowPrice"`           // 24小时最低价
	Volume             string `json:"volume"`             // 24小时成交量
	QuoteVolume        string `json:"quoteVolume"`        // 24小时成交额
	OpenTime           int64  `json:"openTime"`           // 24小时内第一笔交易的发生时间
	CloseTime          int64  `json:"closeTime"`          // 24小时内最后一笔交易的发生时间
	Count              int64  `json:"count"`              // 24小时内交易次数
}

// DepthLevel 深度数据级别
type DepthLevel int

const (
	DepthLevel5    DepthLevel = 5
	DepthLevel10   DepthLevel = 10
	DepthLevel20   DepthLevel = 20
	DepthLevel50   DepthLevel = 50
	DepthLevel100  DepthLevel = 100
	DepthLevel500  DepthLevel = 500
	DepthLevel1000 DepthLevel = 1000
	DepthLevel5000 DepthLevel = 5000
)

// DepthEntry 深度条目
type DepthEntry struct {
	Price    string `json:"price"`    // 价格
	Quantity string `json:"quantity"` // 数量
}

// Depth 深度数据
type Depth struct {
	LastUpdateID int64        `json:"lastUpdateId"` // 最后更新ID
	Bids         []DepthEntry `json:"bids"`         // 买盘
	Asks         []DepthEntry `json:"asks"`         // 卖盘
}

// Kline K线数据
type Kline struct {
	OpenTime                 int64  `json:"t"`   // 开盘时间
	CloseTime                int64  `json:"T"`   // 收盘时间
	Open                     string `json:"o"`   // 开盘价
	High                     string `json:"h"`   // 最高价
	Low                      string `json:"l"`   // 最低价
	Close                    string `json:"c"`   // 收盘价
	Volume                   string `json:"v"`   // 成交量
	QuoteAssetVolume         string `json:"q"`   // 成交额
	TradeNum                 int64  `json:"n"`   // 成交数量
	TakerBuyBaseAssetVolume  string `json:"bb"`  // 主动买入成交量
	TakerBuyQuoteAssetVolume string `json:"qav"` // 主动买入成交额
}

// RecentTrade 最近交易
type RecentTrade struct {
	ID           int64  `json:"id"`           // 交易ID
	Price        string `json:"price"`        // 价格
	Qty          string `json:"qty"`          // 数量
	QuoteQty     string `json:"quoteQty"`     // 成交额
	Time         int64  `json:"time"`         // 时间
	IsBuyerMaker bool   `json:"isBuyerMaker"` // 是否买方挂单
	IsBestMatch  bool   `json:"isBestMatch"`  // 是否最佳匹配
}

// AggTrade 聚合交易
type AggTrade struct {
	AggTradeID  int64  `json:"aggTradeId"`  // 聚合交易ID
	Price       string `json:"price"`       // 价格
	Quantity    string `json:"quantity"`    // 数量
	FirstID     int64  `json:"firstId"`     // 首个交易ID
	LastID      int64  `json:"lastId"`      // 最后交易ID
	Timestamp   int64  `json:"timestamp"`   // 时间戳
	IsBuyer     bool   `json:"isBuyer"`     // 是否买方
	IsBestMatch bool   `json:"isBestMatch"` // 是否最佳匹配
}

// ServerTime 服务器时间
type ServerTime struct {
	ServerTime int64 `json:"serverTime"` // 服务器时间
}

// SymbolInfo 交易对信息
type SymbolInfo struct {
	Symbol                     string   `json:"symbol"`                     // 交易对
	Status                     string   `json:"status"`                     // 交易状态
	BaseAsset                  string   `json:"baseAsset"`                  // 基础资产
	BaseAssetPrecision         int      `json:"baseAssetPrecision"`         // 基础资产精度
	QuoteAsset                 string   `json:"quoteAsset"`                 // 计价资产
	QuoteAssetPrecision        int      `json:"quoteAssetPrecision"`        // 计价资产精度
	OrderTypes                 []string `json:"orderTypes"`                 // 支持的订单类型
	IcebergAllowed             bool     `json:"icebergAllowed"`             // 是否支持冰山订单
	OcoAllowed                 bool     `json:"ocoAllowed"`                 // 是否支持OCO订单
	QuoteOrderQtyMarketAllowed bool     `json:"quoteOrderQtyMarketAllowed"` // 市价单是否支持计价资产数量
	AllowTrailingStop          bool     `json:"allowTrailingStop"`          // 是否支持追踪止损
	CancelReplaceAllowed       bool     `json:"cancelReplaceAllowed"`       // 是否支持取消替换
	IsSpotTradingAllowed       bool     `json:"isSpotTradingAllowed"`       // 是否支持现货交易
	IsMarginTradingAllowed     bool     `json:"isMarginTradingAllowed"`     // 是否支持杠杆交易
	Filters                    []Filter `json:"filters"`                    // 交易规则过滤器
	Permissions                []string `json:"permissions"`                // 权限
}

// Filter 交易规则过滤器
type Filter struct {
	FilterType       string `json:"filterType"`       // 过滤器类型
	MinPrice         string `json:"minPrice"`         // 最小价格
	MaxPrice         string `json:"maxPrice"`         // 最大价格
	TickSize         string `json:"tickSize"`         // 价格精度
	MinQty           string `json:"minQty"`           // 最小数量
	MaxQty           string `json:"maxQty"`           // 最大数量
	StepSize         string `json:"stepSize"`         // 数量精度
	MinNotional      string `json:"minNotional"`      // 最小名义价值
	ApplyToMarket    bool   `json:"applyToMarket"`    // 是否应用于市价单
	AvgPriceMins     int    `json:"avgPriceMins"`     // 平均价格分钟数
	Limit            int    `json:"limit"`            // 限制
	MaxNumOrders     int    `json:"maxNumOrders"`     // 最大订单数
	MaxNumAlgoOrders int    `json:"maxNumAlgoOrders"` // 最大算法订单数
}

// AccountInfo 账户信息
type AccountInfo struct {
	MakerCommission  int       `json:"makerCommission"`  // 挂单手续费率
	TakerCommission  int       `json:"takerCommission"`  // 吃单手续费率
	BuyerCommission  int       `json:"buyerCommission"`  // 买方手续费率
	SellerCommission int       `json:"sellerCommission"` // 卖方手续费率
	CanTrade         bool      `json:"canTrade"`         // 是否可以交易
	CanWithdraw      bool      `json:"canWithdraw"`      // 是否可以提现
	CanDeposit       bool      `json:"canDeposit"`       // 是否可以充值
	UpdateTime       int64     `json:"updateTime"`       // 更新时间
	AccountType      string    `json:"accountType"`      // 账户类型
	Balances         []Balance `json:"balances"`         // 资产余额
	Permissions      []string  `json:"permissions"`      // 权限
	UID              int64     `json:"uid"`              // 用户ID
}

// Balance 资产余额
type Balance struct {
	Asset  string `json:"asset"`  // 资产名称
	Free   string `json:"free"`   // 可用余额
	Locked string `json:"locked"` // 冻结余额
}

// Order 订单信息
type Order struct {
	Symbol             string      `json:"symbol"`              // 交易对
	OrderID            int64       `json:"orderId"`             // 订单ID
	OrderListId        int64       `json:"orderListId"`         // 订单列表ID
	ClientOrderID      string      `json:"clientOrderId"`       // 客户端订单ID
	Price              string      `json:"price"`               // 订单价格
	OrigQty            string      `json:"origQty"`             // 订单数量
	ExecutedQty        string      `json:"executedQty"`         // 已执行数量
	CumulativeQuoteQty string      `json:"cummulativeQuoteQty"` // 累计成交金额
	Status             OrderStatus `json:"status"`              // 订单状态
	TimeInForce        TimeInForce `json:"timeInForce"`         // 订单有效期
	Type               OrderType   `json:"type"`                // 订单类型
	Side               OrderSide   `json:"side"`                // 买卖方向
	StopPrice          string      `json:"stopPrice"`           // 止损价格
	IcebergQty         string      `json:"icebergQty"`          // 冰山数量
	Time               int64       `json:"time"`                // 订单创建时间
	UpdateTime         int64       `json:"updateTime"`          // 订单更新时间
	IsWorking          bool        `json:"isWorking"`           // 是否正在工作
	WorkingTime        int64       `json:"workingTime"`         // 开始工作的时间
	OrigType           OrderType   `json:"origType"`            // 原始订单类型
	PositionSide       string      `json:"positionSide"`        // 持仓方向(仅合约)
	PriceProtect       bool        `json:"priceProtect"`        // 价格保护(仅合约)
}

// NewOrderRequest 新订单请求
type NewOrderRequest struct {
	Symbol        Symbol      `json:"symbol"`        // 交易对
	Side          OrderSide   `json:"side"`          // 买卖方向
	Type          OrderType   `json:"type"`          // 订单类型
	Quantity      string      `json:"quantity"`      // 数量
	Price         string      `json:"price"`         // 价格(限价单必填)
	TimeInForce   TimeInForce `json:"timeInForce"`   // 订单有效期
	StopPrice     string      `json:"stopPrice"`     // 触发价格(止损/止盈)
	ReduceOnly    bool        `json:"reduceOnly"`    // 只减仓(期货)
	ClosePosition bool        `json:"closePosition"` // 全部平仓(期货)
	WorkingType   WorkingType `json:"workingType"`   // 触发价格类型(标记价/合约价)
}

// FormatTime 格式化时间戳为字符串
func FormatTime(timestamp int64) string {
	return time.Unix(0, timestamp*int64(time.Millisecond)).Format("2006-01-02 15:04:05.000")
}

// ParseTime 解析字符串时间戳为int64
func ParseTime(timeStr string) (int64, error) {
	t, err := time.Parse("2006-01-02 15:04:05.000", timeStr)
	if err != nil {
		return 0, err
	}
	return t.UnixNano() / int64(time.Millisecond), nil
}

// OpenInterest 持仓量
type OpenInterest struct {
	Symbol            string `json:"symbol"`            // 交易对
	OpenInterest      string `json:"openInterest"`      // 持仓量
	OpenInterestValue string `json:"openInterestValue"` // 持仓额
	Timestamp         int64  `json:"timestamp"`         // 时间戳
}

// TopLongShortPositionRatio 大户持仓量多空比
type TopLongShortPositionRatio struct {
	Symbol         string `json:"symbol"`         // 交易对
	LongShortRatio string `json:"longShortRatio"` // 大户多空持仓量比值
	LongAccount    string `json:"longAccount"`    // 大户多仓持仓量比例
	ShortAccount   string `json:"shortAccount"`   // 大户空仓持仓量比例
	Timestamp      int64  `json:"timestamp"`      // 时间戳
}

// TopLongShortAccountRatio 大户账户数多空比
type TopLongShortAccountRatio struct {
	Symbol         string `json:"symbol"`         // 交易对
	LongShortRatio string `json:"longShortRatio"` // 大户多空账户数比值
	LongAccount    string `json:"longAccount"`    // 大户多仓账户数比例
	ShortAccount   string `json:"shortAccount"`   // 大户空仓账户数比例
	Timestamp      int64  `json:"timestamp"`      // 时间戳
}

// ClientConfig 客户端配置 - 简化版，直接使用项目配置
type ClientConfig struct {
	APIKey             string
	SecretKey          string
	BaseURL            string
	StreamURL          string
	FuturesBaseURL     string
	FuturesStreamURL   string
	Timeout            int
	Debug              bool
	MaxRetries         int
	RateLimitRateLimit int
	RateLimitInterval  int
	RecvWindow         int
	RetryDelay         int
	RetryBackoff       int
	ProxyURL           string
	ProxyUser          string
	ProxyPass          string
}

// GetFuturesClient 获取期货客户端 - 使用项目配置
func GetFuturesClient() (*FuturesClient, error) {
	return NewFuturesClientFromConfig()
}

// BookTicker 当前最优挂单信息
type BookTicker struct {
	Symbol   string `json:"symbol"`   // 交易对
	BidPrice string `json:"bidPrice"` // 最优买单价
	BidQty   string `json:"bidQty"`   // 最优买单价挂单量
	AskPrice string `json:"askPrice"` // 最优卖单价
	AskQty   string `json:"askQty"`   // 最优卖单价挂单量
	Time     int64  `json:"time"`     // 撮合引擎时间
}
