package binance

import (
	"deeptrade/conf"
	"deeptrade/utils"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/8treenet/freedom/infra/requests"
)

// FuturesClient 合约交易客户端
type FuturesClient struct {
	clientConfig *ClientConfig
	httpClient   *http.Client
	rateLimit    *RequestRateLimit
}

// PositionSide 持仓方向
type PositionSide string

const (
	PositionSideBoth  PositionSide = "BOTH"  // 双向持仓
	PositionSideLong  PositionSide = "LONG"  // 多头
	PositionSideShort PositionSide = "SHORT" // 空头
)

// MarginType 保证金类型
type MarginType string

const (
	MarginTypeIsolated MarginType = "ISOLATED" // 逐仓
	MarginTypeCross    MarginType = "CROSS"    // 全仓
)

// WorkingType 工作类型
type WorkingType string

const (
	WorkingTypeMarkPrice     WorkingType = "MARK_PRICE"     // 标记价格
	WorkingTypeContractPrice WorkingType = "CONTRACT_PRICE" // 合约价格
)

// Position 持仓信息
type Position struct {
	Symbol           string       `json:"symbol"`           // 交易对
	PositionAmt      string       `json:"positionAmt"`      // 持仓数量
	EntryPrice       string       `json:"entryPrice"`       // 开仓均价
	MarkPrice        string       `json:"markPrice"`        // 标记价格
	UnRealizedProfit string       `json:"unRealizedProfit"` // 未实现盈亏
	LiquidationPrice string       `json:"liquidationPrice"` // 强平价格
	Leverage         string       `json:"leverage"`         // 杠杆倍数
	MaxNotionalValue string       `json:"maxNotionalValue"` // 当前杠杆下用户可用的最大名义价值
	MarginType       MarginType   `json:"marginType"`       // 保证金模式
	IsolatedMargin   string       `json:"isolatedMargin"`   // 逐仓保证金
	IsAutoAddMargin  string       `json:"isAutoAddMargin"`  // 是否自动追加保证金
	PositionSide     PositionSide `json:"positionSide"`     // 持仓方向
	Notional         string       `json:"notional"`         // 名义价值
	IsolatedWallet   string       `json:"isolatedWallet"`   // 逐仓钱包余额
	UpdateTime       int64        `json:"updateTime"`       // 更新时间
}

// FuturesAccountInfo 合约账户信息
type FuturesAccountInfo struct {
	FeeTier                     int64      `json:"feeTier"`                     // 手续费等级
	CanTrade                    bool       `json:"canTrade"`                    // 是否可以交易
	CanDeposit                  bool       `json:"canDeposit"`                  // 是否可以充值
	CanWithdraw                 bool       `json:"canWithdraw"`                 // 是否可以提现
	UpdateTime                  int64      `json:"updateTime"`                  // 更新时间
	TotalInitialMargin          string     `json:"totalInitialMargin"`          // 全部起始保证金
	TotalMaintMargin            string     `json:"totalMaintMargin"`            // 全部维持保证金
	TotalWalletBalance          string     `json:"totalWalletBalance"`          // 全部钱包余额
	TotalMarginBalance          string     `json:"totalMarginBalance"`          // 全部保证金余额
	TotalPositionInitialMargin  string     `json:"totalPositionInitialMargin"`  // 全部持仓起始保证金
	TotalOpenOrderInitialMargin string     `json:"totalOpenOrderInitialMargin"` // 全部挂单起始保证金
	TotalCrossWalletBalance     string     `json:"totalCrossWalletBalance"`     // 全仓钱包余额
	AvailableBalance            string     `json:"availableBalance"`            // 可用余额
	MaxWithdrawAmount           string     `json:"maxWithdrawAmount"`           // 最大可转出金额
	Assets                      []Asset    `json:"assets"`                      // 资产信息
	Positions                   []Position `json:"positions"`                   // 持仓信息
}

// Asset 合约资产信息
type Asset struct {
	Asset                  string `json:"asset"`                  // 资产名称
	WalletBalance          string `json:"walletBalance"`          // 钱包余额
	UnrealizedPnl          string `json:"unrealizedPnl"`          // 未实现盈亏
	MarginBalance          string `json:"marginBalance"`          // 保证金余额
	MaintMargin            string `json:"maintMargin"`            // 维持保证金
	InitialMargin          string `json:"initialMargin"`          // 起始保证金
	PositionInitialMargin  string `json:"positionInitialMargin"`  // 持仓起始保证金
	OpenOrderInitialMargin string `json:"openOrderInitialMargin"` // 挂单起始保证金
	CrossWalletBalance     string `json:"crossWalletBalance"`     // 全仓钱包余额
	CrossUnPnl             string `json:"crossUnPnl"`             // 全仓未实现盈亏
	AvailableBalance       string `json:"availableBalance"`       // 可用余额
	MaxWithdrawAmount      string `json:"maxWithdrawAmount"`      // 最大可转出金额
	MarginAvailable        bool   `json:"marginAvailable"`        // 是否可用作保证金
	UpdateTime             int64  `json:"updateTime"`             // 更新时间
}

// FuturesTicker 合约24小时价格统计
type FuturesTicker struct {
	Symbol             string `json:"symbol"`             // 交易对
	PriceChange        string `json:"priceChange"`        // 24小时价格变化
	PriceChangePercent string `json:"priceChangePercent"` // 24小时价格变化百分比
	WeightedAvgPrice   string `json:"weightedAvgPrice"`   // 加权平均价
	LastPrice          string `json:"lastPrice"`          // 最新价格
	LastQty            string `json:"lastQty"`            // 最新成交量
	OpenPrice          string `json:"openPrice"`          // 开盘价
	HighPrice          string `json:"highPrice"`          // 24小时最高价
	LowPrice           string `json:"lowPrice"`           // 24小时最低价
	Volume             string `json:"volume"`             // 24小时成交量
	QuoteVolume        string `json:"quoteVolume"`        // 24小时成交额
	OpenTime           int64  `json:"openTime"`           // 24小时内第一笔交易的发生时间
	CloseTime          int64  `json:"closeTime"`          // 24小时内最后一笔交易的发生时间
	FirstID            int64  `json:"firstId"`            // 首笔成交id
	LastID             int64  `json:"lastId"`             // 末笔成交id
	Count              int64  `json:"count"`              // 24小时内交易次数
}

// MarkPrice 标记价格
type MarkPrice struct {
	Symbol          string `json:"symbol"`          // 交易对
	MarkPrice       string `json:"markPrice"`       // 标记价格
	IndexPrice      string `json:"indexPrice"`      // 指数价格
	EstSettlePrice  string `json:"estSettlePrice"`  // 预估结算价
	LastFundingRate string `json:"lastFundingRate"` // 最近资金费率
	NextFundingTime int64  `json:"nextFundingTime"` // 下次资金费率时间
	InterestRate    string `json:"interestRate"`    // 利率
	Time            int64  `json:"time"`            // 时间
}

// FundingRateHistory 资金费率历史
type FundingRateHistory struct {
	Symbol      string `json:"symbol"`      // 交易对
	FundingRate string `json:"fundingRate"` // 资金费率
	FundingTime int64  `json:"fundingTime"` // 资金费率时间
	MarkPrice   string `json:"markPrice"`   // 资金费对应标记价格
}

// UserTrade 用户成交记录
type UserTrade struct {
	Symbol          string `json:"symbol"`          // 交易对
	ID              int64  `json:"id"`              // 成交ID
	OrderID         int64  `json:"orderId"`         // 订单ID
	Side            string `json:"side"`            // 买卖方向
	Price           string `json:"price"`           // 成交价格
	Qty             string `json:"qty"`             // 成交数量
	RealizedPnl     string `json:"realizedPnl"`     // 已实现盈亏
	MarginAsset     string `json:"marginAsset"`     // 保证金资产
	Commission      string `json:"commission"`      // 手续费
	CommissionAsset string `json:"commissionAsset"` // 手续费资产
	Time            int64  `json:"time"`            // 成交时间
	PositionSide    string `json:"positionSide"`    // 持仓方向
	Buyer           bool   `json:"buyer"`           // 是否买方
	Maker           bool   `json:"maker"`           // 是否挂单方
}

// NewFuturesClient 创建新的合约客户端 (保持向后兼容)
func NewFuturesClient(config interface{}) (*FuturesClient, error) {
	// 这个方法保持向后兼容，但建议使用 NewFuturesClientFromConfig
	return NewFuturesClientFromConfig()
}

// doRequest 执行HTTP请求
func (c *FuturesClient) doRequest(method, endpoint string, params map[string]string, needAuth bool) ([]byte, error) {
	// 速率限制
	c.rateLimit.Wait()

	// 构建URL
	fullURL := c.clientConfig.BaseURL + endpoint

	// 处理参数
	var req *http.Request
	var err error

	// 如果需要认证，添加认证参数
	if needAuth {
		if params == nil {
			params = make(map[string]string)
		}
		c.clientConfig.addAuthParams(params)
	}

	if method == "GET" || method == "DELETE" {
		if params != nil {
			queryString := c.clientConfig.buildQueryString(params)
			if queryString != "" {
				fullURL += "?" + queryString
			}
		}
		req, err = http.NewRequest(method, fullURL, nil)
	} else {
		// POST/PUT请求使用表单格式
		formData := url.Values{}
		for k, v := range params {
			formData.Set(k, v)
		}

		if needAuth {
			// 需要认证的请求，参数作为查询字符串发送
			// 注意：必须使用buildQueryString以确保与签名生成的查询字符串顺序一致
			queryString := c.clientConfig.buildQueryString(params)
			if queryString != "" {
				fullURL += "?" + queryString
			}
			req, err = http.NewRequest(method, fullURL, nil)
		} else {
			// 不需要认证的请求，参数作为表单数据发送
			req, err = http.NewRequest(method, fullURL, strings.NewReader(formData.Encode()))
			if err == nil {
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		}
	}

	if err != nil {
		return nil, NewError(ErrCodeInvalidRequest, "创建请求失败", err.Error(), "")
	}

	// 设置请求头
	req.Header.Set("User-Agent", "deeptrade/binance-go")
	if c.clientConfig.APIKey != "" {
		req.Header.Set("X-MBX-APIKEY", c.clientConfig.APIKey)
	}

	// 调试模式
	if c.clientConfig.Debug {
		//fmt.Printf("[DEBUG] Futures Request: %s %s\n", method, fullURL)
		// 注意：不要在调试中重新计算查询字符串，因为时间戳会变化
	}

	// 执行请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, NewError(ErrCodeDisconnected, "请求失败", err.Error(), "")
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewError(ErrCodeInternalError, "读取响应失败", err.Error(), "")
	}

	// 检查API错误（即使HTTP状态码不是200也尝试解析）
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Code != 0 {
		return nil, ConvertAPIError(&apiErr)
	}

	// 检查HTTP状态码 - 接受所有2xx成功状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ConvertHTTPError(resp.StatusCode, string(body))
	}

	return body, nil
}

// retryRequest 带重试的请求
func (c *FuturesClient) retryRequest(method, endpoint string, params map[string]string, needAuth bool) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.clientConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// 计算退避延迟
			delay := time.Duration(c.clientConfig.RetryDelay) * time.Millisecond
			for i := 1; i < attempt; i++ {
				delay *= time.Duration(c.clientConfig.RetryBackoff)
			}
			time.Sleep(delay)
		}

		body, err := c.doRequest(method, endpoint, params, needAuth)
		if err == nil {
			return body, nil
		}

		lastErr = err

		// 某些错误不需要重试
		if binanceErr, ok := err.(*Error); ok {
			if binanceErr.Code == ErrCodeInvalidRequest ||
				binanceErr.Code == ErrCodeInvalidSymbol ||
				binanceErr.Code == ErrCodeInvalidJSON ||
				binanceErr.Code == ErrCodeUnauthorized ||
				binanceErr.Code == ErrCodeAccountInactive {
				break
			}
		}
	}

	return nil, lastErr
}

// Ping 测试连接
func (c *FuturesClient) Ping() error {
	_, err := c.retryRequest("GET", "/fapi/v1/ping", nil, false)
	if err != nil {
		return err
	}
	return nil
}

// GetServerTime 获取服务器时间
func (c *FuturesClient) GetServerTime() (*ServerTime, error) {
	body, err := c.retryRequest("GET", "/fapi/v1/time", nil, false)
	if err != nil {
		return nil, err
	}

	var serverTime ServerTime
	if err := json.Unmarshal(body, &serverTime); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析服务器时间失败", err.Error(), string(body))
	}

	return &serverTime, nil
}

// GetExchangeInfo 获取交易规则和交易对信息
func (c *FuturesClient) GetExchangeInfo() ([]SymbolInfo, error) {
	body, err := c.retryRequest("GET", "/fapi/v1/exchangeInfo", nil, false)
	if err != nil {
		return nil, err
	}

	var response struct {
		Symbols []SymbolInfo `json:"symbols"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析交易信息失败", err.Error(), string(body))
	}

	return response.Symbols, nil
}

// Get24hrTicker 获取24小时价格变动统计
func (c *FuturesClient) Get24hrTicker(symbol Symbol) (*FuturesTicker, error) {
	params := map[string]string{
		"symbol": string(symbol),
	}

	body, err := c.retryRequest("GET", "/fapi/v1/ticker/24hr", params, false)
	if err != nil {
		return nil, err
	}

	var ticker FuturesTicker
	if err := json.Unmarshal(body, &ticker); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析价格统计失败", err.Error(), string(body))
	}

	return &ticker, nil
}

// GetDepth 获取深度信息
func (c *FuturesClient) GetDepth(symbol Symbol, limit DepthLevel) (*Depth, error) {
	params := map[string]string{
		"symbol": string(symbol),
		"limit":  strconv.Itoa(int(limit)),
	}

	body, err := c.retryRequest("GET", "/fapi/v1/depth", params, false)
	if err != nil {
		return nil, err
	}

	// 币安深度数据返回的是嵌套的字符串数组，需要特殊处理
	var rawDepth struct {
		LastUpdateID int64      `json:"lastUpdateId"`
		Bids         [][]string `json:"bids"`
		Asks         [][]string `json:"asks"`
	}

	if err := json.Unmarshal(body, &rawDepth); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析深度数据失败", err.Error(), string(body))
	}

	// 转换为Depth结构
	depth := &Depth{
		LastUpdateID: rawDepth.LastUpdateID,
		Bids:         make([]DepthEntry, 0, len(rawDepth.Bids)),
		Asks:         make([]DepthEntry, 0, len(rawDepth.Asks)),
	}

	// 转换买盘
	for _, bid := range rawDepth.Bids {
		if len(bid) >= 2 {
			depth.Bids = append(depth.Bids, DepthEntry{
				Price:    bid[0],
				Quantity: bid[1],
			})
		}
	}

	// 转换卖盘
	for _, ask := range rawDepth.Asks {
		if len(ask) >= 2 {
			depth.Asks = append(depth.Asks, DepthEntry{
				Price:    ask[0],
				Quantity: ask[1],
			})
		}
	}

	return depth, nil
}

// GetKlines 获取K线数据
func (c *FuturesClient) GetKlines(symbol Symbol, interval KlineInterval, limit int) ([]Kline, error) {
	params := map[string]string{
		"symbol":   string(symbol),
		"interval": string(interval),
	}

	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}

	body, err := c.retryRequest("GET", "/fapi/v1/klines", params, false)
	if err != nil {
		return nil, err
	}

	var rawKlines [][]interface{}
	if err := json.Unmarshal(body, &rawKlines); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析K线数据失败", err.Error(), string(body))
	}

	var klines []Kline
	for _, raw := range rawKlines {
		if len(raw) < 12 {
			continue
		}

		kline := Kline{
			OpenTime:                 int64(raw[0].(float64)),
			Open:                     raw[1].(string),
			High:                     raw[2].(string),
			Low:                      raw[3].(string),
			Close:                    raw[4].(string),
			Volume:                   raw[5].(string),
			CloseTime:                int64(raw[6].(float64)),
			QuoteAssetVolume:         raw[7].(string),
			TradeNum:                 int64(raw[8].(float64)),
			TakerBuyBaseAssetVolume:  raw[9].(string),
			TakerBuyQuoteAssetVolume: raw[10].(string),
		}

		klines = append(klines, kline)
	}

	// 排除最后一条正在形成的K线（数据不完整）
	// 特别是在K线周期即将结束时，成交量等数据会显示异常值如"2.333"
	if len(klines) > 1 {
		klines = klines[:len(klines)-1]
	}

	return klines, nil
}

// GetKlines1h 获取1小时K线数据
func (c *FuturesClient) GetKlines1h(symbol Symbol, limit int) ([]Kline, error) {
	return c.GetKlines(symbol, KlineInterval1h, limit)
}

// GetKlines4h 获取4小时K线数据
func (c *FuturesClient) GetKlines4h(symbol Symbol, limit int) ([]Kline, error) {
	return c.GetKlines(symbol, KlineInterval4h, limit)
}

// GetRecentTrades 获取最近交易记录
func (c *FuturesClient) GetRecentTrades(symbol Symbol, limit int) ([]RecentTrade, error) {
	params := map[string]string{
		"symbol": string(symbol),
	}

	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}

	body, err := c.retryRequest("GET", "/fapi/v1/trades", params, false)
	if err != nil {
		return nil, err
	}

	var trades []RecentTrade
	if err := json.Unmarshal(body, &trades); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析交易记录失败", err.Error(), string(body))
	}

	return trades, nil
}

// GetMarkPrice 获取标记价格
func (c *FuturesClient) GetMarkPrice(symbol Symbol) (*MarkPrice, error) {
	params := map[string]string{}

	if symbol != "" {
		params["symbol"] = string(symbol)
	}

	body, err := c.retryRequest("GET", "/fapi/v1/premiumIndex", params, false)
	if err != nil {
		return nil, err
	}

	if symbol != "" {
		var markPrice MarkPrice
		if err := json.Unmarshal(body, &markPrice); err != nil {
			return nil, NewError(ErrCodeInvalidJSON, "解析标记价格失败", err.Error(), string(body))
		}
		return &markPrice, nil
	} else {
		// 返回多个交易对的标记价格
		var markPrices []MarkPrice
		if err := json.Unmarshal(body, &markPrices); err != nil {
			return nil, NewError(ErrCodeInvalidJSON, "解析标记价格列表失败", err.Error(), string(body))
		}

		// 查找ETH相关的交易对
		for _, mp := range markPrices {
			if strings.Contains(mp.Symbol, "ETH") {
				return &mp, nil
			}
		}
		return nil, NewError(ErrCodeInvalidSymbol, "未找到ETH相关交易对", "", string(body))
	}
}

// GetLatestFundingRate 获取最新资金费率
func (c *FuturesClient) GetLatestFundingRate(symbol Symbol) (*FundingRateHistory, error) {
	params := map[string]string{
		"symbol": string(symbol),
	}

	body, err := c.retryRequest("GET", "/fapi/v1/fundingRate", params, false)
	if err != nil {
		return nil, err
	}

	var fundingHistory []FundingRateHistory
	if err := json.Unmarshal(body, &fundingHistory); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析资金费率历史失败", err.Error(), string(body))
	}

	// 返回最新的费率（第一个是最近的）
	if len(fundingHistory) > 0 {
		return &fundingHistory[0], nil
	}

	return nil, NewError(ErrCodeInvalidSymbol, "未找到资金费率数据", "", "")
}

// GetFundingRateHistory 获取资金费率历史
func (c *FuturesClient) GetFundingRateHistory(symbol Symbol, limit int, startTime, endTime int64) ([]FundingRateHistory, error) {
	params := map[string]string{
		"symbol": string(symbol),
	}

	// 如果未提供limit，默认使用100
	if limit <= 0 {
		limit = 100
	}
	params["limit"] = strconv.Itoa(limit)

	if startTime > 0 {
		params["startTime"] = strconv.FormatInt(startTime, 10)
	}

	if endTime > 0 {
		params["endTime"] = strconv.FormatInt(endTime, 10)
	}

	body, err := c.retryRequest("GET", "/fapi/v1/fundingRate", params, false)
	if err != nil {
		return nil, err
	}

	var fundingHistory []FundingRateHistory //时间戳正序，最早的在前
	if err := json.Unmarshal(body, &fundingHistory); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析资金费率历史失败", err.Error(), string(body))
	}

	return fundingHistory, nil
}

// GetOpenInterest 获取持仓量（Open Interest）
func (c *FuturesClient) GetOpenInterest(symbol Symbol) (*OpenInterest, error) {
	params := map[string]string{
		"symbol": string(symbol),
	}

	body, err := c.retryRequest("GET", "/fapi/v1/openInterest", params, false)
	if err != nil {
		return nil, err
	}

	var oi OpenInterest
	if err := json.Unmarshal(body, &oi); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析持仓量失败", err.Error(), string(body))
	}

	return &oi, nil
}

// GetAccountInfo 获取合约账户信息（需要API密钥）
func (c *FuturesClient) GetAccountInfo() (*FuturesAccountInfo, error) {
	params := map[string]string{}

	body, err := c.retryRequest("GET", "/fapi/v2/account", params, true)
	if err != nil {
		return nil, err
	}

	var accountInfo FuturesAccountInfo
	if err := json.Unmarshal(body, &accountInfo); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析合约账户信息失败", err.Error(), string(body))
	}

	return &accountInfo, nil
}

// GetPositions 获取持仓信息（需要API密钥）
func (c *FuturesClient) GetPositions(symbol Symbol) ([]Position, error) {
	params := map[string]string{}

	if symbol != "" {
		params["symbol"] = string(symbol)
	}

	body, err := c.retryRequest("GET", "/fapi/v2/positionRisk", params, true)
	if err != nil {
		return nil, err
	}

	var positions []Position
	if err := json.Unmarshal(body, &positions); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析持仓信息失败", err.Error(), string(body))
	}

	return positions, nil
}

// SetLeverage 设置杠杆倍数（需要API密钥）
func (c *FuturesClient) SetLeverage(symbol Symbol, leverage int) error {
	params := map[string]string{
		"symbol":   string(symbol),
		"leverage": strconv.Itoa(leverage),
	}

	_, err := c.retryRequest("POST", "/fapi/v1/leverage", params, true)
	if err != nil {
		return err
	}

	return nil
}

// SetMarginType 设置保证金模式（需要API密钥）
func (c *FuturesClient) SetMarginType(symbol Symbol, marginType MarginType) error {
	params := map[string]string{
		"symbol":     string(symbol),
		"marginType": string(marginType),
	}

	_, err := c.retryRequest("POST", "/fapi/v1/marginType", params, true)
	if err != nil {
		return err
	}

	return nil
}

// SetPositionMode 设置持仓模式（需要API密钥）
func (c *FuturesClient) SetPositionMode(dualSidePosition bool) error {
	params := map[string]string{
		"dualSidePosition": "false",
	}

	if dualSidePosition {
		params["dualSidePosition"] = "true"
	}

	_, err := c.retryRequest("POST", "/fapi/v1/positionSide/dual", params, true)
	if err != nil {
		return err
	}

	return nil
}

// GetPositionMode 获取当前持仓模式（需要API密钥）
func (c *FuturesClient) GetPositionMode() (bool, error) {
	body, err := c.retryRequest("GET", "/fapi/v1/positionSide/dual", nil, true)
	if err != nil {
		return false, err
	}

	var result struct {
		DualSidePosition bool `json:"dualSidePosition"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false, NewError(ErrCodeInvalidJSON, "解析持仓模式失败", err.Error(), string(body))
	}

	return result.DualSidePosition, nil
}

// NewOrder 下新订单（需要API密钥）
func (c *FuturesClient) NewOrder(req *NewOrderRequest, positionSide PositionSide) (*Order, error) {
	params := map[string]string{
		"symbol": string(req.Symbol),
		"side":   string(req.Side),
		"type":   string(req.Type),
	}

	if req.Quantity != "" {
		params["quantity"] = req.Quantity
	}

	if req.Price != "" {
		params["price"] = req.Price
	}

	if req.TimeInForce != "" {
		params["timeInForce"] = string(req.TimeInForce)
	}

	if req.StopPrice != "" {
		params["stopPrice"] = req.StopPrice
	}

	if positionSide != "" {
		params["positionSide"] = string(positionSide)
	}

	// 期货专用参数
	if req.ReduceOnly {
		params["reduceOnly"] = "true"
	}
	if req.ClosePosition {
		params["closePosition"] = "true"
	}
	if req.WorkingType != "" {
		params["workingType"] = string(req.WorkingType)
	}

	body, err := c.retryRequest("POST", "/fapi/v1/order", params, true)
	if err != nil {
		return nil, err
	}

	var order Order
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析订单信息失败", err.Error(), string(body))
	}

	return &order, nil
}

// GetOrder 查询订单（需要API密钥）
func (c *FuturesClient) GetOrder(symbol Symbol, orderId int64, origClientOrderId string) (*Order, error) {
	params := map[string]string{
		"symbol": string(symbol),
	}

	if orderId > 0 {
		params["orderId"] = strconv.FormatInt(orderId, 10)
	}

	if origClientOrderId != "" {
		params["origClientOrderId"] = origClientOrderId
	}

	body, err := c.retryRequest("GET", "/fapi/v1/order", params, true)
	if err != nil {
		return nil, err
	}

	var order Order
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析订单信息失败", err.Error(), string(body))
	}

	return &order, nil
}

// CancelOrder 取消订单（需要API密钥）
func (c *FuturesClient) CancelOrder(symbol Symbol, orderId int64, origClientOrderId string) (*Order, error) {
	params := map[string]string{
		"symbol": string(symbol),
	}

	if orderId > 0 {
		params["orderId"] = strconv.FormatInt(orderId, 10)
	}

	if origClientOrderId != "" {
		params["origClientOrderId"] = origClientOrderId
	}

	body, err := c.retryRequest("DELETE", "/fapi/v1/order", params, true)
	if err != nil {
		return nil, err
	}

	var order Order
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析取消订单信息失败", err.Error(), string(body))
	}

	return &order, nil
}

// CancelAllOpenOrders 取消所有挂单（需要API密钥）
func (c *FuturesClient) CancelAllOpenOrders(symbol Symbol) ([]Order, error) {
	params := map[string]string{
		"symbol": string(symbol),
	}

	body, err := c.retryRequest("DELETE", "/fapi/v1/allOpenOrders", params, true)
	if err != nil {
		return nil, err
	}

	var orders []Order
	if err := json.Unmarshal(body, &orders); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析取消所有订单信息失败", err.Error(), string(body))
	}

	return orders, nil
}

// GetOpenOrders 查询当前挂单（需要API密钥）
func (c *FuturesClient) GetOpenOrders(symbol Symbol) ([]Order, error) {
	params := map[string]string{}

	if symbol != "" {
		params["symbol"] = string(symbol)
	}

	body, err := c.retryRequest("GET", "/fapi/v1/openOrders", params, true)
	if err != nil {
		return nil, err
	}

	var orders []Order
	if err := json.Unmarshal(body, &orders); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析挂单列表失败", err.Error(), string(body))
	}

	return orders, nil
}

// GetOrderHistory 查询所有订单（需要API密钥）
func (c *FuturesClient) GetOrderHistory(symbol Symbol, limit int, orderId, startTime, endTime int64) ([]Order, error) {
	params := map[string]string{
		"symbol": string(symbol),
	}

	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}

	if orderId > 0 {
		params["orderId"] = strconv.FormatInt(orderId, 10)
	}

	if startTime > 0 {
		params["startTime"] = strconv.FormatInt(startTime, 10)
	}

	if endTime > 0 {
		params["endTime"] = strconv.FormatInt(endTime, 10)
	}

	body, err := c.retryRequest("GET", "/fapi/v1/allOrders", params, true)
	if err != nil {
		return nil, err
	}

	var orders []Order
	if err := json.Unmarshal(body, &orders); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析历史订单失败", err.Error(), string(body))
	}

	return orders, nil
}

// GetUserTrades 获取用户成交记录（需要API密钥）
func (c *FuturesClient) GetUserTrades(symbol Symbol, limit int, orderId, startTime, endTime int64) ([]UserTrade, error) {
	params := map[string]string{
		"symbol": string(symbol),
	}

	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}

	if orderId > 0 {
		params["orderId"] = strconv.FormatInt(orderId, 10)
	}

	if startTime > 0 {
		params["startTime"] = strconv.FormatInt(startTime, 10)
	}

	if endTime > 0 {
		params["endTime"] = strconv.FormatInt(endTime, 10)
	}

	body, err := c.retryRequest("GET", "/fapi/v1/userTrades", params, true)
	if err != nil {
		return nil, err
	}

	var trades []UserTrade
	if err := json.Unmarshal(body, &trades); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析用户成交记录失败", err.Error(), string(body))
	}

	return trades, nil
}

// GetIncomeHistory 获取收入历史（需要API密钥）
func (c *FuturesClient) GetIncomeHistory(symbol string, incomeType string, limit int, startTime, endTime int64) ([]FundingRateHistory, error) {
	params := map[string]string{}

	if symbol != "" {
		params["symbol"] = symbol
	}

	if incomeType != "" {
		params["incomeType"] = incomeType
	}

	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}

	if startTime > 0 {
		params["startTime"] = strconv.FormatInt(startTime, 10)
	}

	if endTime > 0 {
		params["endTime"] = strconv.FormatInt(endTime, 10)
	}

	body, err := c.retryRequest("GET", "/fapi/v1/income", params, true)
	if err != nil {
		return nil, err
	}

	var incomeHistory []FundingRateHistory
	if err := json.Unmarshal(body, &incomeHistory); err != nil {
		return nil, NewError(ErrCodeInvalidJSON, "解析收入历史失败", err.Error(), string(body))
	}

	return incomeHistory, nil
}

// GetTopLongShortPositionRatio 获取大户持仓量多空比
func (c *FuturesClient) GetTopLongShortPositionRatio(symbol Symbol, period string, limit int) ([]TopLongShortPositionRatio, error) {
	params := map[string]interface{}{
		"symbol": string(symbol),
		"period": period,
	}

	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}
	var ratios []TopLongShortPositionRatio
	client := utils.GetProxyHTTPClient(conf.Get().Binance.DefaultProxy, conf.Get().Binance.Timeout)
	resp := requests.NewHTTPRequest("https://fapi.binance.com/futures/data/topLongShortPositionRatio").SetClient(client).SetQueryParams(params).ToJSON(&ratios)

	return ratios, resp.Error
}

// GetBookTicker 获取当前最优挂单信息
func (c *FuturesClient) GetBookTicker(symbol Symbol) (*BookTicker, error) {
	params := map[string]string{}

	if symbol != "" {
		params["symbol"] = string(symbol)
	}

	body, err := c.retryRequest("GET", "/fapi/v1/ticker/bookTicker", params, false)
	if err != nil {
		return nil, err
	}

	if symbol != "" {
		// 单个交易对的情况，返回单个对象
		var bookTicker BookTicker
		if err := json.Unmarshal(body, &bookTicker); err != nil {
			return nil, NewError(ErrCodeInvalidJSON, "解析最优挂单信息失败", err.Error(), string(body))
		}
		return &bookTicker, nil
	} else {
		// 没有指定交易对的情况，返回数组
		var bookTickers []BookTicker
		if err := json.Unmarshal(body, &bookTickers); err != nil {
			return nil, NewError(ErrCodeInvalidJSON, "解析最优挂单列表失败", err.Error(), string(body))
		}

		// 如果返回了多个交易对，默认返回ETH相关的
		if len(bookTickers) > 0 {
			for _, ticker := range bookTickers {
				if strings.Contains(ticker.Symbol, "ETH") {
					return &ticker, nil
				}
			}
			// 如果没有找到ETH相关的，返回第一个
			return &bookTickers[0], nil
		}

		return nil, NewError(ErrCodeInvalidSymbol, "未找到最优挂单数据", "", string(body))
	}
}

// GetTopLongShortAccountRatio 获取大户账户数多空比
func (c *FuturesClient) GetTopLongShortAccountRatio(symbol Symbol, period string, limit int) ([]TopLongShortAccountRatio, error) {
	params := map[string]interface{}{
		"symbol": string(symbol),
		"period": period,
	}
	if limit > 0 {
		params["limit"] = strconv.Itoa(limit)
	}

	var ratios []TopLongShortAccountRatio
	client := utils.GetProxyHTTPClient(conf.Get().Binance.DefaultProxy, conf.Get().Binance.Timeout)
	resp := requests.NewHTTPRequest("https://fapi.binance.com/futures/data/topLongShortAccountRatio").SetClient(client).SetQueryParams(params).ToJSON(&ratios)
	return ratios, resp.Error
}
