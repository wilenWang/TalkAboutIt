# TalkAboutIt 实施计划

> **For Hermes:** 使用 `subagent-driven-development` skill 逐任务实施。每个 task 派一个独立子代理，含完整上下文。

**Goal:** 构建独立的圆桌会议应用——用户选择名人 Persona，设定话题，观察他们轮流讨论。

**Architecture:** Go 后端（REST + WebSocket）+ React 前端。Persona 系统与 SillyTavern 角色卡兼容。LLM 通过统一 Gateway 抽象调用。

**Tech Stack:** Go 1.22+ / chi router / gorilla websocket / SQLite / React 18 / Vite / Tailwind CSS

---

## 项目目录结构

```
TalkAboutIt/
├── backend/                    # Go 后端
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── internal/
│   │   ├── api/               # HTTP handlers (router, persona, roundtable, ws)
│   │   ├── engine/            # 讨论引擎 (round-robin)
│   │   ├── persona/           # 角色卡模型 + 加载器
│   │   ├── llm/               # LLM Gateway (provider 接口 + 实现)
│   │   ├── session/           # Session 存储 (SQLite)
│   │   └── config/            # 配置加载
│   ├── personas/              # 角色卡 JSON 文件
│   │   ├── steve-jobs.json
│   │   ├── elon-musk.json
│   │   ├── zhang-yiming.json
│   │   ├── zhang-xiaolong.json
│   │   └── naval-ravikant.json
│   ├── data/                  # 运行时数据 (SQLite DB)
│   ├── docs/                  # 后端文档 (API 设计、架构决策等)
│   ├── config.yaml            # 配置文件
│   ├── go.mod
│   └── go.sum
├── frontend/                  # React 前端
│   ├── src/
│   │   ├── components/
│   │   │   ├── PersonaSelector.tsx
│   │   │   ├── Roundtable.tsx
│   │   │   ├── DiscussionView.tsx
│   │   │   └── MessageBubble.tsx
│   │   ├── hooks/
│   │   │   └── useDiscussion.ts
│   │   ├── App.tsx
│   │   ├── main.tsx
│   │   └── index.css
│   ├── docs/                  # 前端文档 (组件设计、状态管理方案等)
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
└── docs/                      # 项目级文档 (架构图、实施计划)
    ├── implementation-plan.md
    └── talkaboutit-architecture.drawio
```

> **目录规则：** 所有 Go 代码路径相对于 `backend/`，所有前端路径相对于 `frontend/`。go.mod 在 `backend/` 下，package.json 在 `frontend/` 下。`go build` 和 `npm run dev` 均在各自目录下执行。

---

## Phase 1: 项目骨架

### Task 1.1: 创建项目目录结构

**Objective:** 按 `backend/` / `frontend/` / `docs/` 三层分离搭建骨架

**Files:**
- Create: `~/TalkAboutIt/backend/go.mod`
- Create: `~/TalkAboutIt/backend/cmd/server/main.go` (占位)
- Create: `~/TalkAboutIt/backend/internal/` 子目录
- Create: `~/TalkAboutIt/backend/personas/` (空目录)
- Create: `~/TalkAboutIt/backend/data/` (空目录)
- Create: `~/TalkAboutIt/frontend/` (后续 tasks 填充)

**Step 1: 初始化 Go module**

```bash
mkdir -p ~/TalkAboutIt/backend
cd ~/TalkAboutIt/backend
go mod init github.com/wilenwang/talkaboutit
```

**Step 2: 创建占位 main.go**

```go
// backend/cmd/server/main.go
package main

import "fmt"

func main() {
    fmt.Println("TalkAboutIt server starting...")
}
```

**Step 3: 创建 internal 子目录 + 运行时目录 + docs**

```bash
mkdir -p ~/TalkAboutIt/backend/cmd/server
mkdir -p ~/TalkAboutIt/backend/internal/{api,engine,persona,llm,session,config}
mkdir -p ~/TalkAboutIt/backend/personas
mkdir -p ~/TalkAboutIt/backend/data
mkdir -p ~/TalkAboutIt/backend/docs
mkdir -p ~/TalkAboutIt/frontend/docs
```

**Step 4: 验证编译**

```bash
cd ~/TalkAboutIt/backend
go build ./cmd/server/...
```

Expected: 编译成功，运行输出 "TalkAboutIt server starting..."

---

### Task 1.2: 配置文件加载

**Objective:** 定义 Config 结构体，从 `config.yaml` 加载

**Files:**
- Create: `~/TalkAboutIt/backend/config.example.yaml`
- Create: `~/TalkAboutIt/backend/internal/config/config.go`
- Create: `~/TalkAboutIt/backend/internal/config/config_test.go`

**Step 1: 写配置结构体 + 加载**

```go
// internal/config/config.go
package config

import (
    "os"
    "gopkg.in/yaml.v3"
)

type Config struct {
    Server   ServerConfig   `yaml:"server"`
    LLM      LLMConfig      `yaml:"llm"`
    Personas PersonasConfig `yaml:"personas"`
    Session  SessionConfig  `yaml:"session"`
}

type ServerConfig struct {
    Port int    `yaml:"port"`
    Host string `yaml:"host"`
}

type LLMConfig struct {
    Providers map[string]ProviderConfig `yaml:"providers"`
    Default   string                    `yaml:"default"`
}

type ProviderConfig struct {
    Type    string `yaml:"type"`    // "openai", "anthropic"
    APIKey  string `yaml:"api_key"`
    BaseURL string `yaml:"base_url"`
    Model   string `yaml:"model"`
}

type PersonasConfig struct {
    Dir string `yaml:"dir"`
}

type SessionConfig struct {
    DBPath string `yaml:"db_path"`
    MaxRounds int `yaml:"max_rounds"`
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    cfg := &Config{
        Server: ServerConfig{Port: 8080, Host: "localhost"},
        Session: SessionConfig{MaxRounds: 3},
        Personas: PersonasConfig{Dir: "personas"},
    }
    if err := yaml.Unmarshal(data, cfg); err != nil {
        return nil, err
    }
    // 环境变量覆盖 API Key
    for name, p := range cfg.LLM.Providers {
        if envKey := os.Getenv(fmt.Sprintf("LLM_%s_API_KEY", strings.ToUpper(name))); envKey != "" {
            p.APIKey = envKey
            cfg.LLM.Providers[name] = p
        }
    }
    return cfg, nil
}
```

**Step 2: Sample config**

```yaml
# config.example.yaml
server:
  port: 8080
  host: localhost

llm:
  default: deepseek
  providers:
    deepseek:
      type: openai       # DeepSeek 兼容 OpenAI API
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

**Step 3: 单元测试**

```go
// internal/config/config_test.go
func TestLoad(t *testing.T) {
    tmp := t.TempDir()
    path := filepath.Join(tmp, "config.yaml")
    os.WriteFile(path, []byte(`
server:
  port: 9999
session:
  max_rounds: 5
`), 0644)
    cfg, err := Load(path)
    assert.NoError(t, err)
    assert.Equal(t, 9999, cfg.Server.Port)
    assert.Equal(t, 5, cfg.Session.MaxRounds)
}
```

**Step 4: 安装依赖 + 运行测试**

```bash
go get gopkg.in/yaml.v3
go test ./internal/config/ -v
```

Expected: 1 test PASS

---

## Phase 2: Persona 系统

### Task 2.1: 角色卡数据模型

**Objective:** 定义 CharacterCard 结构体，支持 JSON 序列化

**Files:**
- Create: `~/TalkAboutIt/backend/internal/persona/card.go`
- Create: `~/TalkAboutIt/backend/internal/persona/card_test.go`

**Step 1: 写结构体**

```go
// internal/persona/card.go
package persona

type CharacterCard struct {
    Name         string `json:"name"`
    Description  string `json:"description"`
    Personality  string `json:"personality"`
    Scenario     string `json:"scenario"`
    FirstMessage string `json:"first_mes"`
    Examples     string `json:"mes_example"`
    Model        string `json:"model"`        // 指定 LLM provider
    Avatar       string `json:"avatar"`       // emoji or URL
}

// ToSystemPrompt 将角色卡转为 LLM system prompt
func (c *CharacterCard) ToSystemPrompt() string {
    prompt := fmt.Sprintf(`你是 %s。%s

## 你的性格
%s

## 当前场景
%s

## 发言规则
- 用第一人称，像真实对话一样自然交流
- 保持你的性格特点和语言风格
- 可以参考你的经历和理念来回应
- 直接表达观点，不要过度客套`, c.Name, c.Description, c.Personality, c.Scenario)
    return prompt
}
```

**Step 2: 测试 JSON 序列化**

```go
func TestCardJSON(t *testing.T) {
    card := CharacterCard{
        Name:        "Steve Jobs",
        Personality: "Perfectionist",
    }
    data, _ := json.Marshal(card)
    var decoded CharacterCard
    json.Unmarshal(data, &decoded)
    assert.Equal(t, "Steve Jobs", decoded.Name)
}

func TestToSystemPrompt(t *testing.T) {
    card := CharacterCard{
        Name:        "Steve Jobs",
        Description: "Apple co-founder.",
        Personality: "Perfectionist.",
        Scenario:    "Roundtable discussion.",
    }
    prompt := card.ToSystemPrompt()
    assert.Contains(t, prompt, "你是 Steve Jobs")
    assert.Contains(t, prompt, "Perfectionist")
}
```

**Step 3: 测试**

```bash
go test ./internal/persona/ -v
```

---

### Task 2.2: 角色卡加载器

**Objective:** 从 `personas/` 目录加载所有 JSON 角色卡

**Files:**
- Create: `~/TalkAboutIt/backend/internal/persona/loader.go`
- Create: `~/TalkAboutIt/backend/internal/persona/loader_test.go`

**Step 1: Loader**

```go
// internal/persona/loader.go
type Loader struct {
    dir string
    cards map[string]*CharacterCard
}

func NewLoader(dir string) *Loader {
    return &Loader{dir: dir, cards: make(map[string]*CharacterCard)}
}

func (l *Loader) Load() error {
    entries, err := os.ReadDir(l.dir)
    if err != nil {
        return fmt.Errorf("read persona dir: %w", err)
    }
    for _, entry := range entries {
        if !strings.HasSuffix(entry.Name(), ".json") {
            continue
        }
        data, err := os.ReadFile(filepath.Join(l.dir, entry.Name()))
        if err != nil {
            return err
        }
        var card CharacterCard
        if err := json.Unmarshal(data, &card); err != nil {
            return fmt.Errorf("parse %s: %w", entry.Name(), err)
        }
        l.cards[card.Name] = &card
    }
    return nil
}

func (l *Loader) Get(name string) (*CharacterCard, bool) {
    card, ok := l.cards[name]
    return card, ok
}

func (l *Loader) List() []CharacterCard {
    cards := make([]CharacterCard, 0, len(l.cards))
    for _, c := range l.cards {
        cards = append(cards, *c)
    }
    return cards
}
```

**Step 2: 测试**

创建一个临时测试目录，放入测试角色卡 JSON，验证 Load/Get/List。

```bash
go test ./internal/persona/ -v
```

---

### Task 2.3: 编写 5 个名人角色卡

**Objective:** 为乔布斯、马斯克、张一鸣、张小龙、纳瓦尔编写角色卡 JSON

**Files:**
- Create: `~/TalkAboutIt/backend/personas/steve-jobs.json`
- Create: `~/TalkAboutIt/backend/personas/elon-musk.json`
- Create: `~/TalkAboutIt/backend/personas/zhang-yiming.json`
- Create: `~/TalkAboutIt/backend/personas/zhang-xiaolong.json`
- Create: `~/TalkAboutIt/backend/personas/naval-ravikant.json`

**Step 1: 模板**

```json
{
  "name": "Steve Jobs",
  "description": "Apple 联合创始人，被逐出后回归拯救公司，发布 iMac/iPod/iPhone/iPad 等革命性产品。坚信科技与人文的交叉点能创造出伟大的产品。",
  "personality": "极致完美主义者，有「现实扭曲力场」——能用强大信念说服他人。说话直接、尖锐，常用 'This is shit' 或 'It's magic' 这样的两极评价。推崇简约设计，厌恶复杂。相信直觉超过市场调研。",
  "scenario": "你正在参加一个圆桌讨论会，和其他几位企业家一起探讨话题。你可以自由表达你的观点，可以同意或反对他人。",
  "first_mes": "Let me just say one thing — most people don't know what they want until you show it to them.",
  "mes_example": "用户: 你觉得做产品应该听用户的吗？\n乔布斯: People don't know what they want until you show it to them. That's not arrogance — it's reality. Our job is to figure out what they want before they do. When we created the Mac, nobody was asking for a graphical interface. They were using DOS. You think they'd ask for a mouse? No. But once they saw it, they knew.",
  "model": "claude",
  "avatar": "🍎"
}
```

同理写出其他 4 个。每个角色卡的 `mes_example` 需要 2-3 轮对话示例，展示该人物的语言风格和思维模式。

**验证:** Go loader 能成功加载全部 5 张卡。

---

## Phase 3: LLM Gateway

### Task 3.1: LLM Provider 接口定义

**Objective:** 定义统一的 LLM 调用接口

**Files:**
- Create: `~/TalkAboutIt/backend/internal/llm/provider.go`

```go
// internal/llm/provider.go
package llm

import "context"

type Message struct {
    Role    string
    Content string
}

type ChatRequest struct {
    SystemPrompt string
    Messages     []Message
    MaxTokens    int
    Temperature  float64
}

type ChatChunk struct {
    Content string
    Done    bool
}

type Provider interface {
    Chat(ctx context.Context, req ChatRequest) (<-chan ChatChunk, error)
    Name() string
    Model() string
}
```

**无需测试** — 纯接口定义。

---

### Task 3.2: OpenAI Provider 实现

**Objective:** 实现 OpenAI API 调用（同时兼容 DeepSeek 等 OpenAI 兼容 API）

**Files:**
- Create: `~/TalkAboutIt/backend/internal/llm/openai.go`

```go
// internal/llm/openai.go
package llm

import (
    "bufio"
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
)

type OpenAIProvider struct {
    apiKey  string
    baseURL string
    model   string
    client  *http.Client
}

func NewOpenAI(apiKey, baseURL, model string) *OpenAIProvider {
    if baseURL == "" {
        baseURL = "https://api.openai.com/v1"
    }
    return &OpenAIProvider{
        apiKey:  apiKey,
        baseURL: baseURL,
        model:   model,
        client:  &http.Client{},
    }
}

func (p *OpenAIProvider) Name() string { return "openai/" + p.model }
func (p *OpenAIProvider) Model() string { return p.model }

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (<-chan ChatChunk, error) {
    messages := []map[string]string{
        {"role": "system", "content": req.SystemPrompt},
    }
    // 保留最近 20 条历史消息避免 token 溢出
    history := req.Messages
    if len(history) > 20 {
        history = history[len(history)-20:]
    }
    for _, m := range history {
        messages = append(messages, map[string]string{"role": m.Role, "content": m.Content})
    }

    body := map[string]interface{}{
        "model":       p.model,
        "messages":    messages,
        "stream":      true,
        "max_tokens":  req.MaxTokens,
        "temperature": req.Temperature,
    }
    jsonBody, _ := json.Marshal(body)

    httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

    resp, err := p.client.Do(httpReq)
    if err != nil {
        return nil, err
    }

    ch := make(chan ChatChunk, 10)
    go func() {
        defer resp.Body.Close()
        defer close(ch)
        scanner := bufio.NewScanner(resp.Body)
        for scanner.Scan() {
            line := scanner.Text()
            if !strings.HasPrefix(line, "data: ") {
                continue
            }
            data := strings.TrimPrefix(line, "data: ")
            if data == "[DONE]" {
                ch <- ChatChunk{Done: true}
                return
            }
            // 解析 delta content
            var streamResp struct {
                Choices []struct {
                    Delta struct {
                        Content string `json:"content"`
                    } `json:"delta"`
                } `json:"choices"`
            }
            if json.Unmarshal([]byte(data), &streamResp) == nil {
                if len(streamResp.Choices) > 0 {
                    ch <- ChatChunk{Content: streamResp.Choices[0].Delta.Content}
                }
            }
        }
    }()
    return ch, nil
}
```

---

### Task 3.3: Anthropic Provider 实现

**Objective:** 实现 Claude API 调用（Anthropic Messages API + streaming）

**Files:**
- Create: `~/TalkAboutIt/backend/internal/llm/anthropic.go`

类似 OpenAI provider 结构，但适配 Anthropic Messages API 格式。使用 SSE streaming。

**Tip:** Anthropic 需要 `x-api-key` header + `anthropic-version: 2023-06-01`。Streaming 返回 `data: {"type": "content_block_delta", "delta": {"text": "..."}}`

---

### Task 3.4: LLM Gateway 工厂

**Objective:** 从配置创建 Provider 实例，提供统一入口

**Files:**
- Create: `~/TalkAboutIt/backend/internal/llm/gateway.go`

```go
// internal/llm/gateway.go
package llm

import (
    "fmt"
    "talkaboutit/internal/config"
)

type Gateway struct {
    defaultProvider string
    providers       map[string]Provider
}

func NewGateway(cfg config.LLMConfig) (*Gateway, error) {
    g := &Gateway{
        defaultProvider: cfg.Default,
        providers:       make(map[string]Provider),
    }
    for name, pc := range cfg.Providers {
        switch pc.Type {
        case "openai":
            g.providers[name] = NewOpenAI(pc.APIKey, pc.BaseURL, pc.Model)
        case "anthropic":
            g.providers[name] = NewAnthropic(pc.APIKey, pc.BaseURL, pc.Model)
        default:
            return nil, fmt.Errorf("unknown provider type: %s", pc.Type)
        }
    }
    return g, nil
}

func (g *Gateway) Get(providerName string) (Provider, bool) {
    if providerName == "" {
        providerName = g.defaultProvider
    }
    p, ok := g.providers[providerName]
    return p, ok
}
```

---

## Phase 4: 讨论引擎

### Task 4.1: Session 数据模型与存储

**Objective:** 定义 Roundtable 和 Message 模型，实现 SQLite 持久化

**Files:**
- Create: `~/TalkAboutIt/backend/internal/session/store.go`

```go
// internal/session/store.go
package session

import (
    "database/sql"
    "time"
)

type Message struct {
    ID        int64     `json:"id"`
    Round     int       `json:"round"`
    Persona   string    `json:"persona"`
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}

type Roundtable struct {
    ID        string    `json:"id"`
    Topic     string    `json:"topic"`
    Personas  []string  `json:"personas"`
    Status    string    `json:"status"` // "pending", "running", "done"
    MaxRounds int       `json:"max_rounds"`
    CreatedAt time.Time `json:"created_at"`
}

type Store struct {
    db *sql.DB
}

func NewStore(dbPath string) (*Store, error) {
    db, err := sql.Open("sqlite", dbPath)
    if err != nil {
        return nil, err
    }
    s := &Store{db: db}
    return s, s.migrate()
}

func (s *Store) migrate() error {
    _, err := s.db.Exec(`
        CREATE TABLE IF NOT EXISTS roundtables (
            id TEXT PRIMARY KEY,
            topic TEXT NOT NULL,
            personas TEXT NOT NULL,  -- JSON array
            status TEXT DEFAULT 'pending',
            max_rounds INTEGER DEFAULT 3,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        CREATE TABLE IF NOT EXISTS messages (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            roundtable_id TEXT NOT NULL,
            round INTEGER NOT NULL,
            persona TEXT NOT NULL,
            content TEXT NOT NULL,
            timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (roundtable_id) REFERENCES roundtables(id)
        );
    `)
    return err
}

// Create, GetByID, AddMessage, GetMessages 方法...
```

---

### Task 4.2: 讨论引擎 — Round-Robin 编排

**Objective:** 核心引擎：按轮次和人物顺序编排 LLM 调用，流式产出消息

**Files:**
- Create: `~/TalkAboutIt/backend/internal/engine/engine.go`

```go
// internal/engine/engine.go
package engine

import (
    "context"
    "fmt"
    "talkaboutit/internal/llm"
    "talkaboutit/internal/persona"
    "talkaboutit/internal/session"
)

type Engine struct {
    gateway *llm.Gateway
    store   *session.Store
    personas *persona.Loader
}

func New(g *llm.Gateway, s *session.Store, p *persona.Loader) *Engine {
    return &Engine{gateway: g, store: s, personas: p}
}

// RunRoundtable 启动圆桌讨论，返回消息 channel
func (e *Engine) RunRoundtable(ctx context.Context, rt *session.Roundtable) (<-chan *session.Message, <-chan error) {
    msgCh := make(chan *session.Message, 10)
    errCh := make(chan error, 1)

    go func() {
        defer close(msgCh)
        defer close(errCh)

        for round := 1; round <= rt.MaxRounds; round++ {
            for _, personaName := range rt.Personas {
                card, ok := e.personas.Get(personaName)
                if !ok {
                    errCh <- fmt.Errorf("persona not found: %s", personaName)
                    return
                }

                // 获取历史消息
                history, _ := e.store.GetMessages(rt.ID)
                llmMessages := make([]llm.Message, 0, len(history)+1)
                for _, m := range history {
                    role := "assistant"
                    if m.Persona == "_user" {
                        role = "user"
                    }
                    llmMessages = append(llmMessages, llm.Message{
                        Role: role,
                        Content: fmt.Sprintf("[%s]: %s", m.Persona, m.Content),
                    })
                }

                // 选择 provider
                providerName := card.Model
                provider, ok := e.gateway.Get(providerName)
                if !ok {
                    provider, _ = e.gateway.Get("")
                }

                // 调用 LLM
                chunkCh, err := provider.Chat(ctx, llm.ChatRequest{
                    SystemPrompt: card.ToSystemPrompt(),
                    Messages:  llmMessages,
                    MaxTokens:  1024,
                    Temperature: 0.8,
                })
                if err != nil {
                    errCh <- err
                    return
                }

                // 累积并流式发送
                var fullContent strings.Builder
                for chunk := range chunkCh {
                    if chunk.Done {
                        break
                    }
                    fullContent.WriteString(chunk.Content)
                }

                msg := &session.Message{
                    Round:   round,
                    Persona: personaName,
                    Content: fullContent.String(),
                }
                e.store.AddMessage(rt.ID, msg)
                msgCh <- msg
            }
        }
    }()
    return msgCh, errCh
}
```

---

## Phase 5: API 层

### Task 5.1: HTTP Router 与 Persona API

**Objective:** 搭建 chi router，实现 Persona 列表/详情 API

**Files:**
- Create: `~/TalkAboutIt/backend/internal/api/router.go`
- Create: `~/TalkAboutIt/backend/internal/api/persona.go`

```go
// GET /api/personas → List
// GET /api/personas/:name → Get
```

---

### Task 5.2: Roundtable CRUD API

**Objective:** 创建/查询/删除 Roundtable

```
POST   /api/roundtables              → Create
GET    /api/roundtables/:id          → Get (含 messages)
DELETE /api/roundtables/:id          → Delete
GET    /api/roundtables/:id/messages → Messages only
```

---

### Task 5.3: Start Discussion API + WebSocket

**Objective:** POST 启动讨论 + WebSocket 实时推送消息流

```
POST /api/roundtables/:id/start → 启动讨论（后台 goroutine）
GET  /api/roundtables/:id/ws    → WebSocket 实时流
```

WebSocket handler 在讨论进行时，每收到一条引擎消息就推送给前端。

---

### Task 5.4: main.go 启动入口

**Objective:** 串联所有组件，启动 HTTP server

```go
func main() {
    cfg, _ := config.Load("config.yaml")
    gateway, _ := llm.NewGateway(cfg.LLM)
    personaLoader := persona.NewLoader(cfg.Personas.Dir)
    personaLoader.Load()
    store, _ := session.NewStore(cfg.Session.DBPath)
    engine := engine.New(gateway, store, personaLoader)
    // 创建 router, 注入依赖, 启动 server
}
```

---

## Phase 6: Frontend

### Task 6.1: Vite + React 项目初始化

```bash
cd ~/TalkAboutIt
npm create vite@latest frontend -- --template react-ts
cd ~/TalkAboutIt/frontend
npm install
npm install -D tailwindcss @tailwindcss/vite
```

配置 Tailwind，设置 dev proxy → `localhost:8080`。

---

### Task 6.2: 人物选择组件（含边界态）

**Objective:** 展示人物卡片列表，可勾选拖入圆桌

**Files:** `frontend/src/components/PersonaSelector.tsx`

- [ ] 正常态：左侧边栏显示 5 个预设人物，勾选高亮
- [ ] 加载态：Skeleton 卡片占位（灰色闪烁）
- [ ] 空状态：无人物时显示 "暂无可用角色，请先在 Role Tab 中添加"
- [ ] 错误态：API 获取失败显示重试按钮
- [ ] 点击勾选/取消，已选人数实时更新

---

### Task 6.3: 圆桌 UI + 话题输入（含边界态）

**Objective:** 圆桌主视图：选中的人物头像 + 话题输入框 + 开始按钮

**Files:** `frontend/src/components/Roundtable.tsx`

- [ ] 正常态：半环形布局展示已选人物头像
- [ ] 空状态：未选人时显示 "👥 请从左侧选择参与者"
- [ ] 禁用态：选人 < 2 时 Start 按钮灰色不可点击
- [ ] 验证：话题为空时不允许开始
- [ ] 错误态：开始失败时显示 toast 错误提示

---

### Task 6.4: 讨论视图 + 实时消息流（含边界态）

**Objective:** 讨论进行中的消息气泡流 + WebSocket 连接

**Files:**
- `frontend/src/components/DiscussionView.tsx`
- `frontend/src/hooks/useDiscussion.ts`（WebSocket hook + 自动重连）

- [ ] 正常态：消息气泡（不同 persona 不同颜色左侧条）
- [ ] 流式态：打字机效果（收到 chunk 逐字追加）
- [ ] 加载态：历史消息加载中显示骨架屏
- [ ] 空状态：讨论未开始时显示 "等待讨论开始..."
- [ ] 错误态：WebSocket 断开显示黄色横幅 "连接断开，正在重连..."
- [ ] 完成态：讨论结束显示 "✦ 讨论结束" + 总消息数
- [ ] 消息去重：按 seq 判断，不渲染已存在消息
- [ ] 自动重连：断线后 1s/3s/5s 退避重试，最多 3 次

---

### Task 6.5: ErrorBoundary 全局错误组件

**Objective:** React ErrorBoundary 包裹整个应用

**Files:** `frontend/src/components/ErrorBoundary.tsx`

- 捕获未处理的渲染错误
- 显示 "出错了" + 错误信息 + "刷新页面" 按钮
- 避免白屏

---

### Task 6.6: App 组装 + 双 Tab 路由 + 联调

---

## Phase 7: 集成与错误测试

### Task 7.1: 端到端集成测试

**Objective:** 全链路自动化验证

```bash
# 测试脚本 (scripts/e2e_test.sh)
1. 启动后端 server
2. GET /api/personas → 验证返回 5 个角色
3. POST /api/roundtables → 创建 roundtable
4. GET /api/roundtables/{id} → 验证状态 pending
5. POST /api/roundtables/{id}/start → 启动讨论
6. 连接 WS → 验证收到 discussion_started
7. 等待所有轮次完成 → 验证收到 discussion_done
8. GET /api/roundtables/{id} → 验证 status=done, messages 数量正确
9. DELETE /api/roundtables/{id} → 验证 204
```

---

### Task 7.2: 错误注入测试

**Objective:** 验证异常场景下的系统行为

| 测试场景 | 预期行为 |
|----------|---------|
| LLM API Key 无效 | 返回明确错误，不 panic |
| LLM API 超时 | 返回超时错误，跳过当前 persona，继续下一位 |
| LLM API 返回 429 | 指数退避重试，3 次后跳过 |
| WebSocket 断线 | 客户端自动重连，消息去重补全 |
| 并发 roundtable | 多个同时运行不互相影响 |
| SQLite 写满磁盘 | 返回错误，不崩溃 |

---

## Phase 8: 部署

### Task 8.1: 单二进制打包

**Objective:** 编译后单文件可运行，前端静态文件 embed 进二进制

```go
//go:embed all:frontend/dist
var frontendAssets embed.FS

// 生产模式：Go serve 前端静态文件
// 开发模式：Vite proxy 到后端
```

```bash
# 构建
cd backend
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o talkaboutit ./cmd/server/

# 运行（无需 Node、无需 npm）
./talkaboutit
# 打开 http://localhost:8080
```

---

### Task 8.2: Docker 构建（可选）

`backend/Dockerfile` 多阶段构建：Go 编译 → 复制二进制 + 前端 dist + 角色卡 JSON

---

## 技术选型汇总

| 层 | 选型 | 理由 |
|---|---|---|
| Router | chi | 轻量、惯用、兼容 net/http |
| WebSocket | gorilla/websocket | 最成熟的 Go WS 库 |
| SQLite | modernc.org/sqlite | 纯 Go、零 CGo 依赖 |
| LLM HTTP | net/http | 不引入第三方 SDK，可控 |
| 前端构建 | Vite | 快、React 默认 |
| CSS | Tailwind CSS v4 | 快速出 UI |

---

## 人物卡书写指南

每个角色卡重点抓住 3 个要素：

1. **思维模式** — 他怎么看世界？（第一性原理 / 延迟满足 / 简洁至上）
2. **语言风格** — 他怎么说话？（直接 vs 委婉、口头禅、句式偏好）
3. **价值锚点** — 什么对他最重要？（设计 / 工程 / 增长 / 人性）

`mes_example` 是灵魂——2-3 轮对话示例决定 LLM 模仿的准确度，需要精心写。

---

> **下一步：** 确认本计划后，使用 `subagent-driven-development` 逐 task 实施。
