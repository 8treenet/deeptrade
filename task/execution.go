package task

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"

	"deeptrade/binance"
	"deeptrade/indicators"
	"deeptrade/utils"
)

// ExecuteTrade 执行交易（同时支持单向/双向持仓）
func ExecuteTrade(signal *TradingSignal, marketData *MarketData) error {
	log.Printf("[交易执行] 准备执行交易: %s", signal.Action)

	// 基本校验
	if signal == nil {
		return fmt.Errorf("交易信号为空")
	}
	if signal.Action == "HOLD" {
		log.Println("[交易执行] 信号为HOLD，跳过交易")
		return nil
	}
	if utils.InSlice([]string{"CLOSE_LONG", "CLOSE_SHORT", "ADJUST_SL_TP"}, signal.Action) {
		//平仓调仓需要重新拉取持仓，llm处理时间较长可能已经被止损止盈。
		marketData.Positions, _ = binance.GetOnceFuturesClient().GetPositions(binance.ETHUSDT_PERP)
		marketData.PositionInfo = GetPositionInfo(marketData.Positions)
		if !marketData.PositionInfo.HasLong && !marketData.PositionInfo.HasShort {
			log.Println("[交易执行] 持仓已不存在，跳过交易")
			return nil
		}
	}
	// 获取技术指标判断市场环境
	technicalData := PrepareTechnicalData(marketData)

	// 处理动态调整止损止盈的情况

	if isAdjustSLTPAction(signal.Action) {
		log.Println("[交易执行] 信号为ADJUST_SL_TP，开始调整止损止盈")
		return handleAdjustSLTP(signal, marketData, technicalData)
	}

	// 从MarketData中提取所需数据
	currentPrice, err := strconv.ParseFloat(marketData.Ticker.LastPrice, 64)
	if err != nil {
		return fmt.Errorf("解析当前价格失败: %v", err)
	}

	balanceInfo := GetAccountBalanceInfo(marketData)
	availableBalance := balanceInfo.AvailableBalance

	if currentPrice <= 0 {
		return fmt.Errorf("当前价格无效: %.4f", currentPrice)
	}

	// 客户端与交易参数
	client, err := binance.GetFuturesClient()
	if err != nil {
		log.Printf("[交易执行] 创建期货客户端失败: %v", err)
		return err
	}
	symbol := binance.ETHUSDT_PERP

	// 检测持仓模式：dualSide=true 为双向（hedge），false 为单向（one-way）
	dualSide, err := client.GetPositionMode()
	if err != nil {
		log.Printf("[交易执行] 获取持仓模式失败，按单向模式继续: %v", err)
		dualSide = false
	} else {
		if dualSide {
			log.Println("[交易执行] 当前为双向持仓模式")
		} else {
			log.Println("[交易执行] 当前为单向持仓模式")
		}
	}

	// 仓位百分比 - 处理可能的百分号格式
	positionPercent := float64(signal.PositionSize)
	if err != nil {
		log.Printf("[交易执行] 设置仓位百分比: %v, source:%v", err, signal.PositionSize)
		return err
	}

	// 设置杠杆
	leverage := 2
	if signal.Confidence*100 >= 85 && math.Abs(float64(signal.Score)) >= 8 {
		//当评分大于等于8，信心大于等于85 开启10倍杠杆
		leverage = 10
	} else if signal.Confidence*100 >= 75 && math.Abs(float64(signal.Score)) >= 7 {
		leverage = 5
	}
	if !utils.InSlice([]string{"CLOSE_LONG", "CLOSE_SHORT", "ADJUST_SL_TP"}, signal.Action) {
		//非平仓和调整止损止盈需要设置杠杆
		log.Printf("[交易执行] 设置杠杆倍数: %dx", leverage)
		if err := client.SetLeverage(symbol, leverage); err != nil {
			log.Printf("[交易执行] 设置杠杆失败: %v", err)
			return err
		}
	}

	if positionPercent <= 0 {
		positionPercent = 20
	}
	positionPct := positionPercent / 100.0

	// 获取持仓信息
	positionInfo := GetPositionInfo(marketData.Positions)

	// 计算开/加仓目标数量（使用可用余额）
	notional := availableBalance * positionPct * float64(leverage)

	// 仓位按照固定名义金额（移除波动率动态调整）
	openQty := calculateQuantity(notional, currentPrice)

	// 验证平仓操作
	if isCloseAction(signal.Action) {
		if err := ValidatePositionForClose(signal.Action, positionInfo); err != nil {
			jdata, _ := json.Marshal(positionInfo)
			log.Printf("[交易执行] 错误: %v 持仓数据: %v", err, jdata)
			return nil // 不是错误，跳过即可
		}
		openQty, _ = GetCloseQuantity(signal.Action, positionInfo)
	}

	// 对开/加仓，若低于最小步进则不下单
	if isOpenOrAddAction(signal.Action) && openQty < 0.001 {
		log.Printf("[交易执行] 计算得到下单数量过小(%.3f)，跳过", openQty)
		return nil
	}

	// 组装订单参数
	orderParams, err := prepareOrderParams(signal, currentPrice, openQty, positionInfo, dualSide)
	if err != nil {
		return fmt.Errorf("准备订单参数失败: %v", err)
	}

	log.Printf("[交易执行] 交易参数 - 操作: %s, 数量: %s ETH, 当前价格: %.2f, 可用余额: %.2f (钱包: %.2f, 保证金: %.2f), reduceOnly=%v, positionSide=%s",
		orderParams.Description, toQuantityString(orderParams.Quantity), currentPrice, availableBalance,
		balanceInfo.WalletBalance, balanceInfo.MarginBalance,
		orderParams.ReduceOnly, orderParams.PositionSide)

	// 下市价单
	order := &binance.NewOrderRequest{
		Symbol:     symbol,
		Side:       orderParams.Side,
		Type:       binance.OrderTypeMarket,
		Quantity:   toQuantityString(orderParams.Quantity),
		ReduceOnly: !dualSide && orderParams.ReduceOnly,
	}

	// 在双向模式下传递positionSide
	var finalPosSide binance.PositionSide
	if dualSide {
		finalPosSide = orderParams.PositionSide
	} else {
		finalPosSide = ""
	}

	// 下单前先删除可能存在的同方向止盈止损委托单（包含开/加仓与平仓）
	if err := cancelStopLossAndTakeProfitOrders(client, symbol, dualSide, orderParams.PositionSide); err != nil {
		log.Printf("[交易执行] 删除现有止盈止损委托失败: %v", err)
		// 不返回错误，继续执行交易
	}

	orderResult, err := client.NewOrder(order, finalPosSide)
	if err != nil {
		log.Printf("[交易执行] 下单失败: %v", err)
		return err
	}
	log.Printf("[交易执行] 订单执行成功 - 订单ID: %d, 状态: %s", orderResult.OrderID, orderResult.Status)

	// 平仓后删除所有相关的止盈止损委托单
	if orderParams.ReduceOnly {
		if err := cancelStopLossAndTakeProfitOrders(client, symbol, dualSide, orderParams.PositionSide); err != nil {
			log.Printf("[交易执行] 平仓后删除止盈止损委托失败: %v", err)
			// 不返回错误，因为平仓已经成功
		}
	}

	// 仅在开/加仓时设置止损/止盈；平仓不需要
	if !orderParams.ReduceOnly {
		if err := setStopLossAndTakeProfit(signal, marketData, currentPrice, orderParams.Side, technicalData, dualSide, orderParams.PositionSide); err != nil {
			// 不返回错误，因为主订单已经成功
			return err
		}
	}

	log.Println("[交易执行] 交易执行完成")
	return nil
}

// OrderParams 订单参数结构
type OrderParams struct {
	Side         binance.OrderSide
	Quantity     float64
	ReduceOnly   bool
	PositionSide binance.PositionSide
	Description  string
}

// isCloseAction 判断是否为平仓操作
func isCloseAction(action string) bool {
	return action == "CLOSE_LONG" || action == "CLOSE_SHORT"
}

// isOpenOrAddAction 判断是否为开仓或加仓操作
func isOpenOrAddAction(action string) bool {
	return action == "OPEN_LONG" || action == "OPEN_SHORT" || action == "ADD_LONG" || action == "ADD_SHORT"
}

// isAdjustSLTPAction 判断是否为调整止损止盈操作
func isAdjustSLTPAction(action string) bool {
	return action == "ADJUST_SL_TP"
}

// calculateQuantity 计算下单数量（按 0.001 步进取3位小数）
func calculateQuantity(notional, price float64) float64 {
	if notional <= 0 || price <= 0 {
		return 0
	}
	q := notional / price
	// 使用math.Round确保精度，避免大数值溢出
	q = math.Round(q*1000) / 1000.0
	return q
}

// toQuantityString 格式化数量字符串
func toQuantityString(q float64) string {
	return fmt.Sprintf("%.3f", q)
}

// prepareOrderParams 准备订单参数
func prepareOrderParams(signal *TradingSignal, currentPrice float64, quantity float64, positionInfo *PositionInfo, dualSide bool) (*OrderParams, error) {
	var orderParams OrderParams

	switch signal.Action {
	case "OPEN_LONG", "ADD_LONG":
		orderParams.Side = binance.OrderSideBuy
		orderParams.Quantity = quantity
		orderParams.ReduceOnly = false
		orderParams.Description = map[bool]string{true: "加多仓", false: "开多仓"}[signal.Action == "ADD_LONG"]
		if dualSide {
			orderParams.PositionSide = binance.PositionSideLong
		}
	case "OPEN_SHORT", "ADD_SHORT":
		orderParams.Side = binance.OrderSideSell
		orderParams.Quantity = quantity
		orderParams.ReduceOnly = false
		orderParams.Description = map[bool]string{true: "加空仓", false: "开空仓"}[signal.Action == "ADD_SHORT"]
		if dualSide {
			orderParams.PositionSide = binance.PositionSideShort
		}
	case "CLOSE_LONG":
		orderParams.Side = binance.OrderSideSell
		orderParams.ReduceOnly = true
		orderParams.Description = "平多仓"
		if dualSide {
			orderParams.PositionSide = binance.PositionSideLong
		}
		orderParams.Quantity = positionInfo.LongAmt
	case "CLOSE_SHORT":
		orderParams.Side = binance.OrderSideBuy
		orderParams.ReduceOnly = true
		orderParams.Description = "平空仓"
		if dualSide {
			orderParams.PositionSide = binance.PositionSideShort
		}
		orderParams.Quantity = positionInfo.ShortAmt
	default:
		return nil, fmt.Errorf("未知的交易操作类型: %s", signal.Action)
	}

	return &orderParams, nil
}

// setStopLossAndTakeProfit 设置止损止盈
func setStopLossAndTakeProfit(signal *TradingSignal, marketData *MarketData, currentPrice float64, side binance.OrderSide, technicalData *TechnicalAnalysisData, dualSide bool, positionSide binance.PositionSide) (e error) {
	// 根据波动率动态计算止损止盈
	var finalStopLoss, finalTakeProfit float64

	// 获取ATR用于动态止损止盈
	var atr float64
	if len(technicalData.High3m) > 0 && len(technicalData.Low3m) > 0 && len(technicalData.Price3m) > 0 {
		atr = indicators.GetLatestATR(technicalData.High3m, technicalData.Low3m, technicalData.Price3m, 14)
	}

	// 获取波动率
	volatilityPct := CalculateVolatilityForPosition(technicalData)

	// 动态调整止损止盈倍数 - 优化为1:1.5风险回报比（提高止盈达成率，适合短线交易）
	slMultiplier := 1.5
	tpMultiplier := 2.25     // 优化：短线交易1:1.5风险回报比
	if volatilityPct > 5.0 { // 高波动：扩大止损范围
		slMultiplier = 2.0
		tpMultiplier = 3.0 // 优化：保持1:1.5风险回报比
	} else if volatilityPct < 1.0 { // 极低波动：需要扩大止损范围避免被正常波动触发
		slMultiplier = 1.8
		tpMultiplier = 2.7 // 优化：保持1:1.5风险回报比
	} else if volatilityPct < 2.0 { // 低波动：标准设置
		slMultiplier = 1.5
		tpMultiplier = 2.25 // 优化：保持1:1.5风险回报比
	}

	// 如果有ATR，使用ATR计算
	if atr > 0 {
		var isLong bool
		if side == binance.OrderSideBuy {
			isLong = true
		}
		stopLossPrice := indicators.CalculateStopLossPrice(currentPrice, atr, slMultiplier, isLong)
		takeProfitPrice := indicators.CalculateTakeProfitPrice(currentPrice, atr, tpMultiplier, isLong)
		finalStopLoss = stopLossPrice
		finalTakeProfit = takeProfitPrice
	} else {
		// 如果没有ATR，使用固定百分比 - 优化为1:1.5风险回报比（适合短线交易）
		slRatio := 0.02          // 2%
		tpRatio := 0.03          // 3% (1:1.5) - 优化：降低目标提高达成率
		if volatilityPct > 5.0 { // 高波动
			slRatio = 0.03  // 3%
			tpRatio = 0.045 // 4.5% (1:1.5) - 优化：降低目标
		} else if volatilityPct < 2.0 { // 低波动
			slRatio = 0.015  // 1.5%
			tpRatio = 0.0225 // 2.25% (1:1.5) - 优化：降低目标
		}

		var stopLossPrice, takeProfitPrice float64
		if side == binance.OrderSideBuy { // 做多
			stopLossPrice = currentPrice * (1 - slRatio)
			takeProfitPrice = currentPrice * (1 + tpRatio)
		} else { // 做空
			stopLossPrice = currentPrice * (1 + slRatio)
			takeProfitPrice = currentPrice * (1 - tpRatio)
		}
		finalStopLoss = stopLossPrice
		finalTakeProfit = takeProfitPrice
	}

	// 优先使用信号中提供的止损止盈，如果为空则使用动态计算的
	if signal.StopLoss > 0.0 {
		finalStopLoss = signal.StopLoss
	}
	if signal.TakeProfit > 0.0 {
		finalTakeProfit = signal.TakeProfit
	}

	// 获取客户端
	client, err := binance.GetFuturesClient()
	if err != nil {
		return fmt.Errorf("获取期货客户端失败: %v", err)
	}
	symbol := binance.ETHUSDT_PERP

	// 设置止损
	if finalStopLoss > 0.0 {
		var slSide binance.OrderSide
		if side == binance.OrderSideBuy {
			slSide = binance.OrderSideSell
		} else {
			slSide = binance.OrderSideBuy
		}
		slOrder := &binance.NewOrderRequest{
			Symbol:        symbol,
			Side:          slSide,
			Type:          binance.OrderTypeStopMarket,
			StopPrice:     fmt.Sprintf("%.2f", finalStopLoss),
			ClosePosition: true,
			WorkingType:   binance.WorkingTypeMarkPrice,
		}
		var slFinalPosSide binance.PositionSide
		if dualSide {
			slFinalPosSide = positionSide
		} else {
			slFinalPosSide = ""
		}
		if _, err := client.NewOrder(slOrder, slFinalPosSide); err != nil {
			log.Printf("[交易执行] 设置止损单失败: %v", err)
			e = fmt.Errorf("[交易执行] 设置止损单失败: %v", err)
		} else {
			log.Printf("[交易执行] 止损单设置成功，价格: %s (基于波动率%.2f%%)", slOrder.StopPrice, volatilityPct)
		}
	}

	// 设置止盈
	if finalTakeProfit > 0.0 {
		var tpSide binance.OrderSide
		if side == binance.OrderSideBuy {
			tpSide = binance.OrderSideSell
		} else {
			tpSide = binance.OrderSideBuy
		}
		tpOrder := &binance.NewOrderRequest{
			Symbol:        symbol,
			Side:          tpSide,
			Type:          binance.OrderTypeTakeProfitMarket,
			StopPrice:     fmt.Sprintf("%.2f", finalTakeProfit),
			ClosePosition: true,
			WorkingType:   binance.WorkingTypeMarkPrice,
		}
		var tpFinalPosSide binance.PositionSide
		if dualSide {
			tpFinalPosSide = positionSide
		} else {
			tpFinalPosSide = ""
		}
		if _, err := client.NewOrder(tpOrder, tpFinalPosSide); err != nil {
			log.Printf("[交易执行] 设置止盈单失败: %v", err)
			e = fmt.Errorf("[交易执行] 设置止盈单失败: %v", err)
		} else {
			log.Printf("[交易执行] 止盈单设置成功，价格: %s (基于波动率%.2f%%)", tpOrder.StopPrice, volatilityPct)
		}
	}

	return
}

// cancelStopLossAndTakeProfitOrders 删除指定方向的止损止盈委托单
func cancelStopLossAndTakeProfitOrders(client *binance.FuturesClient, symbol binance.Symbol, dualSide bool, positionSide binance.PositionSide) error {
	// 获取当前所有挂单
	orders, err := client.GetOpenOrders(symbol)
	if err != nil {
		return fmt.Errorf("获取挂单失败: %v", err)
	}

	if len(orders) == 0 {
		log.Println("[交易执行] 没有挂单需要删除")
		return nil
	}

	// 筛选并删除止损止盈委托单
	cancelledCount := 0
	for _, order := range orders {
		// 只处理止损和止盈订单
		if order.Type == binance.OrderTypeStopMarket || order.Type == binance.OrderTypeTakeProfitMarket || order.Type == binance.OrderTypeStop || order.Type == binance.OrderTypeTakeProfit {
			// 在双向模式下，只删除对应方向的订单
			if dualSide {
				// 检查订单的持仓方向是否匹配
				if string(order.PositionSide) != string(positionSide) {
					continue
				}
			}

			// 取消单个订单
			_, err := client.CancelOrder(symbol, order.OrderID, "")
			if err != nil {
				log.Printf("[交易执行] 取消订单失败 (ID: %d, Type: %s): %v", order.OrderID, order.Type, err)
			} else {
				log.Printf("[交易执行] 成功取消订单 (ID: %d, Type: %s, Side: %s, Price: %s)",
					order.OrderID, order.Type, order.Side, order.StopPrice)
				cancelledCount++
			}
		}
	}

	if cancelledCount > 0 {
		log.Printf("[交易执行] 共删除 %d 个止盈止损委托单", cancelledCount)
	} else {
		log.Println("[交易执行] 没有找到需要删除的止盈止损委托单")
	}

	return nil
}

// handleAdjustSLTP 处理动态调整止损止盈的操作
func handleAdjustSLTP(signal *TradingSignal, marketData *MarketData, technicalData *TechnicalAnalysisData) error {
	log.Printf("[止损止盈调整] 开始处理动态调整止损止盈: %s", signal.Reasoning)

	// 基本校验
	if signal.StopLoss == 0 && signal.TakeProfit == 0 {
		return fmt.Errorf("调整止损止盈必须提供至少一个价格")
	}

	// 获取客户端
	client, err := binance.GetFuturesClient()
	if err != nil {
		return fmt.Errorf("获取期货客户端失败: %v", err)
	}
	symbol := binance.ETHUSDT_PERP

	// 检测持仓模式：dualSide=true 为双向（hedge），false 为单向（one-way）
	dualSide, err := client.GetPositionMode()
	if err != nil {
		log.Printf("[止损止盈调整] 获取持仓模式失败，按单向模式继续: %v", err)
		dualSide = false
	}

	// 获取持仓信息
	positionInfo := GetPositionInfo(marketData.Positions)

	// 检查是否有持仓
	if !positionInfo.HasLong && !positionInfo.HasShort {
		log.Println("[止损止盈调整] 没有持仓，无法调整止损止盈")
		return nil
	}

	// 获取当前价格
	currentPrice, err := strconv.ParseFloat(marketData.Ticker.LastPrice, 64)
	if err != nil {
		return fmt.Errorf("解析当前价格失败: %v", err)
	}

	// 处理多头持仓的止损止盈调整
	if positionInfo.HasLong {
		log.Printf("[止损止盈调整] 调整多头持仓止损止盈")

		// 删除现有的多头止损止盈订单
		if err := cancelStopLossAndTakeProfitOrders(client, symbol, dualSide, binance.PositionSideLong); err != nil {
			log.Printf("[止损止盈调整] 删除多头现有止损止盈委托失败: %v", err)
			// 继续执行，不返回错误
		}

		if err := setStopLossAndTakeProfit(signal, marketData, currentPrice, binance.OrderSideBuy, technicalData, dualSide, binance.PositionSideLong); err != nil {
			log.Printf("[止损止盈调整] 设置多头新止损止盈失败: %v", err)
			return err
		}
	}

	// 处理空头持仓的止损止盈调整
	if positionInfo.HasShort {
		log.Printf("[止损止盈调整] 调整空头持仓止损止盈")

		// 删除现有的空头止损止盈订单
		if err := cancelStopLossAndTakeProfitOrders(client, symbol, dualSide, binance.PositionSideShort); err != nil {
			log.Printf("[止损止盈调整] 删除空头现有止损止盈委托失败: %v", err)
			// 继续执行，不返回错误
		}

		if err := setStopLossAndTakeProfit(signal, marketData, currentPrice, binance.OrderSideSell, technicalData, dualSide, binance.PositionSideShort); err != nil {
			log.Printf("[止损止盈调整] 设置空头新止损止盈失败: %v", err)
			// 继续执行，不返回错误
			return err
		}
	}

	log.Println("[止损止盈调整] 动态调整止损止盈完成")
	return nil
}
