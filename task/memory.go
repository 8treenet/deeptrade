package task

import (
	"fmt"
	"strings"
	"time"
)

const maxMemoryEntries = 10

type MemoryEntry struct {
	Timestamp time.Time
	Content   string
}

var memory []MemoryEntry

func SetMemory(in string) {
	entry := MemoryEntry{
		Timestamp: time.Now(),
		Content:   in,
	}

	// 添加到末尾
	memory = append(memory, entry)

	// 保持最多10条记录
	if len(memory) > maxMemoryEntries {
		memory = memory[len(memory)-maxMemoryEntries:]
	}
}

func GetMemory() string {
	if len(memory) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, entry := range memory {
		sb.WriteString(fmt.Sprintf("[%s] %s", entry.Timestamp.Format("2006-01-02 15:04:05"), entry.Content))
		if i < len(memory)-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
