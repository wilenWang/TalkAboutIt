// Package llm 提供 TalkAboutIt 的 LLM Provider 抽象与实现。
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OpenAIProvider 实现基于 OpenAI 兼容 API 的流式对话能力。
// 通过修改 baseURL 可支持 DeepSeek / GLM / Kimi / Qwen 等兼容厂商。
type OpenAIProvider struct {
	name    string
	model   string
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenAIProvider 创建 OpenAI 兼容 Provider 实例。
// client 若为 nil，则使用默认 http.Client（自动遵循 HTTPS_PROXY 等环境变量）。
func NewOpenAIProvider(name, model, apiKey, baseURL string, client *http.Client) *OpenAIProvider {
	if client == nil {
		// 标准库 http.DefaultTransport 自动读取 HTTPS_PROXY/HTTP_PROXY 环境变量
		client = &http.Client{Timeout: 60 * time.Second}
	}
	return &OpenAIProvider{
		name:    name,
		model:   model,
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  client,
	}
}

// Name 返回 Provider 名称。
func (o *OpenAIProvider) Name() string { return o.name }

// Model 返回当前模型标识。
func (o *OpenAIProvider) Model() string { return o.model }

// openAIChatRequest 是发往 /v1/chat/completions 的请求体。
type openAIChatRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	Stream      bool            `json:"stream"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// openAIChatResponse 是流式 SSE 中 data 行的 JSON 结构。
type openAIChatResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// splitSSE 是 bufio.Scanner 的 SplitFunc，按空行（\n\n）分割 SSE 事件。
func splitSSE(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if i := bytes.Index(data, []byte("\n\n")); i >= 0 {
		return i + 2, data[:i], nil
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

// sendChunk 尝试向 channel 发送 chunk，同时监听 ctx 取消。
func sendChunk(ctx context.Context, ch chan<- ChatChunk, chunk ChatChunk) bool {
	select {
	case ch <- chunk:
		return true
	case <-ctx.Done():
		select {
		case ch <- ChatChunk{Error: ctx.Err()}:
		default:
		}
		return false
	}
}

// Chat 发起流式对话，返回 channel 逐块读取生成内容。
func (o *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (<-chan ChatChunk, error) {
	// 构造请求体
	messages := make([]openAIMessage, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = openAIMessage{Role: m.Role, Content: m.Content}
	}
	model := o.model
	if req.Model != "" {
		model = req.Model
	}
	body := openAIChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      true,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("序列化请求体失败: %w", err)
	}

	// 构造 HTTP 请求
	chatURL := o.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, chatURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("构造请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	// 发送请求
	resp, err := o.client.Do(httpReq)
	if err != nil {
		// 区分代理/网络错误
		if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
			return nil, &ProviderError{
				Code:        ErrProviderTimeout.Code,
				Recoverable: true,
				UserMessage: ErrProviderTimeout.UserMessage,
				InternalError: err,
			}
		}
		return nil, fmt.Errorf("请求 LLM API 失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, NewProviderError(resp.StatusCode, string(bodyBytes), nil)
	}

	ch := make(chan ChatChunk, 8)

	// 在独立 goroutine 中解析 SSE 流
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		scanner.Split(splitSSE)

		for scanner.Scan() {
			select {
			case <-ctx.Done():
				sendChunk(ctx, ch, ChatChunk{Error: ctx.Err()})
				return
			default:
			}

			event := scanner.Text()
			lines := strings.Split(event, "\n")
			var dataParts []string
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				// 兼容 "data:" 和 "data: " 前缀
				if strings.HasPrefix(line, "data:") {
					data := strings.TrimPrefix(line, "data:")
					dataParts = append(dataParts, strings.TrimSpace(data))
				}
			}

			if len(dataParts) == 0 {
				continue
			}

			// 聚合多行 data
			data := strings.Join(dataParts, "\n")
			data = strings.TrimSpace(data)

			if data == "[DONE]" {
				sendChunk(ctx, ch, ChatChunk{Done: true})
				return
			}

			var streamResp openAIChatResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				sendChunk(ctx, ch, ChatChunk{Error: fmt.Errorf("解析 SSE JSON 失败: %w", err)})
				return
			}

			if len(streamResp.Choices) == 0 {
				continue
			}

			delta := streamResp.Choices[0].Delta.Content
			if delta != "" {
				if !sendChunk(ctx, ch, ChatChunk{Content: delta}) {
					return
				}
			}

			if streamResp.Choices[0].FinishReason != nil && *streamResp.Choices[0].FinishReason != "" {
				sendChunk(ctx, ch, ChatChunk{Done: true})
				return
			}
		}

		// EOF 时处理 scanner 错误
		if err := scanner.Err(); err != nil {
			sendChunk(ctx, ch, ChatChunk{Error: fmt.Errorf("读取 SSE 流失败: %w", err)})
		}
	}()

	return ch, nil
}
