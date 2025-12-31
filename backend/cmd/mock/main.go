package main

import (
	"backend/database/generate"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	mathrand "math/rand"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/argon2"
)

func main() {
	// Seed random
	mathrand.Seed(time.Now().UnixNano())

	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "kong")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "smashqueue")
	dbSSL := getEnv("DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, dbSSL)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("🏸 SmashQueue Mock Data Generator")
	fmt.Println("==================================")
	fmt.Printf("Connected to database: %s@%s:%s/%s\n\n", dbUser, dbHost, dbPort, dbName)

	// Get user input
	var numPlayers int
	fmt.Print("Number of players to generate (10-100): ")
	fmt.Scanln(&numPlayers)
	if numPlayers < 10 {
		numPlayers = 10
	}
	if numPlayers > 100 {
		numPlayers = 100
	}

	var numMatches int
	fmt.Print("Number of matches to generate (20-500): ")
	fmt.Scanln(&numMatches)
	if numMatches < 20 {
		numMatches = 20
	}
	if numMatches > 500 {
		numMatches = 500
	}

	// Create tables if not exist
	if err := createTables(db); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Generate players
	fmt.Printf("\n📝 Generating %d players...\n", numPlayers)
	playerIDs, err := generatePlayers(db, numPlayers)
	if err != nil {
		log.Fatalf("Failed to generate players: %v", err)
	}
	fmt.Printf("✓ Created %d players\n", len(playerIDs))

	// Generate matches with scores
	fmt.Printf("\n🏸 Generating %d matches with scores...\n", numMatches)
	matchCount, err := generateMatches(db, playerIDs, numMatches)
	if err != nil {
		log.Fatalf("Failed to generate matches: %v", err)
	}
	fmt.Printf("✓ Created %d matches\n", matchCount)

	// Update player stats
	fmt.Println("\n📊 Calculating player statistics...")
	if err := updatePlayerStats(db); err != nil {
		log.Fatalf("Failed to update stats: %v", err)
	}
	fmt.Println("✓ Updated all player statistics")

	// Create admin user if not exists
	fmt.Println("\n👤 Creating admin user...")
	if err := createAdminUser(db); err != nil {
		fmt.Printf("  Admin user already exists or error: %v\n", err)
	} else {
		fmt.Println("✓ Admin user created (kong@admin / Admin@123!)")
	}

	// Show summary
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("✅ Mock data generation complete!")
	fmt.Println(strings.Repeat("=", 50))
	showSummary(db)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL DEFAULT '',
			phone VARCHAR(20) DEFAULT '',
			bio TEXT DEFAULT '',
			role VARCHAR(50) NOT NULL DEFAULT 'player',
			avatar_url VARCHAR(500) DEFAULT '',
			is_active BOOLEAN DEFAULT true,
			last_login_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS user_stats (
			user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
			total_matches INTEGER DEFAULT 0,
			wins INTEGER DEFAULT 0,
			losses INTEGER DEFAULT 0,
			win_rate DECIMAL(5,2) DEFAULT 0,
			current_streak INTEGER DEFAULT 0,
			best_streak INTEGER DEFAULT 0,
			skill_level VARCHAR(50) DEFAULT 'Beginner',
			skill_points INTEGER DEFAULT 0,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS matches (
			id SERIAL PRIMARY KEY,
			court VARCHAR(100),
			team1 INTEGER[] NOT NULL,
			team2 INTEGER[] NOT NULL,
			result VARCHAR(50) DEFAULT 'pending',
			started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			ended_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS match_scores (
			id SERIAL PRIMARY KEY,
			match_id INTEGER REFERENCES matches(id) ON DELETE CASCADE,
			game_number INTEGER NOT NULL,
			team1_score INTEGER DEFAULT 0,
			team2_score INTEGER DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS queue_entries (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			position INTEGER NOT NULL,
			status VARCHAR(50) DEFAULT 'waiting',
			joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			called_at TIMESTAMP WITH TIME ZONE
		)`,
		`CREATE TABLE IF NOT EXISTS refresh_tokens (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			revoked_at TIMESTAMP WITH TIME ZONE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone)`,
		`CREATE INDEX IF NOT EXISTS idx_matches_result ON matches(result)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("query failed: %w\nQuery: %s", err, q[:50])
		}
	}
	return nil
}

func hashPassword(password string) string {
	salt := make([]byte, 16)
	rand.Read(salt)
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash)
}

func generatePlayers(db *sql.DB, count int) ([]int64, error) {
	var ids []int64
	defaultPassword := hashPassword("Player@123!")

	for i := 0; i < count; i++ {
		firstName, lastName, nickname := generate.RandomThaiName()
		fullName := firstName + " " + lastName
		phone := generate.RandomPhoneNumber()
		username := strings.ToLower(nickname) + fmt.Sprintf("%d", mathrand.Intn(999))
		createdAt := generate.RandomDate(90) // Within last 90 days

		var id int64
		err := db.QueryRow(`
			INSERT INTO users (username, password_hash, name, phone, bio, role, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, 'player', true, $6, $6)
			ON CONFLICT (username) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, username, defaultPassword, fullName, phone, "Badminton enthusiast 🏸", createdAt).Scan(&id)

		if err != nil {
			return nil, fmt.Errorf("failed to insert player: %w", err)
		}

		// Initialize stats with random skill tier
		skillTier := generate.RandomSkillTier()
		db.Exec(`
			INSERT INTO user_stats (user_id, skill_level, skill_points)
			VALUES ($1, 'Beginner', 0)
			ON CONFLICT (user_id) DO NOTHING
		`, id)

		// Update user with skill_tier
		db.Exec(`UPDATE users SET skill_tier = $1 WHERE id = $2`, skillTier, id)

		ids = append(ids, id)
		fmt.Printf("  [%d/%d] Created: %s (@%s) %s\n", i+1, count, fullName, username, phone)
	}

	return ids, nil
}

func generateMatches(db *sql.DB, playerIDs []int64, count int) (int, error) {
	courts := []string{"Court 1", "Court 2", "Court 3", "Court 4"}
	matchCount := 0

	for i := 0; i < count; i++ {
		// Pick 4 random players for doubles
		if len(playerIDs) < 4 {
			return matchCount, fmt.Errorf("need at least 4 players")
		}

		// Shuffle and pick 4
		shuffled := make([]int64, len(playerIDs))
		copy(shuffled, playerIDs)
		mathrand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		team1 := shuffled[0:2]
		team2 := shuffled[2:4]
		court := courts[mathrand.Intn(len(courts))]
		startedAt := generate.RandomDate(60) // Within last 60 days

		// Generate match result
		games, team1Wins := generate.RandomBadmintonMatch()

		var result string
		if team1Wins {
			result = "team1"
		} else {
			result = "team2"
		}

		endedAt := startedAt.Add(time.Duration(20+mathrand.Intn(40)) * time.Minute)

		// Insert match
		var matchID int64
		err := db.QueryRow(`
			INSERT INTO matches (court, team1, team2, result, started_at, ended_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $5)
			RETURNING id
		`, court, pq(team1), pq(team2), result, startedAt, endedAt).Scan(&matchID)

		if err != nil {
			return matchCount, fmt.Errorf("failed to insert match: %w", err)
		}

		// Insert game scores
		for gameNum, game := range games {
			_, err := db.Exec(`
				INSERT INTO match_scores (match_id, game_number, team1_score, team2_score)
				VALUES ($1, $2, $3, $4)
			`, matchID, gameNum+1, game.Team1, game.Team2)
			if err != nil {
				return matchCount, fmt.Errorf("failed to insert score: %w", err)
			}
		}

		matchCount++

		// Progress output
		if matchCount%10 == 0 {
			fmt.Printf("  Created %d/%d matches...\n", matchCount, count)
		}
	}

	return matchCount, nil
}

// Helper to format int64 slice as PostgreSQL array
func pq(ids []int64) string {
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = fmt.Sprintf("%d", id)
	}
	return "{" + strings.Join(strs, ",") + "}"
}

func updatePlayerStats(db *sql.DB) error {
	// Calculate stats from matches
	_, err := db.Exec(`
		WITH match_results AS (
			SELECT 
				unnest(team1) as user_id,
				CASE WHEN result = 'team1' THEN 1 ELSE 0 END as win,
				CASE WHEN result = 'team2' THEN 1 ELSE 0 END as loss
			FROM matches WHERE result IN ('team1', 'team2')
			UNION ALL
			SELECT 
				unnest(team2) as user_id,
				CASE WHEN result = 'team2' THEN 1 ELSE 0 END as win,
				CASE WHEN result = 'team1' THEN 1 ELSE 0 END as loss
			FROM matches WHERE result IN ('team1', 'team2')
		),
		aggregated AS (
			SELECT 
				user_id,
				COUNT(*) as total_matches,
				SUM(win) as wins,
				SUM(loss) as losses,
				ROUND(SUM(win)::numeric * 100 / NULLIF(COUNT(*), 0), 2) as win_rate
			FROM match_results
			GROUP BY user_id
		)
		INSERT INTO user_stats (user_id, total_matches, wins, losses, win_rate, skill_level, skill_points, updated_at)
		SELECT 
			a.user_id,
			a.total_matches,
			a.wins,
			a.losses,
			a.win_rate,
			CASE 
				WHEN a.win_rate >= 75 THEN 'Expert'
				WHEN a.win_rate >= 55 THEN 'Advanced'
				WHEN a.win_rate >= 40 THEN 'Intermediate'
				ELSE 'Beginner'
			END,
			(a.total_matches * 10 + a.wins * 5),
			NOW()
		FROM aggregated a
		ON CONFLICT (user_id) DO UPDATE SET
			total_matches = EXCLUDED.total_matches,
			wins = EXCLUDED.wins,
			losses = EXCLUDED.losses,
			win_rate = EXCLUDED.win_rate,
			skill_level = EXCLUDED.skill_level,
			skill_points = EXCLUDED.skill_points,
			updated_at = NOW()
	`)

	return err
}

func createAdminUser(db *sql.DB) error {
	adminHash := hashPassword("Admin@123!")

	var id int64
	err := db.QueryRow(`
		INSERT INTO users (username, password_hash, name, phone, bio, role, is_active, created_at, updated_at)
		VALUES ('kong@admin', $1, 'Super Admin', '0899999999', 'System Administrator', 'admin', true, NOW(), NOW())
		RETURNING id
	`, adminHash).Scan(&id)

	if err != nil {
		return err
	}

	// Create stats for admin
	db.Exec(`INSERT INTO user_stats (user_id, skill_level) VALUES ($1, 'Expert') ON CONFLICT DO NOTHING`, id)

	return nil
}

func showSummary(db *sql.DB) {
	var userCount, matchCount, scoreCount int

	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	db.QueryRow("SELECT COUNT(*) FROM matches").Scan(&matchCount)
	db.QueryRow("SELECT COUNT(*) FROM match_scores").Scan(&scoreCount)

	fmt.Println("\n📊 Database Summary:")
	fmt.Printf("   Users:   %d\n", userCount)
	fmt.Printf("   Matches: %d\n", matchCount)
	fmt.Printf("   Scores:  %d\n", scoreCount)

	// Top players
	fmt.Println("\n🏆 Top 5 Players by Win Rate:")
	rows, err := db.Query(`
		SELECT u.name, us.win_rate, us.total_matches, us.skill_level
		FROM users u
		JOIN user_stats us ON u.id = us.user_id
		WHERE us.total_matches >= 5
		ORDER BY us.win_rate DESC, us.total_matches DESC
		LIMIT 5
	`)
	if err == nil {
		defer rows.Close()
		rank := 1
		for rows.Next() {
			var name, skillLevel string
			var winRate float64
			var matches int
			rows.Scan(&name, &winRate, &matches, &skillLevel)
			fmt.Printf("   %d. %s - %.1f%% (%d matches) [%s]\n", rank, name, winRate, matches, skillLevel)
			rank++
		}
	}
}
