.PHONY: help dev dev-backend dev-frontend build up down logs clean test

# Default target
help:
	@echo "SmashQueue - Badminton Queue Management System"
	@echo ""
	@echo "Usage:"
	@echo "  make dev            Start both frontend and backend in development mode"
	@echo "  make dev-backend    Start backend only (Go)"
	@echo "  make dev-frontend   Start frontend only (Next.js)"
	@echo "  make build          Build Docker images"
	@echo "  make up             Start all services with Docker Compose"
	@echo "  make down           Stop all Docker services"
	@echo "  make logs           View Docker logs"
	@echo "  make clean          Remove Docker volumes and images"
	@echo "  make test           Run tests"
	@echo ""

# Development
dev:
	@echo "Starting development servers..."
	@make -j2 dev-backend dev-frontend

dev-backend:
	@echo "Starting Go backend on :8080..."
	cd backend && go run main.go

dev-frontend:
	@echo "Starting Next.js frontend on :3000..."
	cd frontend/astro && npm run dev

# Docker
build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

clean:
	docker compose down -v --rmi local
	@echo "Cleaned up Docker resources"

# Testing
test:
	@echo "Running backend tests..."
	cd backend && go test ./...
	@echo "Running frontend tests..."
	cd frontend/astro && npm test || true

# Database
db-reset:
	docker compose down -v postgres
	docker compose up -d postgres
	@echo "Database reset complete"

# Setup
setup:
	@echo "Setting up SmashQueue..."
	@if [ ! -f .env ]; then cp .env.example .env && echo "Created .env from .env.example"; fi
	@if [ ! -f backend/.env ]; then cp backend/.env.example backend/.env && echo "Created backend/.env"; fi
	@if [ ! -f frontend/astro/.env.local ]; then cp frontend/astro/.env.example frontend/astro/.env.local && echo "Created frontend/.env.local"; fi
	cd backend && go mod tidy
	cd frontend/astro && npm install
	@echo "Setup complete! Run 'make dev' to start development servers."
