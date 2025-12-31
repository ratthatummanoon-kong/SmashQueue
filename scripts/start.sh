#!/bin/bash

# SmashQueue - Start Script
# Starts both frontend and backend services

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🏸 Starting SmashQueue..."
echo ""

# Check if .env exists
if [ ! -f "$PROJECT_DIR/backend/.env" ]; then
    echo -e "${YELLOW}⚠${NC}  No backend/.env found. Run ./scripts/setup.sh first"
    exit 1
fi

# Create logs directory
mkdir -p "$PROJECT_DIR/logs"

# Start Backend
echo -e "${BLUE}→${NC} Starting Go backend on port 8080..."
cd "$PROJECT_DIR/backend"
go run main.go > "$PROJECT_DIR/logs/backend.log" 2>&1 &
BACKEND_PID=$!
echo $BACKEND_PID > "$PROJECT_DIR/logs/backend.pid"
echo -e "${GREEN}✓${NC} Backend started (PID: $BACKEND_PID)"

# Wait for backend to be ready
echo -e "${BLUE}→${NC} Waiting for backend to be ready..."
sleep 2
for i in {1..10}; do
    if curl -s http://localhost:8080/api/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Backend is ready"
        break
    fi
    if [ $i -eq 10 ]; then
        echo -e "${YELLOW}⚠${NC}  Backend health check failed, but continuing..."
    fi
    sleep 1
done

# Start Frontend
echo -e "${BLUE}→${NC} Starting Next.js frontend on port 3000..."
cd "$PROJECT_DIR/frontend/astro"
npm run dev > "$PROJECT_DIR/logs/frontend.log" 2>&1 &
FRONTEND_PID=$!
echo $FRONTEND_PID > "$PROJECT_DIR/logs/frontend.pid"
echo -e "${GREEN}✓${NC} Frontend started (PID: $FRONTEND_PID)"

echo ""
echo "=========================================="
echo -e "${GREEN}✓ SmashQueue is running!${NC}"
echo ""
echo "  Frontend: http://localhost:3000"
echo "  Backend:  http://localhost:8080"
echo "  Health:   http://localhost:8080/api/health"
echo ""
echo "  Logs: ./logs/backend.log, ./logs/frontend.log"
echo "  Stop: ./scripts/stop.sh"
echo ""
