package database

import (
	"backend/model"
	"context"
	"time"
)

// TokenRepository handles refresh token database operations
type TokenRepository struct{}

// NewTokenRepository creates a new token repository
func NewTokenRepository() *TokenRepository {
	return &TokenRepository{}
}

// Create stores a new refresh token
func (r *TokenRepository) Create(ctx context.Context, token *model.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	return DB.QueryRow(ctx, query,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		time.Now(),
	).Scan(&token.ID)
}

// GetByToken finds a refresh token by its value
func (r *TokenRepository) GetByToken(ctx context.Context, tokenValue string) (*model.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at, revoked_at
		FROM refresh_tokens
		WHERE token = $1 AND revoked_at IS NULL
	`

	token := &model.RefreshToken{}
	err := DB.QueryRow(ctx, query, tokenValue).Scan(
		&token.ID,
		&token.UserID,
		&token.Token,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.RevokedAt,
	)

	if err != nil {
		return nil, err
	}

	return token, nil
}

// Revoke invalidates a refresh token
func (r *TokenRepository) Revoke(ctx context.Context, tokenValue string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE token = $1
	`

	_, err := DB.Exec(ctx, query, tokenValue)
	return err
}

// RevokeAllForUser invalidates all refresh tokens for a user
func (r *TokenRepository) RevokeAllForUser(ctx context.Context, userID int64) error {
	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	_, err := DB.Exec(ctx, query, userID)
	return err
}

// CleanupExpired removes expired tokens
func (r *TokenRepository) CleanupExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW() OR revoked_at IS NOT NULL
	`

	result, err := DB.Exec(ctx, query)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}
