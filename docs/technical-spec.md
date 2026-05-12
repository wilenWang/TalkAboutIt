# TalkAboutIt 技术方案

> **版本:** v1.0 | **日期:** 2026-05-10 | **状态:** Draft

---

## 目录

- [Part 1: 后端技术方案](#part-1-后端技术方案)
- [Part 2: 前端技术方案](#part-2-前端技术方案)
- [Part 3: 前后端联调方案](#part-3-前后端联调方案)

---

# Part 1: 后端技术方案

## 1.1 技术选型

| 组件 | 选型 | 版本 | 理由 |
|------|------|------|------|
| 语言 | Go | ≥1.22 | 用户主语言，性能好，并发原生支持 |
| HTTP Router | `chi` | v5 | 轻量、惯用 Go、兼容 `net/http`、中间件链 |
| WebSocket | `gorilla/websocket` | v1.5 | 最成熟的 Go WS 实现，支持逐帧读写 |
| SQLite | `modernc.org/sqlite` | v1 | 纯 Go 实现，零 CGo 依赖，单文件部署 |
| YAML | `gopkg.in/yaml.v3` | v3 | Go 生态标准 YAML 库 |
| HTTP Client | `net/http` | stdlib | 调用 LLM API 用标准库，不引入额外依赖 |

**版本锁定原则:** 所有依赖在 `go.mod` 中精确锁定版本，`go.sum` 进 Git。

---

## 1.2 架构分层

```
┌─────────────────────────────────────────────┐
│                  API Layer                   │
│  chi Router → Handler → Request/Response    │
├─────────────────────────────────────────────┤
│               Engine Layer                   │
│  Roundtable Orchestrator (Round-Robin)      │
├──────────────┬──────────────────────────────┤
│ Persona Mgr  │        LLM Gateway           │
│ Loader +     │  Provider Interface          │
│ Prompt Build │  ├─ OpenAIProvider           │
│              │  └─ AnthropicProvider        │
├──────────────┴──────────────────────────────┤
│              Session Store                   │
│           SQLite (modernc.org)              │
└─────────────────────────────────────────────┘
```

**调用方向：** API → Engine → { Persona Manager, LLM Gateway → External APIs }
**依赖注入：** `main.go` 创建所有组件，通过构造函数注入到 Handler。

---

## 1.3 配置设计

**文件:** `backend/config.yaml`

```yaml
server:
  port: 8080
  host: localhost

llm:
  default: deepseek
  providers:
    deepseek:
      type: openai
      base_url: https://api.deepseek.com/v1
      api_key: ${DEEPSEEK_API_KEY}
      model: deepseek-chat
    openai:
      type: openai
      api_key: ${OPENAI_API_KEY}
      model: gpt-4o
    claude:
      type: anthropic
      api_key: ${ANTHROPIC_API_KEY}
      model: claude-sonnet-4-20250514

personas:
  dir: personas

session:
  db_path: data/sessions.db
  max_rounds: 3
```

**环境变量注入:**
- 每个 provider 的 `api_key` 优先从环境变量 `LLM_{NAME}_API_KEY` 读取
- 其次从 `config.yaml` 直接读取（不建议放明文 key）
- 启动时未配置 key 的 provider 自动跳过（不报错，运行时按需校验）

---

## 1.4 API 设计

### 基础信息

| 项目 | 值 |
|------|-----|
| Base URL | `http://localhost:8080` |
| Content-Type | `application/json` |
| 字符编码 | UTF-8 |

### 1.4.1 Persona API

#### `GET /api/personas`

获取所有可用人物列表。

**Response 200:**
```json
{
  "personas": [
    {
      "name": "Steve Jobs",
      "description": "Apple 联合创始人...",
      "personality": "极致完美主义者...",
      "model": "claude",
      "avatar": "🍎"
    }
  ]
}
```

#### `GET /api/personas/{name}`

获取指定人物详细信息（含完整的系统 prompt 预览）。

**Response 200:**
```json
{
  "name": "Steve Jobs",
  "description": "Apple 联合创始人...",
  "personality": "极致完美主义者...",
  "scenario": "你正在参加圆桌讨论...",
  "first_mes": "Let me just say one thing...",
  "model": "claude",
  "avatar": "🍎",
  "system_prompt": "你是 Steve Jobs。Apple 联合创始人...\n\n## 你的性格\n..."
}
```

### 1.4.2 Roundtable API

#### `POST /api/roundtables`

创建新的圆桌会议。

**Request:**
```json
{
  "topic": "AI 会取代程序员吗？",
  "personas": ["Steve Jobs", "Elon Musk", "Naval Ravikant"],
  "max_rounds": 3
}
```

**Response 201:**
```json
{
  "id": "rt_abc123",
  "topic": "AI 会取代程序员吗？",
  "personas": ["Steve Jobs", "Elon Musk", "Naval Ravikant"],
  "max_rounds": 3,
  "status": "pending",
  "created_at": "2026-05-10T12:00:00Z"
}
```

#### `GET /api/roundtables/{id}`

获取圆桌详情（含所有对话记录）。

**Response 200:**
```json
{
  "id": "rt_abc123",
  "topic": "AI 会取代程序员吗？",
  "personas": ["Steve Jobs", "Elon Musk", "Naval Ravikant"],
  "max_rounds": 3,
  "status": "done",
  "created_at": "2026-05-10T12:00:00Z",
  "messages": [
    {
      "id": 1,
      "round": 1,
      "persona": "Steve Jobs",
      "content": "Look, AI is just a tool...",
      "timestamp": "2026-05-10T12:00:05Z"
    }
  ]
}
```

#### `POST /api/roundtables/{id}/start`

启动讨论（异步）。

**Response 202:**
```json
{
  "id": "rt_abc123",
  "status": "running",
  "message": "Discussion started. Connect via WebSocket for live updates."
}
```

**Errors:**
- `404` — roundtable 不存在
- `409` — 已在运行中
- `400` — 未选择任何 persona

#### `DELETE /api/roundtables/{id}`

删除圆桌及其所有消息。

**Response 204:** No content.

### 1.4.3 WebSocket

#### `GET /api/roundtables/{id}/ws`

实时推送讨论消息流。

**连接:**
```
ws://localhost:8080/api/roundtables/rt_abc123/ws
```

**服务端 → 客户端消息格式:**

```json
// 1. 讨论开始
{
  "type": "discussion_started",
  "roundtable_id": "rt_abc123",
  "topic": "AI 会取代程序员吗？",
  "personas": ["Steve Jobs", "Elon Musk", "Naval Ravikant"],
  "max_rounds": 3
}

// 2. 新一轮开始
{
  "type": "round_start",
  "round": 1,
  "total_rounds": 3
}

// 3. 某人物开始发言
{
  "type": "speaking",
  "persona": "Steve Jobs",
  "avatar": "🍎"
}

// 4. 消息内容流式推送（打字机效果，可多次）
{
  "type": "message_chunk",
  "persona": "Steve Jobs",
  "content": "Look, "
}

// 5. 某人物发言完毕
{
  "type": "message_done",
  "persona": "Steve Jobs",
  "content": "Look, AI is just a tool...",
  "round": 1,
  "timestamp": "2026-05-10T12:00:05Z"
}

// 6. 讨论结束
{
  "type": "discussion_done",
  "total_messages": 9
}

// 7. 错误
{
  "type": "error",
  "message": "LLM API timeout",
  "persona": "Elon Musk"
}
```

**心跳:** 每 30 秒发送 `{"type":"ping"}`，客户端收到后回复 `{"type":"pong"}`。

**断线重连与消息去重:**

```json
// 每条消息携带唯一 sequence 号
{"type":"message_done","seq":5,"persona":"Steve Jobs","content":"...","round":1}

// 客户端重连时发送已收到的最大 seq
// 服务端仅推送 seq > last_seen 的消息
```

**客户端重连流程:**
1. WS 断开 → 3 次指数退避重试（1s/3s/5s）
2. 重连成功 → 发送 `{"type":"sync","last_seq":5}`
3. 服务端推送 `last_seq` 之后的所有消息 + 当前状态
4. 前端去重：按 `seq` 判断，不渲染已存在的消息

**状态恢复:**
- 重连后同时 GET `/api/roundtables/{id}` 补全可能遗漏的消息
- 两条路径去重合并，保证 UI 显示完整且无重复

---

## 1.5 角色卡（Character Card）格式

**文件位置:** `backend/personas/{name}.json`

**完整 Schema:**
```json
{
  "name": "string (required)",
  "description": "string (required) - 1-3句话身份描述",
  "personality": "string (required) - 性格、思维模式、价值观",
  "scenario": "string (optional) - 当前场景描述",
  "first_mes": "string (optional) - 开场白",
  "mes_example": "string (optional) - 2-3轮对话示例，格式: 用户: ...\n角色: ...",
  "model": "string (optional) - 指定 LLM provider，默认用全局 default",
  "avatar": "string (optional) - emoji 或头像 URL"
}
```

**Prompt 构建规则:**
```
System Prompt = 模板(
    身份描述,
    性格特质,
    场景说明,
    发言规则
)
```

发言规则固定为：
1. 用第一人称，像真实对话一样自然交流
2. 保持你的性格特点和语言风格
3. 可以参考你的经历和理念来回应
4. 直接表达观点，不要过度客套
5. 可以同意或反驳其他人的观点，保持真实个性

**示例对话注入:** 如果 `mes_example` 非空，追加到 system prompt 末尾作为 few-shot 参考。

**兼容性:** 与 SillyTavern Character Card v2 格式兼容（忽略 TalkAboutIt 不需要的字段）。

---

## 1.6 LLM Gateway 设计

### Provider 接口

```go
type Provider interface {
    // Chat 发起对话，返回 channel 用于流式读取
    // ctx 取消时会中断请求
    Chat(ctx context.Context, req ChatRequest) (<-chan ChatChunk, error)
    Name() string
    Model() string
}
```

### 多 Provider 支持

| Provider | API 类型 | 适用模型 | 实现要点 |
|----------|---------|---------|---------|
| `OpenAIProvider` | OpenAI Chat Completions + SSE streaming | DeepSeek、GPT-4o、GPT-4.5、GLM、Kimi、Qwen 等兼容厂商 | `base_url` 可覆写以支持兼容 API，MVP 默认指向 DeepSeek |
| `AnthropicProvider` | Anthropic Messages API + SSE streaming | Claude 3.5/4 | 需 `x-api-key` + `anthropic-version` header |

### Provider 工厂与默认策略

- `llm.default` 全局默认值设为 `deepseek`
- `llm.providers.{name}.type` 决定实例化哪个实现
- `type: openai` 统一走 `OpenAIProvider`，通过 `base_url + api_key + model` 切换具体厂商
- `type: anthropic` 走 `AnthropicProvider`
- Engine 选择顺序保持不变：角色卡指定 provider > 全局默认 provider

### 新增 Provider 指南

- OpenAI 兼容厂商：只改 `config.yaml` 的 `base_url + api_key + model`，代码零改动，继续复用 `OpenAIProvider`
- 非 OpenAI 兼容厂商：实现 `Provider` 接口，并在 `gateway.go` 的 provider 工厂里注册新的 `type`
- 约束：所有 provider 都必须暴露 `Chat/Name/Model` 三个统一方法，保证编排层和配置层不感知厂商差异

### 流式处理

```
LLM API (SSE) → 解析 delta → chan ChatChunk → Engine 累积
                                                  ↓
                                           WS 推送前端 (打字机)
```

- 每个 chunk 大小不固定，LLM API 自行决定
- Engine 累积完整回复后才写入数据库
- WS 推送每个 chunk 以实现打字机效果

### 错误处理

| 场景 | 处理 |
|------|------|
| API Key 缺失 | 启动时 warn，调用时返回明确错误 |
| API 超时 (60s) | context 取消，返回超时错误 push 到前端 |
| API 返回 429 (rate limit) | 指数退避重试（1s/2s/4s，最多 3 次），失败则跳过该 persona |
| API 返回 5xx | 重试 1 次，失败则跳过该 persona，继续下一位 |
| 网络断开 | context 取消，清理 goroutine |
| JSON 序列化失败 | 必须处理 error（不可 `_` 忽略），返回错误 |
| SSE 解析失败 | 跳过坏行，不 panic，记录 warn 日志 |
| goroutine 泄露 | context 取消时通过 `select { case <-ctx.Done(): return }` 退出 |

### Provider 实现要点（修订）

```go
// http.Client 必须设超时
client: &http.Client{Timeout: 60 * time.Second}

// JSON 序列化错误必须处理
jsonBody, err := json.Marshal(body)
if err != nil { return nil, fmt.Errorf("marshal request: %w", err) }

// SSE goroutine 必须监听 context 取消
go func() {
    defer resp.Body.Close()
    defer close(ch)
    for {
        select {
        case <-ctx.Done():
            return  // 不泄露 goroutine
        default:
        }
        // ... SSE 解析逻辑
    }
}()

// 全局速率限制（可选，后续迭代）
// 使用 rate.Limiter 限制对同一 provider 的并发请求数
```

### 全局速率限制

为避免多个 roundtable 并发时触发 API 429，Gateway 层增加 per-provider 速率限制：

```go
type Gateway struct {
    limiters map[string]*rate.Limiter  // per-provider 速率限制
    // ...
}

// 默认：每个 provider 最多 3 个并发请求，突发 5 个
func NewGateway(cfg config.LLMConfig) *Gateway {
    return &Gateway{
        limiters: map[string]*rate.Limiter{},
        // 每个 provider 注册一个 limiter: rate.NewLimiter(3, 5)
    }
}
```

---

## 1.7 Discussion Engine 设计

### Round-Robin 算法

```
FOR round = 1 TO max_rounds:
    FOR each persona IN personas:
        1. 构建上下文（topic + 历史消息）
        2. 加载角色卡 → 构建 system prompt
        3. 选择 LLM provider（角色卡指定 > 全局默认）
        4. 调用 LLM API（流式）
        5. 累积完整回复
        6. 写入数据库
        7. WebSocket 推送消息
    END FOR
END FOR
推送 discussion_done
```

### 上下文窗口管理

- 每轮传入所有历史消息
- 历史消息限制最近 **20 条**（≈ 轮数 × 人数，3轮×5人=15条，留余量）
- 消息格式：`[{PersonaName}]: {Content}`，含发言者名字
- 首轮第一条消息前追加 `## 讨论主题\n{topic}\n\n---\n`

### 并发控制

- 同一 roundtable **串行**执行（每人轮流）
- 不同 roundtable **可并发**（各自独立 goroutine）
- 用 `context.WithCancel` 支持取消正在进行的讨论

### 用户介入模式（主持人模式）

在纯自动 Round-Robin 基础上，增加可选的人工介入：

```
FOR round = 1 TO max_rounds:
    FOR each persona IN personas:
        ... (正常发言)
    END FOR
    IF 用户启用「主持人模式」:
        推送 {"type": "waiting_for_user"}
        等待用户输入（或超时 60s 自动继续）
        IF 用户输入:
            作为 _user 消息注入上下文 → 影响后续发言
    END IF
END FOR
```

**实现要点:**
- Engine 增加 `WaitForUserInput` 状态
- WS 双工：服务端推 `waiting_for_user`，客户端发 `{"type":"user_input","content":"..."}`
- 超时保护：60s 无输入自动进入下一轮
- 用户消息存入 messages 表，persona=`_user`

**WS 协议扩展:**
```json
// 服务端 → 客户端：等待用户输入
{"type": "waiting_for_user", "round": 2}

// 客户端 → 服务端：用户输入
{"type": "user_input", "content": "追问：你具体怎么落地这个想法？"}
```

### 讨论摘要机制

当上下文接近 token 上限时，不直接截断——先摘要再续：

```
IF len(history) > THRESHOLD:
    1. 调用一次轻型 LLM 请求生成「讨论摘要」
       - System prompt: "总结以下圆桌讨论的核心观点和分歧"
       - 输入：前 N 轮完整消息
       - 输出：200-300 字结构化摘要
    2. 后续轮次上下文 = 摘要 + 最近 5 条完整消息
    3. 摘要存入 messages 表（persona="_summary"，前端折叠显示）
END IF
```

**触发阈值:** 15 条消息或预估 16K tokens（以先到为准）

**效果：** 5人×10轮也能保证讨论连贯性，上下文稳定在可控范围

---

## 1.8 数据存储设计

### SQLite Schema

```sql
-- 圆桌会议表
CREATE TABLE IF NOT EXISTS roundtables (
    id          TEXT PRIMARY KEY,          -- "rt_" + nanoid(10)
    topic       TEXT NOT NULL,
    personas    TEXT NOT NULL,            -- JSON array: ["Steve Jobs", "Elon Musk"]
    status      TEXT DEFAULT 'pending',   -- pending | running | done | error
    max_rounds  INTEGER DEFAULT 3,
    created_at  DATETIME DEFAULT (datetime('now'))
);

-- 消息表
CREATE TABLE IF NOT EXISTS messages (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    roundtable_id   TEXT NOT NULL,
    round           INTEGER NOT NULL,
    persona         TEXT NOT NULL,
    content         TEXT NOT NULL,
    timestamp       DATETIME DEFAULT (datetime('now')),
    FOREIGN KEY (roundtable_id) REFERENCES roundtables(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_roundtable
    ON messages(roundtable_id, round, id);
```

### 设计决策

- `personas` 字段存 JSON 数组：简单，无需多对多关联表
- ID 用 `rt_` + nanoid：短、URL 安全、避免自增 ID 暴露数量
- 消息按 `(roundtable_id, round, id)` 索引：支持按时间顺序查询
- SQLite WAL 模式：读不阻塞写，适合 WebSocket + HTTP 并发读

---

# Part 2: 前端技术方案

## 2.1 技术选型

| 组件 | 选型 | 版本 | 理由 |
|------|------|------|------|
| 框架 | React | 18.x | 生态成熟，组件化 |
| 构建 | Vite | 6.x | 快，HMR 优秀 |
| 语言 | TypeScript | 5.x | 类型安全 |
| CSS | Tailwind CSS | 4.x | 快速出 UI，暗色模式内置 |
| WebSocket | 原生 `WebSocket` API | — | 零依赖，浏览器内置 |
| HTTP | `fetch` API | — | 零依赖 |
| 状态管理 | React Context + useReducer | — | MVP 不需要 Redux 等重型方案 |
| 路由 | 无需路由 | — | 单页应用，一个页面搞定 |

**无额外依赖:** 不引入 UI 组件库、状态管理库、路由库。MVP 越轻越好。

---

## 2.2 组件树

```
App
├── Header (标题 + GitHub 链接)
├── PersonaSelector (左侧人物选择面板)
│   └── PersonaCard[] (可选中的人物卡片)
├── Roundtable (中央圆桌区域)
│   ├── PersonaAvatar[] (已选人物的头像环)
│   ├── TopicInput (话题输入框)
│   ├── RoundSlider (轮次调节)
│   └── StartButton (开始讨论)
└── DiscussionView (讨论进行时显示，可切换)
    ├── MessageBubble[] (消息气泡流)
    ├── TypingIndicator (某人正在输入)
    └── DoneBanner (讨论结束横幅)
```

**视图状态切换:**

```
┌──────────┐  点击Start  ┌──────────────┐  讨论结束  ┌──────────┐
│  Setup   │ ──────────→ │  Discussing  │ ─────────→ │  Done    │
│  View    │             │  View         │            │  View    │
└──────────┘             └──────────────┘            └──────────┘
```

---

## 2.3 状态管理

### App State (Context)

```typescript
interface AppState {
  // 人物
  personas: Persona[];              // 所有可用人物
  selectedPersonas: string[];       // 已选中的名字列表

  // 圆桌
  topic: string;
  maxRounds: number;
  status: 'setup' | 'discussing' | 'done';

  // 讨论
  roundtableId: string | null;
  currentRound: number;
  currentSpeaker: string | null;
  messages: Message[];
  partialMessage: string;           // 正在输入中的消息片段
}
```

### Actions (useReducer)

```typescript
type Action =
  | { type: 'SET_PERSONAS'; personas: Persona[] }
  | { type: 'TOGGLE_PERSONA'; name: string }
  | { type: 'SET_TOPIC'; topic: string }
  | { type: 'SET_MAX_ROUNDS'; rounds: number }
  | { type: 'START_DISCUSSION'; id: string }
  | { type: 'ROUND_START'; round: number }
  | { type: 'SPEAKING'; persona: string }
  | { type: 'MESSAGE_CHUNK'; persona: string; content: string }
  | { type: 'MESSAGE_DONE'; persona: string; content: string; round: number }
  | { type: 'DISCUSSION_DONE' }
  | { type: 'ERROR'; message: string }
  | { type: 'RESET' };
```

---

## 2.4 组件规格

### 2.4.1 PersonaSelector

**职责:** 展示所有人物，勾选加入圆桌。

```
┌─────────────────────┐
│  👥 选择参与者        │
│                     │
│  ☑️ 🍎 Steve Jobs   │
│  ☑️ 🚀 Elon Musk    │
│  ☐ 🧘 Naval Ravikant│
│  ☐ 🇨🇳 张一鸣        │
│  ☐ 💬 张小龙        │
└─────────────────────┘
```

**交互:**
- 点击卡片切换选中状态
- 至少选 2 人才能开始
- 最多选 5 人（当前）
- 已选人数显示在面板顶部

**Props:** `personas: Persona[], selected: string[], onToggle: (name: string) => void`

### 2.4.2 Roundtable

**职责:** 圆桌主视图，展示已选人物 + 话题输入。

```
             🚀 Elon Musk
    🍎 Steve Jobs     🧘 Naval
        ┌─────────────────┐
        │  讨论话题:       │
        │  [____________]  │
        │                  │
        │  轮次: [===●==] 3│
        │                  │
        │  [ 开始讨论 ]    │
        └─────────────────┘
```

**Props:** `personas: Persona[], topic, maxRounds, onStart, ...`

### 2.4.3 DiscussionView

**职责:** 实时显示讨论对话流。

```
┌─────────────────────────────────────┐
│  🤖 AI 会取代程序员吗？      第 1/3 轮│
│                                     │
│  [🍎 Steve Jobs]                    │
│  ┌─────────────────────────────┐    │
│  │ Look, AI is just a tool...  │    │
│  └─────────────────────────────┘    │
│                                     │
│                     [🚀 Elon Musk]  │
│              ┌────────────────────┐ │
│              │ I disagree. AI is  │ │
│              │ fundamentally...   │ │
│              └────────────────────┘ │
│                                     │
│  🧘 Naval 正在输入...               │
└─────────────────────────────────────┘
```

**消息气泡规则:**
- 自己的发言靠右、他人靠左（或交替）
- 每个 persona 固定一种颜色
- 打字机效果：收到 `message_chunk` 时逐字追加
- 当前发言人显示 "正在输入..." 动画

### 2.4.4 MessageBubble

**Props:**
```typescript
interface MessageBubbleProps {
  persona: string;
  avatar: string;
  content: string;
  isStreaming: boolean;  // 是否还在流式输入中
  side: 'left' | 'right';
}
```

**行为:**
- `isStreaming=true` → 光标闪烁动画
- `isStreaming=false` → 显示完整消息 + 时间戳

---

## 2.5 WebSocket 客户端

**文件:** `frontend/src/hooks/useDiscussion.ts`

```typescript
function useDiscussion(roundtableId: string) {
  // 连接 WS
  // 监听消息，分发到 reducer
  // 自动重连（最多 3 次）
  // 组件卸载时断开

  return { messages, currentSpeaker, currentRound, status, error };
}
```

**重连策略:**
- 断线后延迟 1s / 3s / 5s 重试
- 最多 3 次
- 重连成功后请求 `GET /api/roundtables/{id}` 补全错过的消息

---

## 2.6 开发环境配置

**Vite Config:**
```typescript
// vite.config.ts
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        ws: true,       // WebSocket 代理
      },
    },
  },
});
```

- 前端 `localhost:3000`
- 后端 `localhost:8080`
- Vite proxy 把 `/api/*` 请求转发到后端，**开发时无跨域问题**

---

# Part 3: 前后端联调方案

## 3.1 开发环境拓扑

```
┌──────────────┐     proxy /api/*     ┌──────────────┐
│   Browser    │ ←───────────────────→│  Vite Dev    │
│  :3000       │    ws://:3000/api/   │  Server :3000│
└──────────────┘                      └──────┬───────┘
                                             │ proxy
                                      ┌──────▼───────┐
                                      │  Go Server   │
                                      │  :8080       │
                                      └──────┬───────┘
                                             │ HTTP
                                    ┌────────▼────────┐
                                    │  LLM APIs       │
                                    │  OpenAI/Claude  │
                                    └─────────────────┘
```

**规则:**
1. 浏览器只访问 `localhost:3000`
2. Vite 把 `/api/*` HTTP + WebSocket 代理到 `localhost:8080`
3. **无需 CORS 配置**
4. 前端代码中的 `fetch('/api/...')` 和 `new WebSocket('ws://' + location.host + '/api/...')` 自动走代理

---

## 3.2 API 联调检查清单

### Phase 1: 后端独立测试

```bash
# 1. 启动后端
cd backend && go run ./cmd/server/

# 2. 测试 Persona API
curl http://localhost:8080/api/personas | jq

# 3. 创建 Roundtable
curl -X POST http://localhost:8080/api/roundtables \
  -H 'Content-Type: application/json' \
  -d '{"topic":"测试话题","personas":["Steve Jobs","Elon Musk"],"max_rounds":2}' | jq

# 4. 查看 Roundtable
curl http://localhost:8080/api/roundtables/{id} | jq

# 5. 启动讨论
curl -X POST http://localhost:8080/api/roundtables/{id}/start

# 6. 测试 WebSocket
websocat ws://localhost:8080/api/roundtables/{id}/ws
```

### Phase 2: 前端独立测试

```bash
cd frontend && npm run dev
# 打开 http://localhost:3000
```

Check:
- [ ] PersonaSelector 正确加载人物列表
- [ ] 勾选/取消勾选交互正常
- [ ] Roundtable 视图正确展示已选人物
- [ ] 话题输入和轮次调节正常
- [ ] 未选人时 Start 按钮禁用

### Phase 3: 联调测试

```bash
# Terminal 1
cd backend && go run ./cmd/server/

# Terminal 2
cd frontend && npm run dev
```

端到端流程:
1. [ ] 打开 `localhost:3000`，看到人物列表
2. [ ] 选择 2-3 个人物
3. [ ] 输入话题，设置轮次
4. [ ] 点击 "开始讨论"
5. [ ] 看到 WS 连接建立
6. [ ] 看到 `discussion_started` 消息
7. [ ] 看到第一个人物开始发言（打字机效果）
8. [ ] 发言完毕切换到下一个
9. [ ] 所有轮次完成，看到 `discussion_done`
10. [ ] 讨论结束后可查看完整对话记录
11. [ ] 可以 "新建讨论" 回到 Setup 页

---

## 3.3 WebSocket 联调协议

### 消息时序

```
Client                          Server
  │                                │
  │──── WS Connect ───────────────→│
  │←─── discussion_started ────────│
  │←─── round_start (round=1) ────│
  │←─── speaking (persona=A) ─────│
  │←─── message_chunk ("Hello") ──│
  │←─── message_chunk (" world") ─│
  │←─── message_done (persona=A) ─│
  │←─── speaking (persona=B) ─────│
  │←─── message_chunk ... ────────│
  │←─── message_done (persona=B) ─│
  │←─── round_start (round=2) ────│
  │     ... (重复)                 │
  │←─── discussion_done ──────────│
```

### 调试步骤

1. 先用 `websocat` 或浏览器 DevTools → Network → WS 面板验证消息格式
2. 确认每种 `type` 的消息前端都有对应的 reducer action 处理
3. 测试异常场景：
   - [ ] 讨论中途关闭浏览器 → 后端继续运行，状态不变
   - [ ] 讨论中途刷新页面 → WS 重连，拉取历史消息补全
   - [ ] LLM API 超时 → 前端显示错误提示，讨论继续（跳过该 persona）

---

## 3.4 错误处理约定

### 后端错误码

| HTTP Status | 含义 | 前端处理 |
|-------------|------|---------|
| 400 | 请求参数错误 | 显示错误提示，不重试 |
| 404 | 资源不存在 | 显示 "未找到"，返回首页 |
| 409 | 冲突（如重复启动） | 显示当前状态 |
| 500 | 服务器内部错误 | 显示 "服务器错误，请稍后重试" |
| 502/503 | LLM API 不可用 | 显示 "AI 服务暂时不可用" |

### WebSocket 错误

```json
{
  "type": "error",
  "message": "...",
  "persona": "Elon Musk",     // 可选，关联到具体人物
  "recoverable": true         // true=继续下一位，false=停止讨论
}
```

---

## 3.5 部署备忘（后续）

- Go 后端编译为单二进制：`CGO_ENABLED=0 go build -o talkaboutit ./cmd/server/`
- 前端构建：`npm run build` → `dist/` 静态文件
- 生产环境：Go 后端 serve 前端静态文件（`//go:embed`）或 nginx 反向代理
- SQLite 数据文件与二进制同目录，或通过配置文件指定路径

---

> **文档维护:** 后端 API 变更时更新此文档。前端组件接口变更时更新组件规格。
