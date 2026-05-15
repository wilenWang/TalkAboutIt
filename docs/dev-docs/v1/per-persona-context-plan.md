# Per-Persona Context 实现计划

## 目标

把当前「每次发言都重建 `[system, user_with_history]`」的模式，重构为「1 个共享 Provider + 5 个独立 `ConversationContext`」。每个 persona 持有自己的完整消息序列，使静态 system prompt 长期稳定，后续轮次仅追加新的 user / assistant 消息，从而提高 DeepSeek 上下文缓存命中率。

## 当前实现摘要

- `internal/engine/engine.go`
  - `Run()` 为整个 roundtable 维护一个 `ConversationMemory`
  - 每轮调用 `LLMGenerate()` 时重新构造 `[system, user]`
  - `user` 首行直接拼接历史文本，导致 prompt 前缀每轮变化
- `internal/llm/prompt.go`
  - 只有 `BuildSystemPrompt()`，同时承载静态 persona 定义和动态轮次/语言/示例/去重信息
- `internal/llm/provider.go`
  - `ChatMessage` 只有 `Role` 和 `Content`
- `internal/llm/openai.go`
  - OpenAI 兼容消息未输出 `name`
  - 流式响应未解析 `usage` 中的缓存命中字段
- `internal/persona/state.go`
  - 已有 `PerPersonaState`，可直接作为 per-context 去重状态

## 改动文件清单

### 新增文件

- `backend/internal/persona/context.go`
  - 新增 `ConversationContext`
  - 负责 persona 级消息维护、请求构建、截断

### 修改文件

- `backend/internal/llm/provider.go`
  - `ChatMessage` 增加 `Name`
  - `ChatChunk` 增加可选 `Usage`
- `backend/internal/llm/openai.go`
  - 序列化 `name`
  - 解析流式 usage，包括 `prompt_cache_hit_tokens` / `prompt_cache_miss_tokens`
- `backend/internal/llm/prompt.go`
  - 拆分 `BuildStaticSystemPrompt()` 和 `BuildDynamicContext()`
  - 保留一个轻量兼容层 `BuildSystemPrompt()`，避免一次性打碎全部调用点
- `backend/internal/engine/engine.go`
  - 引入 `map[string]*persona.ConversationContext`
  - 调整 `GenerateFunc` 签名，使其以 context 为输入
  - 逐 persona 追加消息，不再依赖 `ConversationMemory.FormatForPersona()`
  - `ConversationMemory` 降级为兼容性辅助或移除
- `backend/internal/engine/engine_test.go`
  - 适配新的 `GenerateFunc` 签名
  - 增加 per-context 消息流验证
- `backend/internal/llm/prompt_test.go`
  - 更新为测试静态 prompt / 动态 context 两部分
- `backend/internal/llm/llm_test.go`
  - 更新 `BuildSystemPrompt()` 相关断言，拆分验证边界
- `backend/internal/llm/provider_test.go`
  - 增加 `name` / `usage` 编解码测试
- `backend/internal/persona/persona_test.go`
  - 新增 `ConversationContext` 行为测试

## 文件级改动概述

### 1. `internal/persona/context.go`

- 定义：
  - `PersonaID string`
  - `Messages []llm.ChatMessage`
  - `State *PerPersonaState`
  - `Round int`
- 方法：
  - `Append(role, name, content string)`
  - `BuildChatRequest(maxTokens int, temperature float64) llm.ChatRequest`
  - `Truncate(maxMessages int)`，先做基于消息数的滑动窗口
- 初始化约定：
  - 创建时先注入固定 `system`
  - 第 1 轮追加开场 `user`
  - 每轮结束后写入本 persona 的 `assistant`
  - 其他 persona 发言同步追加为 `assistant(name=otherPersonaID)`

### 2. `internal/llm/provider.go`

- `ChatMessage.Name` 用于标记发言人
- `ChatUsage` 结构保存缓存命中指标
- `ChatChunk.Usage` 承载 provider 流末尾返回的 usage

### 3. `internal/llm/openai.go`

- 请求中为每条消息按需输出 `name`
- SSE 响应支持解析：
  - `choices[].delta.content`
  - `choices[].finish_reason`
  - `usage.prompt_cache_hit_tokens`
  - `usage.prompt_cache_miss_tokens`
  - 其他常规 token 字段
- 兼容无 usage 的厂商响应

### 4. `internal/llm/prompt.go`

- `BuildStaticSystemPrompt(...)`
  - 固定包含：身份、立场、核心信念、表达方式、知识边界、互动规则、辩论目标、前置说明、固定发言规则 1-6
- `BuildDynamicContext(...)`
  - 包含：主题、参与者、轮次、语言规则 7、去重提示、示例、当轮任务说明
- 兼容函数：
  - `BuildSystemPrompt(...) = BuildStaticSystemPrompt(...) + BuildDynamicContext(...)`
  - 这样测试与重构可以分阶段迁移

### 5. `internal/engine/engine.go`

- `GenerateFunc` 改为：
  - 输入 persona、topic、peers、language、`*persona.ConversationContext`
  - 输出流式 `ChatChunk`
- `Run()` 流程改造：
  - roundtable 启动时为每个 persona 初始化独立 context
  - context 的第 0 条消息写入静态 system
  - 每轮轮到某 persona 发言时：
    - 构建动态 user 指令并追加到该 persona context
    - 可选执行 `Truncate()`
    - `provider.Chat(ctx, context.BuildChatRequest(...))`
    - 收集 assistant 内容，落库并写回当前 persona context
    - 同步把该发言追加到其他 persona context 作为 `assistant(name=speakerID)`
- `ConversationMemory` 不再作为主路径

## 数据流图

### 改动前

```text
Round N / Persona X
  -> ConversationMemory.FormatForPersona()
  -> 拼接一整段历史文本到 user
  -> BuildSystemPrompt(静态 + 动态全部混在 system)
  -> ChatRequest:
       [0] system: 全量 prompt（含 round / language / dedup）
       [1] user: 历史全文 + 当前任务
  -> Provider.Chat()
```

问题：

- `system` 每轮变化
- `user` 前缀每轮变化
- DeepSeek prompt cache 几乎无法稳定命中

### 改动后

```text
Roundtable Start
  -> 为 5 个 persona 各建一个 ConversationContext
  -> 每个 context:
       [0] system: 仅静态 persona 定义

Round N / Persona X
  -> Append user: 当轮动态任务 + 语言规则 + 去重提示
  -> Provider.Chat(messages for X)
  -> Append assistant(self): X 的本轮发言
  -> 对其他 4 个 context:
       Append assistant(name=X): X 的发言
```

收益：

- `system` 保持稳定
- 绝大部分消息前缀可复用
- 只有尾部 user / assistant 继续增长，符合上下文缓存使用方式

## 实施顺序

1. 新增 `persona/context.go`，落地 `ConversationContext`
2. 扩展 `llm/provider.go` / `llm/openai.go` 的消息与 usage 结构
3. 重构 `llm/prompt.go`，拆出静态 / 动态 prompt
4. 改造 `engine.Run()` 与 `LLMGenerate()`
5. 逐步迁移并修正测试
6. 运行 `go build ./...`
7. 运行 `go test ./...`

## 风险评估

### 风险 1: 消息角色语义变化影响模型表现

- 现方案要求“他人历史发言”以 `assistant(name=other)` 形式注入
- 某些兼容 provider 可能更习惯把非当前模型输出作为 `user`
- 应对：
  - 先按目标方案实现 `assistant(name=other)`，因为它最接近“多说话人聊天记录”
  - 通过测试保证编码层兼容，必要时后续再引入策略开关

### 风险 2: 截断策略过于粗糙导致辩论质量下降

- 需求写了“最早消息用 LLM 摘要替代”，但这会引入额外 provider 调用
- 当前阶段先实现保留 system + 最近窗口 + 本地文本摘要占位，避免把 P1 复杂度带进 P0 主链路

### 风险 3: `GenerateFunc` 签名变更影响测试和调用方

- `engine_test.go` 中有自定义 mock generator
- 应对：
  - 一次性替换所有调用点
  - 保留 mock 行为不变，只改输入参数

### 风险 4: usage 字段并非所有 OpenAI 兼容厂商都会返回

- DeepSeek 可能返回，其他厂商可能缺失
- 应对：
  - 所有 usage 解析字段都定义为可选
  - 缺失时不报错，不影响流式输出

## 验收标准

- `Run()` 不再依赖把完整历史文本拼到单条 user 消息
- 每个 persona 拥有独立 `ConversationContext`
- `ChatMessage` 支持 `Name`
- `BuildStaticSystemPrompt()` 与 `BuildDynamicContext()` 拆分完成
- `go build ./...` 通过
- `go test ./...` 通过
