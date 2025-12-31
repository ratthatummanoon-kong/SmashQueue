package service

import (
	"backend/model"
	"context"
	"database/sql"
	"errors"
)

// UserService handles user profile operations
type UserService struct {
	authService *AuthService
	db          *sql.DB
}

// NewUserService creates a new user service
func NewUserService(authSvc *AuthService, db *sql.DB) *UserService {
	return &UserService{
		authService: authSvc,
		db:          db,
	}
}

// GetProfile returns the user profile with stats
func (s *UserService) GetProfile(userID int64) (*model.UserProfile, error) {
	user, err := s.authService.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	stats, err := s.GetStats(userID)
	if err != nil {
		// Return with empty stats if not found
		stats = &model.UserStats{
			UserID:     userID,
			SkillLevel: "Beginner",
		}
	}

	return &model.UserProfile{
		User:  *user,
		Stats: *stats,
	}, nil
}

// UpdateProfile updates user profile information
func (s *UserService) UpdateProfile(userID int64, req model.UpdateProfileRequest) (*model.User, error) {
	ctx := context.Background()

	var user model.User
	err := s.db.QueryRowContext(ctx, `
		UPDATE users 
		SET name = $2, bio = $3, phone = COALESCE(NULLIF($4, ''), phone), updated_at = NOW()
		WHERE id = $1
		RETURNING id, username, name, phone, bio, role, is_active, created_at, updated_at
	`, userID, req.Name, req.Bio, req.Phone).Scan(
		&user.ID, &user.Username, &user.Name, &user.Phone,
		&user.Bio, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetStats returns the user's performance statistics
func (s *UserService) GetStats(userID int64) (*model.UserStats, error) {
	ctx := context.Background()

	var stats model.UserStats
	err := s.db.QueryRowContext(ctx, `
		SELECT user_id, total_matches, wins, losses, win_rate, current_streak, best_streak, skill_level, skill_points
		FROM user_stats WHERE user_id = $1
	`, userID).Scan(
		&stats.UserID, &stats.TotalMatches, &stats.Wins, &stats.Losses,
		&stats.WinRate, &stats.CurrentStreak, &stats.BestStreak, &stats.SkillLevel, &stats.SkillPoints,
	)

	if err == sql.ErrNoRows {
		// Initialize stats if not found
		s.db.ExecContext(ctx, `
			INSERT INTO user_stats (user_id, skill_level) VALUES ($1, 'Beginner') ON CONFLICT DO NOTHING
		`, userID)
		return &model.UserStats{UserID: userID, SkillLevel: "Beginner"}, nil
	}
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

// UpdateStats updates user statistics after a match
func (s *UserService) UpdateStats(userID int64, won bool) error {
	ctx := context.Background()

	// Get current stats
	stats, err := s.GetStats(userID)
	if err != nil {
		return err
	}

	// Update stats
	stats.TotalMatches++
	if won {
		stats.Wins++
		if stats.CurrentStreak >= 0 {
			stats.CurrentStreak++
		} else {
			stats.CurrentStreak = 1
		}
		if stats.CurrentStreak > stats.BestStreak {
			stats.BestStreak = stats.CurrentStreak
		}
	} else {
		stats.Losses++
		if stats.CurrentStreak <= 0 {
			stats.CurrentStreak--
		} else {
			stats.CurrentStreak = -1
		}
	}

	// Calculate win rate
	if stats.TotalMatches > 0 {
		stats.WinRate = float64(stats.Wins) * 100 / float64(stats.TotalMatches)
	}

	// Calculate skill level
	stats.SkillLevel = calculateSkillLevel(stats.WinRate, stats.TotalMatches)
	stats.SkillPoints = stats.TotalMatches*10 + stats.Wins*5

	// Save to database
	_, err = s.db.ExecContext(ctx, `
		UPDATE user_stats SET
			total_matches = $2, wins = $3, losses = $4, win_rate = $5,
			current_streak = $6, best_streak = $7, skill_level = $8, skill_points = $9,
			updated_at = NOW()
		WHERE user_id = $1
	`, userID, stats.TotalMatches, stats.Wins, stats.Losses, stats.WinRate,
		stats.CurrentStreak, stats.BestStreak, stats.SkillLevel, stats.SkillPoints)

	return err
}

func calculateSkillLevel(winRate float64, matches int) string {
	if matches < 5 {
		return "Beginner"
	}
	if winRate >= 75 {
		return "Expert"
	}
	if winRate >= 55 {
		return "Advanced"
	}
	if winRate >= 40 {
		return "Intermediate"
	}
	return "Beginner"
}
