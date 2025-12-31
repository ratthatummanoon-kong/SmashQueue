#!/bin/bash

# SmashQueue - Mock Data Generator Script
# Generates users, matches, and scores for testing

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🏸 SmashQueue Mock Data Generator${NC}"
echo "=================================="
echo ""

# Load environment variables if .env exists
if [ -f "$PROJECT_DIR/.env" ]; then
    export $(cat "$PROJECT_DIR/.env" | grep -v '^#' | xargs)
    echo -e "${GREEN}✓${NC} Loaded environment from .env"
fi

# Check if database is accessible
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-smashqueue}
DB_USER=${DB_USER:-kong}

echo ""
echo "Database Configuration:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo ""

# Check PostgreSQL connection
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; then
    echo -e "${RED}✗ Cannot connect to database${NC}"
    echo "Please ensure PostgreSQL is running and credentials are correct"
    exit 1
fi

echo -e "${GREEN}✓${NC} Database connection verified"
echo ""

# Run the mock data generator
cd "$PROJECT_DIR/backend/cmd/mock"

echo -e "${BLUE}Running mock data generator...${NC}"
echo ""

go run main.go

EXIT_CODE=$?

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    echo -e "${GREEN}✓ Mock data generated successfully${NC}"
else
    echo -e "${RED}✗ Mock data generation failed (exit code: $EXIT_CODE)${NC}"
    exit $EXIT_CODE
fi

echo ""
echo -e "${BLUE}Summary:${NC}"
echo "  - Users created with random names and stats"
echo "  - Matches generated with scores"
echo "  - Queue entries populated"
echo ""
echo -e "${GREEN}✓ Done!${NC}"
