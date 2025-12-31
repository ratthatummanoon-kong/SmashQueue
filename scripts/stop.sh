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

# Verify all services are stopped
echo ""
echo "🔍 Verifying service status..."
echo ""

BACKEND_RUNNING=false
FRONTEND_RUNNING=false

# Check backend (port 8080 and smashqueue process)
if lsof -ti:8080 > /dev/null 2>&1; then
    BACKEND_RUNNING=true
    echo -e "${YELLOW}⚠ Backend still running on port 8080${NC}"
elif pgrep -f "smashqueue" > /dev/null 2>&1; then
    BACKEND_RUNNING=true
    echo -e "${YELLOW}⚠ Backend process still running${NC}"
else
    echo -e "${GREEN}✓ Backend stopped${NC}"
fi

# Check frontend (port 3000)
if lsof -ti:3000 > /dev/null 2>&1; then
    FRONTEND_RUNNING=true
    echo -e "${YELLOW}⚠ Frontend still running on port 3000${NC}"
else
    echo -e "${GREEN}✓ Frontend stopped${NC}"
fi

echo ""
if [ "$BACKEND_RUNNING" = true ] || [ "$FRONTEND_RUNNING" = true ]; then
    echo -e "${YELLOW}⚠ Some services are still running. Forcing shutdown...${NC}"
    lsof -ti:8080 | xargs kill -9 2>/dev/null
    lsof -ti:3000 | xargs kill -9 2>/dev/null
    pkill -9 -f "smashqueue" 2>/dev/null
    sleep 1
    echo -e "${GREEN}✓ Force shutdown completed${NC}"
else
    echo -e "${GREEN}✓ All services confirmed stopped${NC}"
fi

# Clean up compiled binaries
echo ""
echo "🧹 Cleaning up build artifacts..."
if [ -f "$PROJECT_DIR/backend/smashqueue" ]; then
    rm -f "$PROJECT_DIR/backend/smashqueue"
    echo -e "${GREEN}✓ Removed backend binary${NC}"
fi

if [ -f "$PROJECT_DIR/backend/backend" ]; then
    rm -f "$PROJECT_DIR/backend/backend"
    echo -e "${GREEN}✓ Removed backend binary${NC}"
fi

echo ""
echo -e "${GREEN}✓ Cleanup complete${NC}"
