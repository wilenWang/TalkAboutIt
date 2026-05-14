// Package engine 提供 Roundtable 讨论的编排与执行能力。
package engine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/wilenwang/talkaboutit/internal/llm"
	"github.com/wilenwang/talkaboutit/internal/persona"
	"github.com/wilenwang/talkaboutit/internal/session"
)

// GenerateFunc 是 LLM 生成函数的签名（流式）。
// 返回只读 channel，逐块输出 ChatChunk；channel 关闭时表示生成结束。
type GenerateFunc func(ctx context.Context, p persona.Persona, topic string, peers []string, language string, convo *persona.ConversationContext) (<-chan llm.ChatChunk, error)

// OnEventFunc 是事件写入后的回调签名，用于广播到 SSE 订阅者。
type OnEventFunc func(evt session.Event)

// Engine 编排讨论流程。
type Engine struct {
	store    *session.Store
	loader   *persona.Loader
	generate GenerateFunc
	provider llm.Provider // 可选：真实 LLM provider（用于动态语言模式）
	onEvent  OnEventFunc
}

// NewEngine 创建 Engine 实例。
// 若 generate 为 nil，则使用 DefaultMockGenerate 作为 fallback。
func NewEngine(store *session.Store, loader *persona.Loader, generate GenerateFunc) *Engine {
	if generate == nil {
		generate = DefaultMockGenerate
	}
	return &Engine{
		store:    store,
		loader:   loader,
		generate: generate,
	}
}

// NewEngineWithProvider 创建带真实 LLM provider 的 Engine（支持动态语言模式）。
func NewEngineWithProvider(store *session.Store, loader *persona.Loader, provider llm.Provider) *Engine {
	return &Engine{
		store:    store,
		loader:   loader,
		provider: provider,
	}
}

// SetOnEvent 设置事件回调。
func (e *Engine) SetOnEvent(fn OnEventFunc) {
	e.onEvent = fn
}

// DefaultMockGenerate 是默认的 mock LLM：返回 persona 的 opening_line 作为单一块。
func DefaultMockGenerate(ctx context.Context, p persona.Persona, topic string, peers []string, language string, convo *persona.ConversationContext) (<-chan llm.ChatChunk, error) {
	ch := make(chan llm.ChatChunk, 1)
	ch <- llm.ChatChunk{Content: p.Examples.OpeningLine}
	close(ch)
	return ch, nil
}

// LLMGenerate 包装 llm.Provider 为 Engine 可用的 GenerateFunc。
func LLMGenerate(provider llm.Provider) GenerateFunc {
	return func(ctx context.Context, p persona.Persona, topic string, peers []string, language string, convo *persona.ConversationContext) (<-chan llm.ChatChunk, error) {
		return provider.Chat(ctx, convo.BuildChatRequest(512, 0.8))
	}
}

// sleepBetweenEvents 在事件之间短暂停顿，使 SSE 客户端有时间接收。
func sleepBetweenEvents() {
	time.Sleep(50 * time.Millisecond)
}

func (e *Engine) broadcast(evt session.Event) {
	if e.onEvent != nil {
		e.onEvent(evt)
	}
}

// Run 驱动指定 roundtable 的讨论流程。
func (e *Engine) Run(ctx context.Context, tableID string) error {
	rt, err := e.store.GetRoundtable(ctx, tableID)
	if err != nil {
		return fmt.Errorf("获取 roundtable 失败: %w", err)
	}

	if rt.Status != "running" {
		return fmt.Errorf("roundtable 状态为 %s，无法启动", rt.Status)
	}

	var personaIDs []string
	if err := json.Unmarshal([]byte(rt.PersonasJSON), &personaIDs); err != nil {
		return fmt.Errorf("解析 personas_json 失败: %w", err)
	}
	if len(personaIDs) == 0 {
		return fmt.Errorf("roundtable 缺少 persona")
	}

	personas := make([]persona.Persona, 0, len(personaIDs))
	for _, id := range personaIDs {
		p, err := e.loader.LoadOne(id)
		if err != nil {
			return fmt.Errorf("加载 persona %s 失败: %w", id, err)
		}
		personas = append(personas, p)
	}

	contexts := make(map[string]*persona.ConversationContext, len(personas))
	states := make(map[string]*persona.PerPersonaState)
	for _, p := range personas {
		state := &persona.PerPersonaState{}
		states[p.ID] = state
		contexts[p.ID] = persona.NewConversationContext(p.ID, llm.BuildStaticSystemPrompt(p), state)
	}

	// 动态确定 generate 函数：优先使用 provider，回退到固定 generate
	generate := e.generate
	if e.provider != nil {
		generate = LLMGenerate(e.provider)
	} else if generate == nil {
		generate = DefaultMockGenerate
	}

	// stream_start 事件
	r0 := 0
	si0 := 0
	pid0 := personas[0].ID
	evt, err := e.store.AddEvent(ctx, tableID, "stream_start", &r0, &si0, &pid0, nil, map[string]interface{}{
		"total_rounds": rt.MaxRounds,
	})
	if err != nil {
		return fmt.Errorf("发送 stream_start 失败: %w", err)
	}
	e.broadcast(*evt)
	sleepBetweenEvents()

	totalMessages := 0

	for round := 1; round <= rt.MaxRounds; round++ {
		// round_start
		r := round
		evt, err := e.store.AddEvent(ctx, tableID, "round_start", &r, nil, nil, nil, map[string]interface{}{
			"round":        round,
			"total_rounds": rt.MaxRounds,
		})
		if err != nil {
			return fmt.Errorf("发送 round_start 失败: %w", err)
		}
		e.broadcast(*evt)
		sleepBetweenEvents()

		for i, p := range personas {
			si := i
			pid := p.ID

			// speaking 事件
			evt, err := e.store.AddEvent(ctx, tableID, "speaking", &r, &si, &pid, nil, map[string]interface{}{
				"round":         round,
				"speaker_index": i,
				"persona_id":    p.ID,
				"persona_name":  p.Name,
				"avatar":        p.Avatar,
			})
			if err != nil {
				return fmt.Errorf("发送 speaking 失败: %w", err)
			}
			e.broadcast(*evt)
			sleepBetweenEvents()

			// 每次发言创建子 context，防止提前退出时 goroutine 泄漏
			speakCtx, cancel := context.WithCancel(ctx)

			// 生成内容（流式）
			peers := make([]string, 0, len(personas)-1)
			for _, peer := range personas {
				if peer.ID != p.ID {
					peers = append(peers, peer.Name)
				}
			}
			convo := contexts[p.ID]
			convo.Round = round
			convo.Append("user", "", llm.BuildDynamicContext(p, rt.Topic, peers, round, rt.Language, convo.State))
			convo.Truncate(24)

			chunkCh, err := generate(speakCtx, p, rt.Topic, peers, rt.Language, convo)
			if err != nil {
				cancel()
				// 根据错误类型决定行为
				var perr *llm.ProviderError
				if errors.As(err, &perr) {
					if perr.Recoverable {
						// 可恢复：发送 message_aborted 事件，跳过当前发言，继续下一人
						evt, _ := e.store.AddEvent(ctx, tableID, "message_aborted", &r, &si, &pid, nil, map[string]interface{}{
							"persona_id":      p.ID,
							"round":           round,
							"speaker_index":   i,
							"partial_content": "",
							"code":            perr.Code,
							"error":           perr.UserMessage,
						})
						if evt != nil {
							e.broadcast(*evt)
						}
						continue
					}
					// 不可恢复：终止 roundtable
					return fmt.Errorf("生成内容失败: %w", err)
				}
				return fmt.Errorf("生成内容失败: %w", err)
			}

			// 累积完整内容，同时逐 chunk 推送
			var fullContent strings.Builder
			var done bool
			var chunkErr error
			for chunk := range chunkCh {
				if chunk.Error != nil {
					chunkErr = chunk.Error
					break
				}
				if chunk.Done {
					done = true
					break
				}
				if chunk.Content != "" {
					fullContent.WriteString(chunk.Content)

					evt, err = e.store.AddEvent(ctx, tableID, "message_chunk", &r, &si, &pid, nil, map[string]interface{}{
						"round":         round,
						"speaker_index": i,
						"persona_id":    p.ID,
						"chunk":         chunk.Content,
					})
					if err != nil {
						cancel()
						return fmt.Errorf("发送 message_chunk 失败: %w", err)
					}
					e.broadcast(*evt)
					sleepBetweenEvents()
				}
			}
			cancel() // 确保 goroutine 被清理

			// 处理流式过程中的错误
			if chunkErr != nil {
				var perr *llm.ProviderError
				if errors.As(chunkErr, &perr) {
					if perr.Recoverable {
						// 可恢复：发送 message_aborted 事件，跳过当前发言，继续下一人
						evt, _ := e.store.AddEvent(ctx, tableID, "message_aborted", &r, &si, &pid, nil, map[string]interface{}{
							"persona_id":      p.ID,
							"round":           round,
							"speaker_index":   i,
							"partial_content": fullContent.String(),
							"code":            perr.Code,
							"error":           perr.UserMessage,
						})
						if evt != nil {
							e.broadcast(*evt)
						}
						continue
					}
				}
				return fmt.Errorf("生成内容时出错: %w", chunkErr)
			}

			// 若未收到 Done 标记但 channel 已关闭，也视为正常完成
			_ = done

			content := fullContent.String()
			convo.Append("assistant", p.ID, content)

			// message_done 事件
			msgID := fmt.Sprintf("%s_r%d_s%d", tableID, round, i)
			evt, err = e.store.AddEvent(ctx, tableID, "message_done", &r, &si, &pid, &msgID, map[string]interface{}{
				"round":         round,
				"speaker_index": i,
				"persona_id":    p.ID,
				"persona_name":  p.Name,
				"avatar":        p.Avatar,
				"content":       content,
				"message_id":    msgID,
			})
			if err != nil {
				return fmt.Errorf("发送 message_done 失败: %w", err)
			}
			e.broadcast(*evt)
			totalMessages++

			if st, ok := states[p.ID]; ok {
				st.RecordArgument(persona.ExtractArgument(content))
			}
			for _, peer := range personas {
				if peer.ID == p.ID {
					continue
				}
				contexts[peer.ID].Append("assistant", p.ID, content)
				contexts[peer.ID].Truncate(24)
			}
		}

		// round_end
		evt, err = e.store.AddEvent(ctx, tableID, "round_end", &r, nil, nil, nil, map[string]interface{}{
			"round":        round,
			"total_rounds": rt.MaxRounds,
		})
		if err != nil {
			return fmt.Errorf("发送 round_end 失败: %w", err)
		}
		e.broadcast(*evt)
		sleepBetweenEvents()
	}

	// stream_done
	evt, err = e.store.AddEvent(ctx, tableID, "stream_done", nil, nil, nil, nil, map[string]interface{}{
		"total_rounds":   rt.MaxRounds,
		"total_messages": totalMessages,
		"finished_at":    time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return fmt.Errorf("发送 stream_done 失败: %w", err)
	}
	e.broadcast(*evt)

	// 更新状态为 completed
	if err := e.store.UpdateStatus(ctx, tableID, "completed"); err != nil {
		return fmt.Errorf("更新 completed 状态失败: %w", err)
	}

	return nil
}
