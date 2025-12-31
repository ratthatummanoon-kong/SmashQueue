#!/bin/bash

# SmashQueue - Stop Script
# Stops all running services

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🏸 Stopping SmashQueue..."
echo ""

# Stop Backend
if [ -f "$PROJECT_DIR/logs/backend.pid" ]; then
    PID=$(cat "$PROJECT_DIR/logs/backend.pid")
    if ps -p $PID > /dev/null 2>&1; then
        kill $PID 2>/dev/null
        echo -e "${GREEN}✓${NC} Backend stopped (PID: $PID)"
    else
        echo -e "${YELLOW}→${NC} Backend already stopped"
    fi
    rm -f "$PROJECT_DIR/logs/backend.pid"
else
    # Try to find and kill go processes
    pkill -f "go run main.go" 2>/dev/null && echo -e "${GREEN}✓${NC} Backend processes stopped" || echo -e "${YELLOW}→${NC} No backend process found"
fi

# Stop Frontend
if [ -f "$PROJECT_DIR/logs/frontend.pid" ]; then
    PID=$(cat "$PROJECT_DIR/logs/frontend.pid")
    if ps -p $PID > /dev/null 2>&1; then
        kill $PID 2>/dev/null
        echo -e "${GREEN}✓${NC} Frontend stopped (PID: $PID)"
    else
        echo -e "${YELLOW}→${NC} Frontend already stopped"
    fi
    rm -f "$PROJECT_DIR/logs/frontend.pid"
else
    # Try to find and kill npm dev processes on port 3000
    lsof -ti:3000 | xargs kill 2>/dev/null && echo -e "${GREEN}✓${NC} Frontend processes stopped" || echo -e "${YELLOW}→${NC} No frontend process found"
fi

echo ""
echo -e "${GREEN}✓ All services stopped${NC}"
