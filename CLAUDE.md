# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

TalkAboutIt is a multi-LLM roundtable discussion app. AI personas debate topics via streaming SSE. Stack: Go backend (SQLite, SSE), React 19 + Vite + Tailwind frontend.

## Common Commands

**Run everything (one-click):**
```bash
./start.sh
```
This starts the Go backend on port 8080 and the Vite dev server on port 5173, with API proxying via Vite config.

**Backend only:**
```bash
cd backend
export DEEPSEEK_API_KEY=sk-...
go run ./cmd/server
```

**Frontend only:**
```bash
cd frontend
npm install
npm run dev
```

**Build frontend for production:**
```bash
cd frontend
npm run build   # runs tsc -b && vite build
```

**Backend tests and build:**
```bash
cd backend
./scripts/build.sh test           # unit tests
go test ./... -v -count=1         # run a single test package
go test ./internal/engine -run TestEngine -v
./scripts/build.sh build          # compile binary to bin/
./scripts/build.sh build-all      # cross-compile linux/darwin
./scripts/build.sh integration    # integration tests (tags=integration)
./scripts/build.sh eval           # eval framework (tags=eval)
```

**Prerequisites:** Go 1.26+, Node 18+.

## Architecture

### Backend (Go)

**Module:** `github.com/wilenwang/talkaboutit`  
**SQLite:** Uses `modernc.org/sqlite` (CGO-free). WAL mode enabled. Database file at `backend/data/talkaboutit.db`.

**Debate Engine Flow (`internal/engine/engine.go`):**
The engine drives roundtables through a fixed event lifecycle:
1. `stream_start` → 2. `round_start` → 3. `speaking` → 4. `message_chunk` (streamed) → 5. `message_done` → 6. `round_end` → 7. `stream_done`

Each persona gets an independent `ConversationContext` (`internal/persona/context.go`). The engine maintains one context per persona; after each speaker finishes, their full message is appended to all *other* personas' contexts. Contexts are truncated to 24 messages max via `Truncate(24)`, which keeps the system prompt + a summary message + recent tail.

**SSE Event Bus (`internal/api/sse_handler.go`):**
- `EventBus` manages subscribers per roundtable ID (map of channels)
- Engine broadcasts events via `SetOnEvent` callback hooked into the bus
- SSE handler supports `Last-Event-ID` header/query param for reconnect replay
- History replay is fetched from `GetEventsAfter` before subscribing to live events
- Heartbeat every 15 seconds (`: keepalive`)
- Subscribe happens *before* history replay to avoid race conditions

**Mock Mode:**
If no `DEEPSEEK_API_KEY` (or configured provider key) is set, `llm.NewProvider` fails and the engine falls back to `DefaultMockGenerate`, which simply returns the persona's `opening_line` as a single chunk. This is the normal state for UI development.

**LLM Provider Abstraction (`internal/llm/`):**
- `Provider` interface: `Chat(ctx, req) (<-chan ChatChunk, error)`
- Implementations: `openai.go` (OpenAI-compatible, used for DeepSeek), `anthropic.go`
- `ProviderError` supports `Recoverable` flag; recoverable errors emit `message_aborted` and skip to next speaker
- Prompts split into static (`BuildStaticSystemPrompt`) and dynamic (`BuildDynamicContext`) parts. Static prompt is built once per persona and stored in `ConversationContext`. Dynamic prompt is rebuilt each turn with topic, peers, round number, and deduplication hints.

**Config Loading (`internal/config/config.go`):**
- Loads `backend/config.yaml` (copied from `config.example.yaml` if missing)
- Environment variables override with prefix `TALKABOUTIT_`:
  - `TALKABOUTIT_SERVER_PORT`, `TALKABOUTIT_SERVER_HOST`
  - `TALKABOUTIT_LLM_DEFAULT`
  - `TALKABOUTIT_LLM_{PROVIDER}_API_KEY`, `_BASE_URL`, `_MODEL`
  - `TALKABOUTIT_PERSONAS_DIR`, `TALKABOUTIT_SESSION_DB_PATH`, `TALKABOUTIT_SESSION_MAX_ROUNDS`
- API keys are typically read from env vars, not the YAML file.

**Persona Loader (`internal/persona/loader.go`):**
- Reads/writes JSON files directly from `backend/personas/` directory
- `LoadAll()` scans all `.json` files; `LoadOne(id)` reads `{id}.json`
- `Save` validates via `ValidateJSON` before writing
- Personas are file-based assets, not database rows

**Atomic Roundtable Start (`internal/api/sse_handler.go` `StartRoundtable`):**
- Uses `MarkRunning` (SQL `UPDATE ... WHERE status = 'pending'`) as a CAS operation to prevent concurrent double-starts
- Engine `Run()` is invoked in a goroutine after a 200ms delay to let the SSE client subscribe first

### Frontend (React + TypeScript + Vite)

**Routing:**
No router library. Client-side routing is done manually in `App.tsx` by parsing `window.location.pathname`:
- `/` → talk page
- `/history` → history list
- `/history/:id` → history detail
- `/personas` → persona management
URL sync uses `window.history.pushState` / `popstate`.

**API Client (`frontend/src/api/client.ts`):**
- `BASE = ''` — relies on Vite dev server proxy (`/api` → `http://localhost:8080`)
- Error keys returned match translation keys (e.g., `errFetchPersonas`) and are displayed via `t(key)`

**i18n System (`frontend/src/i18n/`):**
- `SystemLanguage = 'zh-CN' | 'en-US'`
- Translation keys follow strict prefixes:
  - `status.*` — connection badges
  - `action.*` — buttons/CTAs
  - `page.*` — page titles
  - `label.*` — field names
  - `msg.*` — messages/descriptions
  - `input.*` — placeholders
  - `err.*` — error messages
  - `fmt.*` — format patterns (use with `f(key, params)`)
  - `sep.*` — separators
- `t(key)` returns the string for the current language
- `f(key, params)` does typed parameter substitution (e.g., `f('fmtRoundLabel', { n: 3 })`)
- Language is persisted to `localStorage` under key `talkaboutit.system-language`

**SSE Hook (`frontend/src/hooks/useSSE.ts`):**
- Connects to `/api/v1/roundtables/{id}/events`
- Supports `Last-Event-ID` for resume
- Handles reconnection with exponential backoff

**State Management:**
No global state library. `App.tsx` holds the main state (selected personas, topic, rounds, messages, status). Pages receive props and callbacks. `messagesRef` mirrors `messages` state to avoid stale closures in the SSE callback.

## Key Conventions

- **Go module path:** `github.com/wilenwang/talkaboutit`
- **Backend runs on:** `:8080` (override with `PORT` env var)
- **Frontend dev server:** `:5173` (Vite proxy handles `/api`)
- **Persona schema version:** `persona.v1` (see `backend/internal/persona/persona.go`)
- **Debate language:** Controlled by the global system language toggle (upper-right corner). The `language` field on roundtables is set at creation time and determines which language the LLM speaks in.
- **Error handling:** Backend returns Chinese error messages in JSON; frontend maps these via translation keys.
