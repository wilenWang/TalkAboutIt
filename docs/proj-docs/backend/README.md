# TalkAboutIt Backend

## 1. 项目简介

TalkAboutIt 是一个 AI 圆桌讨论（Roundtable）应用，后端使用 Go 语言构建。它允许用户选择多个 AI 角色（Persona），围绕指定话题展开多轮对话。每个角色拥有独立的性格、语言风格和知识背景，由 LLM 驱动生成符合角色特征的发言。

核心能力：

- **多角色圆桌对话**：支持多个 AI 角色同时参与讨论，模拟真实的圆桌会议场景
- **流式响应（SSE）**：通过 Server-Sent Events 实时推送角色发言，提供流畅的用户体验
- **多 LLM 提供商**：支持 OpenAI 兼容接口（DeepSeek、OpenAI 等）和 Anthropic（Claude），通过工厂模式灵活切换
- **角色持久化**：预置角色以 JSON 文件管理，支持自定义扩展
- **会话持久化**：基于 SQLite 存储圆桌会话、消息记录和事件流，支持 SSE 断线重连

---

## 2. 项目目录

```
backend/
├── cmd/
│   └── server/
│       └── main.go                  # 应用入口
├── internal/
│   ├── api/
│   │   ├── router.go                # HTTP 路由注册
│   │   ├── router_test.go
│   │   ├── roundtable_handler.go    # 圆桌 handler（占位，逻辑在 sse_handler）
│   │   └── sse_handler.go           # SSE 事件流 & 圆桌 API handler
│   ├── config/
│   │   ├── config.go                # 配置加载与环境变量覆盖
│   │   └── config_test.go
│   ├── engine/
│   │   ├── engine.go                # 圆桌讨论编排引擎
│   │   └── engine_test.go
│   ├── llm/
│   │   ├── anthropic.go             # Anthropic (Claude) 实现
│   │   ├── errors.go                # LLM 错误类型定义
│   │   ├── factory.go               # Provider 工厂
│   │   ├── llm_test.go
│   │   ├── openai.go                # OpenAI 兼容接口实现
│   │   ├── prompt.go                # Prompt 组装
│   │   ├── prompt_test.go
│   │   ├── provider.go              # Provider 接口定义
│   │   └── provider_test.go
│   ├── persona/
│   │   ├── context.go               # 对话上下文管理
│   │   ├── loader.go                # 角色 JSON 文件加载器
│   │   ├── persona.go               # 角色数据模型
│   │   ├── persona_test.go
│   │   ├── prompt_sections.go       # Prompt 段落构建
│   │   ├── state.go                 # Per-persona 状态管理
│   │   ├── validate.go              # 角色数据校验
│   │   └── validate_test.go
│   └── session/
│       ├── store.go                 # SQLite 会话存储
│       └── store_test.go
├── personas/                        # 预置角色 JSON 文件
│   ├── elon-musk.json
│   ├── naval-ravikant.json
│   ├── steve-jobs.json
│   ├── zhang-xiaolong.json
│   └── zhang-yiming.json
├── scripts/
│   └── build.sh                     # 跨平台构建脚本
├── test/
│   ├── integration_test.go          # 集成测试
│   ├── realllm_test.go              # 真实 LLM 调用测试
│   └── eval/                        # 评测框架
│       ├── baseline.json
│       ├── eval_llm_judge_test.go
│       ├── eval_test.go
│       ├── llm_judge.go
│       ├── persona_audit_test.go
│       ├── questions.json
│       └── scorer.go
├── config.example.yaml              # 配置文件示例
├── Dockerfile                       # Docker 多阶段构建
├── Makefile                         # 构建/测试/格式化命令
├── go.mod                           # Go 模块定义
└── go.sum                           # 依赖校验
```

---

## 3. 各目录主要功能

### `cmd/server/`

应用主入口。负责初始化配置、角色加载器、SQLite 存储、LLM Provider 和 HTTP 路由，最终启动 HTTP 服务监听。

### `internal/api/`

HTTP API 层，负责路由注册和请求处理：

| 功能 | 说明 |
|------|------|
| 角色列表/详情 | `GET /api/v1/personas`、`GET /api/v1/personas/{id}` |
| 创建圆桌 | `POST /api/v1/roundtables` |
| 启动讨论 | `POST /api/v1/roundtables/{id}/start` |
| 圆桌列表/快照 | `GET /api/v1/roundtables`、`GET /api/v1/roundtables/{id}` |
| SSE 事件流 | `GET /api/v1/roundtables/{id}/events` — 流式推送讨论事件 |

内部实现了 `EventBus` 广播机制，支持多客户端实时订阅。

### `internal/config/`

配置管理。从 `config.yaml` 加载 YAML 配置，支持通过 `TALKABOUTIT_*` 前缀的环境变量覆盖配置项（如 API Key、端口等），并提供合理的默认值。

### `internal/engine/`

圆桌讨论编排引擎，核心业务逻辑所在：

- 按轮次（round）和发言顺序驱动各角色依次发言
- 调用 LLM Provider 生成流式响应（chunk by chunk）
- 将事件写入存储并通过 EventBus 广播给 SSE 客户端
- 管理讨论状态（进行中 / 已完成）

### `internal/llm/`

LLM 抽象层，通过接口和工厂模式实现多提供商支持：

- **`provider.go`**：定义 `Provider` 接口（`Chat` 流式方法）和消息类型
- **`factory.go`**：根据配置创建对应的 Provider 实例
- **`openai.go`**：OpenAI 兼容 API 实现（同时支持 DeepSeek 等兼容服务）
- **`anthropic.go`**：Anthropic Claude API 实现
- **`prompt.go`**：角色 Prompt 的组装逻辑
- **`errors.go`**：LLM 相关的错误类型封装

### `internal/persona/`

角色（Persona）领域模型与管理：

- **数据模型**：定义角色的 schema（名称、简介、性格特征、语言风格、知识领域等）
- **加载器**：从 `personas/` 目录读取 JSON 文件并解析为角色对象
- **校验**：对角色数据进行完整性和格式校验
- **上下文/状态**：管理每个角色在对话中的上下文和状态信息
- **Prompt 段落**：将角色属性转换为 LLM 可理解的 prompt 片段

### `internal/session/`

会话持久化层，基于 SQLite 实现：

- `roundtables` 表：存储圆桌会话元数据（话题、参与角色、状态等）
- `messages` 表：存储各角色的发言记录
- `roundtable_events` 表：存储事件流数据，支持 SSE 断线重连（`GetEventsAfter`）

### `personas/`

预置角色资源目录，每个 JSON 文件描述一个 AI 角色的完整配置（性格、语言风格、知识背景等）。当前内置 5 个角色：Elon Musk、Naval Ravikant、Steve Jobs、张小龙、张一鸣。

### `scripts/`

构建辅助脚本，提供本地编译和跨平台交叉编译能力（linux/amd64、darwin/amd64、darwin/arm64），以及测试、评测的快捷命令。

### `test/`

测试目录，包含：

- **集成测试**（`integration_test.go`）：通过 `httptest` 模拟完整的 HTTP + SSE 流程
- **真实 LLM 测试**（`realllm_test.go`）：调用真实 LLM API 的端到端测试

### `test/eval/`

评测框架，用于评估角色发言质量：

- **规则打分**（`scorer.go`）：基于关键词、语言风格等规则对发言打分
- **LLM 裁判**（`llm_judge.go`）：使用 LLM 作为裁判评估发言的角色一致性
- **测试数据**（`questions.json`、`baseline.json`）：评测用的话题集和基线数据

---

## 4. 技术栈

| 类别 | 技术 |
|------|------|
| 编程语言 | Go 1.26 |
| HTTP 服务 | Go 标准库 `net/http` |
| 实时通信 | Server-Sent Events (SSE) |
| 数据库 | SQLite（纯 Go 驱动，无 CGO 依赖） |
| 配置管理 | YAML 配置文件 + 环境变量覆盖 |
| 构建工具 | Makefile + shell 脚本 |
| 容器化 | Docker 多阶段构建（Alpine） |
| 测试 | Go 标准 `testing` + `httptest` |
| LLM 集成 | OpenAI 兼容 API、Anthropic API（HTTP 直接调用，无 SDK 依赖） |

---

## 5. 使用到的第三方包

### 直接依赖

| 包 | 版本 | 用途 |
|----|------|------|
| `gopkg.in/yaml.v3` | v3.0.1 | YAML 配置文件解析 |
| `modernc.org/sqlite` | v1.50.1 | 纯 Go 实现的 SQLite 驱动，无需 CGO |

### 间接依赖（由直接依赖引入）

| 包 | 版本 | 用途 |
|----|------|------|
| `github.com/dustin/go-humanize` | v1.0.1 | 数据格式化（human-readable） |
| `github.com/google/uuid` | v1.6.0 | UUID 生成 |
| `github.com/mattn/go-isatty` | v0.0.20 | 终端检测 |
| `github.com/ncruces/go-strftime` | v1.0.0 | 时间格式化 |
| `github.com/remyoudompheng/bigfft` | v0.0.0-20230129092748 | 大数 FFT 运算（SQLite 内部依赖） |
| `golang.org/x/sys` | v0.42.0 | 系统调用接口 |
| `modernc.org/libc` | v1.72.3 | C 标准库的 Go 实现（SQLite 底层） |
| `modernc.org/mathutil` | v1.7.1 | 数学工具（SQLite 底层） |
| `modernc.org/memory` | v1.11.0 | 内存管理（SQLite 底层） |

> **设计亮点**：项目仅有 2 个直接依赖，LLM API 调用通过 Go 标准库 `net/http` 直接实现，未引入任何 LLM SDK，保持了极简的依赖树和零 CGO 约束。
