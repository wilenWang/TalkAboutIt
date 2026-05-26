#!/usr/bin/env bash
set -e

BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo -e "${BOLD}✦ TalkAboutIt — One-Click Startup${NC}\n"

# ---------- Prerequisites ----------
check_cmd() {
    if ! command -v "$1" &>/dev/null; then
        echo -e "${RED}✗ $1 not found. Please install $1 first.${NC}"
        exit 1
    fi
}

check_cmd go
check_cmd node
check_cmd npm

# ---------- API Key ----------
if [ ! -f "$PROJECT_DIR/.env" ]; then
    if [ -f "$PROJECT_DIR/.env.example" ]; then
        cp "$PROJECT_DIR/.env.example" "$PROJECT_DIR/.env"
    fi
fi

# Source .env if exists
if [ -f "$PROJECT_DIR/.env" ]; then
    set -a; source "$PROJECT_DIR/.env"; set +a
fi

if [ -z "$DEEPSEEK_API_KEY" ]; then
    echo -e "${YELLOW}⚠ DEEPSEEK_API_KEY not set. Running in mock mode (no real LLM).${NC}"
    echo -e "${YELLOW}  Set it in .env or export DEEPSEEK_API_KEY=sk-...${NC}\n"
fi

# ---------- Config ----------
if [ ! -f "$PROJECT_DIR/backend/config.yaml" ]; then
    echo -e "${YELLOW}config.yaml not found, copying from config.example.yaml${NC}"
    cp "$PROJECT_DIR/backend/config.example.yaml" "$PROJECT_DIR/backend/config.yaml"
fi

# ---------- Install Deps ----------
echo -e "${GREEN}➤ Installing frontend dependencies...${NC}"
cd "$PROJECT_DIR/frontend" && npm install --silent

# ---------- Start Backend ----------
echo -e "${GREEN}➤ Starting backend (port 8080)...${NC}"
cd "$PROJECT_DIR/backend"
DEEPSEEK_API_KEY="${DEEPSEEK_API_KEY}" go run ./cmd/server &
BACKEND_PID=$!

# ---------- Start Frontend ----------
echo -e "${GREEN}➤ Starting frontend (port 5173)...${NC}"
cd "$PROJECT_DIR/frontend"
npx vite --port 5173 --host 0.0.0.0 &
FRONTEND_PID=$!

# ---------- Cleanup ----------
cleanup() {
    echo -e "\n${YELLOW}Shutting down...${NC}"
    kill $BACKEND_PID 2>/dev/null
    kill $FRONTEND_PID 2>/dev/null
    wait $BACKEND_PID 2>/dev/null
    wait $FRONTEND_PID 2>/dev/null
    echo -e "${GREEN}Done.${NC}"
}
trap cleanup EXIT INT TERM

echo ""
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✦ TalkAboutIt is running!${NC}"
echo -e "  Frontend: ${BOLD}http://localhost:5173${NC}"
echo -e "  Backend:  ${BOLD}http://localhost:8080${NC}"
echo -e "  Press ${BOLD}Ctrl+C${NC} to stop"
echo -e "${BOLD}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

wait
