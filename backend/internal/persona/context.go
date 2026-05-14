package persona

import "github.com/wilenwang/talkaboutit/internal/llm"

// ConversationContext 维护单个 persona 视角下的完整对话消息序列。
type ConversationContext struct {
	PersonaID string
	Messages  []llm.ChatMessage
	State     *PerPersonaState
	Round     int
}

// NewConversationContext 创建一个带固定 system prompt 的 persona 对话上下文。
func NewConversationContext(personaID string, systemPrompt string, state *PerPersonaState) *ConversationContext {
	if state == nil {
		state = &PerPersonaState{}
	}

	ctx := &ConversationContext{
		PersonaID: personaID,
		State:     state,
	}
	ctx.Append("system", "", systemPrompt)
	return ctx
}

// Append 在上下文末尾追加一条消息。
func (c *ConversationContext) Append(role, name, content string) {
	c.Messages = append(c.Messages, llm.ChatMessage{
		Role:    role,
		Name:    name,
		Content: content,
	})
}

// BuildChatRequest 构造发往 provider 的请求。
func (c *ConversationContext) BuildChatRequest(maxTokens int, temperature float64) llm.ChatRequest {
	messages := make([]llm.ChatMessage, len(c.Messages))
	copy(messages, c.Messages)
	return llm.ChatRequest{
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
		Stream:      true,
	}
}

// Truncate 以 system + 摘要 + 最近消息的方式裁剪上下文。
func (c *ConversationContext) Truncate(maxMessages int) {
	if maxMessages <= 0 || len(c.Messages) <= maxMessages {
		return
	}

	systemMsg := c.Messages[0]
	if maxMessages <= 1 {
		c.Messages = []llm.ChatMessage{systemMsg}
		return
	}

	tailCount := maxMessages - 2
	if tailCount < 0 {
		tailCount = 0
	}
	if tailCount > len(c.Messages)-1 {
		tailCount = len(c.Messages) - 1
	}

	start := len(c.Messages) - tailCount
	if start < 1 {
		start = 1
	}

	summarySource := c.Messages[1:start]
	tail := c.Messages[start:]

	next := make([]llm.ChatMessage, 0, 1+len(tail)+1)
	next = append(next, systemMsg)
	if len(summarySource) > 0 {
		next = append(next, llm.ChatMessage{
			Role:    "system",
			Name:    "",
			Content: summarizeMessages(summarySource),
		})
	}
	next = append(next, tail...)
	c.Messages = next
}

func summarizeMessages(messages []llm.ChatMessage) string {
	const maxSummaryItems = 6

	summary := "以下为更早轮次的摘要：\n"
	count := 0
	for _, message := range messages {
		if message.Content == "" {
			continue
		}
		summary += "- "
		if message.Name != "" {
			summary += message.Name + ": "
		}
		summary += ExtractArgument(message.Content) + "\n"
		count++
		if count >= maxSummaryItems {
			break
		}
	}
	if count == 0 {
		return "以下为更早轮次的摘要：此前有若干轮对话，请延续既有讨论，避免重复。"
	}
	return summary
}
