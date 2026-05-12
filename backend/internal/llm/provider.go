// Package llm 提供 TalkAboutIt 的 LLM Provider 抽象与实现。
package llm

import "context"

// ChatMessage 代表对话中的一条消息。
type ChatMessage struct {
	Role    string `json:"role"`    // system / user / assistant
	Content string `json:"content"` // 消息内容
}

// ChatRequest 是调用 Provider.Chat 时的请求参数。
type ChatRequest struct {
	Messages    []ChatMessage // 对话历史
	Model       string        // 覆盖默认模型（可选）
	MaxTokens   int           // 最大生成 token 数
	Temperature float64       // 采样温度
	Stream      bool          // 是否流式返回
}

// ChatChunk 是流式返回的单个数据块。
type ChatChunk struct {
	Content string // 增量文本内容
	Done    bool   // 是否为结束标记
	Error   error  // 传输过程中的错误（如有）
}

// Provider 是 LLM Gateway 的统一接口。
type Provider interface {
	// Chat 发起对话，返回只读 channel 用于流式读取。
	// ctx 取消时会中断请求并关闭 channel。
	Chat(ctx context.Context, req ChatRequest) (<-chan ChatChunk, error)

	// Name 返回 Provider 名称（如 "deepseek"）。
	Name() string

	// Model 返回当前使用的模型标识（如 "deepseek-chat"）。
	Model() string
}
