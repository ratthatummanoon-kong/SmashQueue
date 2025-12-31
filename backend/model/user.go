package model

import (
	"time"
)

// Role represents user permission levels
type Role string

const (
	RolePlayer    Role = "player"
	RoleOrganizer Role = "organizer"
	RoleAdmin     Role = "admin"
)

// HandPreference represents player's dominant hand
type HandPreference string

const (
	HandRight HandPreference = "right"
	HandLeft  HandPreference = "left"
)

// SkillTier represents player skill levels (Thai badminton ranking style)
type SkillTier string

const (
	SkillBG SkillTier = "BG" // Beginner
	SkillSM SkillTier = "S-" // Sub-Standard minus
	SkillS  SkillTier = "S"  // Standard
	SkillN  SkillTier = "N"  // Normal
	SkillPM SkillTier = "P-" // Pro minus
	SkillP  SkillTier = "P"  // Pro
	SkillPP SkillTier = "P+" // Pro plus
	SkillC  SkillTier = "C"  // Champion
	SkillB  SkillTier = "B"  // Best
	SkillA  SkillTier = "A"  // Ace
)

// User represents a SmashQueue user account
type User struct {
	ID             int64          `json:"id"`
	Username       string         `json:"username"`
	PasswordHash   string         `json:"-"` // Never expose in JSON
	Name           string         `json:"name"`
	Phone          string         `json:"phone"`
	Bio            string         `json:"bio"`
	Role           Role           `json:"role"`
	HandPreference HandPreference `json:"hand_preference"`
	SkillTier      SkillTier      `json:"skill_tier"`
	AvatarURL      string         `json:"avatar_url,omitempty"`
	IsActive       bool           `json:"is_active"`
	LastLoginAt    *time.Time     `json:"last_login_at,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// UserStats holds player performance metrics
type UserStats struct {
	UserID        int64   `json:"user_id"`
	TotalMatches  int     `json:"total_matches"`
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	WinRate       float64 `json:"win_rate"`
	CurrentStreak int     `json:"current_streak"` // Positive = win streak, negative = loss streak
	BestStreak    int     `json:"best_streak"`
	SkillLevel    string  `json:"skill_level"` // Beginner, Intermediate, Advanced, Expert
	SkillPoints   int     `json:"skill_points"`
}

// UserProfile combines user info with stats
type UserProfile struct {
	User  User      `json:"user"`
	Stats UserStats `json:"stats"`
}

// RegisterRequest is the payload for user registration
type RegisterRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Phone           string `json:"phone"`
	Name            string `json:"name,omitempty"`
}

// LoginRequest is the payload for user login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UpdateProfileRequest is the payload for profile updates
type UpdateProfileRequest struct {
	Name  string `json:"name"`
	Bio   string `json:"bio"`
	Phone string `json:"phone,omitempty"`
}

// UpdatePlayerAdminRequest is for admin to update player settings
type UpdatePlayerAdminRequest struct {
	HandPreference string `json:"hand_preference"`
	SkillTier      string `json:"skill_tier"`
}

// AuthResponse is returned after successful authentication
type AuthResponse struct {
	AccessToken string `json:"access_token"`
	User        User   `json:"user"`
}

// UserListItem is a simplified user for admin views
type UserListItem struct {
	ID             int64   `json:"id"`
	Username       string  `json:"username"`
	Name           string  `json:"name"`
	Phone          string  `json:"phone"`
	Role           Role    `json:"role"`
	HandPreference string  `json:"hand_preference"`
	SkillTier      string  `json:"skill_tier"`
	IsActive       bool    `json:"is_active"`
	SkillLevel     string  `json:"skill_level"`
	WinRate        float64 `json:"win_rate"`
	TotalMatches   int     `json:"total_matches"`
	Wins           int     `json:"wins"`
}

// ForgotPasswordRequest is the payload for password reset
type ForgotPasswordRequest struct {
	Username string `json:"username"`
	Phone    string `json:"phone"`
}

// ResetPasswordRequest is the payload for setting new password
type ResetPasswordRequest struct {
	Username    string `json:"username"`
	Phone       string `json:"phone"`
	NewPassword string `json:"new_password"`
}
