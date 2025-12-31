package model

import (
	"time"
)

// TokenPayload contains the claims stored in tokens
type TokenPayload struct {
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	Role      Role      `json:"role"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// IsExpired checks if the token has expired
func (p *TokenPayload) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// RefreshToken represents a stored refresh token
type RefreshToken struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

// IsValid checks if the refresh token is still valid
func (r *RefreshToken) IsValid() bool {
	return r.RevokedAt == nil && time.Now().Before(r.ExpiresAt)
}

// RefreshTokenRequest is the payload for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
