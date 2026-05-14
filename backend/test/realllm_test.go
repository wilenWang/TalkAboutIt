//go:build realllm

// Package test 提供 TalkAboutIt 的真实 LLM 冒烟测试。
// 使用 go test -tags=realllm -run TestRealLLM ./test/ -v -timeout 300s 运行。
package test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wilenwang/talkaboutit/internal/api"
	"github.com/wilenwang/talkaboutit/internal/config"
	"github.com/wilenwang/talkaboutit/internal/engine"
	"github.com/wilenwang/talkaboutit/internal/llm"
	"github.com/wilenwang/talkaboutit/internal/persona"
	"github.com/wilenwang/talkaboutit/internal/session"
)

// setupRealLLMServer 创建使用真实 LLM 的测试服务器。
// 使用 DefaultConfig（DeepSeek + DEEPSEEK_API_KEY 环境变量）。
func setupRealLLMServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	cfg := config.DefaultConfig()
	cfg.Personas.Dir = "../personas" // test/ 目录相对 backend/

	prov, err := llm.NewProvider(*cfg)
	if err != nil {
		t.Skipf("跳过真实 LLM 测试：创建 provider 失败 (%v)", err)
	}

	loader := persona.NewLoader(cfg.Personas.Dir)
	dbDir := t.TempDir()
	store, err := session.NewStore(dbDir + "/realllm_test.db")
	if err != nil {
		t.Fatalf("NewStore failed: %v", err)
	}

	eng := engine.NewEngineWithProvider(store, loader, prov)
	h := api.NewHandler(loader, store, eng)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	srv := httptest.NewServer(mux)
	cleanup := func() {
		srv.Close()
		store.Close()
	}
	return srv, cleanup
}

// sseEvent 表示解析后的 SSE 事件。
type sseEvent struct {
	ID    string
	Event string
	Data  string
}

// collectSSEEvents 收集 SSE 事件直到 stream_done 或超时。
func collectSSEEventsReal(t *testing.T, serverURL, roundtableID string, timeout time.Duration) []sseEvent {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	url := serverURL + "/api/v1/roundtables/" + roundtableID + "/events"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("构造 SSE 请求失败: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("SSE 连接失败: %v", err)
	}
	defer resp.Body.Close()

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
			return events
		}
		line = strings.TrimRight(line, "\n")

		if line == "" {
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
	}
}

// TestRealLLM_JobsVsMusk 真实 LLM 冒烟测试：Steve Jobs vs Elon Musk 2 轮辩论。
func TestRealLLM_JobsVsMusk(t *testing.T) {
	srv, cleanup := setupRealLLMServer(t)
	defer cleanup()

	// 创建 roundtable：Jobs vs Musk，2 轮，英文话题
	body := `{"topic":"Should AI development be slowed down for safety reasons?","personas":["steve-jobs","elon-musk"],"max_rounds":2,"language":"en-US"}`
	resp, err := http.Post(srv.URL+"/api/v1/roundtables", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("创建 roundtable 失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Fatalf("创建 roundtable 失败：状态码 %d，响应: %s", resp.StatusCode, string(bodyBytes))
	}

	var createResult struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&createResult); err != nil {
		t.Fatalf("解析创建响应失败: %v", err)
	}
	if createResult.ID == "" {
		t.Fatalf("创建 roundtable 失败：返回空 ID")
	}
	rtID := createResult.ID
	t.Logf("✅ 创建 roundtable: %s", rtID)

	// 启动
	resp2, err := http.Post(srv.URL+"/api/v1/roundtables/"+rtID+"/start", "application/json", nil)
	if err != nil {
		t.Fatalf("启动 roundtable 失败: %v", err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("启动 roundtable 失败：状态码 %d", resp2.StatusCode)
	}

	// 收集 SSE 事件（等待 180s）
	t.Log("⏳ 等待真实 LLM 辩论（最多 180s）...")
	startTime := time.Now()
	events := collectSSEEventsReal(t, srv.URL, rtID, 180*time.Second)
	elapsed := time.Since(startTime)
	t.Logf("📊 收到 %d 个 SSE 事件（耗时 %v）", len(events), elapsed.Round(time.Second))

	// 验证事件序列
	var (
		foundStreamStart, foundStreamDone bool
		speakingCount                     int
		messageDoneCount                  int
		messageContents                   []string
	)

	for _, evt := range events {
		switch evt.Event {
		case "stream_start":
			foundStreamStart = true
		case "stream_done":
			foundStreamDone = true
		case "speaking":
			speakingCount++
		case "message_done":
			messageDoneCount++
			var payload struct {
				Content     string `json:"content"`
				PersonaName string `json:"persona_name"`
				Round       int    `json:"round"`
			}
			if err := json.Unmarshal([]byte(evt.Data), &payload); err == nil {
				content := strings.TrimSpace(payload.Content)
				if content != "" {
					messageContents = append(messageContents,
						fmt.Sprintf("[R%d] %s: %s", payload.Round, payload.PersonaName, truncate(content, 150)))
				}
			}
		}
	}

	// 断言
	if !foundStreamStart {
		t.Error("❌ 缺少 stream_start 事件")
	}
	if !foundStreamDone {
		t.Error("❌ 缺少 stream_done 事件")
	}
	if speakingCount < 2 {
		t.Errorf("❌ speaking 事件数不足：期望 ≥2，得到 %d", speakingCount)
	}
	if messageDoneCount < 2 {
		t.Errorf("❌ message_done 事件数不足：期望 ≥2，得到 %d", messageDoneCount)
	}

	// 打印对话内容
	t.Log("═══ 辩论内容 ═══")
	for _, mc := range messageContents {
		t.Log(mc)
	}

	// 质量检查：每条消息不能为空
	for i, mc := range messageContents {
		if strings.TrimSpace(mc) == "" {
			t.Errorf("❌ 消息 #%d 为空", i)
		}
	}

	// 检查是否有两方参与
	var hasJobs, hasMusk bool
	for _, mc := range messageContents {
		if strings.Contains(mc, "Steve Jobs") {
			hasJobs = true
		}
		if strings.Contains(mc, "Elon Musk") {
			hasMusk = true
		}
	}
	if !hasJobs {
		t.Error("❌ Steve Jobs 没有发言")
	}
	if !hasMusk {
		t.Error("❌ Elon Musk 没有发言")
	}

	t.Logf("✅ 真实 LLM 冒烟测试通过：%d 条消息，耗时 %v", messageDoneCount, elapsed.Round(time.Second))
}

func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}
