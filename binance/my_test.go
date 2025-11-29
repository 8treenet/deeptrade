package binance_test

import (
	"deeptrade/binance"
	"deeptrade/conf"
	"deeptrade/task"
	"encoding/json"
	"log"
	"os"
	"testing"
)

func init() {
	// 设置配置文件目录路径，确保能够找到项目配置
	os.Setenv("deeptrade_conf", "/Users/ys/work/my/deeptrade/conf")
	// 初始化项目配置系统
	conf.EntryPoint()
}

// GetFuturesClient
func GetFuturesClient() *binance.FuturesClient {
	// 使用测试网配置创建期货客户端，避免使用真实资金
	futuresClient, err := binance.NewFuturesClientFromConfig()
	if err != nil {
		log.Fatal("创建客户端失败:", err)
	}
	return futuresClient
}

func TestFuturesGetPositions(t *testing.T) {
	client := GetFuturesClient()
	data, _ := client.GetPositions(binance.ETHUSDT_PERP)
	t.Log(jsondata(data))
}

func TestGetFundingRateHistory(t *testing.T) {
	client := GetFuturesClient()
	data, _ := client.GetFundingRateHistory(binance.ETHUSDT_PERP, 6, 0, 0)
	t.Log(jsondata(data))
}

func TestGetBookTicker(t *testing.T) {
	client := GetFuturesClient()
	data, _ := client.GetBookTicker(binance.ETHUSDT_PERP)
	t.Log(jsondata(data))
}

func TestGetKlines(t *testing.T) {
	client := GetFuturesClient()
	data, _ := client.GetKlines(binance.ETHUSDT_PERP, binance.KlineInterval3m, 30)
	t.Log(task.CSVData(data))
}

func TestGetRecentTrades(t *testing.T) {
	client := GetFuturesClient()
	data, _ := client.GetRecentTrades(binance.ETHUSDT_PERP, 1000)
	t.Log(jsondata(data))
}

func TestGetOpenOrders(t *testing.T) {
	client := GetFuturesClient()
	data, _ := client.GetOpenOrders(binance.ETHUSDT_PERP)
	t.Log(jsondata(data))
}

// jsondata 保留原有的JSON格式化函数，以备不时之需
func jsondata(data interface{}) string {
	jdata, _ := json.Marshal(data)
	return string(jdata)
}
