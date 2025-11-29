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
