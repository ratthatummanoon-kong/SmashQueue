#!/bin/bash

# SmashQueue - Setup Script
# Run this once to initialize the project

set -e

echo "🏸 SmashQueue Setup"
echo "==================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
echo ""
echo "Checking prerequisites..."

check_command() {
    if command -v $1 &> /dev/null; then
        echo -e "${GREEN}✓${NC} $1 found"
        return 0
    else
        echo -e "${YELLOW}✗${NC} $1 not found"
        return 1
    fi
}

check_command "go" || { echo "Please install Go 1.22+"; exit 1; }
check_command "node" || { echo "Please install Node.js LTS"; exit 1; }
check_command "npm" || { echo "Please install npm"; exit 1; }
check_command "docker" || echo "Docker not found (optional, for containerized deployment)"

# Create environment files
echo ""
echo "Setting up environment files..."

if [ ! -f .env ]; then
    cp .env.example .env
    echo -e "${GREEN}✓${NC} Created .env from .env.example"
else
    echo -e "${YELLOW}→${NC} .env already exists, skipping"
fi

if [ ! -f backend/.env ]; then
    cp backend/.env.example backend/.env
    echo -e "${GREEN}✓${NC} Created backend/.env"
else
    echo -e "${YELLOW}→${NC} backend/.env already exists, skipping"
fi

if [ ! -f frontend/astro/.env.local ]; then
    cp frontend/astro/.env.example frontend/astro/.env.local
    echo -e "${GREEN}✓${NC} Created frontend/astro/.env.local"
else
    echo -e "${YELLOW}→${NC} frontend/astro/.env.local already exists, skipping"
fi

# Install backend dependencies
echo ""
echo "Installing backend dependencies..."
cd backend
go mod tidy
echo -e "${GREEN}✓${NC} Backend dependencies installed"
cd ..

# Install frontend dependencies
echo ""
echo "Installing frontend dependencies..."
cd frontend/astro
npm install
echo -e "${GREEN}✓${NC} Frontend dependencies installed"
cd ../..

echo ""
echo "=========================================="
echo -e "${GREEN}✓ Setup complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Edit .env files with your configuration"
echo "  2. Start PostgreSQL (or use: ./scripts/db.sh start)"
echo "  3. Run: ./scripts/start.sh"
echo ""
