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
	os.Setenv("deeptrade_conf", "/Users/ys/work/config/deeptrade")
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

	msg := schema.UserMessage(`你是什么模型？知识库截止时间是多少？版本号是多少？上下文长度是多少？`)
	//etOpt := openai.WithExtraFields(map[string]any{"enable_thinking": true})
	out, err := llmModel.Generate(context.Background(), []*schema.Message{msg}, opts...)
	if err != nil {
		panic(err)
	}
	t.Log("model:", model)
	t.Log("ReasoningContent:", out.ReasoningContent)
	t.Log("Content:", out.Content)
}

func TestMsg(t *testing.T) {
	llmobj, _ := utils.GetOpenAIChatModel(false)
	input := []*schema.Message{schema.UserMessage(`你是什么模型？知识库截止时间是多少？版本号是多少？上下文长度是多少？`)}
	out, e := llmobj.Generate(context.Background(), input)
	if e != nil {
		panic(e)
	}
	t.Log("Content:", out.Content, e)
}
