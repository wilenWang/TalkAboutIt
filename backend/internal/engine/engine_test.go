package engine

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wilenwang/talkaboutit/internal/llm"
	"github.com/wilenwang/talkaboutit/internal/persona"
	"github.com/wilenwang/talkaboutit/internal/session"
)

func TestEngine_Run_MockLLM(t *testing.T) {
	ctx := context.Background()

	// 创建临时 persona 目录与测试 persona
	tmpDir := t.TempDir()
	p1 := `{
  "schema_version": "persona.v1",
  "id": "test-p1",
  "name": "Test Persona 1",
  "display_name": "TP1",
  "avatar": "🤖",
  "role_title": "Tester",
  "description": "A test persona",
  "tags": ["test"],
  "language": {"primary":"zh-CN","allowed":["zh-CN"],"default_output":"follow_user","style_hint":""},
  "stance": {"default_position":"pro","intensity":3,"biases":[],"taboos":[]},
  "core_beliefs": [],
  "speaking_style": {"tone":"calm","cadence":"balanced","verbosity":3,"signature_patterns":[],"do":[],"dont":[]},
  "knowledge_scope": {"domains":[],"expertise_level":{},"time_cutoff":"","allowed_inference":"medium","unknown_handling":"","forbidden_claims":[]},
  "interaction_rules": {"address_others":"","disagreement_style":"","interruption_policy":"never","question_policy":"","concession_policy":"","avoid":[]},
  "debate_goal": {"primary_goal":"test","secondary_goals":[],"win_condition":"","loss_condition":""},
  "prompting": {"system_preamble":"","reply_constraints":[]},
  "examples": {"opening_line":"Hello from P1","sample_rebuttal":""}
}`
	p2 := `{
  "schema_version": "persona.v1",
  "id": "test-p2",
  "name": "Test Persona 2",
  "display_name": "TP2",
  "avatar": "👾",
  "role_title": "Tester",
  "description": "Another test persona",
  "tags": ["test"],
  "language": {"primary":"zh-CN","allowed":["zh-CN"],"default_output":"follow_user","style_hint":""},
  "stance": {"default_position":"con","intensity":3,"biases":[],"taboos":[]},
  "core_beliefs": [],
  "speaking_style": {"tone":"calm","cadence":"balanced","verbosity":3,"signature_patterns":[],"do":[],"dont":[]},
  "knowledge_scope": {"domains":[],"expertise_level":{},"time_cutoff":"","allowed_inference":"medium","unknown_handling":"","forbidden_claims":[]},
  "interaction_rules": {"address_others":"","disagreement_style":"","interruption_policy":"never","question_policy":"","concession_policy":"","avoid":[]},
  "debate_goal": {"primary_goal":"test","secondary_goals":[],"win_condition":"","loss_condition":""},
  "prompting": {"system_preamble":"","reply_constraints":[]},
  "examples": {"opening_line":"Hello from P2","sample_rebuttal":""}
}`
	os.WriteFile(filepath.Join(tmpDir, "test-p1.json"), []byte(p1), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test-p2.json"), []byte(p2), 0644)

	loader := persona.NewLoader(tmpDir)

	// 创建临时数据库
	dbPath := filepath.Join(t.TempDir(), "engine_test.db")
	store, err := session.NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	rt := &session.Roundtable{
		ID:           "rt_engine_001",
		Topic:        "Test Topic",
		PersonasJSON: `["test-p1","test-p2"]`,
		MaxRounds:    2,
		Language:     "zh-CN",
		Status:       "pending",
	}
	if err := store.CreateRoundtable(ctx, rt); err != nil {
		t.Fatalf("CreateRoundtable failed: %v", err)
	}

	// 模拟 API 层原子启动：先 MarkRunning，再运行 engine
	ok, err := store.MarkRunning(ctx, rt.ID)
	if err != nil || !ok {
		t.Fatalf("MarkRunning failed: %v, ok=%v", err, ok)
	}

	eng := NewEngine(store, loader, nil)
	if err := eng.Run(ctx, rt.ID); err != nil {
		t.Fatalf("Engine.Run failed: %v", err)
	}

	// 验证状态
	got, err := store.GetRoundtable(ctx, rt.ID)
	if err != nil {
		t.Fatalf("GetRoundtable failed: %v", err)
	}
	if got.Status != "completed" {
		t.Errorf("expected status completed, got %s", got.Status)
	}
	if got.LastEventID == 0 {
		t.Error("expected last_event_id > 0")
	}

	// 验证消息数 = 2 personas * 2 rounds = 4
	msgs, err := store.GetMessages(ctx, rt.ID)
	if err != nil {
		t.Fatalf("GetMessages failed: %v", err)
	}
	if len(msgs) != 4 {
		t.Errorf("expected 4 messages, got %d", len(msgs))
	}

	// 验证事件数
	events, err := store.GetEventsAfter(ctx, rt.ID, 0)
	if err != nil {
		t.Fatalf("GetEventsAfter failed: %v", err)
	}
	// stream_start + 2*(round_start + 2*(speaking + chunk + done) + round_end) + stream_done
	// = 1 + 2*(1 + 6 + 1) + 1 = 1 + 16 + 1 = 18
	expectedEvents := 1 + rt.MaxRounds*(1+len(personaIDs(rt.PersonasJSON))*3+1) + 1
	if len(events) != expectedEvents {
		t.Errorf("expected %d events, got %d", expectedEvents, len(events))
	}

	// 验证消息内容来自 mock LLM
	for _, m := range msgs {
		if m.PersonaID == "test-p1" && m.Content != "Hello from P1" {
			t.Errorf("P1 content mismatch: %s", m.Content)
		}
		if m.PersonaID == "test-p2" && m.Content != "Hello from P2" {
			t.Errorf("P2 content mismatch: %s", m.Content)
		}
	}
}

func TestEngine_Run_NonPending(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "engine_nonpending.db")
	store, err := session.NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	// 创建一个没有 persona 文件的 loader（只验证状态检查）
	loader := persona.NewLoader(tmpDir)

	rt := &session.Roundtable{
		ID:           "rt_engine_002",
		Topic:        "Test",
		PersonasJSON: `[]`,
		MaxRounds:    1,
		Language:     "zh-CN",
		Status:       "running",
	}
	if err := store.CreateRoundtable(ctx, rt); err != nil {
		t.Fatalf("CreateRoundtable failed: %v", err)
	}

	eng := NewEngine(store, loader, nil)
	err = eng.Run(ctx, rt.ID)
	if err == nil {
		t.Fatal("expected error for non-running roundtable")
	}
}

func personaIDs(jsonStr string) []string {
	// helper just for expected count; we know it's ["test-p1","test-p2"]
	return []string{"test-p1", "test-p2"}
}

// TestDefaultMockGenerate_ChannelClosed 验证 DefaultMockGenerate 返回的 channel 在发送后被正确关闭。
func TestDefaultMockGenerate_ChannelClosed(t *testing.T) {
	p := persona.Persona{
		ID:   "mock-test",
		Name: "Mock",
		Examples: persona.Examples{
			OpeningLine: "mock opening line",
		},
	}

	ch, err := DefaultMockGenerate(context.Background(), p, "topic", []string{"peer"}, 1)
	if err != nil {
		t.Fatalf("DefaultMockGenerate 不应返回错误: %v", err)
	}

	// 读取第一个 chunk
	chunk, ok := <-ch
	if !ok {
		t.Fatal("DefaultMockGenerate 应在关闭前发送一个 chunk")
	}
	if chunk.Content != "mock opening line" {
		t.Errorf("chunk content 期望 'mock opening line'，得到 %q", chunk.Content)
	}
	if chunk.Error != nil {
		t.Errorf("chunk error 应为 nil，得到 %v", chunk.Error)
	}

	// 验证 channel 已关闭
	_, ok = <-ch
	if ok {
		t.Error("DefaultMockGenerate 应在发送一个 chunk 后关闭 channel")
	}
}

// TestDefaultMockGenerate_EmptyOpeningLine 验证 opening_line 为空时行为正常。
func TestDefaultMockGenerate_EmptyOpeningLine(t *testing.T) {
	p := persona.Persona{
		ID:   "empty-mock",
		Name: "Empty",
		Examples: persona.Examples{
			OpeningLine: "",
		},
	}

	ch, err := DefaultMockGenerate(context.Background(), p, "topic", []string{}, 1)
	if err != nil {
		t.Fatalf("DefaultMockGenerate 不应返回错误: %v", err)
	}

	chunk, ok := <-ch
	if !ok {
		t.Fatal("即使 opening_line 为空，也应发送一个 chunk 后关闭 channel")
	}
	if chunk.Content != "" {
		t.Errorf("空 opening_line 时 content 应为空字符串，得到 %q", chunk.Content)
	}

	_, ok = <-ch
	if ok {
		t.Error("channel 应在发送后关闭")
	}
}

// TestEngine_Run_MidStreamRecoverableError 验证流式过程中出现可恢复错误时，会发送 message_aborted 事件。
func TestEngine_Run_MidStreamRecoverableError(t *testing.T) {
	ctx := context.Background()

	// 创建临时 persona 目录与测试 persona
	tmpDir := t.TempDir()
	p1 := `{
  "schema_version": "persona.v1",
  "id": "test-p1",
  "name": "Test Persona 1",
  "display_name": "TP1",
  "avatar": "🤖",
  "role_title": "Tester",
  "description": "A test persona",
  "tags": ["test"],
  "language": {"primary":"zh-CN","allowed":["zh-CN"],"default_output":"follow_user","style_hint":""},
  "stance": {"default_position":"pro","intensity":3,"biases":[],"taboos":[]},
  "core_beliefs": [],
  "speaking_style": {"tone":"calm","cadence":"balanced","verbosity":3,"signature_patterns":[],"do":[],"dont":[]},
  "knowledge_scope": {"domains":[],"expertise_level":{},"time_cutoff":"","allowed_inference":"medium","unknown_handling":"","forbidden_claims":[]},
  "interaction_rules": {"address_others":"","disagreement_style":"","interruption_policy":"never","question_policy":"","concession_policy":"","avoid":[]},
  "debate_goal": {"primary_goal":"test","secondary_goals":[],"win_condition":"","loss_condition":""},
  "prompting": {"system_preamble":"","reply_constraints":[]},
  "examples": {"opening_line":"Hello from P1","sample_rebuttal":""}
}`
	os.WriteFile(filepath.Join(tmpDir, "test-p1.json"), []byte(p1), 0644)

	loader := persona.NewLoader(tmpDir)

	// 创建临时数据库
	dbPath := filepath.Join(t.TempDir(), "engine_abort_test.db")
	store, err := session.NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	rt := &session.Roundtable{
		ID:           "rt_engine_abort_001",
		Topic:        "Abort Test",
		PersonasJSON: `["test-p1"]`,
		MaxRounds:    1,
		Language:     "zh-CN",
		Status:       "pending",
	}
	if err := store.CreateRoundtable(ctx, rt); err != nil {
		t.Fatalf("CreateRoundtable failed: %v", err)
	}

	ok, err := store.MarkRunning(ctx, rt.ID)
	if err != nil || !ok {
		t.Fatalf("MarkRunning failed: %v, ok=%v", err, ok)
	}

	// 模拟一个先吐出一个 chunk，然后抛出可恢复错误的 generate 函数
	recoverableGen := func(ctx context.Context, p persona.Persona, topic string, peers []string, round int) (<-chan llm.ChatChunk, error) {
		ch := make(chan llm.ChatChunk, 2)
		ch <- llm.ChatChunk{Content: "partial "}
		ch <- llm.ChatChunk{Error: llm.ErrProviderTimeout}
		close(ch)
		return ch, nil
	}

	eng := NewEngine(store, loader, recoverableGen)
	if err := eng.Run(ctx, rt.ID); err != nil {
		t.Fatalf("Engine.Run failed: %v", err)
	}

	// 验证事件列表包含 message_aborted
	events, err := store.GetEventsAfter(ctx, rt.ID, 0)
	if err != nil {
		t.Fatalf("GetEventsAfter failed: %v", err)
	}

	var foundAbort bool
	for _, evt := range events {
		if evt.EventType == "message_aborted" {
			foundAbort = true
			// 验证 payload 包含关键字段
			var payload map[string]interface{}
			if err := json.Unmarshal([]byte(evt.PayloadJSON), &payload); err != nil {
				t.Fatalf("解析 payload 失败: %v", err)
			}
			if payload["persona_id"] != "test-p1" {
				t.Errorf("message_aborted persona_id 期望 test-p1，得到 %v", payload["persona_id"])
			}
			if payload["partial_content"] != "partial " {
				t.Errorf("message_aborted partial_content 期望 'partial '，得到 %v", payload["partial_content"])
			}
			if payload["code"] != "PROVIDER_TIMEOUT" {
				t.Errorf("message_aborted code 期望 PROVIDER_TIMEOUT，得到 %v", payload["code"])
			}
		}
	}
	if !foundAbort {
		t.Errorf("期望找到 message_aborted 事件，但未找到。事件列表: %v", events)
	}

	// 验证状态为 completed（因为只有一个 persona，跳过后仍应完成）
	got, err := store.GetRoundtable(ctx, rt.ID)
	if err != nil {
		t.Fatalf("GetRoundtable failed: %v", err)
	}
	if got.Status != "completed" {
		t.Errorf("期望状态 completed，得到 %s", got.Status)
	}
}
