package session

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestStore_CreateGetUpdate(t *testing.T) {
	ctx := context.Background()
	dbPath := "/tmp/test_talkaboutit_" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(dbPath)

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	rt := &Roundtable{
		ID:           "rt_test_001",
		Topic:        "AI 会取代程序员吗？",
		PersonasJSON: `["steve-jobs","elon-musk"]`,
		MaxRounds:    3,
		Language:     "zh-CN",
		Status:       "pending",
		LastEventID:  0,
	}

	if err := store.CreateRoundtable(ctx, rt); err != nil {
		t.Fatalf("CreateRoundtable failed: %v", err)
	}

	got, err := store.GetRoundtable(ctx, rt.ID)
	if err != nil {
		t.Fatalf("GetRoundtable failed: %v", err)
	}
	if got.ID != rt.ID || got.Topic != rt.Topic || got.Status != "pending" {
		t.Errorf("roundtable mismatch: got %+v", got)
	}

	if err := store.UpdateStatus(ctx, rt.ID, "running"); err != nil {
		t.Fatalf("UpdateStatus failed: %v", err)
	}

	got, err = store.GetRoundtable(ctx, rt.ID)
	if err != nil {
		t.Fatalf("GetRoundtable after update failed: %v", err)
	}
	if got.Status != "running" {
		t.Errorf("expected status running, got %s", got.Status)
	}
	if got.StartedAt == nil {
		t.Error("expected started_at to be set")
	}
}

func TestStore_AddEventAndMessages(t *testing.T) {
	ctx := context.Background()
	dbPath := "/tmp/test_talkaboutit_event_" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(dbPath)

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	rt := &Roundtable{
		ID:           "rt_event_001",
		Topic:        "测试话题",
		PersonasJSON: `["p1","p2"]`,
		MaxRounds:    2,
		Language:     "zh-CN",
		Status:       "pending",
	}
	if err := store.CreateRoundtable(ctx, rt); err != nil {
		t.Fatalf("CreateRoundtable failed: %v", err)
	}

	// stream_start
	r1 := 1
	si0 := 0
	p1 := "p1"
	mid1 := "msg_001"
	e1, err := store.AddEvent(ctx, rt.ID, "stream_start", &r1, &si0, &p1, nil, map[string]interface{}{"total_rounds": 2})
	if err != nil {
		t.Fatalf("AddEvent stream_start failed: %v", err)
	}
	if e1.EventID != 1 {
		t.Errorf("expected event_id 1, got %d", e1.EventID)
	}

	// message_done
	e2, err := store.AddEvent(ctx, rt.ID, "message_done", &r1, &si0, &p1, &mid1, map[string]interface{}{
		"content": "Hello world",
	})
	if err != nil {
		t.Fatalf("AddEvent message_done failed: %v", err)
	}
	if e2.EventID != 2 {
		t.Errorf("expected event_id 2, got %d", e2.EventID)
	}

	// 验证消息已写入
	msgs, err := store.GetMessages(ctx, rt.ID)
	if err != nil {
		t.Fatalf("GetMessages failed: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != "Hello world" {
		t.Errorf("expected content 'Hello world', got %s", msgs[0].Content)
	}

	// 验证 last_event_id 更新
	got, _ := store.GetRoundtable(ctx, rt.ID)
	if got.LastEventID != 2 {
		t.Errorf("expected last_event_id 2, got %d", got.LastEventID)
	}

	// 验证 GetEventsAfter
	events, err := store.GetEventsAfter(ctx, rt.ID, 0)
	if err != nil {
		t.Fatalf("GetEventsAfter failed: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	events, err = store.GetEventsAfter(ctx, rt.ID, 1)
	if err != nil {
		t.Fatalf("GetEventsAfter(1) failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event after 1, got %d", len(events))
	}
	if events[0].EventID != 2 {
		t.Errorf("expected event_id 2, got %d", events[0].EventID)
	}
}

func TestStore_GetEventsAfter_NilPointers(t *testing.T) {
	ctx := context.Background()
	dbPath := "/tmp/test_talkaboutit_nil_" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(dbPath)

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	rt := &Roundtable{
		ID:           "rt_nil_001",
		Topic:        "测试",
		PersonasJSON: `[]`,
		MaxRounds:    1,
		Language:     "zh-CN",
		Status:       "pending",
	}
	if err := store.CreateRoundtable(ctx, rt); err != nil {
		t.Fatalf("CreateRoundtable failed: %v", err)
	}

	// 事件不含 round/speaker_index/persona_id
	e1, err := store.AddEvent(ctx, rt.ID, "stream_done", nil, nil, nil, nil, map[string]interface{}{"total_messages": 0})
	if err != nil {
		t.Fatalf("AddEvent stream_done failed: %v", err)
	}
	if e1.EventID != 1 {
		t.Errorf("expected event_id 1, got %d", e1.EventID)
	}

	events, err := store.GetEventsAfter(ctx, rt.ID, 0)
	if err != nil {
		t.Fatalf("GetEventsAfter failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].EventType != "stream_done" {
		t.Errorf("expected stream_done, got %s", events[0].EventType)
	}
	if events[0].Round != nil {
		t.Error("expected round to be nil")
	}
	if events[0].PersonaID != nil {
		t.Error("expected persona_id to be nil")
	}
}

// TestStore_ListRoundtables_LimitBounds 验证 limit 边界值（0、负数、超大值）。
func TestStore_ListRoundtables_LimitBounds(t *testing.T) {
	ctx := context.Background()
	dbPath := "/tmp/test_talkaboutit_limit_" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(dbPath)

	store, err := NewStore(dbPath)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store.Close()

	// 创建 5 条记录
	for i := 0; i < 5; i++ {
		rt := &Roundtable{
			ID:           fmt.Sprintf("rt_limit_%d", i),
			Topic:        fmt.Sprintf("Topic %d", i),
			PersonasJSON: `["steve-jobs"]`,
			MaxRounds:    2,
			Language:     "zh-CN",
			Status:       "completed",
		}
		if err := store.CreateRoundtable(ctx, rt); err != nil {
			t.Fatalf("CreateRoundtable failed: %v", err)
		}
	}

	// limit = 0 时应被修正为 1
	list, err := store.ListRoundtables(ctx, "", 0)
	if err != nil {
		t.Fatalf("ListRoundtables(limit=0) failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("limit=0 应返回 1 条，得到 %d", len(list))
	}

	// limit = -5 时应被修正为 1
	list, err = store.ListRoundtables(ctx, "", -5)
	if err != nil {
		t.Fatalf("ListRoundtables(limit=-5) failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("limit=-5 应返回 1 条，得到 %d", len(list))
	}

	// limit = 9999 时应被硬上限截断为 100
	list, err = store.ListRoundtables(ctx, "", 9999)
	if err != nil {
		t.Fatalf("ListRoundtables(limit=9999) failed: %v", err)
	}
	if len(list) != 5 {
		t.Errorf("limit=9999 应返回全部 5 条（被上限 100 截断后仍有 5 条），得到 %d", len(list))
	}

	// 验证上限确实生效：创建 120 条记录，limit=9999 应只返回 100 条
	dbPath2 := "/tmp/test_talkaboutit_limit2_" + time.Now().Format("20060102150405") + ".db"
	defer os.Remove(dbPath2)
	store2, err := NewStore(dbPath2)
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}
	defer store2.Close()
	for i := 0; i < 120; i++ {
		rt := &Roundtable{
			ID:           fmt.Sprintf("rt_bulk_%d", i),
			Topic:        fmt.Sprintf("Bulk %d", i),
			PersonasJSON: `[]`,
			MaxRounds:    1,
			Language:     "zh-CN",
			Status:       "pending",
		}
		if err := store2.CreateRoundtable(ctx, rt); err != nil {
			t.Fatalf("CreateRoundtable failed: %v", err)
		}
	}
	list, err = store2.ListRoundtables(ctx, "", 9999)
	if err != nil {
		t.Fatalf("ListRoundtables(limit=9999) bulk failed: %v", err)
	}
	if len(list) != 100 {
		t.Errorf("limit=9999 应被硬上限截断为 100 条，得到 %d", len(list))
	}
}
