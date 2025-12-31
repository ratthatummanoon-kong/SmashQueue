#!/bin/bash

# SmashQueue - Data Cleanup Script
# Deletes all data except super admin user

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${RED}🗑️  SmashQueue Data Cleanup${NC}"
echo "=================================="
echo ""

# Load environment variables if .env exists
if [ -f "$PROJECT_DIR/.env" ]; then
    export $(cat "$PROJECT_DIR/.env" | grep -v '^#' | xargs)
    echo -e "${GREEN}✓${NC} Loaded environment from .env"
fi

# Database configuration
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

# Count current data
echo "Current database contents:"
USER_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM users;" 2>/dev/null | xargs)
MATCH_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM matches;" 2>/dev/null | xargs)
SCORE_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM match_scores;" 2>/dev/null | xargs)

echo "  Users: $USER_COUNT"
echo "  Matches: $MATCH_COUNT"
echo "  Scores: $SCORE_COUNT"
echo ""

# Warning
echo -e "${YELLOW}⚠️  WARNING:${NC}"
echo "This will delete ALL data from the database"
echo "Only the super admin (kong@admin) will be preserved"
echo ""
read -p "Are you sure? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
    echo -e "${BLUE}Operation cancelled${NC}"
    exit 0
fi

echo ""
echo -e "${BLUE}Cleaning database...${NC}"

# Delete data in correct order (respecting foreign keys)
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" <<EOF
-- Start transaction
BEGIN;

-- Save admin password hash
CREATE TEMP TABLE temp_admin AS 
SELECT password_hash FROM users WHERE username = 'kong@admin';

-- Delete all data (cascade will handle foreign keys)
DELETE FROM match_scores;
DELETE FROM matches;
DELETE FROM queue_entries;
DELETE FROM refresh_tokens;
DELETE FROM user_stats;
DELETE FROM users;

-- Reset all sequences to 1
ALTER SEQUENCE users_id_seq RESTART WITH 1;
ALTER SEQUENCE matches_id_seq RESTART WITH 1;
ALTER SEQUENCE match_scores_id_seq RESTART WITH 1;
ALTER SEQUENCE queue_entries_id_seq RESTART WITH 1;
ALTER SEQUENCE refresh_tokens_id_seq RESTART WITH 1;

-- Recreate admin user with ID 1
INSERT INTO users (username, password_hash, name, phone, bio, role, skill_tier, is_active, created_at, updated_at)
SELECT 'kong@admin', password_hash, 'Super Admin', '0899999999', 'System Administrator', 'admin', 'A', true, NOW(), NOW()
FROM temp_admin;

-- Create admin stats
INSERT INTO user_stats (user_id, total_matches, wins, losses, win_rate, skill_level, skill_points)
VALUES (1, 0, 0, 0, 0, 'Expert', 0);

-- Commit transaction
COMMIT;

-- Show summary
SELECT 'Cleanup complete!' as message;
SELECT COUNT(*) as remaining_users FROM users;
SELECT COUNT(*) as remaining_matches FROM matches;
SELECT COUNT(*) as remaining_scores FROM match_scores;
EOF

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ Database cleaned successfully${NC}"
    echo ""
    echo "Remaining data:"
    REMAINING_USERS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM users;" 2>/dev/null | xargs)
    REMAINING_MATCHES=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM matches;" 2>/dev/null | xargs)
    echo "  Users: $REMAINING_USERS (super admin only)"
    echo "  Matches: $REMAINING_MATCHES"
    echo ""
    echo -e "${BLUE}Admin credentials:${NC}"
    echo "  Username: kong@admin"
    echo "  Password: Admin@123!"
else
    echo ""
    echo -e "${RED}✗ Failed to clean database${NC}"
    exit 1
fi
