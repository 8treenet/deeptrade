package task_test

import (
	"deeptrade/conf"
	"deeptrade/task"
	"encoding/json"
	"os"
	"testing"
)

func init() {
	// 设置配置文件目录路径，确保能够找到项目配置
	os.Setenv("deeptrade_conf", "/Users/ys/work/my/deeptrade/conf")
	// 初始化项目配置系统
	conf.EntryPoint()
}

func TestGetMarketData(t *testing.T) {
	// t.Log("全部钱包余额:", 4581.4791268)
	// t.Log("可用余额:", 4581.4791268)
	// return
	data, _ := task.GetMarketData()
	t.Log("全部钱包余额:", data.Account.TotalWalletBalance)
	t.Log("可用余额:", data.Account.AvailableBalance)
	PositionsData, _ := json.Marshal(data.Positions)
	t.Log(string(PositionsData))
}

func TestCloseOrder(t *testing.T) {
	data, _ := task.GetMarketData()
	signal := &task.TradingSignal{Action: "CLOSE_LONG", PositionSize: 100}
	t.Log(task.ExecuteTrade(signal, data))
}

func TestExecuteTradeWithMemory_OpenLong(t *testing.T) {
	data, err := task.GetMarketData()
	if err != nil {
		t.Fatalf("获取市场数据失败: %v", err)
	}

	stopLoss := 1500.00
	takeProfit := 2500.00
	signal := &task.TradingSignal{
		Action:       "OPEN_LONG",
		Score:        8,
		Confidence:   0.85,
		StopLoss:     stopLoss,
		TakeProfit:   takeProfit,
		PositionSize: 20,
		Memory:       "测试开多仓",
	}

	err = task.ExecuteTradeWithMemory(signal, data, "test-decision-id-001")
	if err != nil {
		t.Logf("交易执行结果: %v", err)
	} else {
		t.Log("交易执行成功")
	}

	t.Logf("当前价格: %s", data.Ticker.LastPrice)
	t.Logf("可用余额: %s", data.Account.AvailableBalance)
	t.Logf("止损价格: %.2f (10%%)", stopLoss)
	t.Logf("止盈价格: %.2f (10%%)", takeProfit)
}
