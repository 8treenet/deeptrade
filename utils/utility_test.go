package utils_test

import (
	"context"
	"deeptrade/conf"
	"deeptrade/utils"
	"os"
	"testing"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

func init() {
	// 设置配置文件目录路径，确保能够找到项目配置
	os.Setenv("deeptrade_conf", "/Users/ys/work/my/deeptrade/conf")
	// 初始化项目配置系统
	conf.EntryPoint()
}

func TestRun(t *testing.T) {
	t.Log(time.Now().Weekday() == 0)
	return
	t.Log(utils.Run(false, schema.UserMessage("你好，我想测试下思考的传参")))
}

func TestDeepSeek(t *testing.T) {
	opts := []model.Option{}
	model := ""
	model = conf.Get().GetLLM(false).Model
	llmModel, prefx := utils.GetOpenAIChatModel(false)
	if len(prefx) > 0 {
		etOpt := openai.WithExtraFields(prefx)
		opts = append(opts, etOpt)
	}
	t.Log(prefx)

	msg := schema.UserMessage(`分析下这个记录
	## 最近交易记录
─────────────────────────────────────────────────────────
1. 平多仓 ✅ 已成交 (已平仓)
   订单ID: 7064834189
   数量: 0.276 ETH, 价格: 市价成交
   方向: SELL
   持仓方向: LONG
   时间: 2025-11-11 14:22:31
─────────────────────────────────────────────────────────
2. 开多仓 ✅ 已成交 (已平仓)
   订单ID: 7064798356
   数量: 0.276 ETH, 价格: 市价成交
   方向: BUY
   持仓方向: LONG
   时间: 2025-11-11 14:21:54
─────────────────────────────────────────────────────────
3. 平多仓 ✅ 已成交 (已平仓)
   订单ID: 7062801089
   数量: 3.961 ETH, 价格: 市价成交
   方向: SELL
   持仓方向: LONG
   时间: 2025-11-11 13:38:17
─────────────────────────────────────────────────────────
4. 开多仓 ✅ 已成交 (已平仓)
   订单ID: 7039724666
   数量: 0.676 ETH, 价格: 市价成交
   方向: BUY
   持仓方向: LONG
   时间: 2025-11-11 01:47:29
─────────────────────────────────────────────────────────
5. 开多仓 ✅ 已成交 (已平仓)
   订单ID: 7038935116
   数量: 3.285 ETH, 价格: 市价成交
   方向: BUY
   持仓方向: LONG
   时间: 2025-11-11 01:27:14
─────────────────────────────────────────────────────────`)
	//etOpt := openai.WithExtraFields(map[string]any{"enable_thinking": true})
	out, err := llmModel.Generate(context.Background(), []*schema.Message{msg}, opts...)
	if err != nil {
		panic(err)
	}
	t.Log("model:", model)
	t.Log("ReasoningContent:", out.ReasoningContent)
	t.Log("Content:", out.Content)
}
