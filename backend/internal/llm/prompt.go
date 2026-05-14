// Package llm 提供 TalkAboutIt 的 LLM Provider 抽象与实现。
package llm

import (
	"fmt"
	"strings"
)

// PromptPersona 描述 prompt 构建所需的 persona 信息。
type PromptPersona interface {
	PromptIdentitySection() string
	PromptStanceSection() string
	PromptCoreBeliefsSection() string
	PromptSpeakingStyleSection() string
	PromptKnowledgeScopeSection() string
	PromptInteractionRulesSection() string
	PromptDebateGoalSection() string
	PromptPreambleSection() string
	PromptReplyConstraintsSection() string
	PromptOpeningLine() string
	PromptSampleRebuttal() string
}

// DedupHintBuilder 描述去重提示的构建能力。
type DedupHintBuilder interface {
	BuildDedupHint() string
}

// BuildStaticSystemPrompt 构建不会随轮次变化的静态 system prompt。
func BuildStaticSystemPrompt(p PromptPersona) string {
	var b strings.Builder

	appendSection(&b, p.PromptIdentitySection())
	appendSection(&b, p.PromptStanceSection())
	appendSection(&b, p.PromptCoreBeliefsSection())
	appendSection(&b, p.PromptSpeakingStyleSection())
	appendSection(&b, p.PromptKnowledgeScopeSection())
	appendSection(&b, p.PromptInteractionRulesSection())
	appendSection(&b, p.PromptDebateGoalSection())
	appendSection(&b, p.PromptPreambleSection())
	appendSection(&b, p.PromptReplyConstraintsSection())

	b.WriteString("## 发言规则\n")
	b.WriteString("1. 用第一人称，像真实对话一样自然交流\n")
	b.WriteString("2. 保持你的性格特点和语言风格\n")
	b.WriteString("3. 可以参考你的经历和理念来回应\n")
	b.WriteString("4. 直接表达观点，不要过度客套\n")
	b.WriteString("5. 可以同意或反驳其他人的观点，保持真实个性\n")
	b.WriteString("6. 每次发言控制在 200-500 字之间\n")

	return strings.TrimSpace(b.String())
}

// BuildDynamicContext 构建会随轮次变化的动态 user 指令。
func BuildDynamicContext(p PromptPersona, topic string, peers []string, round int, language string, state DedupHintBuilder) string {
	var b strings.Builder

	b.WriteString("## 当前讨论\n")
	b.WriteString(fmt.Sprintf("主题：%s\n", topic))
	if len(peers) > 0 {
		b.WriteString(fmt.Sprintf("参与者：%s\n", strings.Join(peers, ", ")))
	}
	b.WriteString(fmt.Sprintf("当前轮次：第 %d 轮\n", round))
	b.WriteString("\n")

	if language == "en-US" {
		b.WriteString("7. 你必须用英文发言。无论你的 persona 原本使用什么语言，本轮讨论统一使用英文。\n")
	} else {
		b.WriteString("7. 你必须用中文发言。无论你的 persona 原本使用什么语言，本轮讨论统一使用中文（简体）。\n")
	}

	if state != nil {
		dedupHint := state.BuildDedupHint()
		if dedupHint != "" {
			b.WriteString("\n")
			b.WriteString(dedupHint)
		}
	}

	if p.PromptOpeningLine() != "" || p.PromptSampleRebuttal() != "" {
		b.WriteString("\n## 参考示例\n")
		if p.PromptOpeningLine() != "" {
			b.WriteString(fmt.Sprintf("开场示例：%s\n", p.PromptOpeningLine()))
		}
		if p.PromptSampleRebuttal() != "" {
			b.WriteString(fmt.Sprintf("反驳示例：%s\n", p.PromptSampleRebuttal()))
		}
	}

	b.WriteString("\n## 本轮任务\n")
	if round <= 1 {
		b.WriteString("请围绕主题发表开场观点。\n")
	} else {
		b.WriteString(fmt.Sprintf("现在是第 %d 轮。请直接回应其他人的最新观点，推进讨论，并避免重复你之前已经表达过的论点。\n", round))
	}

	return strings.TrimSpace(b.String())
}

// BuildSystemPrompt 兼容旧调用方，返回静态 system 与动态上下文的拼接结果。
func BuildSystemPrompt(p PromptPersona, topic string, peers []string, round int, language string, state DedupHintBuilder) string {
	return BuildStaticSystemPrompt(p) + "\n\n" + BuildDynamicContext(p, topic, peers, round, language, state)
}

func appendSection(b *strings.Builder, section string) {
	section = strings.TrimSpace(section)
	if section == "" {
		return
	}
	b.WriteString(section)
	b.WriteString("\n\n")
}
