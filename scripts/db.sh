#!/bin/bash

# SmashQueue - Database Script
# Manage PostgreSQL database

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Load environment
if [ -f "$PROJECT_DIR/.env" ]; then
    export $(grep -v '^#' "$PROJECT_DIR/.env" | xargs)
fi

# Defaults
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
DB_NAME=${DB_NAME:-smashqueue}

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

usage() {
    echo "SmashQueue Database Manager"
    echo ""
    echo "Usage: $0 <command>"
    echo ""
    echo "Commands:"
    echo "  start      Start PostgreSQL container"
    echo "  stop       Stop PostgreSQL container"
    echo "  status     Check database status"
    echo "  create     Create the smashqueue database"
    echo "  migrate    Run database migrations"
    echo "  seed       Seed database with sample data"
    echo "  reset      Reset database (drop and recreate)"
    echo "  connect    Connect to database via psql"
    echo ""
}

start_db() {
    # First check if PostgreSQL is already running locally
    if pg_isready -h $DB_HOST -p $DB_PORT > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} PostgreSQL is already running at $DB_HOST:$DB_PORT"
        echo -e "${YELLOW}→${NC} No need to start Docker container"
        return 0
    fi

    # Try Docker if local PostgreSQL is not available
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}✗${NC} Docker not found and PostgreSQL not running locally"
        echo "Please install PostgreSQL or Docker"
        return 1
    fi

    echo -e "${BLUE}→${NC} Starting PostgreSQL container..."
    docker run -d \
        --name smashqueue-postgres \
        -e POSTGRES_USER=$DB_USER \
        -e POSTGRES_PASSWORD=$DB_PASSWORD \
        -e POSTGRES_DB=$DB_NAME \
        -p $DB_PORT:5432 \
        -v smashqueue_pgdata:/var/lib/postgresql/data \
        postgres:16-alpine
    
    echo -e "${BLUE}→${NC} Waiting for PostgreSQL to be ready..."
    sleep 3
    for i in {1..30}; do
        if docker exec smashqueue-postgres pg_isready -U $DB_USER > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC} PostgreSQL is ready"
            return 0
        fi
        sleep 1
    done
    echo -e "${RED}✗${NC} PostgreSQL failed to start"
    return 1
}

stop_db() {
    echo -e "${BLUE}→${NC} Stopping PostgreSQL container..."
    docker stop smashqueue-postgres 2>/dev/null || true
    docker rm smashqueue-postgres 2>/dev/null || true
    echo -e "${GREEN}✓${NC} PostgreSQL stopped"
}

status_db() {
    if docker ps --filter "name=smashqueue-postgres" --format "{{.Names}}" | grep -q smashqueue-postgres; then
        echo -e "${GREEN}✓${NC} PostgreSQL container is running"
        docker exec smashqueue-postgres pg_isready -U $DB_USER
    else
        echo -e "${YELLOW}→${NC} PostgreSQL container is not running"
        echo ""
        echo "Checking for local PostgreSQL..."
        if pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER > /dev/null 2>&1; then
            echo -e "${GREEN}✓${NC} Local PostgreSQL is running at $DB_HOST:$DB_PORT"
        else
            echo -e "${YELLOW}→${NC} No PostgreSQL found. Run: $0 start"
        fi
    fi
}

create_db() {
    echo -e "${BLUE}→${NC} Creating database $DB_NAME..."
    
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "CREATE DATABASE $DB_NAME;" 2>/dev/null || \
        echo -e "${YELLOW}→${NC} Database already exists or connection failed"
    
    echo -e "${GREEN}✓${NC} Database ready"
}

migrate_db() {
    echo -e "${BLUE}→${NC} Running database migrations..."
    
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << 'EOF'
-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL DEFAULT '',
    bio TEXT DEFAULT '',
    role VARCHAR(50) NOT NULL DEFAULT 'player',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User stats table
CREATE TABLE IF NOT EXISTS user_stats (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_matches INTEGER DEFAULT 0,
    wins INTEGER DEFAULT 0,
    losses INTEGER DEFAULT 0,
    win_rate DECIMAL(5,2) DEFAULT 0,
    current_streak INTEGER DEFAULT 0,
    skill_level VARCHAR(50) DEFAULT 'Beginner',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Matches table
CREATE TABLE IF NOT EXISTS matches (
    id SERIAL PRIMARY KEY,
    court VARCHAR(100),
    team1 INTEGER[] NOT NULL,
    team2 INTEGER[] NOT NULL,
    result VARCHAR(50) DEFAULT 'pending',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Match scores table
CREATE TABLE IF NOT EXISTS match_scores (
    id SERIAL PRIMARY KEY,
    match_id INTEGER REFERENCES matches(id) ON DELETE CASCADE,
    game_number INTEGER NOT NULL,
    team1_score INTEGER DEFAULT 0,
    team2_score INTEGER DEFAULT 0
);

-- Queue entries table
CREATE TABLE IF NOT EXISTS queue_entries (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    status VARCHAR(50) DEFAULT 'waiting',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    called_at TIMESTAMP WITH TIME ZONE
);

-- Refresh tokens table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    revoked_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_queue_entries_status ON queue_entries(status);
CREATE INDEX IF NOT EXISTS idx_matches_result ON matches(result);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);

-- Updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to users table
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

EOF

    echo -e "${GREEN}✓${NC} Migrations complete"
}

seed_db() {
    echo -e "${BLUE}→${NC} Seeding database with sample data..."
    
    # The password hash is for "Admin@123!" using Argon2id
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME << 'EOF'
-- Insert admin user (password: Admin@123!)
INSERT INTO users (username, password_hash, name, bio, role)
VALUES ('kong@admin', 'PLACEHOLDER_HASH', 'Super Admin', 'System Administrator', 'admin')
ON CONFLICT (username) DO NOTHING;

-- Insert sample players
INSERT INTO users (username, password_hash, name, bio, role)
VALUES 
    ('player1', 'PLACEHOLDER_HASH', 'John Doe', 'Love badminton!', 'player'),
    ('player2', 'PLACEHOLDER_HASH', 'Jane Smith', 'Doubles specialist', 'player'),
    ('organizer1', 'PLACEHOLDER_HASH', 'Mike Organizer', 'Hua Guan', 'organizer')
ON CONFLICT (username) DO NOTHING;

-- Initialize stats for users
INSERT INTO user_stats (user_id, total_matches, wins, losses, win_rate, skill_level)
SELECT id, 0, 0, 0, 0, 'Beginner' FROM users
ON CONFLICT (user_id) DO NOTHING;

EOF

    echo -e "${GREEN}✓${NC} Database seeded"
    echo -e "${YELLOW}!${NC} Note: Update password hashes for seeded users"
}

reset_db() {
    echo -e "${RED}⚠${NC}  This will DELETE all data. Are you sure? (y/N)"
    read -r response
    if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
        echo -e "${BLUE}→${NC} Dropping database..."
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c "DROP DATABASE IF EXISTS $DB_NAME;" 2>/dev/null || true
        create_db
        migrate_db
        echo -e "${GREEN}✓${NC} Database reset complete"
    else
        echo "Cancelled"
    fi
}

connect_db() {
    echo -e "${BLUE}→${NC} Connecting to database..."
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME
}

# Main
case "${1:-}" in
    start)
        start_db
        ;;
    stop)
        stop_db
        ;;
    status)
        status_db
        ;;
    create)
        create_db
        ;;
    migrate)
        migrate_db
        ;;
    seed)
        seed_db
        ;;
    reset)
        reset_db
        ;;
    connect)
        connect_db
        ;;
    *)
        usage
        ;;
esac
