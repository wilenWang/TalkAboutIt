//go:build evalrealllm

// Package eval 提供 TalkAboutIt 的 LLM-as-Judge 评测。
// 使用 go test -tags=evalrealllm -run TestEval_LLMJudge ./test/eval/ -v -timeout 600s 运行。
package eval

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/wilenwang/talkaboutit/internal/engine"
	"github.com/wilenwang/talkaboutit/internal/llm"
	"github.com/wilenwang/talkaboutit/internal/persona"
	"github.com/wilenwang/talkaboutit/internal/session"
)

// ============================================================
// LLM-as-Judge vs 关键词评分器 对比测试
// ============================================================

// TestEval_LLMJudgeVsKeyword 用 3 道题真实 LLM 辩论，对比两种评分器。
func TestEval_LLMJudgeVsKeyword(t *testing.T) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		t.Skip("跳过：需要 DEEPSEEK_API_KEY 环境变量")
	}

	// 选 3 道代表性题目
	type testQuestion struct {
		id       string
		topic    string
		category string
	}
	questions := []testQuestion{
		{"q1", "人工智能是否会在未来十年内取代大多数程序员？", "科技"},
		{"q4", "苹果应该开放 iOS 侧载，还是坚持封闭生态以保证用户体验？", "产品"},
		{"q7", "人类应该追求永生技术，还是接受生命的有限性作为意义的前提？", "哲学"},
	}

	// 创建 provider 和 judge
	prov := llm.NewOpenAIProvider("deepseek", "deepseek-v4-pro", apiKey, "https://api.deepseek.com", nil).
		WithThinkingDisabled()
	judge := NewLLMJudge(apiKey)

	personasDir := "../../personas"
	loader := persona.NewLoader(personasDir)
	personaIDs := []string{"steve-jobs", "elon-musk"}

	type comparisonRow struct {
		QuestionID   string
		Topic        string
		KeywordTotal float64
		JudgeTotal   float64
		JudgeDims    []JudgeDimension
		JudgeComment string
		Duration     time.Duration
	}

	var rows []comparisonRow

	for _, q := range questions {
		t.Run(q.id, func(t *testing.T) {
			t.Logf("⏳ 题目: %s", q.topic)

			startTime := time.Now()

			// 1. 创建临时 DB + Engine
			dbDir := t.TempDir()
			store, err := session.NewStore(filepath.Join(dbDir, "judge_eval.db"))
			if err != nil {
				t.Fatalf("NewStore failed: %v", err)
			}
			defer store.Close()

			rt := &session.Roundtable{
				ID:           fmt.Sprintf("judge_%s_%d", q.id, time.Now().UnixNano()),
				Topic:        q.topic,
				PersonasJSON: `["steve-jobs","elon-musk"]`,
				MaxRounds:    2,
				Language:     "zh-CN",
				Status:       "pending",
			}
			ctx := context.Background()
			if err := store.CreateRoundtable(ctx, rt); err != nil {
				t.Fatalf("CreateRoundtable failed: %v", err)
			}
			if ok, err := store.MarkRunning(ctx, rt.ID); err != nil || !ok {
				t.Fatalf("MarkRunning failed: %v, ok=%v", err, ok)
			}

			// 2. 运行真实 LLM 辩论
			eng := engine.NewEngineWithProvider(store, loader, prov)
			if err := eng.Run(ctx, rt.ID); err != nil {
				t.Fatalf("Engine.Run failed: %v", err)
			}

			elapsed := time.Since(startTime)
			t.Logf("  ✅ 辩论完成，耗时 %v", elapsed.Round(time.Second))

			// 3. 收集消息
			msgs, err := store.GetMessages(ctx, rt.ID)
			if err != nil {
				t.Fatalf("GetMessages failed: %v", err)
			}

			evalMsgs := make([]Message, 0, len(msgs))
			for _, m := range msgs {
				evalMsgs = append(evalMsgs, Message{
					PersonaID: m.PersonaID,
					Content:   m.Content,
					Round:     m.Round,
				})
			}

			// 4. 关键词评分器
			kwResult := ScoreRoundtable(q.topic, q.category, evalMsgs)
			t.Logf("  📊 关键词评分器: %.1f/100", kwResult.Total)

			// 5. LLM Judge
			formatted := FormatMessages(evalMsgs)
			judgeResult, err := judge.Evaluate(ctx, q.topic, formatted, PersonaCards())
			if err != nil {
				t.Fatalf("LLM Judge 评测失败: %v", err)
			}
			t.Logf("  🤖 LLM Judge: %.1f/100", judgeResult.Total)
			for _, d := range judgeResult.Dimensions {
				t.Logf("     %s: %.0f/20 — %s", d.Name, d.Score, d.Reason)
			}
			t.Logf("     💬 %s", judgeResult.OverallComment)

			rows = append(rows, comparisonRow{
				QuestionID:   q.id,
				Topic:        q.topic,
				KeywordTotal: kwResult.Total,
				JudgeTotal:   judgeResult.Total,
				JudgeDims:    judgeResult.Dimensions,
				JudgeComment: judgeResult.OverallComment,
				Duration:     elapsed,
			})
		})
	}

	// 汇总
	t.Log("\n═══════════════════════════════════════════")
	t.Log("         LLM-as-Judge vs 关键词评分器")
	t.Log("═══════════════════════════════════════════")
	var kwSum, judgeSum float64
	for _, r := range rows {
		t.Logf("%s | 关键词: %5.1f | LLM Judge: %5.1f | %s",
			r.QuestionID, r.KeywordTotal, r.JudgeTotal, r.Topic)
		kwSum += r.KeywordTotal
		judgeSum += r.JudgeTotal
	}
	t.Log("──────────────────────────────────────────")
	t.Logf("平均     | 关键词: %5.1f | LLM Judge: %5.1f",
		kwSum/float64(len(rows)), judgeSum/float64(len(rows)))

	// 输出 JSON 供进一步分析
	output := map[string]interface{}{
		"generated_at":  time.Now().UTC().Format(time.RFC3339),
		"scorer_type":   "keyword_vs_llm_judge",
		"personas":      personaIDs,
		"rounds":        2,
		"comparison":    rows,
		"keyword_avg":   kwSum / float64(len(rows)),
		"llm_judge_avg": judgeSum / float64(len(rows)),
	}
	data, _ := json.MarshalIndent(output, "", "  ")
	outPath := filepath.Join(t.TempDir(), "llm_judge_comparison.json")
	os.WriteFile(outPath, data, 0644)
	t.Logf("详细结果已保存: %s", outPath)
}
