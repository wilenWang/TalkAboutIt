// Package llm 提供 TalkAboutIt 的 LLM Provider 抽象与实现。
package llm

import (
	"context"
	"errors"
)

// AnthropicProvider 是 Anthropic Claude 的 Provider 占位实现。
// Phase 3.1 仅提供骨架，Phase 3.3 再完整接入 Messages API + SSE streaming。
type AnthropicProvider struct {
	name  string
	model string
}

// NewAnthropicProvider 创建 Anthropic Provider 占位实例。
func NewAnthropicProvider(name, model string) *AnthropicProvider {
	return &AnthropicProvider{name: name, model: model}
}

// Name 返回 Provider 名称。
func (a *AnthropicProvider) Name() string { return a.name }

// Model 返回当前模型标识。
func (a *AnthropicProvider) Model() string { return a.model }

// Chat 返回占位错误，提示尚未接入。
func (a *AnthropicProvider) Chat(ctx context.Context, req ChatRequest) (<-chan ChatChunk, error) {
	return nil, errors.New("Anthropic 尚未接入")
}
