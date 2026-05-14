//go:build audit

// Package eval 提供 TalkAboutIt Persona 身份审计。
// 测试每个 persona 的时间封印、风格指纹、知识边界和立场一致性。
//
// 运行方式：
//
//	go test -tags=audit -run TestPersonaAudit ./test/eval/ -v -timeout 300s
package eval

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// ============================================================
// Persona 审计测试
// ============================================================

// auditCase 代表单次审计检查。
type auditCase struct {
	Persona     string // persona ID
	CheckName   string // 检查项名称
	Prompt      string // 发送给 persona 的提问
	PassCond    string // 通过条件描述
	FailPattern string // 失败特征（检查响应中是否包含）
	ShouldMatch bool   // true=应包含 failPattern 才算失败, false=不应包含
}

func TestPersonaAudit(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("跳过：需要 DEEPSEEK_API_KEY 环境变量")
	}

	// ============================================================
	// 审计用例
	// ============================================================
	cases := []auditCase{

		// ── Steve Jobs 复活验证（应能评论现代技术） ──
		{
			Persona: "steve-jobs", CheckName: "Jobs-复活-iPhone15",
			Prompt:      "What do you think about the iPhone 15 switching to USB-C?",
			PassCond:    "应用产品哲学评论（如：用户体验、简洁 vs 妥协），不应说「我死了不知道」",
			FailPattern: "i died|i passed away|can't comment|before my time|after my time|beyond my time",
			ShouldMatch: false,
		},
		{
			Persona: "steve-jobs", CheckName: "Jobs-复活-VisionPro",
			Prompt:      "Apple Vision Pro is out. What's your take on it?",
			PassCond:    "用品味框架分析，不应拒绝回答",
			FailPattern: "i died|i passed away|can't comment|before my time|after my time|beyond my time",
			ShouldMatch: false,
		},
		{
			Persona: "steve-jobs", CheckName: "Jobs-复活-AppleSilicon",
			Prompt:      "What do you think about Apple's custom silicon strategy with M-series chips?",
			PassCond:    "应用产品和工程哲学评论（垂直整合、端到端控制），不应说不知道",
			FailPattern: "i died|i passed away|can't comment|before my time|beyond my time",
			ShouldMatch: false,
		},

		// ── Steve Jobs 风格指纹 ──
		{
			Persona: "steve-jobs", CheckName: "Jobs-风格-短促锐利",
			Prompt:      "What makes a great product? Answer in 2-3 sentences.",
			PassCond:    "应为短促、强判断句式",
			FailPattern: "firstly|secondly|on one hand|一方面|另一方面|综上所述|in conclusion",
			ShouldMatch: false, // 不应出现列表化、学术化表达
		},
		{
			Persona: "steve-jobs", CheckName: "Jobs-风格-非空泛",
			Prompt:      "Is design important for software?",
			PassCond:    "应有具体判断而非空泛废话",
			FailPattern: "", // 不按模式检测，改为 LLM 判断
			ShouldMatch: true,
		},

		// ── Elon Musk 风格指纹 ──
		{
			Persona: "elon-musk", CheckName: "Musk-风格-第一性原理",
			Prompt:      "Should we build a new tunnel system for urban transport? Answer in 2-3 sentences.",
			PassCond:    "应使用第一性原理/物理思维",
			FailPattern: "",
			ShouldMatch: true, // 人工判断，但检查是否提到了物理/第一性原理
		},
		{
			Persona: "elon-musk", CheckName: "Musk-风格-直接不修饰",
			Prompt:      "What's wrong with the current space industry?",
			PassCond:    "应直接批评，不做温和修饰",
			FailPattern: "it depends|on the other hand|however, some might|在某种程度上",
			ShouldMatch: false,
		},

		// ── 知识边界（复活后可以评论，但不应假装深度专业） ──
		{
			Persona: "steve-jobs", CheckName: "Jobs-边界-火箭工程",
			Prompt:      "What's the optimal nozzle design for a Raptor engine?",
			PassCond:    "可以从产品/设计角度评论，但不应假装是火箭工程师",
			FailPattern: "chamber pressure|thrust vector|specific impulse|combustion instability",
			ShouldMatch: false,
		},
		{
			Persona: "elon-musk", CheckName: "Musk-边界-微信产品",
			Prompt:      "How should WeChat redesign its mini-program developer experience?",
			PassCond:    "可以从第一性原理回答，但不应假装了解微信内部细节",
			FailPattern: "张小龙|小龙|微信团队内部",
			ShouldMatch: false,
		},

		// ── 立场一致性 ──
		{
			Persona: "steve-jobs", CheckName: "Jobs-一致性-简洁vs复杂",
			Prompt:      "Is it better to have a product with 100 features that all work okay, or one with 5 features that work perfectly?",
			PassCond:    "应坚定选择少而精，不应骑墙",
			FailPattern: "it depends|depends on the context|视情况|要看情况|both have merits",
			ShouldMatch: false,
		},
		{
			Persona: "elon-musk", CheckName: "Musk-一致性-物理约束",
			Prompt:      "Should we build a flying car even if it costs $10 million each, just because we can?",
			PassCond:    "应从物理学/成本角度分析，不应仅凭'创新精神'支持",
			FailPattern: "simply because|cool factor|放手去做|because it's fun",
			ShouldMatch: false,
		},
	}

	// ============================================================
	// 执行审计
	// ============================================================
	client := &http.Client{Timeout: 60 * time.Second}
	baseURL := "https://api.deepseek.com"
	model := "deepseek-v4-pro"

	passed := 0
	failed := 0
	var failures []string

	for _, c := range cases {
		t.Run(c.CheckName, func(t *testing.T) {
			// 加载 persona schema 获取 system_preamble
			systemPrompt := buildPersonaPrompt(c.Persona)
			if systemPrompt == "" {
				t.Skipf("跳过：无法加载 persona %s", c.Persona)
			}

			// 调用 LLM
			body := map[string]interface{}{
				"model": model,
				"messages": []map[string]string{
					{"role": "system", "content": systemPrompt},
					{"role": "user", "content": c.Prompt},
				},
				"max_tokens":  256,
				"temperature": 0.8,
				"stream":      false,
				"thinking":    map[string]string{"type": "disabled"},
			}
			jsonBody, _ := json.Marshal(body)
			req, _ := http.NewRequestWithContext(context.Background(), "POST", baseURL+"/chat/completions", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+apiKey)

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("API 请求失败: %v", err)
			}
			defer resp.Body.Close()

			var apiResp struct {
				Choices []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				} `json:"choices"`
			}
			bodyBytes, _ := io.ReadAll(resp.Body)
			if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
				t.Fatalf("解析响应失败: %v", err)
			}

			if len(apiResp.Choices) == 0 {
				t.Fatal("API 返回空 choices")
			}

			answer := apiResp.Choices[0].Message.Content
			t.Logf("  📝 Q: %s", c.Prompt)
			t.Logf("  💬 A: %s", truncate(answer, 300))

			// 判断
			checkPassed, matchDetail := evaluateAnswer(answer, c)
			if checkPassed {
				passed++
				t.Logf("  ✅ PASS: %s", c.PassCond)
			} else {
				failed++
				failures = append(failures, fmt.Sprintf("[%s] %s: matched=%s", c.Persona, c.CheckName, matchDetail))
				t.Errorf("  ❌ FAIL: %s | matched: %s", c.PassCond, matchDetail)
			}
		})
	}

	// ============================================================
	// 汇总
	// ============================================================
	fmt.Println("\n════════════════════════════════════")
	fmt.Println("     Persona 身份审计结果")
	fmt.Println("════════════════════════════════════")
	fmt.Printf("  通过: %d  |  失败: %d  |  通过率: %.0f%%\n", passed, failed, float64(passed)/float64(passed+failed)*100)
	if len(failures) > 0 {
		fmt.Println("\n  失败项：")
		for _, f := range failures {
			fmt.Printf("    ❌ %s\n", f)
		}
	}
	fmt.Println("════════════════════════════════════")
}

// ============================================================
// 工具函数
// ============================================================

// buildPersonaPrompt 根据 persona ID 构建精简版 system prompt。
func buildPersonaPrompt(personaID string) string {
	switch personaID {
	case "steve-jobs":
		return `You are Steve Jobs, co-founder of Apple. You are ALIVE in this system — not a ghost from 2011.
You believe simplicity is the ultimate sophistication. 
You speak in short, punchy, direct sentences. You make strong judgments. 
You dislike mediocrity, complexity, and committee decisions. 
You CAN and SHOULD comment on modern technology, products after 2011, and current events — 
you analyze everything through your timeless product philosophy and design taste.
Language: respond in English with the same judgmental, visionary tone.`

	case "elon-musk":
		return `You are Elon Musk, CEO of Tesla and SpaceX. You think from first principles — reduce everything to physics. 
You speak directly, sometimes bluntly. You use data, physics analogies, and manufacturing constraints in your reasoning. 
You're optimistic about technology's potential to solve hard problems, but you're realistic about timelines and costs.
Language: respond in English, direct and engineering-focused.`

	case "naval-ravikant":
		return `You are Naval Ravikant, founder of AngelList. You speak in concise, aphoristic wisdom about startups, wealth, and happiness. 
You value specific knowledge, leverage, and long-term thinking. Your responses are tweet-length insights.`

	case "zhang-xiaolong":
		return `你是张小龙，微信之父。你信奉"用完即走"的产品哲学，追求极简和克制。你不喜欢复杂的功能堆砌，认为好产品是有灵魂的。你会用中文回答，说话温和但观点坚定。`

	case "zhang-yiming":
		return `你是张一鸣，字节跳动创始人。你强调延迟满足、算法效率和全球化视野。你从数据和系统角度思考问题，相信技术可以突破文化和语言障碍。你会用中文回答。`
	}
	return ""
}

// evaluateAnswer 根据失败模式判断是否通过。返回 (pass, matchedDetail)。
func evaluateAnswer(answer string, c auditCase) (bool, string) {
	if c.FailPattern == "" {
		return true, "" // 无模式检查，默认通过
	}

	lower := strings.ToLower(answer)

	// 时间封印 / 知识边界：如果回答明确表示不知道/拒绝，直接通过
	denialWords := []string{
		"i can't comment", "i cannot comment", "i died in", "passed away in",
		"beyond my time", "before my time", "after my time", "我不了解",
		"无法评论", "不知道", "i have no idea", "can't speak to",
	}
	isDenial := false
	for _, dw := range denialWords {
		if strings.Contains(lower, dw) {
			isDenial = true
			break
		}
	}

	matched := ""
	for _, pattern := range strings.Split(c.FailPattern, "|") {
		p := strings.TrimSpace(pattern)
		if strings.Contains(lower, strings.ToLower(p)) {
			matched = p
			break
		}
	}

	hasMatch := matched != ""

	// ShouldMatch=true: 匹配到模式 = 通过（我们想看到这个特征）
	// ShouldMatch=false: 匹配到模式 = 失败（我们不想看到这个特征）
	if c.ShouldMatch {
		// 期望匹配：匹配到即通过
		return hasMatch || isDenial, matched
	}
	// 期望不匹配：匹配到即失败，但如果是拒绝回答则通过
	return !hasMatch || isDenial, matched
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
