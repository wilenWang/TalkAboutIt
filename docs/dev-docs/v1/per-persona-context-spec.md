# Per-Persona Context 技术规格

## 设计目标

- 让每个 persona 持有稳定增长的 `messages[]`
- 把静态 persona 定义固定在 `system` 消息中
- 把每轮变化信息收敛到新增 `user` 消息，而不是重写整段历史
- 为 DeepSeek / OpenAI 兼容接口保留 `name` 和缓存 usage 信息

## 新数据结构定义

### `internal/llm/provider.go`

```go
type ChatMessage struct {
    Role    string `json:"role"`
    Name    string `json:"name,omitempty"`
    Content string `json:"content"`
}

type ChatUsage struct {
    PromptTokens          int `json:"prompt_tokens,omitempty"`
    CompletionTokens      int `json:"completion_tokens,omitempty"`
    TotalTokens           int `json:"total_tokens,omitempty"`
    PromptCacheHitTokens  int `json:"prompt_cache_hit_tokens,omitempty"`
    PromptCacheMissTokens int `json:"prompt_cache_miss_tokens,omitempty"`
}

type ChatChunk struct {
    Content string
    Done    bool
    Error   error
    Usage   *ChatUsage
}
```

约定：

- `Name` 仅在有明确发言人标识时使用
- `Usage` 通常只在流末尾返回一次；中间 chunk 可为 `nil`

### `internal/persona/context.go`

```go
type ConversationContext struct {
    PersonaID string
    Messages  []llm.ChatMessage
    State     *PerPersonaState
    Round     int
}
```

职责：

- 保存某个 persona 视角下的完整消息序列
- 负责追加消息
- 负责构建最终 `ChatRequest`
- 负责执行窗口截断

## 新接口与函数签名

### `persona.NewConversationContext`

```go
func NewConversationContext(personaID string, systemPrompt string, state *PerPersonaState) *ConversationContext
```

输入约定：

- `personaID` 必须非空
- `systemPrompt` 会作为第一条 `system` 消息写入
- `state` 允许为 `nil`，内部会补默认值

输出约定：

- 返回的 context 至少包含 1 条消息：
  - `Messages[0] = {Role: "system", Content: systemPrompt}`

### `(*ConversationContext).Append`

```go
func (c *ConversationContext) Append(role, name, content string)
```

输入约定：

- `role` 允许值：`system` / `user` / `assistant`
- `name` 可为空
- `content` 空字符串也允许，便于保持流式逻辑简单

行为约定：

- 直接在 `Messages` 末尾追加 1 条消息

### `(*ConversationContext).BuildChatRequest`

```go
func (c *ConversationContext) BuildChatRequest(maxTokens int, temperature float64) llm.ChatRequest
```

输入约定：

- `maxTokens <= 0` 时由上层自行保证默认值；该函数不做 provider 级推断
- `temperature` 原样透传

输出约定：

- 返回：

```go
llm.ChatRequest{
    Messages:    c.Messages,
    MaxTokens:   maxTokens,
    Temperature: temperature,
    Stream:      true,
}
```

### `(*ConversationContext).Truncate`

```go
func (c *ConversationContext) Truncate(maxMessages int)
```

输入约定：

- `maxMessages <= 0` 时不做处理
- `maxMessages` 指系统消息之外允许保留的近期消息上限，或直接作为总消息上限，最终实现会在代码中固定一种语义并由测试锁定

行为约定：

- 始终保留首条 `system`
- 若超出窗口：
  - 保留最近若干条消息
  - 把被移除历史压缩成 1 条摘要消息插回 `system` 后
- 初版摘要不额外调用 LLM，采用本地字符串聚合摘要

## Prompt 构建规格

### `BuildStaticSystemPrompt`

```go
func BuildStaticSystemPrompt(p persona.Persona) string
```

必须包含：

- 身份
- 立场
- 核心信念
- 表达方式
- 知识边界
- 互动规则
- 辩论目标
- 前置说明
- 发言规则 1-6

必须不包含：

- 当前轮次
- 当前 topic
- 当前 peers
- 语言规则 7
- 去重提示
- 当轮动态任务

### `BuildDynamicContext`

```go
func BuildDynamicContext(
    p persona.Persona,
    topic string,
    peers []string,
    round int,
    language string,
    state *persona.PerPersonaState,
) string
```

必须包含：

- 当前主题
- 当前参与者
- 当前轮次
- 语言规则 7
- 去重提示
- 示例
- 当轮指令

轮次指令规则：

- `round == 1`
  - 输出开场任务，例如“请围绕主题发表开场观点”
- `round > 1`
  - 输出继续辩论任务，例如“请回应其他人的最新观点，并避免重复自己之前的论点”

### 兼容函数 `BuildSystemPrompt`

```go
func BuildSystemPrompt(
    p persona.Persona,
    topic string,
    peers []string,
    round int,
    language string,
    state *persona.PerPersonaState,
) string
```

行为定义：

- 返回 `BuildStaticSystemPrompt(p) + "\n\n" + BuildDynamicContext(...)`
- 仅作为兼容层，供旧测试与临时调用点过渡使用

## Engine 侧接口变更

### `GenerateFunc`

旧签名：

```go
type GenerateFunc func(
    ctx context.Context,
    p persona.Persona,
    topic string,
    peers []string,
    round int,
    memory *ConversationMemory,
) (<-chan llm.ChatChunk, error)
```

新签名：

```go
type GenerateFunc func(
    ctx context.Context,
    p persona.Persona,
    topic string,
    peers []string,
    language string,
    convo *persona.ConversationContext,
) (<-chan llm.ChatChunk, error)
```

变更原因：

- `round` 与历史已经体现在 `ConversationContext`
- `ConversationMemory` 不再是主输入
- `language` 仍由 roundtable 决定，需要显式传入

### `LLMGenerate`

```go
func LLMGenerate(provider llm.Provider) GenerateFunc
```

行为约定：

- 从 `convo.Round` 读取当前轮次
- 追加前由调用方准备好最新 user 消息
- `LLMGenerate` 只负责把 `convo.BuildChatRequest(...)` 发给 provider

## ConversationContext 消息追加约定

### 初始化

每个 persona context 初始为：

```text
[0] system     BuildStaticSystemPrompt(persona)
```

### 第 1 轮，persona X 发言前

```text
[1] user       BuildDynamicContext(... round=1 ...)
```

### persona X 发言完成后

```text
[2] assistant  name="" 或 personaID，自身回复正文
```

### 该轮同步给其他 persona Y

```text
append assistant name="persona-x" content="X 的发言"
```

说明：

- 为了保留多说话人信息，其他人的历史统一落成 `assistant(name=speakerID)`
- 当前 persona 自己的回复也可带 `name=personaID`，但不是强制

## OpenAI 兼容协议规格

### 请求消息序列化

```json
{
  "role": "assistant",
  "name": "steve-jobs",
  "content": "..."
}
```

规则：

- `name == ""` 时省略字段
- 其他请求字段维持现有行为

### 流式响应 usage 解析

兼容以下响应片段：

```json
{
  "choices": [...],
  "usage": {
    "prompt_tokens": 123,
    "completion_tokens": 45,
    "total_tokens": 168,
    "prompt_cache_hit_tokens": 100,
    "prompt_cache_miss_tokens": 23
  }
}
```

行为：

- 如 usage 存在，则在对应 `ChatChunk` 上挂载 `Usage`
- 不依赖 usage 存在才结束流

## 输入输出约定

### `Engine.Run`

输入：

- `tableID` 对应已 `running` 的 roundtable

输出：

- 持续写入：
  - `speaking`
  - `message_chunk`
  - `message_done`
  - `message_aborted`
  - `round_start`
  - `round_end`
  - `stream_start`
  - `stream_done`

内部状态变化：

- 每个 persona context 独立增长
- 每次 persona 产出后：
  - 当前 context 追加 self assistant
  - 其他 context 追加 cross-persona assistant
  - `PerPersonaState.RecordArgument()` 更新去重状态

## 缓存命中预期

### 改造前

- `system` 每轮变化
- `user` 首行历史文本每轮变化
- 高概率为 `prompt_cache_miss_tokens` 持续偏高

### 改造后

- `system` 完全稳定
- 大部分历史消息为 append-only
- 后续轮次预计表现为：
  - `prompt_cache_hit_tokens` 随轮次增长而增加
  - `prompt_cache_miss_tokens` 主要集中在新增尾部 user / assistant 消息

### 定性预期

- 第 1 轮命中率低，属正常
- 第 2 轮开始，同一 persona 的请求前缀应具备显著缓存复用条件
- 5 个 persona 间不会共享缓存前缀，因为 system 不同；缓存收益主要来自“同一 persona 跨轮次”

## 非目标

- 本次不引入额外摘要 LLM provider
- 本次不实现跨 persona 的共享上下文压缩
- 本次不修改 session schema 或前端事件协议
