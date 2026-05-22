// Package api 提供 TalkAboutIt 的 HTTP API 路由与处理器。
package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/wilenwang/talkaboutit/internal/engine"
	"github.com/wilenwang/talkaboutit/internal/persona"
	"github.com/wilenwang/talkaboutit/internal/session"
)

// Handler 持有所有 API 依赖。
type Handler struct {
	loader persona.Repository
	store  *session.Store
	bus    *EventBus
	engine *engine.Engine
}

// NewHandler 创建 API Handler。
func NewHandler(loader persona.Repository, store *session.Store, eng *engine.Engine) *Handler {
	h := &Handler{
		loader: loader,
		store:  store,
		bus:    NewEventBus(store),
		engine: eng,
	}
	// 将 engine 的事件回调绑定到 bus
	eng.SetOnEvent(func(evt session.Event) {
		h.bus.Publish(evt.RoundtableID, evt)
	})
	return h
}

// corsMiddleware 为所有 API 路由统一添加 CORS 头。
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Last-Event-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// RegisterRoutes 向 http.ServeMux 注册所有 API 路由。
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/personas", corsMiddleware(h.ListPersonas))
	mux.HandleFunc("GET /api/v1/personas/{id}", corsMiddleware(h.GetPersona))
	mux.HandleFunc("POST /api/v1/personas", corsMiddleware(h.CreatePersona))
	mux.HandleFunc("PUT /api/v1/personas/{id}", corsMiddleware(h.UpdatePersona))
	mux.HandleFunc("DELETE /api/v1/personas/{id}", corsMiddleware(h.DeletePersona))

	mux.HandleFunc("POST /api/v1/roundtables", corsMiddleware(h.CreateRoundtable))
	mux.HandleFunc("POST /api/v1/roundtables/{id}/start", corsMiddleware(h.StartRoundtable))
	mux.HandleFunc("GET /api/v1/roundtables", corsMiddleware(h.ListRoundtables))
	mux.HandleFunc("GET /api/v1/roundtables/{id}", corsMiddleware(h.GetRoundtable))
	mux.HandleFunc("GET /api/v1/roundtables/{id}/events", corsMiddleware(h.SSEHandler))
}

// PersonaSummary 是列表接口返回的摘要结构。
type PersonaSummary struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Avatar      string   `json:"avatar"`
	RoleTitle   string   `json:"role_title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Archetype   string   `json:"archetype"`
}

// ListPersonas 返回所有预置 Persona 的摘要列表。
func (h *Handler) ListPersonas(w http.ResponseWriter, r *http.Request) {
	personas, err := h.loader.LoadAll()
	if err != nil {
		http.Error(w, `{"error":"加载 persona 失败"}`, http.StatusInternalServerError)
		return
	}

	summaries := make([]PersonaSummary, 0, len(personas))
	for _, p := range personas {
		summaries = append(summaries, PersonaSummary{
			ID:          p.ID,
			Name:        p.Name,
			DisplayName: p.DisplayName,
			Avatar:      p.Avatar,
			RoleTitle:   p.RoleTitle,
			Description: p.Description,
			Tags:        p.Tags,
			Archetype:   p.Archetype,
		})
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(summaries)
}

// GetPersona 返回指定 ID 的完整 Persona Schema。
func (h *Handler) GetPersona(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"缺少 persona ID"}`, http.StatusBadRequest)
		return
	}

	p, err := h.loader.LoadOne(id)
	if err != nil {
		http.Error(w, `{"error":"persona 不存在"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(p)
}

// CreatePersona 创建新的 Persona。
func (h *Handler) CreatePersona(w http.ResponseWriter, r *http.Request) {
	var p persona.Persona
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, `{"error":"请求体解析失败"}`, http.StatusBadRequest)
		return
	}
	if p.ID == "" {
		http.Error(w, `{"error":"id 不能为空"}`, http.StatusBadRequest)
		return
	}
	// 检查是否已存在
	if _, err := h.loader.LoadOne(p.ID); err == nil {
		http.Error(w, `{"error":"persona 已存在"}`, http.StatusConflict)
		return
	}
	if err := h.loader.Save(p); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

// UpdatePersona 更新已有 Persona。
func (h *Handler) UpdatePersona(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"缺少 persona ID"}`, http.StatusBadRequest)
		return
	}
	var p persona.Persona
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, `{"error":"请求体解析失败"}`, http.StatusBadRequest)
		return
	}
	if p.ID != id {
		http.Error(w, `{"error":"URL ID 与请求体 ID 不一致"}`, http.StatusBadRequest)
		return
	}
	if _, err := h.loader.LoadOne(id); err != nil {
		http.Error(w, `{"error":"persona 不存在"}`, http.StatusNotFound)
		return
	}
	if err := h.loader.Save(p); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(p)
}

// DeletePersona 删除指定 Persona。
func (h *Handler) DeletePersona(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, `{"error":"缺少 persona ID"}`, http.StatusBadRequest)
		return
	}
	if _, err := h.loader.LoadOne(id); err != nil {
		http.Error(w, `{"error":"persona 不存在"}`, http.StatusNotFound)
		return
	}
	if err := h.loader.Delete(id); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
