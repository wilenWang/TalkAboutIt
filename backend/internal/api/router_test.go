package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/wilenwang/talkaboutit/internal/engine"
	"github.com/wilenwang/talkaboutit/internal/persona"
	"github.com/wilenwang/talkaboutit/internal/session"
)

func setupTestHandler(t *testing.T) (*Handler, *session.Store, string) {
	t.Helper()
	tmpDir := t.TempDir()
	loader := persona.NewLoader("../../personas")

	dbPath := filepath.Join(tmpDir, "api_test.db")
	store, err := session.NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	eng := engine.NewEngine(store, loader, nil)
	h := NewHandler(loader, store, eng)
	return h, store, dbPath
}

func TestListPersonas(t *testing.T) {
	h, _, _ := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/personas", nil)
	rec := httptest.NewRecorder()

	h.ListPersonas(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，得到 %d", rec.Code)
	}

	var summaries []PersonaSummary
	if err := json.Unmarshal(rec.Body.Bytes(), &summaries); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if len(summaries) == 0 {
		t.Error("期望返回至少一个 persona 摘要")
	}

	found := false
	for _, s := range summaries {
		if s.ID == "steve-jobs" {
			found = true
			if s.Name != "Steve Jobs" {
				t.Errorf("steve-jobs 的 name 期望 Steve Jobs，得到 %s", s.Name)
			}
		}
	}
	if !found {
		t.Error("期望返回 steve-jobs 的摘要")
	}
}

func TestGetPersona(t *testing.T) {
	h, _, _ := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/personas/steve-jobs", nil)
	req.SetPathValue("id", "steve-jobs")
	rec := httptest.NewRecorder()

	h.GetPersona(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，得到 %d", rec.Code)
	}

	var p persona.Persona
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if p.ID != "steve-jobs" {
		t.Errorf("id 期望 steve-jobs，得到 %s", p.ID)
	}
	if p.Name != "Steve Jobs" {
		t.Errorf("name 期望 Steve Jobs，得到 %s", p.Name)
	}
}

func TestGetPersonaNotFound(t *testing.T) {
	h, _, _ := setupTestHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/personas/nonexistent", nil)
	req.SetPathValue("id", "nonexistent")
	rec := httptest.NewRecorder()

	h.GetPersona(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("期望状态码 404，得到 %d", rec.Code)
	}
}

func TestRegisterRoutes(t *testing.T) {
	h, _, _ := setupTestHandler(t)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// 测试列表路由
	req := httptest.NewRequest(http.MethodGet, "/api/v1/personas", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("列表路由期望 200，得到 %d", rec.Code)
	}

	// 测试详情路由
	req = httptest.NewRequest(http.MethodGet, "/api/v1/personas/steve-jobs", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("详情路由期望 200，得到 %d", rec.Code)
	}
}

func TestCreateRoundtable(t *testing.T) {
	h, _, _ := setupTestHandler(t)

	body := `{"topic":"AI 会取代程序员吗？","personas":["steve-jobs","elon-musk"],"max_rounds":3,"language":"zh-CN"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/roundtables", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.CreateRoundtable(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("期望状态码 201，得到 %d, body: %s", rec.Code, rec.Body.String())
	}

	var resp CreateRoundtableResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	if resp.Topic != "AI 会取代程序员吗？" {
		t.Errorf("话题不匹配: %s", resp.Topic)
	}
	if resp.Status != "pending" {
		t.Errorf("期望状态 pending, 得到 %s", resp.Status)
	}
	if len(resp.Personas) != 2 {
		t.Errorf("期望 2 个 persona, 得到 %d", len(resp.Personas))
	}
}

func TestCreateRoundtable_Validation(t *testing.T) {
	h, _, _ := setupTestHandler(t)

	// 缺少话题
	body := `{"personas":["steve-jobs","elon-musk"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/roundtables", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.CreateRoundtable(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("期望 400, 得到 %d", rec.Code)
	}

	// persona 不足
	body = `{"topic":"test","personas":["steve-jobs"]}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/roundtables", strings.NewReader(body))
	rec = httptest.NewRecorder()
	h.CreateRoundtable(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("期望 400, 得到 %d", rec.Code)
	}
}

func TestGetRoundtable_Snapshot(t *testing.T) {
	ctx := context.Background()
	h, store, _ := setupTestHandler(t)

	rt := &session.Roundtable{
		ID:           "rt_snap_001",
		Topic:        "Test",
		PersonasJSON: `["steve-jobs"]`,
		MaxRounds:    1,
		Language:     "zh-CN",
		Status:       "pending",
	}
	store.CreateRoundtable(ctx, rt)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/roundtables/rt_snap_001", nil)
	req.SetPathValue("id", "rt_snap_001")
	rec := httptest.NewRecorder()
	h.GetRoundtable(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("期望 200, 得到 %d", rec.Code)
	}

	var snap RoundtableSnapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if snap.ID != "rt_snap_001" {
		t.Errorf("id 不匹配")
	}
	if snap.Status != "pending" {
		t.Errorf("status 不匹配")
	}
}

func TestStartRoundtable(t *testing.T) {
	ctx := context.Background()
	h, store, _ := setupTestHandler(t)

	// 创建临时 persona 文件用于 engine 运行
	tmpDir := t.TempDir()
	p1 := `{"schema_version":"persona.v1","id":"test-s1","name":"S1","display_name":"S1","avatar":"🤖","role_title":"T","description":"D","tags":[],"language":{"primary":"zh-CN","allowed":["zh-CN"],"default_output":"follow_user","style_hint":""},"stance":{"default_position":"pro","intensity":3,"biases":[],"taboos":[]},"core_beliefs":[],"speaking_style":{"tone":"calm","cadence":"balanced","verbosity":3,"signature_patterns":[],"do":[],"dont":[]},"knowledge_scope":{"domains":[],"expertise_level":{},"time_cutoff":"","allowed_inference":"medium","unknown_handling":"","forbidden_claims":[]},"interaction_rules":{"address_others":"","disagreement_style":"","interruption_policy":"never","question_policy":"","concession_policy":"","avoid":[]},"debate_goal":{"primary_goal":"test","secondary_goals":[],"win_condition":"","loss_condition":""},"prompting":{"system_preamble":"","reply_constraints":[]},"examples":{"opening_line":"hi","sample_rebuttal":""}}`
	p2 := `{"schema_version":"persona.v1","id":"test-s2","name":"S2","display_name":"S2","avatar":"👾","role_title":"T","description":"D","tags":[],"language":{"primary":"zh-CN","allowed":["zh-CN"],"default_output":"follow_user","style_hint":""},"stance":{"default_position":"con","intensity":3,"biases":[],"taboos":[]},"core_beliefs":[],"speaking_style":{"tone":"calm","cadence":"balanced","verbosity":3,"signature_patterns":[],"do":[],"dont":[]},"knowledge_scope":{"domains":[],"expertise_level":{},"time_cutoff":"","allowed_inference":"medium","unknown_handling":"","forbidden_claims":[]},"interaction_rules":{"address_others":"","disagreement_style":"","interruption_policy":"never","question_policy":"","concession_policy":"","avoid":[]},"debate_goal":{"primary_goal":"test","secondary_goals":[],"win_condition":"","loss_condition":""},"prompting":{"system_preamble":"","reply_constraints":[]},"examples":{"opening_line":"hello","sample_rebuttal":""}}`
	os.WriteFile(filepath.Join(tmpDir, "test-s1.json"), []byte(p1), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test-s2.json"), []byte(p2), 0644)

	loader := persona.NewLoader(tmpDir)
	eng := engine.NewEngine(store, loader, nil)
	h = NewHandler(loader, store, eng)

	rt := &session.Roundtable{
		ID:           "rt_start_001",
		Topic:        "Test",
		PersonasJSON: `["test-s1","test-s2"]`,
		MaxRounds:    1,
		Language:     "zh-CN",
		Status:       "pending",
	}
	store.CreateRoundtable(ctx, rt)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/roundtables/rt_start_001/start", nil)
	req.SetPathValue("id", "rt_start_001")
	rec := httptest.NewRecorder()
	h.StartRoundtable(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("期望 200, 得到 %d, body: %s", rec.Code, rec.Body.String())
	}

	// 等待 engine 异步完成（含 200ms 预延迟 + 事件间隔）
	time.Sleep(1200 * time.Millisecond)

	got, err := store.GetRoundtable(ctx, rt.ID)
	if err != nil {
		t.Fatalf("GetRoundtable failed: %v", err)
	}
	if got.Status != "completed" {
		t.Errorf("期望 completed, 得到 %s", got.Status)
	}
}

// TestGetRoundtable_SnapshotWithMessages 验证快照接口包含消息列表。
func TestGetRoundtable_SnapshotWithMessages(t *testing.T) {
	ctx := context.Background()
	h, store, _ := setupTestHandler(t)

	// 创建临时 persona 文件
	tmpDir := t.TempDir()
	p1 := `{"schema_version":"persona.v1","id":"test-s1","name":"S1","display_name":"S1","avatar":"🤖","role_title":"T","description":"D","tags":[],"language":{"primary":"zh-CN","allowed":["zh-CN"],"default_output":"follow_user","style_hint":""},"stance":{"default_position":"pro","intensity":3,"biases":[],"taboos":[]},"core_beliefs":[],"speaking_style":{"tone":"calm","cadence":"balanced","verbosity":3,"signature_patterns":[],"do":[],"dont":[]},"knowledge_scope":{"domains":[],"expertise_level":{},"time_cutoff":"","allowed_inference":"medium","unknown_handling":"","forbidden_claims":[]},"interaction_rules":{"address_others":"","disagreement_style":"","interruption_policy":"never","question_policy":"","concession_policy":"","avoid":[]},"debate_goal":{"primary_goal":"test","secondary_goals":[],"win_condition":"","loss_condition":""},"prompting":{"system_preamble":"","reply_constraints":[]},"examples":{"opening_line":"hi","sample_rebuttal":""}}`
	p2 := `{"schema_version":"persona.v1","id":"test-s2","name":"S2","display_name":"S2","avatar":"👾","role_title":"T","description":"D","tags":[],"language":{"primary":"zh-CN","allowed":["zh-CN"],"default_output":"follow_user","style_hint":""},"stance":{"default_position":"con","intensity":3,"biases":[],"taboos":[]},"core_beliefs":[],"speaking_style":{"tone":"calm","cadence":"balanced","verbosity":3,"signature_patterns":[],"do":[],"dont":[]},"knowledge_scope":{"domains":[],"expertise_level":{},"time_cutoff":"","allowed_inference":"medium","unknown_handling":"","forbidden_claims":[]},"interaction_rules":{"address_others":"","disagreement_style":"","interruption_policy":"never","question_policy":"","concession_policy":"","avoid":[]},"debate_goal":{"primary_goal":"test","secondary_goals":[],"win_condition":"","loss_condition":""},"prompting":{"system_preamble":"","reply_constraints":[]},"examples":{"opening_line":"hello","sample_rebuttal":""}}`
	os.WriteFile(filepath.Join(tmpDir, "test-s1.json"), []byte(p1), 0644)
	os.WriteFile(filepath.Join(tmpDir, "test-s2.json"), []byte(p2), 0644)

	loader := persona.NewLoader(tmpDir)
	eng := engine.NewEngine(store, loader, nil)
	h = NewHandler(loader, store, eng)

	rt := &session.Roundtable{
		ID:           "rt_snap_msg_001",
		Topic:        "Snapshot Test",
		PersonasJSON: `["test-s1","test-s2"]`,
		MaxRounds:    1,
		Language:     "zh-CN",
		Status:       "pending",
	}
	store.CreateRoundtable(ctx, rt)

	// 启动讨论并等待完成
	req := httptest.NewRequest(http.MethodPost, "/api/v1/roundtables/rt_snap_msg_001/start", nil)
	req.SetPathValue("id", "rt_snap_msg_001")
	rec := httptest.NewRecorder()
	h.StartRoundtable(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("启动失败: %d, %s", rec.Code, rec.Body.String())
	}
	time.Sleep(1200 * time.Millisecond)

	// 查询快照
	req = httptest.NewRequest(http.MethodGet, "/api/v1/roundtables/rt_snap_msg_001", nil)
	req.SetPathValue("id", "rt_snap_msg_001")
	rec = httptest.NewRecorder()
	h.GetRoundtable(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("期望 200, 得到 %d", rec.Code)
	}

	var snap RoundtableSnapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snap); err != nil {
		t.Fatalf("解析快照失败: %v", err)
	}
	if snap.ID != "rt_snap_msg_001" {
		t.Errorf("id 不匹配")
	}
	if snap.Status != "completed" {
		t.Errorf("status 期望 completed, 得到 %s", snap.Status)
	}
	if len(snap.Messages) == 0 {
		t.Errorf("期望快照包含消息列表")
	}
	// 验证消息内容
	for _, m := range snap.Messages {
		if m.RoundtableID != "rt_snap_msg_001" {
			t.Errorf("消息 roundtable_id 不匹配")
		}
		if m.Content == "" {
			t.Errorf("消息内容不应为空")
		}
	}
}

// TestSSEHandler_LastEventIDReconnect 验证 SSE 断连后通过 Last-Event-ID 重连，事件不丢不重。
func TestSSEHandler_LastEventIDReconnect(t *testing.T) {
	ctx := context.Background()
	h, store, _ := setupTestHandler(t)

	// 创建 roundtable 并手动写入若干事件
	rt := &session.Roundtable{
		ID:           "rt_sse_reconnect_001",
		Topic:        "SSE Reconnect Test",
		PersonasJSON: `["steve-jobs","elon-musk"]`,
		MaxRounds:    1,
		Language:     "zh-CN",
		Status:       "running",
	}
	if err := store.CreateRoundtable(ctx, rt); err != nil {
		t.Fatalf("创建 roundtable 失败: %v", err)
	}

	// 写入 3 个事件到 DB（模拟历史）
	for i := 0; i < 3; i++ {
		_, err := store.AddEvent(ctx, rt.ID, "message_chunk", nil, nil, nil, nil, map[string]interface{}{
			"index": i,
		})
		if err != nil {
			t.Fatalf("写入事件失败: %v", err)
		}
	}

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// 第一次连接：不带 Last-Event-ID
	req1, cancel1 := context.WithCancel(ctx)
	rec1 := httptest.NewRecorder()

	done1 := make(chan struct{})
	go func() {
		mux.ServeHTTP(rec1, httptest.NewRequest(http.MethodGet, "/api/v1/roundtables/rt_sse_reconnect_001/events", nil).WithContext(req1))
		close(done1)
	}()

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 通过 bus 发布 event 4（仅广播到已连接客户端，不写入 DB）
	h.bus.Publish(rt.ID, session.Event{
		RoundtableID: rt.ID,
		EventID:      4,
		EventType:    "message_chunk",
		PayloadJSON:  `{"index":3}`,
	})

	time.Sleep(100 * time.Millisecond)

	// 关闭第一个连接
	cancel1()
	<-done1

	// 将 event 4 和 5 写入 DB（模拟 engine 在广播后持久化的事件）
	_, _ = store.AddEvent(ctx, rt.ID, "message_chunk", nil, nil, nil, nil, map[string]interface{}{"index": 3})
	_, _ = store.AddEvent(ctx, rt.ID, "message_chunk", nil, nil, nil, nil, map[string]interface{}{"index": 4})

	// 第二次连接：带 Last-Event-ID=2，应补发 DB 中的 event 3、4、5
	req2, cancel2 := context.WithCancel(ctx)
	rec2 := httptest.NewRecorder()

	done2 := make(chan struct{})
	go func() {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/roundtables/rt_sse_reconnect_001/events", nil).WithContext(req2)
		r.Header.Set("Last-Event-ID", "2")
		mux.ServeHTTP(rec2, r)
		close(done2)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel2()
	<-done2

	body2 := rec2.Body.String()
	// 应包含 id: 3、4、5
	if !strings.Contains(body2, "id: 3") {
		t.Errorf("重连后应补发 event 3，响应体:\n%s", body2)
	}
	if !strings.Contains(body2, "id: 4") {
		t.Errorf("重连后应补发 event 4，响应体:\n%s", body2)
	}
	if !strings.Contains(body2, "id: 5") {
		t.Errorf("重连后应补发 event 5，响应体:\n%s", body2)
	}
	// 不应重复 event 1 和 2
	if strings.Contains(body2, "id: 1") || strings.Contains(body2, "id: 2") {
		t.Errorf("重连后不应重复 event 1/2，响应体:\n%s", body2)
	}
}

// TestListRoundtables 验证列表接口支持 status 过滤和全量查询。
func TestListRoundtables(t *testing.T) {
	ctx := context.Background()
	h, store, _ := setupTestHandler(t)

	// 创建 3 条记录：pending、completed、completed
	for i, status := range []string{"pending", "completed", "completed"} {
		rt := &session.Roundtable{
			ID:           fmt.Sprintf("rt_list_%d", i),
			Topic:        fmt.Sprintf("Topic %d", i),
			PersonasJSON: `["steve-jobs"]`,
			MaxRounds:    2,
			Language:     "zh-CN",
			Status:       status,
		}
		if err := store.CreateRoundtable(ctx, rt); err != nil {
			t.Fatalf("创建 roundtable 失败: %v", err)
		}
	}

	// 全量查询
	req := httptest.NewRequest(http.MethodGet, "/api/v1/roundtables", nil)
	rec := httptest.NewRecorder()
	h.ListRoundtables(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("期望 200, 得到 %d", rec.Code)
	}
	var all []RoundtableListItem
	if err := json.Unmarshal(rec.Body.Bytes(), &all); err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("期望 3 条，得到 %d", len(all))
	}

	// 按 completed 过滤
	req = httptest.NewRequest(http.MethodGet, "/api/v1/roundtables?status=completed", nil)
	rec = httptest.NewRecorder()
	h.ListRoundtables(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("期望 200, 得到 %d", rec.Code)
	}
	var completed []RoundtableListItem
	if err := json.Unmarshal(rec.Body.Bytes(), &completed); err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(completed) != 2 {
		t.Errorf("期望 2 条 completed，得到 %d", len(completed))
	}
	for _, item := range completed {
		if item.Status != "completed" {
			t.Errorf("期望 status completed, 得到 %s", item.Status)
		}
	}
}

// TestListRoundtables_SortOrder 验证列表按 created_at DESC 排序。
func TestListRoundtables_SortOrder(t *testing.T) {
	ctx := context.Background()
	h, store, _ := setupTestHandler(t)

	// 创建 3 条记录，按时间顺序插入；SQLite CURRENT_TIMESTAMP 精度为秒，需间隔 >1s
	ids := []string{"rt_older", "rt_mid", "rt_newer"}
	for i, id := range ids {
		rt := &session.Roundtable{
			ID:           id,
			Topic:        fmt.Sprintf("Topic %d", i),
			PersonasJSON: `["steve-jobs"]`,
			MaxRounds:    2,
			Language:     "zh-CN",
			Status:       "completed",
		}
		if err := store.CreateRoundtable(ctx, rt); err != nil {
			t.Fatalf("创建 roundtable 失败: %v", err)
		}
		// 确保时间有差异（SQLite 默认秒级精度）
		time.Sleep(1100 * time.Millisecond)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/roundtables", nil)
	rec := httptest.NewRecorder()
	h.ListRoundtables(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("期望 200, 得到 %d", rec.Code)
	}
	var list []RoundtableListItem
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(list) != 3 {
		t.Fatalf("期望 3 条，得到 %d", len(list))
	}
	// 验证 DESC 排序：最新的在前
	if list[0].ID != "rt_newer" {
		t.Errorf("第一条期望 rt_newer，得到 %s", list[0].ID)
	}
	if list[2].ID != "rt_older" {
		t.Errorf("最后一条期望 rt_older，得到 %s", list[2].ID)
	}
}

// TestListRoundtables_InvalidStatus 验证非法 status 返回空列表。
func TestListRoundtables_InvalidStatus(t *testing.T) {
	ctx := context.Background()
	h, store, _ := setupTestHandler(t)

	rt := &session.Roundtable{
		ID:           "rt_status_test",
		Topic:        "Status Test",
		PersonasJSON: `["steve-jobs"]`,
		MaxRounds:    2,
		Language:     "zh-CN",
		Status:       "completed",
	}
	if err := store.CreateRoundtable(ctx, rt); err != nil {
		t.Fatalf("创建 roundtable 失败: %v", err)
	}

	// 使用非法 status 查询
	req := httptest.NewRequest(http.MethodGet, "/api/v1/roundtables?status=nonexistent", nil)
	rec := httptest.NewRecorder()
	h.ListRoundtables(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("期望 200, 得到 %d", rec.Code)
	}
	var list []RoundtableListItem
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("非法 status 应返回空列表，得到 %d 条", len(list))
	}
}

// TestMethodNotAllowed 验证对不存在的路由使用错误的 HTTP 方法时返回 405。
func TestMethodNotAllowed(t *testing.T) {
	h, _, _ := setupTestHandler(t)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// POST /api/v1/personas 不存在，应返回 405（Go 1.22+ ServeMux 对方法不匹配返回 405）
	req := httptest.NewRequest(http.MethodPost, "/api/v1/personas", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /api/v1/personas 期望 405，得到 %d", rec.Code)
	}

	// PUT /api/v1/roundtables/{id} 不存在，应返回 405
	req = httptest.NewRequest(http.MethodPut, "/api/v1/roundtables/rt_001", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("PUT /api/v1/roundtables/rt_001 期望 405，得到 %d", rec.Code)
	}

	// DELETE /api/v1/roundtables/{id}/start 不存在，应返回 405
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/roundtables/rt_001/start", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("DELETE /api/v1/roundtables/rt_001/start 期望 405，得到 %d", rec.Code)
	}
}

// TestSSEHandler_QueryParamLastEventID 验证通过 Query 参数传递 lastEventId 也能重连补发。
func TestSSEHandler_QueryParamLastEventID(t *testing.T) {
	ctx := context.Background()
	h, store, _ := setupTestHandler(t)

	rt := &session.Roundtable{
		ID:           "rt_sse_query_001",
		Topic:        "SSE Query Reconnect Test",
		PersonasJSON: `["steve-jobs"]`,
		MaxRounds:    1,
		Language:     "zh-CN",
		Status:       "running",
	}
	store.CreateRoundtable(ctx, rt)

	// 写入 2 个事件
	for i := 0; i < 2; i++ {
		_, err := store.AddEvent(ctx, rt.ID, "message_chunk", nil, nil, nil, nil, map[string]interface{}{
			"index": i,
		})
		if err != nil {
			t.Fatalf("写入事件失败: %v", err)
		}
	}

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	reqCtx, cancel := context.WithCancel(ctx)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/roundtables/rt_sse_query_001/events?lastEventId=1", nil).WithContext(reqCtx)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		mux.ServeHTTP(rec, req)
		close(done)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	body := rec.Body.String()
	if !strings.Contains(body, "id: 2") {
		t.Errorf("Query 参数重连应补发 event 2，响应体:\n%s", body)
	}
	if strings.Contains(body, "id: 1") {
		t.Errorf("Query 参数重连不应重复 event 1，响应体:\n%s", body)
	}
}
