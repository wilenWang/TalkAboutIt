// Package llm_test 提供 TalkAboutIt 的 LLM Provider 抽象与实现测试。
package llm_test

import (
	"strings"
	"testing"

	"github.com/wilenwang/talkaboutit/internal/llm"
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

	prompt := llm.BuildSystemPrompt(p, topic, peers, round, "zh-CN", nil)

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

// TestBuildSystemPrompt_DedupHint 验证已用论点去重提示能正确注入。
func TestBuildSystemPrompt_DedupHint(t *testing.T) {
	p := persona.Persona{
		ID:        "dedup-test",
		Name:      "Dedup",
		RoleTitle: "Tester",
		Language: persona.Language{
			Primary: "zh-CN",
		},
	}

	state := &persona.PerPersonaState{}
	state.RecordArgument("论点一")
	state.RecordArgument("论点二")

	prompt := llm.BuildSystemPrompt(p, "测试主题", nil, 1, "zh-CN", state)

	if !strings.Contains(prompt, "注意：避免重复") {
		t.Error("prompt 应包含去重提示标题")
	}
	if !strings.Contains(prompt, "论点一") {
		t.Error("prompt 应包含已用论点一")
	}
	if !strings.Contains(prompt, "论点二") {
		t.Error("prompt 应包含已用论点二")
	}
}

// TestBuildSystemPrompt_SteveJobs 验证 Steve Jobs 的 system prompt 构建符合预期。
func TestBuildSystemPrompt_SteveJobs(t *testing.T) {
	loader := persona.NewLoader("../../personas")
	p, err := loader.LoadOne("steve-jobs")
	if err != nil {
		t.Fatalf("加载 steve-jobs 失败: %v", err)
	}

	prompt := llm.BuildSystemPrompt(p, "AI 会取代产品经理吗？", []string{"Elon Musk", "Naval Ravikant"}, 1, "zh-CN", nil)

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

func TestBuildStaticAndDynamicPromptSplit(t *testing.T) {
	p := persona.Persona{
		ID:          "split-test",
		Name:        "Split Test",
		RoleTitle:   "Tester",
		Description: "用于验证静态与动态 prompt 拆分。",
		Language: persona.Language{
			Primary: "zh-CN",
		},
		Examples: persona.Examples{
			OpeningLine: "开场示例",
		},
	}
	state := &persona.PerPersonaState{}
	state.RecordArgument("不要重复这个论点")

	staticPrompt := llm.BuildStaticSystemPrompt(p)
	if strings.Contains(staticPrompt, "当前轮次") {
		t.Fatal("静态 prompt 不应包含当前轮次")
	}
	if strings.Contains(staticPrompt, "你必须用中文发言") {
		t.Fatal("静态 prompt 不应包含语言规则 7")
	}

	dynamicPrompt := llm.BuildDynamicContext(p, "测试主题", []string{"Peer A"}, 2, "zh-CN", state)
	for _, check := range []string{"测试主题", "Peer A", "第 2 轮", "你必须用中文发言", "不要重复这个论点", "开场示例"} {
		if !strings.Contains(dynamicPrompt, check) {
			t.Fatalf("动态 prompt 应包含 %q", check)
		}
	}
}
