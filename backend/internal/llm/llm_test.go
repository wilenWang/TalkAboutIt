// Package llm 提供 TalkAboutIt 的 LLM Provider 抽象与实现。
package llm

import (
	"strings"
	"testing"

	"github.com/wilenwang/talkaboutit/internal/persona"
)

// TestBuildSystemPrompt_ContainsAllPersonaFields 验证 BuildSystemPrompt 输出包含所有 persona 字段。
func TestBuildSystemPrompt_ContainsAllPersonaFields(t *testing.T) {
	p := persona.Persona{
		SchemaVersion: "persona.v1",
		ID:            "test-all-fields",
		Name:          "Test Name",
		DisplayName:   "Test Display",
		Avatar:        "🧪",
		RoleTitle:     "Test Role",
		Description:   "测试描述内容",
		Tags:          []string{"tagA", "tagB"},
		Language: persona.Language{
			Primary: "zh-CN",
			Allowed: []string{"zh-CN", "en-US"},
		},
		Stance: persona.Stance{
			DefaultPosition: "默认立场测试",
			Intensity:       4,
			Biases:          []string{"偏见一", "偏见二"},
			Taboos:          []string{"禁忌一"},
		},
		CoreBeliefs: []persona.CoreBelief{
			{Belief: "信念A", Priority: 5, Rationale: "理由A"},
			{Belief: "信念B", Priority: 3, Rationale: "理由B"},
		},
		SpeakingStyle: persona.SpeakingStyle{
			Tone:              "direct",
			Cadence:           "short_punchy",
			Verbosity:         2,
			SignaturePatterns: []string{"模式一", "模式二"},
			Do:                []string{"应该做A"},
			Dont:              []string{"不应该做B"},
		},
		KnowledgeScope: persona.KnowledgeScope{
			Domains:          []string{"领域X", "领域Y"},
			ExpertiseLevel:   map[string]int{"领域X": 5, "领域Y": 3},
			TimeCutoff:       "2024-01-01",
			AllowedInference: "medium",
			UnknownHandling:  "未知处理说明",
			ForbiddenClaims:  []string{"禁止声明一"},
		},
		InteractionRules: persona.InteractionRules{
			AddressOthers:      "直接称呼",
			DisagreementStyle:  "直接反驳",
			InterruptionPolicy: "allowed",
			QuestionPolicy:     "尖锐提问",
			ConcessionPolicy:   "有限让步",
			Avoid:              []string{"避免一"},
		},
		DebateGoal: persona.DebateGoal{
			PrimaryGoal:    "主要目标",
			SecondaryGoals: []string{"次要目标一"},
			WinCondition:   "胜利条件",
			LossCondition:  "失败条件",
		},
		Prompting: persona.Prompting{
			SystemPreamble:   "前置说明",
			ReplyConstraints: []string{"约束一"},
		},
		Examples: persona.Examples{
			OpeningLine:    "开场示例",
			SampleRebuttal: "反驳示例",
		},
	}

	topic := "测试主题"
	peers := []string{"Peer A", "Peer B"}
	round := 3

	prompt := BuildSystemPrompt(p, topic, peers, round)

	// 定义所有应出现的字段检查项
	checks := []string{
		// 身份定位
		"Test Name",
		"Test Role",
		"测试描述内容",
		"tagA",
		"tagB",
		"zh-CN",
		"en-US",
		// 立场
		"默认立场测试",
		"4/5",
		"偏见一",
		"偏见二",
		"禁忌一",
		// 核心信念
		"信念A",
		"理由A",
		"信念B",
		"理由B",
		// 表达方式
		"direct",
		"short_punchy",
		"2/5",
		"模式一",
		"应该做A",
		"不应该做B",
		// 知识边界
		"领域X",
		"领域Y",
		"领域X: 5/5",
		"领域Y: 3/5",
		"2024-01-01",
		"medium",
		"未知处理说明",
		"禁止声明一",
		// 互动规则
		"直接称呼",
		"直接反驳",
		"allowed",
		"尖锐提问",
		"有限让步",
		"避免一",
		// 辩论目标
		"主要目标",
		"次要目标一",
		"胜利条件",
		"失败条件",
		// 额外提示词约束
		"前置说明",
		"约束一",
		// 讨论上下文
		"测试主题",
		"Peer A",
		"Peer B",
		"第 3 轮",
		// 参考示例
		"开场示例",
		"反驳示例",
		// 发言规则（固定追加）
		"用第一人称",
		"200-500 字",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Errorf("system prompt 中应包含 %q", check)
		}
	}
}

// TestBuildSystemPrompt_EmptyFields 验证空字段不会导致异常输出。
func TestBuildSystemPrompt_EmptyFields(t *testing.T) {
	p := persona.Persona{
		SchemaVersion: "persona.v1",
		ID:            "empty-test",
		Name:          "Empty Test",
		RoleTitle:     "Tester",
		Description:   "",
		Tags:          []string{},
		Language: persona.Language{
			Primary: "zh-CN",
			Allowed: []string{},
		},
		Stance: persona.Stance{
			DefaultPosition: "",
			Intensity:       0,
			Biases:          []string{},
			Taboos:          []string{},
		},
		CoreBeliefs: []persona.CoreBelief{},
		SpeakingStyle: persona.SpeakingStyle{
			Tone:              "",
			Cadence:           "",
			Verbosity:         0,
			SignaturePatterns: []string{},
			Do:                []string{},
			Dont:              []string{},
		},
		KnowledgeScope: persona.KnowledgeScope{
			Domains:          []string{},
			ExpertiseLevel:   map[string]int{},
			TimeCutoff:       "",
			AllowedInference: "",
			UnknownHandling:  "",
			ForbiddenClaims:  []string{},
		},
		InteractionRules: persona.InteractionRules{
			AddressOthers:      "",
			DisagreementStyle:  "",
			InterruptionPolicy: "",
			QuestionPolicy:     "",
			ConcessionPolicy:   "",
			Avoid:              []string{},
		},
		DebateGoal: persona.DebateGoal{
			PrimaryGoal:    "",
			SecondaryGoals: []string{},
			WinCondition:   "",
			LossCondition:  "",
		},
		Prompting: persona.Prompting{
			SystemPreamble:   "",
			ReplyConstraints: []string{},
		},
		Examples: persona.Examples{
			OpeningLine:    "",
			SampleRebuttal: "",
		},
	}

	prompt := BuildSystemPrompt(p, "空字段测试", []string{}, 1)

	// 验证基本结构仍然存在
	if !strings.Contains(prompt, "Empty Test") {
		t.Error("prompt 应包含角色名")
	}
	if !strings.Contains(prompt, "空字段测试") {
		t.Error("prompt 应包含主题")
	}
	if !strings.Contains(prompt, "第 1 轮") {
		t.Error("prompt 应包含轮次")
	}

	// 验证空字段对应的区块没有输出无意义内容
	if strings.Contains(prompt, "## 你的立场\n\n") {
		t.Error("空立场不应输出空区块")
	}
	if strings.Contains(prompt, "## 你的核心信念\n\n") {
		t.Error("空核心信念不应输出空区块")
	}
	if strings.Contains(prompt, "## 本轮目标\n\n") {
		t.Error("空辩论目标不应输出空区块")
	}
	if strings.Contains(prompt, "## 参考示例\n\n") {
		t.Error("空示例不应输出空区块")
	}
}
