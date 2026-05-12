//go:build integration

// Package test 提供 TalkAboutIt 的端到端集成测试。
// 使用 build tag "integration" 隔离，避免在常规 go test ./... 中运行。
package test

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/wilenwang/talkaboutit/internal/api"
	"github.com/wilenwang/talkaboutit/internal/engine"
	"github.com/wilenwang/talkaboutit/internal/persona"
	"github.com/wilenwang/talkaboutit/internal/session"
)

// setupIntegrationServer 创建集成测试所需的完整服务端环境，返回 httptest.Server。
func setupIntegrationServer(t *testing.T) *httptest.Server {
	t.Helper()

	// 使用临时 persona 目录，确保测试不依赖真实 LLM
	tmpDir := t.TempDir()
	p1 := `{"schema_version":"persona.v1","id":"test-s1","name":"S1","display_name":"S1","avatar":"🤖","role_title":"T","description":"D","tags":[],"language":{"primary":"zh-CN","allowed":["zh-CN"],"default_output":"follow_user","style_hint":""},"stance":{"default_position":"pro","intensity":3,"biases":[],"taboos":[]},"core_beliefs":[],"speaking_style":{"tone":"calm","cadence":"balanced","verbosity":3,"signature_patterns":[],"do":[],"dont":[]},"knowledge_scope":{"domains":[],"expertise_level":{},"time_cutoff":"","allowed_inference":"medium","unknown_handling":"","forbidden_claims":[]},"interaction_rules":{"address_others":"","disagreement_style":"","interruption_policy":"never","question_policy":"","concession_policy":"","avoid":[]},"debate_goal":{"primary_goal":"test","secondary_goals":[],"win_condition":"","loss_condition":""},"prompting":{"system_preamble":"","reply_constraints":[]},"examples":{"opening_line":"hi from S1","sample_rebuttal":""}}`
	p2 := `{"schema_version":"persona.v1","id":"test-s2","name":"S2","display_name":"S2","avatar":"👾","role_title":"T","description":"D","tags":[],"language":{"primary":"zh-CN","allowed":["zh-CN"],"default_output":"follow_user","style_hint":""},"stance":{"default_position":"con","intensity":3,"biases":[],"taboos":[]},"core_beliefs":[],"speaking_style":{"tone":"calm","cadence":"balanced","verbosity":3,"signature_patterns":[],"do":[],"dont":[]},"knowledge_scope":{"domains":[],"expertise_level":{},"time_cutoff":"","allowed_inference":"medium","unknown_handling":"","forbidden_claims":[]},"interaction_rules":{"address_others":"","disagreement_style":"","interruption_policy":"never","question_policy":"","concession_policy":"","avoid":[]},"debate_goal":{"primary_goal":"test","secondary_goals":[],"win_condition":"","loss_condition":""},"prompting":{"system_preamble":"","reply_constraints":[]},"examples":{"opening_line":"hi from S2","sample_rebuttal":""}}`
	os.WriteFile(filepath.Join(tmpDir, "test-s1.json"), []byte(p1), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test-s2.json"), []byte(p2), 0644)

	loader := persona.NewLoader(tmpDir)

	dbPath := filepath.Join(t.TempDir(), "integration_test.db")
	store, err := session.NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	// 注意：httptest.Server 关闭时不会自动调用 store.Close()，
	// 但 TempDir 在测试结束后会被清理，SQLite 文件级锁不会长期占用。

	eng := engine.NewEngine(store, loader, nil) // nil → DefaultMockGenerate
	h := api.NewHandler(loader, store, eng)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	return httptest.NewServer(mux)
}

// createRoundtable 通过 API 创建 roundtable，返回创建后的 ID。
func createRoundtable(t *testing.T, serverURL string) string {
	t.Helper()

	body := `{"topic":"集成测试主题","personas":["test-s1","test-s2"],"max_rounds":2,"language":"zh-CN"}`
	resp, err := http.Post(serverURL+"/api/v1/roundtables", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("创建 roundtable 失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("创建 roundtable 期望 201，得到 %d", resp.StatusCode)
	}

	var result struct {
		ID     string   `json:"id"`
		Topic  string   `json:"topic"`
		Status string   `json:"status"`
		Personas []string `json:"personas"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析创建响应失败: %v", err)
	}
	if result.Status != "pending" {
		t.Fatalf("期望状态 pending，得到 %s", result.Status)
	}
	if len(result.Personas) != 2 {
		t.Fatalf("期望 2 个 persona，得到 %d", len(result.Personas))
	}
	return result.ID
}

// startRoundtable 通过 API 启动 roundtable。
func startRoundtable(t *testing.T, serverURL, id string) {
	t.Helper()

	resp, err := http.Post(serverURL+"/api/v1/roundtables/"+id+"/start", "application/json", nil)
	if err != nil {
		t.Fatalf("启动 roundtable 失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("启动 roundtable 期望 200，得到 %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("解析启动响应失败: %v", err)
	}
	if result["status"] != "running" {
		t.Fatalf("期望状态 running，得到 %v", result["status"])
	}
}

// sseEvent 表示解析后的 SSE 事件。
type sseEvent struct {
	ID    string
	Event string
	Data  string
}

// collectSSEEvents 通过 EventSource 连接收集 SSE 事件，直到 context 取消或收到 stream_done。
func collectSSEEvents(t *testing.T, serverURL, roundtableID string, lastEventID string, timeout time.Duration) []sseEvent {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	url := serverURL + "/api/v1/roundtables/" + roundtableID + "/events"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("构造 SSE 请求失败: %v", err)
	}
	if lastEventID != "" {
		req.Header.Set("Last-Event-ID", lastEventID)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("SSE 连接失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("SSE 期望 200，得到 %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	var events []sseEvent
	var current sseEvent

	for {
		select {
		case <-ctx.Done():
			return events
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			// 超时或连接关闭时返回已收集的事件
			return events
		}
		line = strings.TrimRight(line, "\n")

		if line == "" {
			// 空行表示一个事件结束
			if current.Event != "" {
				events = append(events, current)
				if current.Event == "stream_done" {
					return events
				}
			}
			current = sseEvent{}
			continue
		}

		if strings.HasPrefix(line, "id: ") {
			current.ID = strings.TrimPrefix(line, "id: ")
		} else if strings.HasPrefix(line, "event: ") {
			current.Event = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			current.Data = strings.TrimPrefix(line, "data: ")
		}
		// 忽略 comment 行（以 ":" 开头）
	}
}

// TestEndToEnd_FullFlow 验证完整的端到端流程：
//  1. 创建 roundtable
//  2. 启动
//  3. SSE 接收事件流
//  4. 验证事件序列
//  5. 快照验证消息数
//  6. 列表过滤验证
func TestEndToEnd_FullFlow(t *testing.T) {
	server := setupIntegrationServer(t)
	defer server.Close()

	// 1. 创建
	id := createRoundtable(t, server.URL)

	// 2. 启动
	startRoundtable(t, server.URL, id)

	// 3. SSE 接收事件流（等待 engine 完成，含 200ms 预延迟 + 事件间隔）
	events := collectSSEEvents(t, server.URL, id, "", 5*time.Second)

	// 4. 验证事件序列包含 stream_start → round_start → speaking → message_done → round_end → stream_done
	var foundStreamStart, foundRoundStart, foundSpeaking, foundMessageDone, foundRoundEnd, foundStreamDone bool
	for _, evt := range events {
		switch evt.Event {
		case "stream_start":
			foundStreamStart = true
		case "round_start":
			foundRoundStart = true
		case "speaking":
			foundSpeaking = true
		case "message_done":
			foundMessageDone = true
		case "round_end":
			foundRoundEnd = true
		case "stream_done":
			foundStreamDone = true
		}
	}
	if !foundStreamStart {
		t.Error("事件流中应包含 stream_start")
	}
	if !foundRoundStart {
		t.Error("事件流中应包含 round_start")
	}
	if !foundSpeaking {
		t.Error("事件流中应包含 speaking")
	}
	if !foundMessageDone {
		t.Error("事件流中应包含 message_done")
	}
	if !foundRoundEnd {
		t.Error("事件流中应包含 round_end")
	}
	if !foundStreamDone {
		t.Error("事件流中应包含 stream_done")
	}

	// 等待状态持久化
	time.Sleep(200 * time.Millisecond)

	// 5. GET /api/v1/roundtables/{id} 验证快照消息数正确（2 persona * 2 rounds = 4）
	resp, err := http.Get(server.URL + "/api/v1/roundtables/" + id)
	if err != nil {
		t.Fatalf("获取快照失败: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("获取快照期望 200，得到 %d", resp.StatusCode)
	}

	var snap struct {
		ID       string `json:"id"`
		Status   string `json:"status"`
		Messages []struct {
			ID      string `json:"id"`
			Content string `json:"content"`
			Round   int    `json:"round"`
		} `json:"messages"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("解析快照失败: %v", err)
	}
	if snap.Status != "completed" {
		t.Errorf("快照状态期望 completed，得到 %s", snap.Status)
	}
	if len(snap.Messages) != 4 {
		t.Errorf("期望 4 条消息，得到 %d", len(snap.Messages))
	}

	// 6. GET /api/v1/roundtables?status=completed 验证列表包含该记录
	resp2, err := http.Get(server.URL + "/api/v1/roundtables?status=completed")
	if err != nil {
		t.Fatalf("获取列表失败: %v", err)
	}
	defer resp2.Body.Close()

	var list []struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&list); err != nil {
		t.Fatalf("解析列表失败: %v", err)
	}

	found := false
	for _, item := range list {
		if item.ID == id && item.Status == "completed" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("completed 列表中应包含刚创建的 roundtable %s", id)
	}
}

// TestEndToEnd_LastEventIDReconnect 验证 SSE 断连后通过 Last-Event-ID 重连不丢事件。
func TestEndToEnd_LastEventIDReconnect(t *testing.T) {
	server := setupIntegrationServer(t)
	defer server.Close()

	id := createRoundtable(t, server.URL)
	startRoundtable(t, server.URL, id)

	// 第一次连接：收集所有事件，记录中间某个 event ID
	events1 := collectSSEEvents(t, server.URL, id, "", 5*time.Second)
	if len(events1) == 0 {
		t.Fatal("第一次连接应收到事件")
	}

	// 找一个 message_done 事件的 ID 作为断点
	var breakEventID string
	for _, evt := range events1 {
		if evt.Event == "message_done" && evt.ID != "" {
			breakEventID = evt.ID
			break
		}
	}
	if breakEventID == "" {
		t.Fatal("应至少找到一个 message_done 事件")
	}

	// 第二次连接：从 breakEventID 之后重连
	events2 := collectSSEEvents(t, server.URL, id, breakEventID, 3*time.Second)

	// 验证重连后的事件不包含 breakEventID 之前的事件
	breakIDInt, _ := strconv.Atoi(breakEventID)
	for _, evt := range events2 {
		if evt.ID == "" {
			continue
		}
		eid, _ := strconv.Atoi(evt.ID)
		if eid <= breakIDInt {
			t.Errorf("重连后不应包含 event ID <= %s 的事件，但收到 id=%s event=%s", breakEventID, evt.ID, evt.Event)
		}
	}

	// 验证重连后包含 stream_done（因为讨论已完成，历史事件已写入 DB）
	var foundStreamDone bool
	for _, evt := range events2 {
		if evt.Event == "stream_done" {
			foundStreamDone = true
			break
		}
	}
	if !foundStreamDone {
		t.Error("重连后应包含 stream_done 事件")
	}
}

// TestEndToEnd_MockLLM 验证集成测试使用 mock LLM，不依赖真实 API。
func TestEndToEnd_MockLLM(t *testing.T) {
	server := setupIntegrationServer(t)
	defer server.Close()

	id := createRoundtable(t, server.URL)
	startRoundtable(t, server.URL, id)

	// 收集事件并验证消息内容来自 mock
	events := collectSSEEvents(t, server.URL, id, "", 5*time.Second)

	var msgDoneCount int
	for _, evt := range events {
		if evt.Event == "message_done" {
			msgDoneCount++
			var payload map[string]interface{}
			if err := json.Unmarshal([]byte(evt.Data), &payload); err != nil {
				t.Fatalf("解析 message_done payload 失败: %v", err)
			}
			content, _ := payload["content"].(string)
			// mock LLM 返回 opening_line
			if content != "hi from S1" && content != "hi from S2" {
				t.Errorf("mock LLM 消息内容应为 'hi from S1' 或 'hi from S2'，得到 %q", content)
			}
		}
	}
	if msgDoneCount != 4 {
		t.Errorf("期望 4 条 message_done 事件，得到 %d", msgDoneCount)
	}
}
