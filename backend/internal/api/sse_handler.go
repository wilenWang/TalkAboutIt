// Package api 提供 TalkAboutIt 的 HTTP API 路由与处理器。
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/wilenwang/talkaboutit/internal/llm"
	"github.com/wilenwang/talkaboutit/internal/session"
)

// EventBus 管理 SSE 订阅者，支持按 roundtable ID 广播。
type EventBus struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan session.Event]struct{} // roundtableID -> set of channels
	store       *session.Store
}

// NewEventBus 创建事件总线。
func NewEventBus(store *session.Store) *EventBus {
	return &EventBus{
		subscribers: make(map[string]map[chan session.Event]struct{}),
		store:       store,
	}
}

// Subscribe 为指定 roundtable 订阅事件流，返回接收通道和取消函数。
func (b *EventBus) Subscribe(roundtableID string) (chan session.Event, func()) {
	ch := make(chan session.Event, 64)
	b.mu.Lock()
	if b.subscribers[roundtableID] == nil {
		b.subscribers[roundtableID] = make(map[chan session.Event]struct{})
	}
	b.subscribers[roundtableID][ch] = struct{}{}
	b.mu.Unlock()

	cancel := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if subs, ok := b.subscribers[roundtableID]; ok {
			delete(subs, ch)
			if len(subs) == 0 {
				delete(b.subscribers, roundtableID)
			}
		}
		// 不主动 close(ch)，避免与 Publish 并发时 send on closed channel panic
	}
	return ch, cancel
}

// Publish 向指定 roundtable 的所有订阅者广播事件。
func (b *EventBus) Publish(roundtableID string, evt session.Event) {
	b.mu.RLock()
	subs, ok := b.subscribers[roundtableID]
	if !ok {
		b.mu.RUnlock()
		return
	}
	// 复制订阅者列表避免在发送时持有读锁
	chans := make([]chan session.Event, 0, len(subs))
	for ch := range subs {
		chans = append(chans, ch)
	}
	b.mu.RUnlock()

	for _, ch := range chans {
		select {
		case ch <- evt:
		default:
			// 通道满则丢弃，避免阻塞
		}
	}
}

// SSEHandler 处理 GET /api/v1/roundtables/{id}/events
func (h *Handler) SSEHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"缺少 roundtable ID"}`, http.StatusBadRequest)
		return
	}

	// 验证 roundtable 存在
	_, err := h.store.GetRoundtable(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"roundtable 不存在"}`, http.StatusNotFound)
		return
	}

	// SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, `{"error":"不支持流式输出"}`, http.StatusInternalServerError)
		return
	}

	// 解析 Last-Event-ID（重连补发），支持 Header 和 Query 参数
	lastEventID := 0
	if leid := r.Header.Get("Last-Event-ID"); leid != "" {
		if v, err := strconv.Atoi(leid); err == nil {
			lastEventID = v
		}
	}
	if lastEventID == 0 {
		if leid := r.URL.Query().Get("lastEventId"); leid != "" {
			if v, err := strconv.Atoi(leid); err == nil {
				lastEventID = v
			}
		}
	}

	// 订阅实时事件（必须先订阅，再补历史，避免竞态漏事件）
	ch, cancel := h.bus.Subscribe(id)
	defer cancel()

	// 补发历史事件，并用 seen map 去重
	seen := make(map[int]struct{})
	if lastEventID > 0 {
		history, err := h.store.GetEventsAfter(r.Context(), id, lastEventID)
		if err != nil {
			http.Error(w, `{"error":"历史事件读取失败"}`, http.StatusInternalServerError)
			return
		}
		for _, evt := range history {
			seen[evt.EventID] = struct{}{}
			writeEvent(w, evt)
		}
		flusher.Flush()
	}

	// 心跳 ticker
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// 发送 stream_start 给新连接（如果没有 Last-Event-ID）
	if lastEventID == 0 {
		fmt.Fprintf(w, ": connected\n\n")
		flusher.Flush()
	}

	for {
		select {
		case evt := <-ch:
			if _, ok := seen[evt.EventID]; ok {
				continue
			}
			writeEvent(w, evt)
			flusher.Flush()

		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()

		case <-r.Context().Done():
			return
		}
	}
}

func writeEvent(w http.ResponseWriter, evt session.Event) {
	fmt.Fprintf(w, "id: %d\n", evt.EventID)
	fmt.Fprintf(w, "event: %s\n", evt.EventType)
	fmt.Fprintf(w, "data: %s\n\n", evt.PayloadJSON)
}

// BroadcastEvent 将事件写入 DB 并广播到所有订阅者。
func (h *Handler) BroadcastEvent(ctx context.Context, roundtableID string, eventType string,
	round, speakerIndex *int, personaID, messageID *string, payload map[string]interface{}) (*session.Event, error) {

	evt, err := h.store.AddEvent(ctx, roundtableID, eventType, round, speakerIndex, personaID, messageID, payload)
	if err != nil {
		return nil, err
	}
	if h.bus != nil {
		h.bus.Publish(roundtableID, *evt)
	}
	return evt, nil
}

// CreateRoundtableRequest 创建 roundtable 的请求体。
type CreateRoundtableRequest struct {
	Topic      string   `json:"topic"`
	Personas   []string `json:"personas"`
	MaxRounds  int      `json:"max_rounds"`
	Language   string   `json:"language"`
}

// CreateRoundtableResponse 创建 roundtable 的响应体。
type CreateRoundtableResponse struct {
	ID        string   `json:"id"`
	Topic     string   `json:"topic"`
	Personas  []string `json:"personas"`
	MaxRounds int      `json:"max_rounds"`
	Language  string   `json:"language"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
}

// RoundtableSnapshot 是 GET /roundtables/{id} 的响应。
type RoundtableSnapshot struct {
	ID           string            `json:"id"`
	Topic        string            `json:"topic"`
	Personas     []string          `json:"personas"`
	MaxRounds    int               `json:"max_rounds"`
	Language     string            `json:"language"`
	Status       string            `json:"status"`
	CreatedAt    string            `json:"created_at"`
	StartedAt    *string           `json:"started_at,omitempty"`
	FinishedAt   *string           `json:"finished_at,omitempty"`
	LastEventID  int               `json:"last_event_id"`
	Messages     []session.Message `json:"messages"`
}

// CreateRoundtable 创建新的圆桌讨论。
func (h *Handler) CreateRoundtable(w http.ResponseWriter, r *http.Request) {
	var req CreateRoundtableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"请求体解析失败"}`, http.StatusBadRequest)
		return
	}

	if req.Topic == "" || strings.TrimSpace(req.Topic) == "" {
		http.Error(w, `{"error":"话题不能为空"}`, http.StatusBadRequest)
		return
	}
	if len(req.Personas) < 2 || len(req.Personas) > 4 {
		http.Error(w, `{"error":"参与者数量必须在 2 到 4 之间"}`, http.StatusBadRequest)
		return
	}
	// 去重并验证 persona 存在
	seen := make(map[string]struct{})
	for _, pid := range req.Personas {
		if _, ok := seen[pid]; ok {
			http.Error(w, `{"error":"参与者不能重复"}`, http.StatusBadRequest)
			return
		}
		seen[pid] = struct{}{}
		if _, err := h.loader.LoadOne(pid); err != nil {
			http.Error(w, `{"error":"参与者不存在: `+pid+`"}`, http.StatusBadRequest)
			return
		}
	}
	if req.MaxRounds < 1 || req.MaxRounds > 5 {
		req.MaxRounds = 3
	}
	if req.Language == "" {
		req.Language = "zh-CN"
	}

	personasJSON, _ := json.Marshal(req.Personas)
	rt := &session.Roundtable{
		ID:           generateID(),
		Topic:        req.Topic,
		PersonasJSON: string(personasJSON),
		MaxRounds:    req.MaxRounds,
		Language:     req.Language,
		Status:       "pending",
	}

	if err := h.store.CreateRoundtable(r.Context(), rt); err != nil {
		http.Error(w, `{"error":"创建失败"}`, http.StatusInternalServerError)
		return
	}

	// 重新读取以获取数据库生成的 created_at
	rt, err := h.store.GetRoundtable(r.Context(), rt.ID)
	if err != nil {
		http.Error(w, `{"error":"创建后读取失败"}`, http.StatusInternalServerError)
		return
	}

	resp := CreateRoundtableResponse{
		ID:        rt.ID,
		Topic:     rt.Topic,
		Personas:  req.Personas,
		MaxRounds: rt.MaxRounds,
		Language:  rt.Language,
		Status:    rt.Status,
		CreatedAt: rt.CreatedAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// StartRoundtable 启动圆桌讨论。
func (h *Handler) StartRoundtable(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"缺少 roundtable ID"}`, http.StatusBadRequest)
		return
	}

	// 原子 CAS：只有 pending 才能切到 running，防止并发双启动
	ok, err := h.store.MarkRunning(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"启动失败"}`, http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, `{"error":"roundtable 不在 pending 状态"}`, http.StatusConflict)
		return
	}

	// 先让 SSE 客户端有时间建立订阅，再启动 engine
	go func() {
		time.Sleep(200 * time.Millisecond)
		ctx := context.Background()
		if err := h.engine.Run(ctx, id); err != nil {
			// 发送 error 事件并更新状态为 failed
			// 不可恢复错误使用统一用户友好文案，不直接暴露内部错误
			var userMsg string
			var perr *llm.ProviderError
			if errors.As(err, &perr) {
				userMsg = perr.UserMessage
			} else {
				userMsg = "讨论发生不可恢复错误，已终止。请稍后重试或联系管理员。"
			}
			h.BroadcastEvent(ctx, id, "error", nil, nil, nil, nil, map[string]interface{}{
				"error":       userMsg,
				"recoverable": false,
			})
			h.store.UpdateStatus(ctx, id, "failed")
		}
	}()

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     id,
		"status": "running",
	})
}

// GetRoundtable 返回圆桌讨论快照（含消息列表）。
func (h *Handler) GetRoundtable(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"缺少 roundtable ID"}`, http.StatusBadRequest)
		return
	}

	rt, err := h.store.GetRoundtable(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"roundtable 不存在"}`, http.StatusNotFound)
		return
	}

	msgs, err := h.store.GetMessages(r.Context(), id)
	if err != nil {
		http.Error(w, `{"error":"读取消息失败"}`, http.StatusInternalServerError)
		return
	}

	var personas []string
	json.Unmarshal([]byte(rt.PersonasJSON), &personas)

	snap := RoundtableSnapshot{
		ID:          rt.ID,
		Topic:       rt.Topic,
		Personas:    personas,
		MaxRounds:   rt.MaxRounds,
		Language:    rt.Language,
		Status:      rt.Status,
		CreatedAt:   rt.CreatedAt.Format(time.RFC3339),
		LastEventID: rt.LastEventID,
		Messages:    msgs,
	}
	if rt.StartedAt != nil {
		s := rt.StartedAt.Format(time.RFC3339)
		snap.StartedAt = &s
	}
	if rt.FinishedAt != nil {
		s := rt.FinishedAt.Format(time.RFC3339)
		snap.FinishedAt = &s
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(snap)
}

// RoundtableListItem 是列表接口返回的摘要项。
type RoundtableListItem struct {
	ID         string   `json:"id"`
	Topic      string   `json:"topic"`
	Personas   []string `json:"personas"`
	MaxRounds  int      `json:"max_rounds"`
	Status     string   `json:"status"`
	CreatedAt  string   `json:"created_at"`
	FinishedAt *string  `json:"finished_at,omitempty"`
}

// ListRoundtables 返回圆桌讨论列表，支持按 status 过滤。
func (h *Handler) ListRoundtables(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	list, err := h.store.ListRoundtables(r.Context(), status, 50)
	if err != nil {
		http.Error(w, `{"error":"查询列表失败"}`, http.StatusInternalServerError)
		return
	}

	items := make([]RoundtableListItem, 0, len(list))
	for _, rt := range list {
		var personas []string
		json.Unmarshal([]byte(rt.PersonasJSON), &personas)

		item := RoundtableListItem{
			ID:        rt.ID,
			Topic:     rt.Topic,
			Personas:  personas,
			MaxRounds: rt.MaxRounds,
			Status:    rt.Status,
			CreatedAt: rt.CreatedAt.Format(time.RFC3339),
		}
		if rt.FinishedAt != nil {
			s := rt.FinishedAt.Format(time.RFC3339)
			item.FinishedAt = &s
		}
		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(items)
}

func generateID() string {
	return fmt.Sprintf("rt_%d", time.Now().UnixNano())
}
