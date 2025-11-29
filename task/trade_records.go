package task

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"deeptrade/binance"
	"deeptrade/utils"
)

// TradeRecord 交易记录结构
type TradeRecord struct {
	OrderID         int64               `json:"order_id"`         // 订单ID
	Symbol          string              `json:"symbol"`           // 交易对
	Side            binance.OrderSide   `json:"side"`             // 买卖方向
	Type            binance.OrderType   `json:"type"`             // 订单类型
	Quantity        float64             `json:"quantity"`         // 成交数量
	Price           string              `json:"price"`            // 成交价格
	ExecutedQty     float64             `json:"executed_qty"`     // 已执行数量
	CumulativeQuote float64             `json:"cumulative_quote"` // 累计成交金额
	Status          binance.OrderStatus `json:"status"`           // 订单状态
	Time            time.Time           `json:"time"`             // 创建时间
	UpdateTime      time.Time           `json:"update_time"`      // 更新时间
	PositionSide    string              `json:"position_side"`    // 持仓方向
	TradeType       string              `json:"trade_type"`       // 交易类型描述
	RealizedPnl     string              // 已实现盈亏
	Commission      string              // 手续费
	CommissionAsset string              // 手续费资产
}

// GetTradeRecordsFromMarketData 从MarketData中的OrderHistory生成交易记录（推荐使用）
func GetTradeRecordsFromMarketData(marketData *MarketData, limit int) []*TradeRecord {
	if limit <= 0 {
		limit = 6
	}

	if marketData == nil || len(marketData.OrderHistory) == 0 {
		return []*TradeRecord{}
	}

	// 使用MarketData中已有的持仓信息
	return ProcessOrderHistoryToTradeRecords(marketData.OrderHistory, marketData.Positions, limit)
}

// ProcessOrderHistoryToTradeRecords 处理订单历史数据，转换为交易记录
func ProcessOrderHistoryToTradeRecords(orders []binance.Order, positions []binance.Position, limit int) []*TradeRecord {
	client, _ := binance.GetFuturesClient()
	// 转换为交易记录并过滤已成交的订单
	var tradeRecords []*TradeRecord
	// 按时间排序（最新的在前，即时间倒序）
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].Time > orders[j].Time
	})

	for _, order := range orders {
		// 只处理已完全成交的实际交易订单
		if order.Status != binance.OrderStatusFilled {
			continue
		}

		// 过滤掉止损单、止盈单等非实际交易订单
		if order.Type == binance.OrderTypeStopMarket || order.Type == binance.OrderTypeTakeProfitMarket {
			continue
		}

		// 解析数量和价格
		qty := utils.ParseFloatSafe(order.ExecutedQty, 0)
		if qty <= 0 {
			continue
		}

		// 解析成交金额
		cumulative := parseFloat(order.CumulativeQuoteQty)

		tradeType := getTradeTypeDescription(order.Side, order.PositionSide, order.Type)
		realizedPnl, commission, commissionAsset, price := GetgetTradeRealizedPnl(client, order.OrderID)
		record := &TradeRecord{
			OrderID:         order.OrderID,
			Symbol:          order.Symbol,
			Side:            order.Side,
			Type:            order.Type,
			Quantity:        qty,
			Price:           price,
			ExecutedQty:     qty,
			CumulativeQuote: cumulative,
			Status:          order.Status,
			Time:            time.Unix(0, order.Time*int64(time.Millisecond)),
			UpdateTime:      time.Unix(0, order.UpdateTime*int64(time.Millisecond)),
			PositionSide:    order.PositionSide,
			TradeType:       tradeType,
			RealizedPnl:     realizedPnl,
			Commission:      commission,
			CommissionAsset: commissionAsset,
		}

		tradeRecords = append(tradeRecords, record)
		if len(tradeRecords) >= limit {
			break //满足后跳出
		}
	}

	// 限制返回数量
	if len(tradeRecords) > limit {
		tradeRecords = tradeRecords[:limit]
	}

	return tradeRecords
}

func GetgetTradeRealizedPnl(client *binance.FuturesClient, orderId int64) (realizedPnl, commission, commissionAsset, price string) {
	if client == nil {
		return
	}
	list, err := client.GetUserTrades(binance.ETHUSDT, 20, orderId, 0, 0)
	if err != nil {
		return
	}
	if len(list) == 0 {
		return
	}

	// 实现list数组的遍历，realizedPnl需要累计，commission需要累计，price需要平均值
	var totalRealizedPnl, totalCommission float64
	var totalPrice float64
	var count int

	for _, trade := range list {
		// 累计已实现盈亏
		totalRealizedPnl += utils.ParseFloatSafe(trade.RealizedPnl, 0)

		// 累计手续费
		totalCommission += utils.ParseFloatSafe(trade.Commission, 0)

		// 累计价格用于计算平均值
		totalPrice += utils.ParseFloatSafe(trade.Price, 0)

		// 记录手续费资产（使用第一个交易的资产）
		if commissionAsset == "" {
			commissionAsset = trade.CommissionAsset
		}

		count++
	}

	// 计算平均价格
	var avgPrice float64
	if count > 0 {
		avgPrice = totalPrice / float64(count)
	}

	// 转换结果为字符串
	realizedPnl = fmt.Sprintf("%g", totalRealizedPnl)
	commission = fmt.Sprintf("%g", totalCommission)
	price = fmt.Sprintf("%g", avgPrice)
	if totalRealizedPnl < 0.00000001 && totalRealizedPnl > -0.00000001 {
		realizedPnl = ""
	}
	return
}

// getTradeTypeDescription 获取交易类型描述
func getTradeTypeDescription(side binance.OrderSide, positionSide string, orderType binance.OrderType) string {
	if orderType == binance.OrderTypeStopMarket {
		return "止损单"
	}
	if orderType == binance.OrderTypeTakeProfitMarket {
		return "止盈单"
	}
	if positionSide == string(binance.PositionSideLong) {
		if side == binance.OrderSideSell {
			return "平多仓"
		} else {
			return "开多仓"
		}
	} else if positionSide == string(binance.PositionSideShort) {
		if side == binance.OrderSideBuy {
			return "平空仓"
		} else {
			return "开空仓"
		}
	}

	// 默认描述
	if side == binance.OrderSideBuy {
		return "买入"
	} else {
		return "卖出"
	}
}

// FormatTradeRecords 格式化交易记录输出
func FormatTradeRecords(records []*TradeRecord) string {
	if len(records) == 0 {
		return "暂无交易记录"
	}

	result := "─────────────────────────────────────────────────────────\n"

	for i, record := range records {
		status := "✅ 已成交"

		result += fmt.Sprintf("%d. %s %s\n", i+1, record.TradeType, status)
		result += fmt.Sprintf("   订单ID: %d\n", record.OrderID)
		// 价格显示优化
		var priceDisplay string
		if record.Price != "" {
			priceDisplay = fmt.Sprintf("%s USDT", toFixed3(record.Price))
		} else {
			priceDisplay = "市价成交"
		}

		result += fmt.Sprintf("   数量: %.3f ETH, 成交价: %s\n", record.Quantity, priceDisplay)
		result += fmt.Sprintf("   方向: %s\n", record.Side)
		if record.PositionSide != "" {
			result += fmt.Sprintf("   持仓方向: %s\n", record.PositionSide)
		}
		if record.RealizedPnl != "" {
			result += fmt.Sprintf("   已实现盈亏:%s %s, 手续费:%s %s\n", toFixed3(record.RealizedPnl), record.CommissionAsset, toFixed3(record.Commission), record.CommissionAsset)
		}
		result += fmt.Sprintf("   时间: %s\n", record.Time.Format("2006-01-02 15:04:05"))
		result += "─────────────────────────────────────────────────────────\n"
	}

	return result
}

// 辅助函数
func parseFloat(s string) float64 {
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}
	return 0
}

// 把任意字符串形式的浮点数转成保留 3 位小数的字符串
func toFixed3(s string) string {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%.3f", f)
}
