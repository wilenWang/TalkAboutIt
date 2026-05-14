# TalkAboutIt — 多 LLM 名人圆桌辩论

> 让 AI 名人同台辩论。Steve Jobs 谈设计品味，Elon Musk 谈物理约束，Naval Ravikant 谈长期复利。

TalkAboutIt 将传奇人物以 AI 人格的形式"复活"，让他们围坐在虚拟圆桌前，对你选定的话题展开多轮辩论。每个角色保持独特的声音、信念和辩论风格，并且会真正倾听和回应彼此。

[English](README.md)

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

## 🏗 架构

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

## 📁 项目结构

```
TalkAboutIt/
├── backend/                   # Go API 服务
│   ├── cmd/server/            # 入口
│   ├── internal/
│   │   ├── api/               # HTTP handlers + SSE
│   │   ├── config/            # YAML 配置加载
│   │   ├── engine/            # 辩论编排器
│   │   ├── llm/               # Provider 抽象 (OpenAI/Anthropic)
│   │   ├── persona/           # 人物 Schema + 对话上下文
│   │   └── session/           # SQLite 持久化
│   ├── personas/              # 人物 JSON 定义
│   └── config.example.yaml
├── frontend/                  # React SPA
│   └── src/
│       ├── components/        # UI 组件
│       ├── pages/             # 页面组件
│       └── i18n/              # 语言上下文
├── docs/                      # 设计文档
├── start.sh                   # 一键启动
└── .env.example               # 环境变量模板
```

## 🔧 添加新人物

在 `backend/personas/` 下创建 JSON 文件：

```json
{
  "id": "alan-turing",
  "name": "Alan Turing",
  "display_name": "Alan Turing",
  "avatar": "⚙️",
  "role_title": "Mathematician, Computer Science Pioneer",
  "description": "计算机科学和人工智能之父...",
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

## 📄 开源协议

MIT
