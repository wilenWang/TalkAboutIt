// Package llm 提供 TalkAboutIt 的 LLM Provider 抽象与实现。
package llm

import (
	"fmt"
	"strings"

	"github.com/wilenwang/talkaboutit/internal/persona"
)

// BuildSystemPrompt 根据 persona、讨论主题、同伴列表和当前轮次构建完整的 system prompt。
// 要求角色以第一人称发言，每次发言限制在 200-500 字。
func BuildSystemPrompt(p persona.Persona, topic string, peers []string, round int) string {
	var b strings.Builder

	// 身份定位
	b.WriteString(fmt.Sprintf("你是 %s（%s）。\n", p.Name, p.RoleTitle))
	if p.DisplayName != "" && p.DisplayName != p.Name {
		b.WriteString(fmt.Sprintf("显示名称：%s\n", p.DisplayName))
	}
	if p.Description != "" {
		b.WriteString(fmt.Sprintf("%s\n", p.Description))
	}
	if len(p.Tags) > 0 {
		b.WriteString(fmt.Sprintf("标签：%s\n", strings.Join(p.Tags, ", ")))
	}
	if p.Language.Primary != "" {
		b.WriteString(fmt.Sprintf("主要语言：%s", p.Language.Primary))
		if len(p.Language.Allowed) > 0 {
			b.WriteString(fmt.Sprintf("（允许：%s）", strings.Join(p.Language.Allowed, ", ")))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// 立场
	if p.Stance.DefaultPosition != "" {
		b.WriteString("## 你的立场\n")
		b.WriteString(fmt.Sprintf("默认立场：%s\n", p.Stance.DefaultPosition))
		if p.Stance.Intensity > 0 {
			b.WriteString(fmt.Sprintf("立场强度：%d/5\n", p.Stance.Intensity))
		}
		if len(p.Stance.Biases) > 0 {
			b.WriteString("倾向：\n")
			for _, bias := range p.Stance.Biases {
				b.WriteString(fmt.Sprintf("- %s\n", bias))
			}
		}
		if len(p.Stance.Taboos) > 0 {
			b.WriteString("禁忌：\n")
			for _, taboo := range p.Stance.Taboos {
				b.WriteString(fmt.Sprintf("- %s\n", taboo))
			}
		}
		b.WriteString("\n")
	}

	// 核心信念
	if len(p.CoreBeliefs) > 0 {
		b.WriteString("## 你的核心信念\n")
		for _, cb := range p.CoreBeliefs {
			b.WriteString(fmt.Sprintf("- %s", cb.Belief))
			if cb.Priority > 0 {
				b.WriteString(fmt.Sprintf("（优先级 %d/5）", cb.Priority))
			}
			if cb.Rationale != "" {
				b.WriteString(fmt.Sprintf(" —— %s", cb.Rationale))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// 表达方式
	b.WriteString("## 你的表达方式\n")
	if p.SpeakingStyle.Tone != "" {
		b.WriteString(fmt.Sprintf("语气：%s\n", p.SpeakingStyle.Tone))
	}
	if p.SpeakingStyle.Cadence != "" {
		b.WriteString(fmt.Sprintf("节奏：%s\n", p.SpeakingStyle.Cadence))
	}
	if p.SpeakingStyle.Verbosity > 0 {
		b.WriteString(fmt.Sprintf("详细程度：%d/5\n", p.SpeakingStyle.Verbosity))
	}
	if len(p.SpeakingStyle.SignaturePatterns) > 0 {
		b.WriteString("标志性表达模式：\n")
		for _, pat := range p.SpeakingStyle.SignaturePatterns {
			b.WriteString(fmt.Sprintf("- %s\n", pat))
		}
	}
	if len(p.SpeakingStyle.Do) > 0 {
		b.WriteString("应该：\n")
		for _, d := range p.SpeakingStyle.Do {
			b.WriteString(fmt.Sprintf("- %s\n", d))
		}
	}
	if len(p.SpeakingStyle.Dont) > 0 {
		b.WriteString("不应该：\n")
		for _, d := range p.SpeakingStyle.Dont {
			b.WriteString(fmt.Sprintf("- %s\n", d))
		}
	}
	b.WriteString("\n")

	// 知识边界
	b.WriteString("## 你的知识边界\n")
	if len(p.KnowledgeScope.Domains) > 0 {
		b.WriteString(fmt.Sprintf("领域：%s\n", strings.Join(p.KnowledgeScope.Domains, ", ")))
	}
	if len(p.KnowledgeScope.ExpertiseLevel) > 0 {
		b.WriteString("专业水平：\n")
		for domain, level := range p.KnowledgeScope.ExpertiseLevel {
			b.WriteString(fmt.Sprintf("- %s: %d/5\n", domain, level))
		}
	}
	if p.KnowledgeScope.TimeCutoff != "" {
		b.WriteString(fmt.Sprintf("时间截止：%s\n", p.KnowledgeScope.TimeCutoff))
	}
	if p.KnowledgeScope.AllowedInference != "" {
		b.WriteString(fmt.Sprintf("允许推断程度：%s\n", p.KnowledgeScope.AllowedInference))
	}
	if p.KnowledgeScope.UnknownHandling != "" {
		b.WriteString(fmt.Sprintf("未知内容处理：%s\n", p.KnowledgeScope.UnknownHandling))
	}
	if len(p.KnowledgeScope.ForbiddenClaims) > 0 {
		b.WriteString("禁止声明：\n")
		for _, fc := range p.KnowledgeScope.ForbiddenClaims {
			b.WriteString(fmt.Sprintf("- %s\n", fc))
		}
	}
	b.WriteString("\n")

	// 互动规则
	if p.InteractionRules.AddressOthers != "" || p.InteractionRules.DisagreementStyle != "" ||
		p.InteractionRules.InterruptionPolicy != "" || len(p.InteractionRules.Avoid) > 0 {
		b.WriteString("## 互动规则\n")
		if p.InteractionRules.AddressOthers != "" {
			b.WriteString(fmt.Sprintf("称呼他人：%s\n", p.InteractionRules.AddressOthers))
		}
		if p.InteractionRules.DisagreementStyle != "" {
			b.WriteString(fmt.Sprintf("不同意时的风格：%s\n", p.InteractionRules.DisagreementStyle))
		}
		if p.InteractionRules.InterruptionPolicy != "" {
			b.WriteString(fmt.Sprintf("打断策略：%s\n", p.InteractionRules.InterruptionPolicy))
		}
		if p.InteractionRules.QuestionPolicy != "" {
			b.WriteString(fmt.Sprintf("提问策略：%s\n", p.InteractionRules.QuestionPolicy))
		}
		if p.InteractionRules.ConcessionPolicy != "" {
			b.WriteString(fmt.Sprintf("让步策略：%s\n", p.InteractionRules.ConcessionPolicy))
		}
		if len(p.InteractionRules.Avoid) > 0 {
			b.WriteString("避免：\n")
			for _, av := range p.InteractionRules.Avoid {
				b.WriteString(fmt.Sprintf("- %s\n", av))
			}
		}
		b.WriteString("\n")
	}

	// 辩论目标
	if p.DebateGoal.PrimaryGoal != "" {
		b.WriteString("## 本轮目标\n")
		b.WriteString(fmt.Sprintf("主要目标：%s\n", p.DebateGoal.PrimaryGoal))
		if len(p.DebateGoal.SecondaryGoals) > 0 {
			b.WriteString("次要目标：\n")
			for _, sg := range p.DebateGoal.SecondaryGoals {
				b.WriteString(fmt.Sprintf("- %s\n", sg))
			}
		}
		if p.DebateGoal.WinCondition != "" {
			b.WriteString(fmt.Sprintf("胜利条件：%s\n", p.DebateGoal.WinCondition))
		}
		if p.DebateGoal.LossCondition != "" {
			b.WriteString(fmt.Sprintf("失败条件：%s\n", p.DebateGoal.LossCondition))
		}
		b.WriteString("\n")
	}

	// 额外提示词约束
	if p.Prompting.SystemPreamble != "" {
		b.WriteString(fmt.Sprintf("## 前置说明\n%s\n\n", p.Prompting.SystemPreamble))
	}
	if len(p.Prompting.ReplyConstraints) > 0 {
		b.WriteString("## 回复约束\n")
		for _, rc := range p.Prompting.ReplyConstraints {
			b.WriteString(fmt.Sprintf("- %s\n", rc))
		}
		b.WriteString("\n")
	}

	// 讨论上下文
	b.WriteString("## 当前讨论\n")
	b.WriteString(fmt.Sprintf("主题：%s\n", topic))
	if len(peers) > 0 {
		b.WriteString(fmt.Sprintf("参与者：%s\n", strings.Join(peers, ", ")))
	}
	b.WriteString(fmt.Sprintf("当前轮次：第 %d 轮\n", round))
	b.WriteString("\n")

	// 发言规则（固定追加）
	b.WriteString("## 发言规则\n")
	b.WriteString("1. 用第一人称，像真实对话一样自然交流\n")
	b.WriteString("2. 保持你的性格特点和语言风格\n")
	b.WriteString("3. 可以参考你的经历和理念来回应\n")
	b.WriteString("4. 直接表达观点，不要过度客套\n")
	b.WriteString("5. 可以同意或反驳其他人的观点，保持真实个性\n")
	b.WriteString("6. 每次发言控制在 200-500 字之间\n")

	// 示例
	if p.Examples.OpeningLine != "" || p.Examples.SampleRebuttal != "" {
		b.WriteString("\n## 参考示例\n")
		if p.Examples.OpeningLine != "" {
			b.WriteString(fmt.Sprintf("开场示例：%s\n", p.Examples.OpeningLine))
		}
		if p.Examples.SampleRebuttal != "" {
			b.WriteString(fmt.Sprintf("反驳示例：%s\n", p.Examples.SampleRebuttal))
		}
	}

	return b.String()
}
