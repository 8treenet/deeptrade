# Binance Go SDK

生产级别的币安交易所 Go 语言 SDK，专注于 ETH 量化交易，支持现货和合约交易。

## 特性

- 🚀 **高性能**: 内置连接池、请求重试、速率限制
- 📊 **完整的市场数据**: 现货/合约 K线、深度、成交记录等
- 🔌 **实时数据流**: WebSocket 实时价格、深度、账户更新推送
- 💼 **交易功能**: 现货/合约下单、撤单、查询订单
- 🛡️ **安全可靠**: 完善的错误处理、签名验证、超时控制
- 📈 **量化友好**: 专为量化交易设计的市场数据管理器
- 🧪 **完整测试**: 全面的单元测试覆盖

## 安装

```bash
go get deeptrade/binance
```

## 快速开始

### 环境配置

本SDK支持两个环境：

**🏭 真实盘 (Production)**
- 实际交易环境，使用真实资金
- API地址：`https://api.binance.com` (现货), `https://fapi.binance.com` (合约)
- 用于生产环境交易

**🧪 模拟盘 (Testnet)**
- 测试环境，使用虚拟资金，无风险
- API地址：`https://testnet.binance.vision` (现货), `https://testnet.binancefuture.com` (合约)
- 用于策略测试和开发

### 基本配置

#### 方法1: 使用项目配置系统（推荐）

项目的币安配置集成在现有的配置系统中，通过 `conf/config.toml` 文件管理：

```go
package main

import (
    "fmt"
    "log"
    "deeptrade/binance"
)

func main() {
    // 从项目配置系统加载当前环境配置
    config, err := binance.LoadCurrentEnvironmentConfigFromProject()
    if err != nil {
        log.Fatal(err)
    }

    // 创建客户端
    spotClient, err := binance.NewSpotClient(config)
    if err != nil {
        log.Fatal(err)
    }

    // 获取价格
    ticker, err := spotClient.Get24hrTicker(binance.ETHUSDT)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("ETH价格: %s USDT\n", ticker.LastPrice)
}
```

#### 方法2: 强制使用特定环境

```go
// 强制使用模拟盘配置
testnetConfig, err := binance.LoadTestnetConfigFromProject()
if err != nil {
    log.Fatal(err)
}

// 强制使用真实盘配置
productionConfig, err := binance.LoadProductionConfigFromProject()
if err != nil {
    log.Fatal(err)
}
```

#### 方法3: 代码配置

```go
package main

import (
    "fmt"
    "log"
    "deeptrade/binance"
)

func main() {
    // 真实盘配置
    prodConfig := binance.ProductionConfig("your_api_key", "your_secret_key")

    // 模拟盘配置（推荐用于测试）
    testnetConfig := binance.TestnetConfig("your_api_key", "your_secret_key")

    // 带代理的配置（可选）
    config := binance.TestnetConfigWithProxy("your_api_key", "your_secret_key", "http://127.0.0.1:33210")
    config.Timeout = 10
    config.MaxRetries = 3

    // 创建客户端
    spotClient, err := binance.NewSpotClient(config)
    if err != nil {
        log.Fatal(err)
    }

    // 获取价格
    ticker, err := spotClient.Get24hrTicker(binance.ETHUSDT)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("ETH价格: %s USDT\n", ticker.LastPrice)
}
```

### 项目配置文件

在项目根目录的 `conf/config.toml` 文件中添加币安配置：

```toml
[binance]
# 当前环境: testnet, production
current_environment = "testnet"
# 默认代理设置
default_proxy = "http://127.0.0.1:33210"

# 环境配置
[binance.environments]

[binance.environments.testnet]
name = "testnet"
api_key = "your_testnet_api_key"
secret_key = "your_testnet_secret_key"
spot_base_url = "https://testnet.binance.vision"
futures_base_url = "https://testnet.binancefuture.com"
spot_stream_url = "wss://testnet.binance.vision"
futures_stream_url = "wss://stream.binancefuture.com"
timeout = 30
max_retries = 3
debug = true
proxy_url = "http://127.0.0.1:33210"

[binance.environments.production]
name = "production"
api_key = "your_production_api_key"
secret_key = "your_production_secret_key"
spot_base_url = "https://api.binance.com"
futures_base_url = "https://fapi.binance.com"
spot_stream_url = "wss://stream.binance.com:9443"
futures_stream_url = "wss://fstream.binance.com"
timeout = 10
max_retries = 5
debug = false
proxy_url = "http://127.0.0.1:33210"
```

### 配置函数

#### 项目配置系统函数

| 函数 | 说明 | 用途 |
|------|------|------|
| `LoadCurrentEnvironmentConfigFromProject()` | 从项目配置系统加载当前环境 | 通用配置加载 |
| `LoadTestnetConfigFromProject()` | 从项目配置系统加载模拟盘配置 | 强制使用模拟盘 |
| `LoadProductionConfigFromProject()` | 从项目配置系统加载真实盘配置 | 强制使用真实盘 |

#### 传统代码配置函数

| 函数 | 说明 | 用途 |
|------|------|------|
| `ProductionConfig(apiKey, secretKey)` | 创建真实盘配置 | 真实盘交易 |
| `TestnetConfig(apiKey, secretKey)` | 创建模拟盘配置 | 模拟盘测试 |
| `*ConfigWithProxy(apiKey, secretKey, proxyURL)` | 创建带代理的配置 | 通过代理访问 |

### 环境选择指南

| 场景 | 推荐环境 | 配置方式 | 说明 |
|------|----------|----------|------|
| 开发测试 | 模拟盘 | 项目配置 + `current_environment = "testnet"` | 使用虚拟资金，无风险 |
| 策略回测 | 模拟盘 | `LoadTestnetConfigFromProject()` | 强制使用模拟盘测试策略 |
| 生产交易 | 真实盘 | 项目配置 + `current_environment = "production"` | 实际交易，需谨慎 |
| 灵活切换 | 任意 | 修改 `config.toml` 中的 `current_environment` | 重启应用后生效 |

### 市场数据

```go
// 获取K线数据
klines, err := spotClient.GetKlines(binance.ETHUSDT, binance.KlineInterval1h, 100)
if err != nil {
    log.Fatal(err)
}

// 获取深度信息
depth, err := spotClient.GetDepth(binance.ETHUSDT, binance.DepthLevel20)
if err != nil {
    log.Fatal(err)
}

// 获取最近交易
trades, err := spotClient.GetRecentTrades(binance.ETHUSDT, 10)
if err != nil {
    log.Fatal(err)
}
```

### 交易操作

```go
// 下限价单
orderRequest := &binance.NewOrderRequest{
    Symbol:      binance.ETHUSDT,
    Side:        binance.OrderSideBuy,
    Type:        binance.OrderTypeLimit,
    Quantity:    "0.1",
    Price:       "2000.00",
    TimeInForce: binance.TimeInForceGTC,
}

order, err := spotClient.NewOrder(orderRequest)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("订单ID: %d\n", order.OrderID)

// 查询订单
orderInfo, err := spotClient.GetOrder(binance.ETHUSDT, order.OrderID, "")
if err != nil {
    log.Fatal(err)
}

// 取消订单
cancelledOrder, err := spotClient.CancelOrder(binance.ETHUSDT, order.OrderID, "")
if err != nil {
    log.Fatal(err)
}
```

### 合约交易

```go
// 创建合约客户端
futuresClient, err := binance.NewFuturesClient(config)
if err != nil {
    log.Fatal(err)
}

// 设置杠杆
err = futuresClient.SetLeverage(binance.ETHUSDT_PERP, 20)
if err != nil {
    log.Fatal(err)
}

// 下合约订单
order, err := futuresClient.NewOrder(orderRequest, binance.PositionSideLong)
if err != nil {
    log.Fatal(err)
}

// 获取持仓信息
positions, err := futuresClient.GetPositions(binance.ETHUSDT_PERP)
if err != nil {
    log.Fatal(err)
}

// 注意：GetPositions 返回所有持仓记录，包括positionAmt="0.000"的空持仓
// 需要调用者自己判断是否有实际持仓（注意导入 strconv 包）
for _, pos := range positions {
    positionAmt, _ := strconv.ParseFloat(pos.PositionAmt, 64)
    if positionAmt != 0 {
        fmt.Printf("持仓: %s %s, 盈亏: %s\n",
            pos.Symbol, pos.PositionAmt, pos.UnRealizedProfit)
    }
}
```

### 实时数据流

```go
// 创建WebSocket客户端
wsClient := binance.NewWSClient(config, false) // false表示现货

// 设置消息处理器
wsClient.SetMessageHandler(binance.WSTicker, func(data interface{}) {
    if wsTicker, ok := data.(map[string]interface{}); ok {
        symbol := wsTicker["s"].(string)
        price := wsTicker["c"].(string)
        fmt.Printf("价格更新: %s = %s\n", symbol, price)
    }
})

// 订阅市场数据
symbols := []binance.Symbol{binance.ETHUSDT}
messageTypes := []binance.WebSocketMessageType{binance.WSTicker}

err = wsClient.SubscribeMarketData(symbols, messageTypes)
if err != nil {
    log.Fatal(err)
}

// 程序结束时断开连接
defer wsClient.Disconnect()
```

### 市场数据管理器

```go
// 创建市场数据管理器
managerConfig := &binance.MarketDataConfig{
    Symbols:        []binance.Symbol{binance.ETHUSDT, binance.ETHUSDT_PERP},
    Intervals:      []binance.KlineInterval{binance.KlineInterval1m, binance.KlineInterval5m},
    EnableRealtime: true,
}

manager := binance.NewMarketDataManager(spotClient, futuresClient, managerConfig)

// 启动管理器
err := manager.Start()
if err != nil {
    log.Fatal(err)
}
defer manager.Stop()

// 获取实时价格更新
tickerChan := manager.GetRealtimeTicker()
go func() {
    for ticker := range tickerChan {
        fmt.Printf("实时价格: %s = %s\n", ticker.Symbol, ticker.LastPrice)
    }
}()
```

## 配置选项

### 基础配置

```go
config := &binance.Config{
    APIKey:     "your_api_key",
    SecretKey:  "your_secret_key", 
    Timeout:    30,              // 超时时间(秒)
    Debug:      false,           // 调试模式
    MaxRetries: 3,               // 最大重试次数
    RecvWindow: 5000,            // 接收窗口(毫秒)
}
```

### 环境配置

```go
// 生产环境配置
config := binance.ProductionConfig(apiKey, secretKey)

// 测试环境配置  
config := binance.TestConfig(apiKey, secretKey)

// 默认配置
config := binance.DefaultConfig()
```

## 错误处理

SDK 提供了完善的错误处理机制：

```go
ticker, err := spotClient.Get24hrTicker(binance.ETHUSDT)
if err != nil {
    // 检查错误类型
    if binance.IsTimeoutError(err) {
        log.Println("请求超时")
    } else if binance.IsRateLimitError(err) {
        log.Println("请求频率过高")
    } else if binance.IsAuthError(err) {
        log.Println("认证失败")
    } else if binance.IsOrderError(err) {
        log.Println("订单相关错误")
    }
    
    // 获取详细错误信息
    if binanceErr, ok := err.(*binance.Error); ok {
        log.Printf("错误代码: %d, 消息: %s, 详情: %s", 
            binanceErr.Code, binanceErr.Message, binanceErr.Details)
    }
}
```

## 数据类型

### 主要常量

```go
// 交易对
const (
    ETHUSDT     binance.Symbol = "ETHUSDT"
    ETHUSDT_PERP binance.Symbol = "ETHUSDT" // 永续合约
)

// 订单类型
const (
    OrderTypeMarket      binance.OrderType = "MARKET"
    OrderTypeLimit       binance.OrderType = "LIMIT" 
    OrderTypeStopLoss    binance.OrderType = "STOP_LOSS"
    OrderTypeTakeProfit  binance.OrderType = "TAKE_PROFIT"
)

// K线间隔
const (
    KlineInterval1m  binance.KlineInterval = "1m"
    KlineInterval5m  binance.KlineInterval = "5m"
    KlineInterval15m binance.KlineInterval = "15m"
    KlineInterval1h  binance.KlineInterval = "1h"
    KlineInterval4h  binance.KlineInterval = "4h"
    KlineInterval1d  binance.KlineInterval = "1d"
)
```

### 核心结构体

- `Ticker` - 24小时价格统计
- `Depth` - 深度信息
- `Kline` - K线数据
- `Order` - 订单信息
- `Balance` - 账户余额
- `Position` - 持仓信息（合约）

## WebSocket 消息类型

### 现货数据流

- `WSTicker` - 24小时价格统计
- `WSDepth` - 深度数据
- `WSKline` - K线数据
- `WSTrade` - 交易数据
- `WSAggTrade` - 聚合交易数据

### 合约数据流

- `WSMarkPrice` - 标记价格
- `WSFundingRate` - 资金费率
- `WSContinuousKline` - 连续K线数据

### 用户数据流

- 账户更新 (`accountUpdate`)
- 订单更新 (`orderUpdate`)

## 最佳实践

### 1. 安全配置

```go
// 从环境变量获取API密钥
apiKey := os.Getenv("BINANCE_API_KEY")
secretKey := os.Getenv("BINANCE_SECRET_KEY")

// 使用生产环境配置
config := binance.ProductionConfig(apiKey, secretKey)
```

### 2. 错误处理

```go
// 总是检查错误
ticker, err := client.Get24hrTicker(binance.ETHUSDT)
if err != nil {
    log.Printf("获取价格失败: %v", err)
    return
}

// 使用类型断言获取详细错误信息
if binanceErr, ok := err.(*binance.Error); ok {
    log.Printf("错误详情: %s", binanceErr.Details)
}
```

### 3. 资源管理

```go
// 使用defer确保资源释放
wsClient := binance.NewWSClient(config, false)
defer wsClient.Disconnect()

manager := binance.NewMarketDataManager(spotClient, futuresClient, config)
manager.Start()
defer manager.Stop()
```

### 4. 并发安全

```go
// 市场数据管理器是并发安全的
// 可以在多个goroutine中安全使用
go func() {
    for ticker := range manager.GetRealtimeTicker() {
        // 处理实时价格更新
    }
}()

go func() {
    for depth := range manager.GetRealtimeDepth() {
        // 处理实时深度更新  
    }
}()
```

### 5. 性能优化

```go
// 复用客户端连接
client, _ := binance.NewSpotClient(config)

// 使用市场数据管理器缓存数据
manager := binance.NewMarketDataManager(client, nil, config)
manager.Start()

// 从缓存获取数据，减少API调用
ticker, _ := manager.GetTicker(binance.ETHUSDT)
```

## 示例代码

查看 `examples/` 目录获取完整的示例代码：

- `basic_usage.go` - 基本功能演示
- `websocket_example.go` - WebSocket 实时数据演示

## 测试

运行单元测试：

```bash
go test ./binance/...
```

运行特定测试：

```bash
go test ./binance/ -run TestConfig
go test ./binance/ -run TestError
go test ./binance/ -run Benchmark
```

## 注意事项

1. **API 密钥安全**: 请妥善保管 API 密钥，不要硬编码在代码中
2. **速率限制**: 注意币安 API 的速率限制，避免请求过于频繁
3. **网络延迟**: 设置合适的超时时间和重试机制
4. **数据精度**: 价格和数量使用字符串类型以避免精度丢失
5. **时间同步**: 确保服务器时间同步，避免时间戳错误

## 许可证

本项目采用 MIT 许可证。

## 贡献

欢迎提交 Issue 和 Pull Request！

## 更新日志

### v1.0.0

- 初始版本发布
- 支持现货和合约交易
- WebSocket 实时数据流
- 市场数据管理器
- 完整的错误处理
- 单元测试覆盖

---

如有问题，请提交 Issue 或联系维护者。