package service

import (
	"backend/config"
	"backend/model"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserExists         = errors.New("username already exists")
	ErrPhoneExists        = errors.New("phone number already registered")
	ErrWeakPassword       = errors.New("password does not meet requirements")
	ErrPasswordMatch      = errors.New("passwords do not match")
	ErrPasswordSameAsUser = errors.New("password cannot be the same as username")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrInvalidPhone       = errors.New("invalid phone number format")
	ErrUserNotFound       = errors.New("user not found")
	ErrPhoneMismatch      = errors.New("phone number does not match")
)

// AuthService handles authentication logic
type AuthService struct {
	config *config.Config
	db     *sql.DB
}

// NewAuthService creates a new auth service
func NewAuthService(cfg *config.Config, db *sql.DB) *AuthService {
	svc := &AuthService{
		config: cfg,
		db:     db,
	}

	if db != nil {
		svc.ensureAdminExists()
		svc.ensureTablesUpdated()
	}

	return svc
}

func (s *AuthService) ensureAdminExists() {
	ctx := context.Background()
	var count int
	s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE username = 'kong@admin'").Scan(&count)
	if count == 0 {
		hash := s.hashPassword("Admin@123!")
		s.db.ExecContext(ctx, `
			INSERT INTO users (username, password_hash, name, phone, bio, role, hand_preference, skill_tier, is_active, created_at, updated_at)
			VALUES ('kong@admin', $1, 'Super Admin', '0899999999', 'System Administrator', 'admin', 'right', 'A', true, NOW(), NOW())
		`, hash)
	}
}

func (s *AuthService) ensureTablesUpdated() {
	ctx := context.Background()
	// Add new columns if they don't exist
	s.db.ExecContext(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS hand_preference VARCHAR(10) DEFAULT 'right'`)
	s.db.ExecContext(ctx, `ALTER TABLE users ADD COLUMN IF NOT EXISTS skill_tier VARCHAR(5) DEFAULT 'N'`)
}

// Register creates a new user account
func (s *AuthService) Register(req model.RegisterRequest) (*model.AuthResponse, error) {
	if req.Password != req.ConfirmPassword {
		return nil, ErrPasswordMatch
	}

	if err := s.validatePassword(req.Username, req.Password); err != nil {
		return nil, err
	}

	if req.Phone != "" && !isValidPhone(req.Phone) {
		return nil, ErrInvalidPhone
	}

	ctx := context.Background()

	var existingCount int
	s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE username = $1", req.Username).Scan(&existingCount)
	if existingCount > 0 {
		return nil, ErrUserExists
	}

	if req.Phone != "" {
		var phoneCount int
		s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE phone = $1", req.Phone).Scan(&phoneCount)
		if phoneCount > 0 {
			return nil, ErrPhoneExists
		}
	}

	hash := s.hashPassword(req.Password)

	name := req.Name
	if name == "" {
		name = req.Username
	}

	var user model.User
	now := time.Now()
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO users (username, password_hash, name, phone, bio, role, hand_preference, skill_tier, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, '', 'player', 'right', 'N', true, $5, $5)
		RETURNING id, username, name, phone, bio, role, hand_preference, skill_tier, is_active, created_at, updated_at
	`, req.Username, hash, name, req.Phone, now).Scan(
		&user.ID, &user.Username, &user.Name, &user.Phone, &user.Bio,
		&user.Role, &user.HandPreference, &user.SkillTier, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	s.db.ExecContext(ctx, `
		INSERT INTO user_stats (user_id, skill_level, skill_points)
		VALUES ($1, 'Beginner', 0)
		ON CONFLICT (user_id) DO NOTHING
	`, user.ID)

	accessToken, err := s.generateAccessToken(&user)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken: accessToken,
		User:        user,
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(req model.LoginRequest) (*model.AuthResponse, string, error) {
	ctx := context.Background()

	var user model.User
	var passwordHash string
	var handPref, skillTier sql.NullString

	err := s.db.QueryRowContext(ctx, `
		SELECT id, username, password_hash, name, phone, bio, role, 
		       COALESCE(hand_preference, 'right'), COALESCE(skill_tier, 'N'), 
		       is_active, created_at, updated_at
		FROM users WHERE username = $1
	`, req.Username).Scan(
		&user.ID, &user.Username, &passwordHash, &user.Name, &user.Phone,
		&user.Bio, &user.Role, &handPref, &skillTier,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, "", ErrInvalidCredentials
	}
	if err != nil {
		return nil, "", err
	}

	user.HandPreference = model.HandPreference(handPref.String)
	user.SkillTier = model.SkillTier(skillTier.String)

	if !s.verifyPassword(req.Password, passwordHash) {
		return nil, "", ErrInvalidCredentials
	}

	s.db.ExecContext(ctx, "UPDATE users SET last_login_at = NOW() WHERE id = $1", user.ID)

	accessToken, err := s.generateAccessToken(&user)
	if err != nil {
		return nil, "", err
	}

	refreshToken, err := s.generateRefreshToken(&user)
	if err != nil {
		return nil, "", err
	}

	return &model.AuthResponse{
		AccessToken: accessToken,
		User:        user,
	}, refreshToken, nil
}

// VerifyUserForReset verifies username and phone for password reset
func (s *AuthService) VerifyUserForReset(req model.ForgotPasswordRequest) error {
	ctx := context.Background()

	var storedPhone string
	err := s.db.QueryRowContext(ctx, `
		SELECT phone FROM users WHERE username = $1
	`, req.Username).Scan(&storedPhone)

	if err == sql.ErrNoRows {
		return ErrUserNotFound
	}
	if err != nil {
		return err
	}

	if storedPhone != req.Phone {
		return ErrPhoneMismatch
	}

	return nil
}

// ResetPassword resets password after verification
func (s *AuthService) ResetPassword(req model.ResetPasswordRequest) error {
	// First verify
	if err := s.VerifyUserForReset(model.ForgotPasswordRequest{
		Username: req.Username,
		Phone:    req.Phone,
	}); err != nil {
		return err
	}

	// Validate new password
	if err := s.validatePassword(req.Username, req.NewPassword); err != nil {
		return err
	}

	ctx := context.Background()
	hash := s.hashPassword(req.NewPassword)

	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET password_hash = $2, updated_at = NOW() WHERE username = $1
	`, req.Username, hash)

	return err
}

// RefreshAccessToken generates a new access token from a refresh token
func (s *AuthService) RefreshAccessToken(refreshToken string) (*model.AuthResponse, error) {
	ctx := context.Background()

	var token model.RefreshToken
	var userID int64
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, expires_at, revoked_at
		FROM refresh_tokens WHERE token = $1
	`, refreshToken).Scan(&token.ID, &userID, &token.ExpiresAt, &token.RevokedAt)

	if err == sql.ErrNoRows {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}

	if !token.IsValid() {
		return nil, ErrInvalidToken
	}

	user, err := s.GetUserByID(userID)
	if err != nil {
		return nil, ErrInvalidToken
	}

	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		AccessToken: accessToken,
		User:        *user,
	}, nil
}

// ValidateToken verifies and decodes an access token
func (s *AuthService) ValidateToken(tokenString string) (*model.TokenPayload, error) {
	var payload model.TokenPayload

	v2 := paseto.NewV2()
	key := []byte(s.config.Auth.SecretKey)[:32]

	err := v2.Decrypt(tokenString, key, &payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if payload.IsExpired() {
		return nil, ErrInvalidToken
	}

	return &payload, nil
}

// GetUserByID returns a user by ID
func (s *AuthService) GetUserByID(userID int64) (*model.User, error) {
	ctx := context.Background()

	var user model.User
	var handPref, skillTier sql.NullString

	err := s.db.QueryRowContext(ctx, `
		SELECT id, username, name, phone, bio, role, 
		       COALESCE(hand_preference, 'right'), COALESCE(skill_tier, 'N'),
		       is_active, created_at, updated_at
		FROM users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Username, &user.Name, &user.Phone,
		&user.Bio, &user.Role, &handPref, &skillTier,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}

	user.HandPreference = model.HandPreference(handPref.String)
	user.SkillTier = model.SkillTier(skillTier.String)

	return &user, nil
}

// GetAllUsers returns all users with stats (for admin)
func (s *AuthService) GetAllUsers() ([]model.UserListItem, error) {
	ctx := context.Background()
	rows, err := s.db.QueryContext(ctx, `
		SELECT u.id, u.username, u.name, u.phone, u.role, 
		       COALESCE(u.hand_preference, 'right'), COALESCE(u.skill_tier, 'N'),
		       u.is_active, 
		       COALESCE(us.skill_level, 'Beginner'), COALESCE(us.win_rate, 0),
		       COALESCE(us.total_matches, 0), COALESCE(us.wins, 0)
		FROM users u
		LEFT JOIN user_stats us ON u.id = us.user_id
		ORDER BY u.name ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.UserListItem
	for rows.Next() {
		var u model.UserListItem
		if err := rows.Scan(&u.ID, &u.Username, &u.Name, &u.Phone, &u.Role,
			&u.HandPreference, &u.SkillTier, &u.IsActive, &u.SkillLevel, &u.WinRate,
			&u.TotalMatches, &u.Wins); err != nil {
			continue
		}
		users = append(users, u)
	}

	return users, nil
}

// UpdatePlayerAdmin updates player hand preference and skill tier (admin only)
func (s *AuthService) UpdatePlayerAdmin(userID int64, req model.UpdatePlayerAdminRequest) error {
	ctx := context.Background()

	// Validate skill tier
	validTiers := []string{"BG", "S-", "S", "N", "P-", "P", "P+", "C", "B", "A"}
	validTier := false
	for _, t := range validTiers {
		if req.SkillTier == t {
			validTier = true
			break
		}
	}
	if !validTier {
		return errors.New("invalid skill tier")
	}

	// Validate hand preference
	if req.HandPreference != "left" && req.HandPreference != "right" {
		return errors.New("invalid hand preference")
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET hand_preference = $2, skill_tier = $3, updated_at = NOW()
		WHERE id = $1
	`, userID, req.HandPreference, req.SkillTier)

	return err
}

func isValidPhone(phone string) bool {
	matched, _ := regexp.MatchString(`^0[89]\d{8}$`, phone)
	return matched
}

func (s *AuthService) validatePassword(username, password string) error {
	if len(password) < 8 {
		return ErrWeakPassword
	}

	if !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return ErrWeakPassword
	}

	if !regexp.MustCompile(`[A-Z]`).MatchString(password) {
		return ErrWeakPassword
	}

	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return ErrWeakPassword
	}

	hasSpecial := false
	for _, c := range password {
		if unicode.IsPunct(c) || unicode.IsSymbol(c) {
			hasSpecial = true
			break
		}
	}
	if !hasSpecial {
		return ErrWeakPassword
	}

	if strings.EqualFold(password, username) {
		return ErrPasswordSameAsUser
	}

	return nil
}

func (s *AuthService) hashPassword(password string) string {
	salt := make([]byte, 16)
	rand.Read(salt)
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash)
}

func (s *AuthService) verifyPassword(password, storedHash string) bool {
	parts := strings.Split(storedHash, ":")
	if len(parts) != 2 {
		return false
	}

	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}

	expectedHash, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}

	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	if len(hash) != len(expectedHash) {
		return false
	}

	for i := range hash {
		if hash[i] != expectedHash[i] {
			return false
		}
	}

	return true
}

func (s *AuthService) generateAccessToken(user *model.User) (string, error) {
	now := time.Now()
	payload := model.TokenPayload{
		UserID:    user.ID,
		Username:  user.Username,
		Role:      user.Role,
		IssuedAt:  now,
		ExpiresAt: now.Add(s.config.Auth.AccessTokenDuration),
	}

	v2 := paseto.NewV2()
	key := []byte(s.config.Auth.SecretKey)[:32]

	return v2.Encrypt(key, payload, nil)
}

func (s *AuthService) generateRefreshToken(user *model.User) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(s.config.Auth.RefreshTokenDuration)

	ctx := context.Background()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
	`, user.ID, token, expiresAt)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) Logout(refreshToken string) error {
	ctx := context.Background()
	_, err := s.db.ExecContext(ctx, `
		UPDATE refresh_tokens SET revoked_at = NOW() WHERE token = $1
	`, refreshToken)
	return err
}
