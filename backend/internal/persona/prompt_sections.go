package persona

import (
	"fmt"
	"strings"
)

// PromptIdentitySection 返回 persona 的身份定位区块。
func (p Persona) PromptIdentitySection() string {
	var b strings.Builder
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
	return b.String()
}

// PromptStanceSection 返回 persona 的立场区块。
func (p Persona) PromptStanceSection() string {
	if p.Stance.DefaultPosition == "" {
		return ""
	}

	var b strings.Builder
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
	return b.String()
}

// PromptCoreBeliefsSection 返回 persona 的核心信念区块。
func (p Persona) PromptCoreBeliefsSection() string {
	if len(p.CoreBeliefs) == 0 {
		return ""
	}

	var b strings.Builder
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
	return b.String()
}

// PromptSpeakingStyleSection 返回 persona 的表达方式区块。
func (p Persona) PromptSpeakingStyleSection() string {
	var b strings.Builder
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
	return b.String()
}

// PromptKnowledgeScopeSection 返回 persona 的知识边界区块。
func (p Persona) PromptKnowledgeScopeSection() string {
	var b strings.Builder
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
	return b.String()
}

// PromptInteractionRulesSection 返回 persona 的互动规则区块。
func (p Persona) PromptInteractionRulesSection() string {
	if p.InteractionRules.AddressOthers == "" && p.InteractionRules.DisagreementStyle == "" &&
		p.InteractionRules.InterruptionPolicy == "" && p.InteractionRules.QuestionPolicy == "" &&
		p.InteractionRules.ConcessionPolicy == "" && len(p.InteractionRules.Avoid) == 0 {
		return ""
	}

	var b strings.Builder
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
	return b.String()
}

// PromptDebateGoalSection 返回 persona 的辩论目标区块。
func (p Persona) PromptDebateGoalSection() string {
	if p.DebateGoal.PrimaryGoal == "" {
		return ""
	}

	var b strings.Builder
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
	return b.String()
}

// PromptPreambleSection 返回 persona 的前置说明区块。
func (p Persona) PromptPreambleSection() string {
	if p.Prompting.SystemPreamble == "" {
		return ""
	}
	return fmt.Sprintf("## 前置说明\n%s\n", p.Prompting.SystemPreamble)
}

// PromptReplyConstraintsSection 返回 persona 的回复约束区块。
func (p Persona) PromptReplyConstraintsSection() string {
	if len(p.Prompting.ReplyConstraints) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("## 回复约束\n")
	for _, rc := range p.Prompting.ReplyConstraints {
		b.WriteString(fmt.Sprintf("- %s\n", rc))
	}
	return b.String()
}

// PromptOpeningLine 返回 persona 的开场示例。
func (p Persona) PromptOpeningLine() string { return p.Examples.OpeningLine }

// PromptSampleRebuttal 返回 persona 的反驳示例。
func (p Persona) PromptSampleRebuttal() string { return p.Examples.SampleRebuttal }
