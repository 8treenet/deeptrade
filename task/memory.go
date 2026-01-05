package task

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const maxMemoryEntries = 30 // 增加到30条

// MemoryEntry 结构化的记忆条目
type MemoryEntry struct {
	// 时间戳
	Timestamp time.Time `json:"timestamp"`

	// 决策信息
	Action     string  `json:"action"`     // 操作类型: OPEN_LONG, OPEN_SHORT, CLOSE_LONG, CLOSE_SHORT, HOLD等
	Score      int     `json:"score"`      // 评分: -10到+10
	Confidence float64 `json:"confidence"` // 置信度: 0.0-1.0
	Reasoning  string  `json:"reasoning"`  // 决策理由

	// 市场状态快照
	MarketState MarketStateSnapshot `json:"market_state"`

	// 执行结果 (后续更新)
	Executed bool `json:"executed"` // 是否执行
	// 决策ID (用于关联开仓和平仓)
	DecisionID string `json:"decision_id"`
}

// MarketStateSnapshot 市场状态快照
type MarketStateSnapshot struct {
	Price           float64 `json:"price"`            // 当时价格
	Trend           string  `json:"trend"`            // 趋势: UP/DOWN/SIDEWAYS
	TrendStrength   float64 `json:"trend_strength"`   // 趋势强度 0-10
	Volatility      float64 `json:"volatility"`       // 波动率百分比
	RSI             float64 `json:"rsi"`              // RSI值
	MACDSignal      string  `json:"macd_signal"`      // MACD信号: BULLISH/BEARISH/NEUTRAL
	SupportLevel    float64 `json:"support_level"`    // 最近支撑位
	ResistanceLevel float64 `json:"resistance_level"` // 最近阻力位
	FundingRate     float64 `json:"funding_rate"`     // 资金费率
}

var memory []*MemoryEntry

// SetMemory 添加新的记忆条目
func SetMemory(action string, score int, confidence float64, reasoning string, marketState MarketStateSnapshot) string {
	decisionID := generateDecisionID(action)

	entry := &MemoryEntry{
		Timestamp:   time.Now(),
		Action:      action,
		Score:       score,
		Confidence:  confidence,
		Reasoning:   reasoning,
		MarketState: marketState,
		Executed:    false,
		DecisionID:  decisionID,
	}

	// 添加到切片
	memory = append(memory, entry)

	// 保持最多maxMemoryEntries条记录，按重要性保留
	if len(memory) > maxMemoryEntries {
		// 按置信度和评分排序，保留重要决策
		sort.Slice(memory, func(i, j int) bool {
			// 高置信度优先
			return memory[i].Confidence > memory[j].Confidence
		})
		memory = memory[:maxMemoryEntries]
	}

	return decisionID
}

// SetMemoryFromSignal 从TradingSignal创建记忆条目
func SetMemoryFromSignal(signal *TradingSignal, marketState MarketStateSnapshot) string {
	return SetMemory(signal.Action, signal.Score, signal.Confidence, signal.Memory, marketState)
}

// UpdateMemoryResult 更新记忆条目的执行结果
func UpdateMemoryResult(decisionID string, entryPrice, exitPrice, realizedPnL float64, outcome string) {
	// for _, entry := range memory {
	// 	if entry.DecisionID == decisionID {
	// 		entry.Executed = true
	// 		entry.EntryPrice = entryPrice
	// 		entry.ExitPrice = exitPrice
	// 		entry.RealizedPnL = realizedPnL
	// 		entry.Outcome = outcome
	// 		return
	// 	}
	// }
}

// MarkMemoryExecuted 标记记忆条目已执行
func MarkMemoryExecuted(decisionID string, entryPrice float64) {
	for _, entry := range memory {
		if entry.DecisionID == decisionID {
			entry.Executed = true
			return
		}
	}
}

// GetMemory 获取格式化的记忆字符串
func GetMemory() string {
	if len(memory) == 0 {
		return "暂无历史记忆"
	}

	count := 0
	var sb strings.Builder
	// 显示最近的记忆（倒序，最新的在前）
	for i := len(memory) - 1; i >= 0; i-- {
		entry := memory[i]
		sb.WriteString(formatMemoryEntry(entry))
		count += 1
		if count >= 15 {
			break
		}
	}

	return sb.String()
}

// formatMemoryEntry 格式化单个记忆条目
func formatMemoryEntry(entry *MemoryEntry) string {
	var sb strings.Builder

	// 时间和操作
	sb.WriteString(fmt.Sprintf("[%s] %s", entry.Timestamp.Format("01-02 15:04"), entry.Action))

	// 评分和置信度
	sb.WriteString(fmt.Sprintf(" | score:%d confidence:%.0f%%\n", entry.Score, entry.Confidence*100))

	// 市场状态
	if entry.MarketState.Price > 0 {
		sb.WriteString(fmt.Sprintf("  价格:%.2f 趋势:%s(%.1f) RSI:%.1f\n",
			entry.MarketState.Price, entry.MarketState.Trend, entry.MarketState.TrendStrength, entry.MarketState.RSI))
	}

	// 决策理由（截取前100字符）
	if entry.Reasoning != "" {
		sb.WriteString(fmt.Sprintf("  记忆: %s\n", entry.Reasoning))
	}

	sb.WriteString("───────────────────────────────────────────────────────────\n")
	return sb.String()
}

// generateDecisionID 生成决策ID
func generateDecisionID(action string) string {
	return fmt.Sprintf("%s_%d", action, time.Now().UnixNano())
}

// abs 返回绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// GetLastOpenDecision 获取最近的开仓决策
func GetLastOpenDecision() *MemoryEntry {
	for i := len(memory) - 1; i >= 0; i-- {
		if memory[i].Action == "OPEN_LONG" || memory[i].Action == "OPEN_SHORT" ||
			memory[i].Action == "ADD_LONG" || memory[i].Action == "ADD_SHORT" {
			return memory[i]
		}
	}
	return nil
}
