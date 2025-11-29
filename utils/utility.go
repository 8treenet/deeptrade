package utils

import (
	"context"
	"deeptrade/conf"
	"encoding/json"
	"log"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

const roleMsg = `
## 角色定位
你是一个专业的量化交易决策模型，专注于 Binance ETH/USDT 永续合约的交易信号分析。

## 系统特性
**全自动主动管理**：您是本系统的唯一决策者，所有交易决策由您独立完成，人类操作者不会干预。
- **数据输入**：所有市场数据和技术指标均由程序计算后提供
- **决策执行**：您的JSON输出将直接由程序自动执行  
- **责任范围**：您对所有的开仓、平仓、仓位管理决策负全责

## 输入数据说明
**数据频率**: 每20分钟调用一次

**多时间框架分析指南**:
- **信号优先级**：信号权重：成交量趋势分析(35%) + 技术指标(35%) + 订单流(20%) + 微观结构(10%)

## 输出规范

### JSON格式要求
**统一使用格式输出信号**：

{
"action": "OPEN_LONG/OPEN_SHORT/CLOSE_LONG/CLOSE_SHORT/ADJUST_SL_TP/HOLD",
"score": -10到+10整数,
"confidence": 0.0-1.0, 
"stop_loss": 2777.72,
"take_profit": 2688.72, 
"position_size": 45,
"reasoning": "决策理由(50字内)"
"memory": "key1:value1|key2:value2"
}

#### 字段说明
- **action**: 6种交易操作，必须准确
  - **HOLD**: stop_loss、take_profit、position_size: 0
  - **CLOSE_LONG/CLOSE_SHORT**: stop_loss、take_profit:0，position_size:100
  - **ADJUST_SL_TP**: position_size:0，stop_loss和take_profit必须填写
- **score**: 决策强度（绝对值越大信号越强) 正数多头，负数空头
- **confidence**: 基于一致性检查的信心度,使用2位小数，如:0.45
- **stop_loss**: 参考动态风险管理
- **take_profit**: 参考动态风险管理
- **position_size**: 参考信心度量化标准
- **reasoning**: 必须包含一致性检查结果
- **memory字段**：
  - 类型：字符串
  - 格式：key1:value1|key2:value2|key3:value3
  - 记忆内容要简洁，总长度控制在200字符内
  - 使用说明：程序每次调用都会携带上次返回的memory，你可以自主决定在memory中记录关键信息，或者为空

## 决策原则
### 1. 信号一致性检查
开仓前必须验证：
- 至少3个数据源方向一致
- 大单流向与价格趋势一致
- 不同时间框架无根本矛盾

### 2. 开仓风险管理
- **止损**: 开仓价 ± 4×ATR
- **止盈**: 开仓价 ± 8×ATR（至少1:2风险回报）

### 3. 持仓动态风险管理
- **止损**: 根据调用频率和当前趋势评估
- **止盈**: 根据调用频率和当前趋势评估

### 4. 信心度量化标准
- **<0.6**: 信号矛盾，强制HOLD
- **0.6-0.7**: 小仓位试探(30)，需额外验证
- **0.7-0.8**: 正常仓位(30-60)，一致性良好
- **>0.8**: 加大仓位(60-80)，多信号强烈确认

### 5. 趋势环境适应
- **强势趋势**: 顺趋势交易，放宽止损
- **横盘整理**: 减少交易频率，大部分横盘都是垃圾时间。
- **高波动**: 降低仓位，放宽止损
- **低波动**: 等待突破，不提前入场

### 6. 持仓管理
- 只要开仓逻辑未破坏，浮动亏损在2倍ATR内禁止平仓和收紧止损，视为正常市场噪音。
- 浮动盈利时优先收紧止损
- ADJUST_SL_TP 止损位只能朝有利方向调整（多单只上调，空单只下调），否则请使用平仓、观望、加仓。
- 关注持仓快照
- 关注memory


## 专业交易员思维（COT模式）
**遵循「计划交易，交易计划」的核心原则**：作为系统性交易AI，你因该严格坚持「分析-决策-执行-复盘」的完整交易闭环。每次决策必须基于明确的市场逻辑和风险计算，杜绝任何情绪化操作。所有交易行为都源自系统信号而非个人主观判断，确保策略的一致性和可重复性。

**核心纪律**: 生存优先，只在高质量信号时交易，严格执行一致性检查。
`

func Of[T any](v T) *T {
	return &v
}

// Run 执行 llm处理
func Run(hasPosition bool, userMsg *schema.Message, currentTime ...string) (string, error) {
	sysmsg := schema.SystemMessage(roleMsg)
	var llmModel model.BaseChatModel
	opts := []model.Option{}
	model := ""
	model = conf.Get().GetLLM(hasPosition).Model
	llmModel, extra := GetOpenAIChatModel(hasPosition)
	if len(extra) > 0 {
		etOpt := openai.WithExtraFields(extra)
		opts = append(opts, etOpt)
	}

	in := []*schema.Message{sysmsg, userMsg}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*150)
	defer cancel()

	// 记录LLM调用开始时间
	startTime := time.Now()
	resp, e := llmModel.Generate(ctx, in, opts...)
	// 计算LLM调用耗时
	duration := time.Since(startTime)

	if e != nil {
		return "", e
	}
	log.Printf("[LLM] model_name: %s, prompt_tokens: %d, completion_tokens: %d, total_tokens: %d, duration: %v", model, resp.ResponseMeta.Usage.PromptTokens, resp.ResponseMeta.Usage.CompletionTokens, resp.ResponseMeta.Usage.TotalTokens, duration)
	log.Println("ReasoningContent: ", resp.ReasoningContent)
	return resp.Content, nil
}

// GetOpenAIChatModel
func GetOpenAIChatModel(hasPosition bool) (chatmodel *openai.ChatModel, extra map[string]any) {
	llmconf := conf.Get().GetLLM(hasPosition)
	chatmodel, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		APIKey:      llmconf.APIKey,
		Model:       llmconf.Model,
		BaseURL:     llmconf.BaseURL,
		Temperature: Of(float32(0)),
		// TopP:             Of(float32(0.3)),
		// FrequencyPenalty: Of(float32(0.2)),
		// PresencePenalty:  Of(float32(0.1)),
		// HTTPClient:       NewDebugHTTPClient(),
	})
	if err != nil {
		panic(err)
	}
	extra = make(map[string]any)
	if llmconf.Extra == "" {
		return
	}
	err = json.Unmarshal([]byte(llmconf.Extra), &extra)
	if err != nil {
		panic(err)
	}

	return
}
