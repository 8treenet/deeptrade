# DeepTrade - ETH期货量化交易Agent

[![Go Version](https://img.shields.io/badge/Go-1.24.0+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[📊 Agent生成的报告预览相关](report.md)

一个基于人工智能的加密货币量化交易系统，专门用于ETH期货交易。系统集成了多种LLM提供商（DeepSeek、Qwen等）进行智能决策，并实现了高级技术指标分析和实时市场数据处理。

## 🌟 主要特性

- 🤖 **AI驱动决策**: 集成多个LLM提供商（DeepSeek、Qwen、Kimi）进行智能交易分析
- 📊 **全面技术分析**: 支持RSI、MACD、布林带、ATR、EMA、Stochastic、CCI、Williams %R、ROC等多种技术指标
- ⚡ **实时市场数据**: 并发获取价格、K线、订单簿、持仓、账户信息等实时数据
- 🔄 **智能交易执行**: 自动开仓、平仓、加仓、止损止盈等交易操作
- 🛡️ **风险管理**: 内置仓位管理、杠杆控制和交易时间限制
- 📈 **交易流分析**: 专业的交易流分析和成交量趋势分析
- 💾 **记忆系统**: 保存交易历史和分析记忆，提高决策连续性
- 🌐 **多环境支持**: 支持测试网和生产环境切换
- 📧 **邮件通知**: 重要事件和错误邮件提醒

## 🏗️ 系统架构

```
DeepTrade
├── main.go                 # 系统入口点，包含交易定时器
├── task/                   # 核心交易逻辑
│   ├── quantitative_trading.go  # 量化交易主流程
│   ├── market_data.go          # 市场数据获取
│   ├── analysis.go             # LLM分析处理
│   ├── execution.go            # 交易执行逻辑
│   ├── technical_analysis.go   # 技术指标分析
│   ├── position.go             # 持仓管理
│   ├── memory.go               # 交易记忆系统
│   └── types.go                # 核心数据结构
├── binance/                # Binance API集成
│   ├── futures_client.go       # 期货交易客户端
│   ├── client.go               # 通用客户端
│   └── types.go                # API类型定义
├── indicators/             # 技术指标实现
│   ├── rsi.go                  # RSI指标
│   ├── macd.go                 # MACD指标
│   ├── bollinger.go            # 布林带
│   ├── atr.go                  # ATR指标
│   ├── sma_ema.go              # 移动平均线
│   ├── stochastic.go           # 随机指标
│   ├── cci.go                  # CCI指标
│   ├── williams_r.go           # Williams %R
│   ├── roc.go                  # ROC指标
│   ├── volume_analysis.go      # 成交量分析
│   └── trade_flow_analysis.go  # 交易流分析
├── utils/                  # 工具函数
│   ├── utility.go              # LLM集成和通用工具
│   ├── email.go                # 邮件通知
│   └── proxy_client.go         # 代理客户端
└── conf/                   # 配置管理
    ├── config.go               # 配置结构定义
    └── config.toml             # 配置文件
```

## 🚀 快速开始

### 环境要求

- Go 1.24.0+
- Binance API密钥（测试网或生产网）
- LLM API密钥（DeepSeek、Qwen等）

### 安装步骤

1. **克隆项目**
```bash
git clone <repository-url>
cd deeptrade
```

2. **安装依赖**
```bash
go mod download
```

3. **配置设置**
编辑 [`conf/config.toml`](conf/config.toml:1) 文件，配置API密钥和交易参数：

```toml
[binance]
# 当前环境: testnet, production
current_environment = "testnet"
timeout = 30
max_retries = 3

[binance.testnet]
api_key = "your_testnet_api_key"
secret_key = "your_testnet_secret_key"
futures_base_url = "https://testnet.binancefuture.com"

[binance.production]
api_key = "your_production_api_key"
secret_key = "your_production_secret_key"
futures_base_url = "https://fapi.binance.com"

# LLM配置 - 示例配置
# 开仓决策模型（推荐使用推理能力强的模型）
[[llm]]
api_key = "your_deepseek_api_key"
model = "deepseek-reasoner"
base_url = "https://api.deepseek.com/v1"
entry_enable = true   # 用于开仓决策
track_enable = false  # 持仓时不使用

# 持仓跟踪模型（推荐使用响应速度快的模型）
[[llm]]
api_key = "your_qwen_api_key"
model = "qwen-plus"
base_url = "https://dashscope.aliyuncs.com/compatible-mode/v1"
entry_enable = false  # 开仓时不使用
track_enable = true   # 用于持仓跟踪

# 交易配置
[trading]
trigger_time = 20  # 交易周期（分钟）
```

4. **构建和运行**
```bash
# 构建应用
go build main.go

# 直接运行
./main

# 或使用脚本运行（后台模式）
./run.sh
```

## 📊 核心功能

### 1. 市场数据获取

系统并发获取多种市场数据：

- **实时价格**: 24小时价格统计和最新成交价
- **K线数据**: 1分钟和3分钟K线数据
- **订单簿深度**: 买卖盘深度信息
- **持仓信息**: 当前持仓和未实现盈亏
- **账户信息**: 可用余额和保证金状态
- **资金费率**: 当前和历史资金费率
- **持仓量**: 市场未平仓合约数量

### 2. 技术指标分析

系统实现了完整的技术指标库：

- **趋势指标**: EMA、MACD、布林带
- **动量指标**: RSI、Stochastic、CCI、Williams %R、ROC
- **波动率指标**: ATR、布林带宽度
- **成交量分析**: 成交量趋势和交易流分析

### 3. AI决策系统

集成多个LLM提供商进行智能分析：

- **DeepSeek**: 主要决策模型，支持推理模式
- **Qwen**: 备用决策模型
- **Kimi**: 辅助分析模型

系统将市场数据和技术指标转换为结构化提示，发送给LLM获取交易信号。

### 4. 交易执行

支持多种交易操作：

- **开仓**: 多头/空头开仓
- **平仓**: 多头/空头平仓
- **加仓**: 现有仓位增加
- **止损止盈**: 自动风险控制

### 5. 风险管理

- **交易时间限制**: 只在特定时间段交易（周一至周五 9:00-23:59，周二至周六 0:00-2:59）
- **仓位控制**: 基于账户余额的百分比仓位管理
- **杠杆控制**: 可配置的交易杠杆
- **持仓监控**: 实时监控持仓状态和盈亏

## ⚙️ 配置说明

### 环境配置

系统支持两个环境：

- **测试网 (testnet)**: 使用虚拟资金，无风险测试
- **生产网 (production)**: 实际交易环境，使用真实资金

在 [`conf/config.toml`](conf/config.toml:3) 中设置 `current_environment` 切换环境。

### LLM配置

系统支持配置多个LLM提供商，并通过 `entry_enable` 和 `track_enable` 参数控制不同场景下的模型选择：

```toml
# 开仓决策模型 - 用于分析市场数据并生成开仓信号
[[llm]]
api_key = "your_deepseek_api_key"
model = "deepseek-reasoner"
base_url = "https://api.deepseek.com/v1"
entry_enable = true   # 开仓时使用此模型
track_enable = false   # 持仓时不使用此模型

# 持仓跟踪模型 - 用于管理现有持仓和调整止损止盈
[[llm]]
api_key = "your_qwen_api_key"
model = "qwen-plus"
base_url = "https://dashscope.aliyuncs.com/compatible-mode/v1"
entry_enable = false   # 开仓时不使用此模型
track_enable = true    # 持仓时使用此模型
```

#### 模型选择逻辑

- **开仓决策** (`entry_enable = true`): 系统会查找第一个启用了 `entry_enable` 的模型进行市场分析和开仓信号生成
- **持仓跟踪** (`track_enable = true`): 当有持仓时，系统会查找第一个启用了 `track_enable` 的模型进行持仓管理和调整决策

#### 推荐配置策略

1. **开仓模型**: 选择推理能力强、分析深度高的模型（如 DeepSeek Reasoner）
2. **持仓模型**: 选择响应速度快、成本较低的模型（如 Qwen Plus）

**⚠️ 免责声明**: 本系统仅用于教育和研究目的。加密货币交易具有高风险，可能导致资金损失。使用本系统进行实际交易的风险由用户自行承担。