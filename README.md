# TalkAboutIt — Multi-LLM Roundtable Discussion

> Watch AI personas debate. Steve Jobs on design. Elon Musk on physics. Naval Ravikant on leverage. In one virtual roundtable.

TalkAboutIt brings legendary thinkers back to life as AI personas, seating them around a virtual table to debate the topics you choose. Each persona maintains a distinct voice, beliefs, and debating style — and they actually listen to and respond to each other.

[📖 中文文档](README.zh-CN.md)

## ✨ Features

- **5 built-in personas** — Steve Jobs, Elon Musk, Naval Ravikant, Zhang Xiaolong (WeChat), Zhang Yiming (ByteDance)
- **Multi-round debates** — Personas reference each other's arguments and evolve their positions
- **Per-persona conversation context** — Uses DeepSeek's automatic KV-cache for efficient multi-turn reasoning
- **System-level i18n** — English & Simplified Chinese UI, with debate language following the selected system language
- **Real-time streaming** — Watch the debate unfold via Server-Sent Events (SSE)
- **History replay** — Browse and replay past debates
- **Zero-cost caching** — DeepSeek Context Caching is enabled by default, no configuration needed

## 🚀 Quick Start

### Prerequisites

- **Go** 1.26+
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
  db_path: data/sessions.db    # SQLite database path
  max_rounds: 3              # 1-5 rounds per debate
```

### Development checks

```bash
cd backend && go test ./...
cd frontend && npm run lint && npm run test && npm run build
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
│  React Router + Zustand      │     │  Engine                   │
│  ├─ TopicPanel               │ SSE │  ├─ Run()                 │
│  ├─ PersonaSelector          │◄───►│  ├─ LLMGenerate()         │
│  ├─ Header Language Menu     │     │  └─ ConversationContext   │
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
