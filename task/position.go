package task

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"deeptrade/binance"
)

// PositionWithTime 带时间戳的持仓记录
type PositionWithTime struct {
	binance.Position
	recordTime time.Time // 记录时间（PositionCache的时间）
}

type PositionCache struct {
	list []PositionWithTime
}

var (
	positionQueueMutex sync.Mutex
	positionQueue      []PositionCache
	positionStopChan   chan struct{}
)

// GetPositionInfo 获取持仓信息
func GetPositionInfo(positions []binance.Position) *PositionInfo {
	info := &PositionInfo{IsDualSide: true}

	// 双向持仓模式：分别统计LONG和SHORT
	for _, pos := range positions {
		amt, err := strconv.ParseFloat(pos.PositionAmt, 64)
		if err != nil {
			continue
		}

		if amt < 0 {
			amt = -amt // 取绝对值
		}

		if pos.PositionSide == binance.PositionSideLong && amt > 0 {
			info.HasLong = true
			info.LongAmt = amt
		}
		if pos.PositionSide == binance.PositionSideShort && amt > 0 {
			info.HasShort = true
			info.ShortAmt = amt
		}

		if amt > 0 {
			positionTime := time.Unix(pos.UpdateTime/1000, 0)
			now := time.Now()
			duration := now.Sub(positionTime)
			info.Duration = duration
			info.UnRealizedProfit = pos.UnRealizedProfit
		}
	}
	return info
}

// HasRealPosition 检查是否有实际持仓
func HasRealPosition(positions []binance.Position) bool {
	for _, pos := range positions {
		positionAmt, err := strconv.ParseFloat(pos.PositionAmt, 64)
		if err == nil && positionAmt != 0 {
			return true
		}
	}
	return false
}

// AccountBalanceInfo 账户余额信息
type AccountBalanceInfo struct {
	WalletBalance    float64 // 钱包余额
	AvailableBalance float64 // 可用余额
	MarginBalance    float64 // 保证金余额
}

// GetAccountBalanceInfo 获取账户余额信息
func GetAccountBalanceInfo(data *MarketData) *AccountBalanceInfo {
	info := &AccountBalanceInfo{}

	// 解析钱包余额
	if balance, err := strconv.ParseFloat(data.Account.TotalWalletBalance, 64); err == nil {
		info.WalletBalance = balance
	}

	// 解析可用余额
	if balance, err := strconv.ParseFloat(data.Account.AvailableBalance, 64); err == nil {
		info.AvailableBalance = balance
	}

	// 解析保证金余额
	if balance, err := strconv.ParseFloat(data.Account.TotalMarginBalance, 64); err == nil {
		info.MarginBalance = balance
	}
	return info
}

// ValidatePositionForClose 验证平仓操作的持仓条件
func ValidatePositionForClose(action string, positionInfo *PositionInfo) error {
	switch action {
	case "CLOSE_LONG":
		if positionInfo.IsDualSide {
			if positionInfo.LongAmt <= 0 {
				return fmt.Errorf("无多头可平")
			}
		} else {
			if positionInfo.NetAmt <= 0 {
				return fmt.Errorf("无多头可平")
			}
		}
	case "CLOSE_SHORT":
		if positionInfo.IsDualSide {
			if positionInfo.ShortAmt <= 0 {
				return fmt.Errorf("无空头可平")
			}
		} else {
			if positionInfo.NetAmt >= 0 {
				return fmt.Errorf("无空头可平")
			}
		}
	default:
		return fmt.Errorf("未知的平仓操作类型: %s", action)
	}
	return nil
}

// GetCloseQuantity 获取平仓数量
func GetCloseQuantity(action string, positionInfo *PositionInfo) (float64, error) {
	switch action {
	case "CLOSE_LONG":
		if positionInfo.IsDualSide {
			return positionInfo.LongAmt, nil
		} else {
			return positionInfo.NetAmt, nil
		}
	case "CLOSE_SHORT":
		if positionInfo.IsDualSide {
			return positionInfo.ShortAmt, nil
		} else {
			return -positionInfo.NetAmt, nil
		}
	default:
		return 0, fmt.Errorf("未知的平仓操作类型: %s", action)
	}
}

// FormatPositionWithSLTP 格式化持仓信息并包含止损止盈
func FormatPositionWithSLTP(positions []binance.Position, orders []binance.Order) string {
	var analysis strings.Builder
	// 创建订单映射，方便查找止损止盈订单
	stopLossOrders := make(map[string]*binance.Order)
	takeProfitOrders := make(map[string]*binance.Order)

	for i := range orders {
		order := &orders[i]
		if order.Type == binance.OrderTypeStopMarket || order.Type == binance.OrderTypeStop {
			stopLossOrders[string(order.PositionSide)] = order
		} else if order.Type == binance.OrderTypeTakeProfitMarket || order.Type == binance.OrderTypeTakeProfit {
			takeProfitOrders[string(order.PositionSide)] = order
		}
	}

	realPositions := 0
	for _, pos := range positions {
		positionAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)
		if positionAmt != 0 {
			realPositions++
		}
	}

	if realPositions > 0 {
		analysis.WriteString("【双向持仓模式】当前持仓信息:\n")
		for _, pos := range positions {
			positionAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)
			if positionAmt != 0 {
				// 计算并显示盈亏比例
				entryPrice, _ := strconv.ParseFloat(pos.EntryPrice, 64)
				markPrice, _ := strconv.ParseFloat(pos.MarkPrice, 64)
				leverage, _ := strconv.ParseFloat(pos.Leverage, 64)

				var pnlPercent float64
				if entryPrice > 0 { // 避免除零错误
					if pos.PositionSide == binance.PositionSideLong {
						pnlPercent = ((markPrice - entryPrice) / entryPrice) * leverage * 100
					} else {
						pnlPercent = ((entryPrice - markPrice) / entryPrice) * leverage * 100
					}
				}
				// 计算持仓持续时间
				positionDuration := calculatePositionDuration(pos.UpdateTime)

				analysis.WriteString(fmt.Sprintf("  持仓数量: %s, 方向: %s\n", pos.PositionAmt, pos.PositionSide))
				analysis.WriteString(fmt.Sprintf("  开仓价: %s\n", pos.EntryPrice))
				analysis.WriteString(fmt.Sprintf("  标记价: %s\n", pos.MarkPrice))
				analysis.WriteString(fmt.Sprintf("  杠杆倍率: %s\n", pos.Leverage))
				analysis.WriteString(fmt.Sprintf("  盈亏比例: %.2f%%\n", pnlPercent))
				analysis.WriteString(fmt.Sprintf("  当前未实现盈亏: %s\n", pos.UnRealizedProfit))
				analysis.WriteString(fmt.Sprintf("  持仓时间: %s\n", positionDuration))

				// 添加止损止盈信息
				positionSide := string(pos.PositionSide)
				if slOrder, exists := stopLossOrders[positionSide]; exists {
					setTime := formatOrderTime(slOrder.Time)
					slPrice, _ := strconv.ParseFloat(slOrder.StopPrice, 64)
					analysis.WriteString(fmt.Sprintf("   🛑 止损: %.2f (设置时间: %s)\n", slPrice, setTime))
				} else {
					analysis.WriteString("   🛑 止损: 未设置\n")
				}

				if tpOrder, exists := takeProfitOrders[positionSide]; exists {
					setTime := formatOrderTime(tpOrder.Time)
					tpPrice, _ := strconv.ParseFloat(tpOrder.StopPrice, 64)
					analysis.WriteString(fmt.Sprintf("   🎯 止盈: %.2f (设置时间: %s)\n", tpPrice, setTime))
				} else {
					analysis.WriteString("   🎯 止盈: 未设置\n")
				}

				// 添加近期未实现盈亏快照
				analysis.WriteString("   📊 未实现盈亏快照:\n")
				history := GetPositionHistory(pos)
				// 最多显示8条最新记录（最新的在最后）
				start := 0
				if len(history) > 10 {
					start = len(history) - 10
				}
				for i := start; i < len(history); i++ {
					histPos := history[i]
					pnl, _ := strconv.ParseFloat(histPos.UnRealizedProfit, 64)
					updateTime := histPos.recordTime.Format("15:04:05")
					analysis.WriteString(fmt.Sprintf("     • 盈亏: %.4f, 时间: %s\n", pnl, updateTime))
				}
				if len(history) == 0 {
					analysis.WriteString("     • 暂无历史记录\n")
				}

			}
		}
	} else {
		analysis.WriteString("【双向持仓模式】当前无实际持仓\n")
	}

	return analysis.String()
}

// calculatePositionDuration 计算持仓持续时间
func calculatePositionDuration(updateTime int64) string {
	if updateTime == 0 {
		return "未知"
	}

	// 将毫秒时间戳转换为时间
	positionTime := time.Unix(updateTime/1000, 0)
	now := time.Now()
	duration := now.Sub(positionTime)

	// 格式化持续时间
	if duration.Hours() < 1 {
		return fmt.Sprintf("%.0f分钟", duration.Minutes())
	} else if duration.Hours() < 24 {
		return fmt.Sprintf("%.1f小时", duration.Hours())
	} else {
		days := int(duration.Hours() / 24)
		hours := int(duration.Hours()) % 24
		if hours > 0 {
			return fmt.Sprintf("%d天%d小时", days, hours)
		}
		return fmt.Sprintf("%d天", days)
	}
}

// formatOrderTime 格式化订单时间
func formatOrderTime(orderTime int64) string {
	if orderTime == 0 {
		return "未知"
	}

	// 将毫秒时间戳转换为时间
	t := time.Unix(orderTime/1000, 0)
	return t.Format("01-02 15:04")
}

// GetPositionsWithSLTP 获取包含止损止盈的持仓信息
func GetPositionsWithSLTP() (string, error) {
	// 获取期货客户端
	client, err := binance.GetFuturesClient()
	if err != nil {
		return "", fmt.Errorf("创建期货客户端失败: %v", err)
	}

	symbol := binance.ETHUSDT_PERP

	// 获取持仓信息
	positions, err := client.GetPositions(symbol)
	if err != nil {
		return "", fmt.Errorf("获取持仓信息失败: %v", err)
	}

	// 获取当前挂单信息
	orders, err := client.GetOpenOrders(symbol)
	if err != nil {
		return "", fmt.Errorf("获取挂单信息失败: %v", err)
	}

	// 格式化输出
	return FormatPositionWithSLTP(positions, orders), nil
}

// GetPositionHistory 持仓历史
func GetPositionHistory(pos binance.Position) (result []PositionWithTime) {
	positionQueueMutex.Lock()
	defer positionQueueMutex.Unlock()
	for _, v := range positionQueue {
		for _, p := range v.list {
			if pos.Symbol != p.Symbol || pos.PositionSide != p.PositionSide {
				continue
			}
			result = append(result, p)
		}
	}
	return
}

func CloseFetchPosition() {
	positionQueueMutex.Lock()
	defer positionQueueMutex.Unlock()
	// 关闭通道以发送停止信号
	if positionStopChan != nil {
		close(positionStopChan)
		positionStopChan = nil
	}

	positionQueue = []PositionCache{}
}

func StartFetchPosition() {
	positionQueueMutex.Lock()
	if positionStopChan != nil {
		positionQueueMutex.Unlock()
		return //正在执行中
	}
	positionQueue = []PositionCache{}
	positionStopChan = make(chan struct{})
	positionQueueMutex.Unlock()
	go func() {
		for {
			client := binance.GetOnceFuturesClient()
			pos, err := client.GetPositions(binance.ETHUSDT_PERP)
			if err != nil {
				log.Println(err)
			}
			positionQueueMutex.Lock()
			posinfo := GetPositionInfo(pos)
			if !posinfo.HasLong && !posinfo.HasShort {
				//如果获取的持仓没有数量，说明已经被止盈止损了
				close(positionStopChan)
				positionStopChan = nil
				log.Println("[量化交易] 未获取到持仓盈亏,关闭拉取持仓信息")
				positionQueueMutex.Unlock()
				setOffSystem(posinfo)
				return
			}

			poswt := []PositionWithTime{}
			for _, p := range pos {
				poswt = append(poswt, PositionWithTime{
					Position:   p,
					recordTime: time.Now(),
				})
			}
			positionQueue = append(positionQueue, PositionCache{
				list: poswt,
			})
			if len(positionQueue) > 15 {
				positionQueue = positionQueue[1:]
			}
			positionQueueMutex.Unlock()
			side := "多头"
			if posinfo.HasShort {
				side = "空头"
			}
			log.Printf("[量化交易] 拉取持仓信息成功 方向 :%v 未实现盈亏: %v\n", side, posinfo.UnRealizedProfit)

			// 使用select实现实时关闭功能
			select {
			case <-positionStopChan:
				// 收到停止信号，退出循环
				log.Println("[量化交易] 关闭拉取持仓信息")
				return
			case <-time.After(3 * time.Minute):
				// 默认等待2分钟后继续执行
			}
		}
	}()
}
