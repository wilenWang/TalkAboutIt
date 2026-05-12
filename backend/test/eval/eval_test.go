//go:build eval

// Package eval 提供 TalkAboutIt 的评测入口。
// 使用 build tag "eval" 隔离，避免在常规测试中运行。
//
// 运行方式：
//
//	go test -tags=eval ./test/eval/ -v -count=1
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
	"github.com/wilenwang/talkaboutit/internal/persona"
	"github.com/wilenwang/talkaboutit/internal/session"
)

// question 代表评测题目。
type question struct {
	ID       string `json:"id"`
	Topic    string `json:"topic"`
	Category string `json:"category"`
}

// baselineRecord 代表单条评测记录。
type baselineRecord struct {
	QuestionID string           `json:"question_id"`
	Topic      string           `json:"topic"`
	Category   string           `json:"category"`
	Result     EvaluationResult `json:"result"`
}

// baselineOutput 是评测结果的整体输出结构。
type baselineOutput struct {
	GeneratedAt string           `json:"generated_at"`
	Personas    []string         `json:"personas"`
	Rounds      int              `json:"rounds"`
	Records     []baselineRecord `json:"records"`
	Average     float64          `json:"average"`
}

// TestEval_Baseline 是评测主入口：加载题目、运行 roundtable、评分、输出 baseline.json。
func TestEval_Baseline(t *testing.T) {
	// 1. 加载 questions.json
	questions, err := loadQuestions()
	if err != nil {
		t.Fatalf("加载评测题目失败: %v", err)
	}

	// 2. 准备环境（使用真实 persona：Steve Jobs + Elon Musk）
	loader := persona.NewLoader("../../personas")
	personas := []string{"steve-jobs", "elon-musk"}
	rounds := 2

	// 验证 persona 存在
	for _, pid := range personas {
		if _, err := loader.LoadOne(pid); err != nil {
			t.Fatalf("加载 persona %s 失败: %v", pid, err)
		}
	}

	var records []baselineRecord
	var totalScore float64

	for _, q := range questions {
		t.Run(q.ID, func(t *testing.T) {
			result := runEvaluation(t, q, loader, personas, rounds)
			records = append(records, baselineRecord{
				QuestionID: q.ID,
				Topic:      q.Topic,
				Category:   q.Category,
				Result:     result,
			})
			totalScore += result.Total
		})
	}

	avg := totalScore / float64(len(records))

	output := baselineOutput{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Personas:    personas,
		Rounds:      rounds,
		Records:     records,
		Average:     avg,
	}

	// 3. 输出 baseline.json
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatalf("序列化 baseline 失败: %v", err)
	}

	outPath := "baseline.json"
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		t.Fatalf("写入 baseline.json 失败: %v", err)
	}

	t.Logf("评测完成，平均分: %.2f/100，结果已写入 %s", avg, outPath)
	fmt.Printf("评测完成，平均分: %.2f/100，结果已写入 %s\n", avg, outPath)
}

// loadQuestions 从 questions.json 加载评测题目。
func loadQuestions() ([]question, error) {
	data, err := os.ReadFile("questions.json")
	if err != nil {
		return nil, err
	}
	var qs []question
	if err := json.Unmarshal(data, &qs); err != nil {
		return nil, err
	}
	return qs, nil
}

// runEvaluation 为单个题目创建 roundtable、运行 engine、收集消息并评分。
func runEvaluation(t *testing.T, q question, loader *persona.Loader, personas []string, rounds int) EvaluationResult {
	t.Helper()

	// 创建临时数据库
	dbPath := filepath.Join(t.TempDir(), fmt.Sprintf("eval_%s.db", q.ID))
	store, err := session.NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	// 创建 roundtable
	ctx := context.Background()
	personasJSON, _ := json.Marshal(personas)
	rt := &session.Roundtable{
		ID:           fmt.Sprintf("eval_%s_%d", q.ID, time.Now().UnixNano()),
		Topic:        q.Topic,
		PersonasJSON: string(personasJSON),
		MaxRounds:    rounds,
		Language:     "zh-CN",
		Status:       "pending",
	}
	if err := store.CreateRoundtable(ctx, rt); err != nil {
		t.Fatalf("CreateRoundtable failed: %v", err)
	}

	// 启动（模拟 API 层的 MarkRunning + Run）
	ok, err := store.MarkRunning(ctx, rt.ID)
	if err != nil || !ok {
		t.Fatalf("MarkRunning failed: %v, ok=%v", err, ok)
	}

	// 使用 mock LLM（nil → DefaultMockGenerate）
	eng := engine.NewEngine(store, loader, nil)
	if err := eng.Run(ctx, rt.ID); err != nil {
		t.Fatalf("Engine.Run failed: %v", err)
	}

	// 收集消息
	msgs, err := store.GetMessages(ctx, rt.ID)
	if err != nil {
		t.Fatalf("GetMessages failed: %v", err)
	}

	// 转换为评分器需要的格式
	evalMsgs := make([]Message, 0, len(msgs))
	for _, m := range msgs {
		evalMsgs = append(evalMsgs, Message{
			PersonaID: m.PersonaID,
			Content:   m.Content,
			Round:     m.Round,
		})
	}

	return ScoreRoundtable(q.Topic, q.Category, evalMsgs)
}
