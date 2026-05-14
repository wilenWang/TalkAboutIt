package persona

import (
	"fmt"
	"strings"
)

// PerPersonaState 维护单个人物在当前 session 中的运行时状态。
type PerPersonaState struct {
	UsedArguments []string // 最近使用过的核心论点（最多保留 5 条）
}

// RecordArgument 记录一条已使用的论点。
func (s *PerPersonaState) RecordArgument(argument string) {
	s.UsedArguments = append(s.UsedArguments, argument)
	if len(s.UsedArguments) > 5 {
		s.UsedArguments = s.UsedArguments[len(s.UsedArguments)-5:]
	}
}

// BuildDedupHint 生成去重提示，注入到 system prompt 中。
// 如果没有已用论点则返回空字符串。
func (s *PerPersonaState) BuildDedupHint() string {
	if len(s.UsedArguments) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("## 注意：避免重复\n")
	b.WriteString("你已经在之前的发言中使用过以下核心论点，请不要逐字重复：\n")
	for _, arg := range s.UsedArguments {
		b.WriteString(fmt.Sprintf("- %s\n", arg))
	}
	b.WriteString("请从新的角度、用新的案例或用不同的措辞来表达你的立场。\n")
	return b.String()
}

// ExtractArgument 从发言内容中简单提取核心论点（取前 80 字符作为摘要）。
// 后续可以替换为 LLM 摘要，但当前用简单截断即可。
func ExtractArgument(content string) string {
	runes := []rune(content)
	if len(runes) > 80 {
		return string(runes[:80]) + "..."
	}
	return content
}
