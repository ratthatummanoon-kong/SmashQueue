package database

import (
	"backend/model"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// UserRepository handles user database operations
type UserRepository struct{}

// NewUserRepository creates a new user repository
func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, password_hash, name, bio, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	now := time.Now()
	err := DB.QueryRow(ctx, query,
		user.Username,
		user.PasswordHash,
		user.Name,
		user.Bio,
		user.Role,
		now,
		now,
	).Scan(&user.ID)

	if err != nil {
		return err
	}

	user.CreatedAt = now
	user.UpdatedAt = now

	// Initialize user stats
	statsQuery := `
		INSERT INTO user_stats (user_id, skill_level)
		VALUES ($1, 'Beginner')
		ON CONFLICT (user_id) DO NOTHING
	`
	_, _ = DB.Exec(ctx, statsQuery, user.ID)

	return nil
}

// GetByUsername finds a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, name, bio, role, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	user := &model.User{}
	err := DB.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Name,
		&user.Bio,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID finds a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*model.User, error) {
	query := `
		SELECT id, username, password_hash, name, bio, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	user := &model.User{}
	err := DB.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Name,
		&user.Bio,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Update updates a user's profile
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET name = $2, bio = $3, updated_at = $4
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()
	_, err := DB.Exec(ctx, query, user.ID, user.Name, user.Bio, user.UpdatedAt)
	return err
}

// UsernameExists checks if a username is already taken
func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	var exists bool
	err := DB.QueryRow(ctx, query, username).Scan(&exists)
	return exists, err
}

// GetStats returns a user's statistics
func (r *UserRepository) GetStats(ctx context.Context, userID int64) (*model.UserStats, error) {
	query := `
		SELECT user_id, total_matches, wins, losses, win_rate, current_streak, skill_level
		FROM user_stats
		WHERE user_id = $1
	`

	stats := &model.UserStats{}
	err := DB.QueryRow(ctx, query, userID).Scan(
		&stats.UserID,
		&stats.TotalMatches,
		&stats.Wins,
		&stats.Losses,
		&stats.WinRate,
		&stats.CurrentStreak,
		&stats.SkillLevel,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		// Return default stats
		return &model.UserStats{
			UserID:     userID,
			SkillLevel: "Beginner",
		}, nil
	}
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// UpdateStats updates a user's statistics
func (r *UserRepository) UpdateStats(ctx context.Context, stats *model.UserStats) error {
	query := `
		INSERT INTO user_stats (user_id, total_matches, wins, losses, win_rate, current_streak, skill_level, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			total_matches = $2,
			wins = $3,
			losses = $4,
			win_rate = $5,
			current_streak = $6,
			skill_level = $7,
			updated_at = NOW()
	`

	_, err := DB.Exec(ctx, query,
		stats.UserID,
		stats.TotalMatches,
		stats.Wins,
		stats.Losses,
		stats.WinRate,
		stats.CurrentStreak,
		stats.SkillLevel,
	)

	return err
}
