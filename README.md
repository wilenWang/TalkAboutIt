# TalkAboutIt — Multi-LLM Roundtable Discussion

> Watch AI personas debate. Steve Jobs on design. Elon Musk on physics. Naval Ravikant on leverage. In one virtual roundtable.

TalkAboutIt brings legendary thinkers back to life as AI personas, seating them around a virtual table to debate the topics you choose. Each persona maintains a distinct voice, beliefs, and debating style — and they actually listen to and respond to each other.

[English](#english) | [简体中文](#简体中文)

---

<a name="english"></a>

## ✨ Features

- **5 built-in personas** — Steve Jobs, Elon Musk, Naval Ravikant, Zhang Xiaolong (WeChat), Zhang Yiming (ByteDance)
- **Multi-round debates** — Personas reference each other's arguments and evolve their positions
- **Per-persona conversation context** — Uses DeepSeek's automatic KV-cache for efficient multi-turn reasoning
- **System-level i18n** — English & Simplified Chinese UI, debate language independently selectable
- **Real-time streaming** — Watch the debate unfold via Server-Sent Events (SSE)
- **History replay** — Browse and replay past debates
- **Zero-cost caching** — DeepSeek Context Caching is enabled by default, no configuration needed

## 🚀 Quick Start

### Prerequisites

- **Go** 1.21+
- **Node.js** 18+
- **DeepSeek API Key** — [Get one free](https://platform.deepseek.com/api_keys)

### One-Click

```bash
# 1. Set your API key
export DEEPSEEK_API_KEY=sk-your-key-here

# 2. Start everything
chmod +x start.sh && ./start.sh
```

Then open **http://localhost:5173** and start a debate.

### Manual

```bash
# Terminal 1 — Backend
cd backend
export DEEPSEEK_API_KEY=sk-your-key-here
go run ./cmd/server

# Terminal 2 — Frontend
cd frontend
npm install
npm run dev
```

## ⚙️ Configuration

Copy and edit the config:

```bash
cp backend/config.example.yaml backend/config.yaml
```

### Key settings

```yaml
llm:
  default: deepseek          # deepseek | openai | claude
  providers:
    deepseek:
      type: openai
      base_url: https://api.deepseek.com
      model: deepseek-v4-pro
      # api_key is read from DEEPSEEK_API_KEY env var

session:
  max_rounds: 3              # 1-5 rounds per debate
```

### Supported Providers

| Provider | Env Variable | Model Example |
|----------|-------------|---------------|
| DeepSeek | `DEEPSEEK_API_KEY` | `deepseek-v4-pro` |
| OpenAI | `OPENAI_API_KEY` | `gpt-4o` |
| Anthropic | `ANTHROPIC_API_KEY` | `claude-sonnet-4-20250514` |

> **Tip:** If no API key is set, TalkAboutIt falls back to mock mode — perfect for UI development.

## 🏗 Architecture

```
frontend (React + TypeScript + Vite)     backend (Go)
┌─────────────────────────────┐     ┌──────────────────────────┐
│  App.tsx                     │     │  Engine                   │
│  ├─ PersonaSelector          │ SSE │  ├─ Run()                 │
│  ├─ TopicInput               │◄───►│  ├─ LLMGenerate()         │
│  ├─ LanguageToggle           │     │  └─ ConversationContext   │
│  ├─ MessageStream            │     │                           │
│  └─ i18n/LanguageContext     │     │  Per-Persona Context      │
│                              │     │  ┌───────┬───────┬───────┐│
│                              │     │  │ Steve │ Elon  │ Naval ││
│                              │     │  │ [sys] │ [sys] │ [sys] ││
│                              │     │  │ [usr] │ [usr] │ [usr] ││
│   DeepSeek KV-Cache          │     │  │ [asst]│ [asst]│ [asst]││
│   (auto, free, disk-based)   │     │  └───────┴───────┴───────┘│
└─────────────────────────────┘     └──────────────────────────┘
```

## 📁 Project Structure

```
TalkAboutIt/
├── backend/                   # Go API server
│   ├── cmd/server/            # Entry point
│   ├── internal/
│   │   ├── api/               # HTTP handlers + SSE
│   │   ├── config/            # YAML config loader
│   │   ├── engine/            # Debate orchestrator
│   │   ├── llm/               # Provider abstraction (OpenAI/Anthropic)
│   │   ├── persona/           # Persona schema + conversation context
│   │   └── session/           # SQLite persistence
│   ├── personas/              # Persona JSON definitions
│   └── config.example.yaml
├── frontend/                  # React SPA
│   └── src/
│       ├── components/        # UI components
│       ├── pages/             # Page-level components
│       └── i18n/              # Language context
├── docs/                      # Design documents
├── start.sh                   # One-click startup
└── .env.example               # Environment template
```

## 🔧 Adding a New Persona

Create a JSON file in `backend/personas/`:

```json
{
  "id": "alan-turing",
  "name": "Alan Turing",
  "display_name": "Alan Turing",
  "avatar": "⚙️",
  "role_title": "Mathematician, Computer Science Pioneer",
  "description": "Father of theoretical computer science and AI...",
  "language": { "primary": "en-US", "allowed": ["en-US", "zh-CN"] },
  "stance": { "default_position": "...", "intensity": 4, "biases": [...], "taboos": [...] },
  "core_beliefs": [
    { "belief": "...", "priority": 5, "rationale": "..." }
  ],
  "speaking_style": { "tone": "analytical, precise", "cadence": "measured", "verbosity": 3 },
  "knowledge_scope": { "domains": [...], "expertise_level": {...} },
  "interaction_rules": { "address_others": "...", "disagreement_style": "..." },
  "debate_goal": { "primary_goal": "..." },
  "prompting": { "system_preamble": "..." },
  "examples": { "opening_line": "...", "sample_rebuttal": "..." }
}
```

## 📄 License

MIT

---

<a name="简体中文"></a>

## ✨ 功能特性

- **5 个内置人物** — Steve Jobs、Elon Musk、Naval Ravikant、张小龙、张一鸣
- **多轮辩论** — 人物互相引用、观点逐步演进
- **按人物拆分对话上下文** — 利用 DeepSeek 自动 KV-Cache，跨轮次高效复用
- **系统级国际化** — 英文 / 简体中文 UI，辩论语言独立可选
- **实时流式输出** — 通过 SSE 实时观看辩论过程
- **历史记录回放** — 浏览和回放过往辩论
- **零成本缓存** — DeepSeek 上下文缓存默认开启，无需额外配置

## 🚀 快速启动

### 环境要求

- **Go** 1.21+
- **Node.js** 18+
- **DeepSeek API Key** — [免费获取](https://platform.deepseek.com/api_keys)

### 一键启动

```bash
export DEEPSEEK_API_KEY=sk-your-key-here
chmod +x start.sh && ./start.sh
```

浏览器打开 **http://localhost:5173**，即可开始辩论。

### 手动启动

```bash
# 终端 1 — 后端
cd backend
export DEEPSEEK_API_KEY=sk-your-key-here
go run ./cmd/server

# 终端 2 — 前端
cd frontend
npm install
npm run dev
```

## ⚙️ 核心配置

```bash
cp backend/config.example.yaml backend/config.yaml
```

```yaml
llm:
  default: deepseek          # 可选 deepseek | openai | claude
  providers:
    deepseek:
      base_url: https://api.deepseek.com
      model: deepseek-v4-pro

session:
  max_rounds: 3              # 每场辩论 1-5 轮
```

### 支持的模型供应商

| 供应商 | 环境变量 | 模型示例 |
|--------|---------|---------|
| DeepSeek | `DEEPSEEK_API_KEY` | `deepseek-v4-pro` |
| OpenAI | `OPENAI_API_KEY` | `gpt-4o` |
| Anthropic | `ANTHROPIC_API_KEY` | `claude-sonnet-4-20250514` |

> 未设置 API Key 时自动降级为 mock 模式，适合前端开发调试。

## 🔧 添加新人物

在 `backend/personas/` 下创建 JSON 文件，格式见上方 [Adding a New Persona](#-adding-a-new-persona)。

## 📄 开源协议

MIT
