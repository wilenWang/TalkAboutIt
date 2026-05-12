// Package llm 提供 TalkAboutIt 的 LLM Provider 抽象与实现。
package llm

import (
	"strings"
	"testing"

	"github.com/wilenwang/talkaboutit/internal/persona"
)

// TestBuildSystemPrompt 验证 llm.BuildSystemPrompt 能正确展开所有模板字段。
func TestBuildSystemPrompt(t *testing.T) {
	p := persona.Persona{
		SchemaVersion: "persona.v1",
		ID:            "test-persona",
		Name:          "Test Persona",
		DisplayName:   "Test",
		Avatar:        "🧪",
		RoleTitle:     "Test Role",
		Description:   "这是一个测试角色，用于验证 system prompt 构建逻辑。",
		Tags:          []string{"tag1", "tag2"},
		Language: persona.Language{
			Primary: "zh-CN",
			Allowed: []string{"zh-CN", "en-US"},
		},
		Stance: persona.Stance{
			DefaultPosition: "测试默认立场。",
			Intensity:       3,
			Biases:          []string{"倾向 A", "倾向 B"},
			Taboos:          []string{"禁忌 A"},
		},
		CoreBeliefs: []persona.CoreBelief{
			{Belief: "信念一", Priority: 5, Rationale: "理由一"},
		},
		SpeakingStyle: persona.SpeakingStyle{
			Tone:              "direct",
			Cadence:           "short_punchy",
			Verbosity:         2,
			SignaturePatterns: []string{"模式一"},
			Do:                []string{"应该做"},
			Dont:              []string{"不应该做"},
		},
		KnowledgeScope: persona.KnowledgeScope{
			Domains:          []string{"领域 A"},
			ExpertiseLevel:   map[string]int{"领域 A": 4, "领域 B": 2},
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
			SecondaryGoals: []string{"次要目标"},
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

	topic := "AI 会取代程序员吗？"
	peers := []string{"Steve Jobs", "Elon Musk"}
	round := 2

	prompt := BuildSystemPrompt(p, topic, peers, round)

	// 验证关键字段是否出现在 prompt 中
	checks := []string{
		"Test Persona",
		"Test Role",
		"这是一个测试角色",
		"测试默认立场",
		"信念一",
		"AI 会取代程序员吗？",
		"Steve Jobs",
		"Elon Musk",
		"第 2 轮",
		"开场示例",
		"反驳示例",
		// 新增字段
		"tag1",
		"zh-CN",
		"en-US",
		"领域 A: 4/5",
		"领域 B: 2/5",
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check) {
			t.Errorf("system prompt 中应包含 %q", check)
		}
	}
}

// TestBuildSystemPrompt_SteveJobs 验证 Steve Jobs 的 system prompt 构建符合预期。
func TestBuildSystemPrompt_SteveJobs(t *testing.T) {
	loader := persona.NewLoader("../../personas")
	p, err := loader.LoadOne("steve-jobs")
	if err != nil {
		t.Fatalf("加载 steve-jobs 失败: %v", err)
	}

	prompt := BuildSystemPrompt(p, "AI 会取代产品经理吗？", []string{"Elon Musk", "Naval Ravikant"}, 1)

	// 验证关键内容
	if !strings.Contains(prompt, "Steve Jobs") {
		t.Error("prompt 应包含角色名")
	}
	if !strings.Contains(prompt, "AI 会取代产品经理吗？") {
		t.Error("prompt 应包含讨论主题")
	}
	if !strings.Contains(prompt, "Elon Musk") {
		t.Error("prompt 应包含同伴")
	}
	if !strings.Contains(prompt, "Simplicity is the ultimate sophistication") {
		t.Error("prompt 应包含核心信念")
	}
}
