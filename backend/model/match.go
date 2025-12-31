package model

import (
	"time"
)

// MatchResult represents the outcome of a match
type MatchResult string

const (
	MatchResultPending MatchResult = "pending"
	MatchResultTeam1   MatchResult = "team1"
	MatchResultTeam2   MatchResult = "team2"
	MatchResultDraw    MatchResult = "draw"
)

// Match represents a badminton game
type Match struct {
	ID        int64       `json:"id"`
	Court     string      `json:"court"`
	Team1     []int64     `json:"team1"`  // Player IDs
	Team2     []int64     `json:"team2"`  // Player IDs
	Scores    []GameScore `json:"scores"` // Score per game
	Result    string      `json:"result"`
	StartedAt time.Time   `json:"started_at"`
	EndedAt   *time.Time  `json:"ended_at,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}

// GameScore represents the score of a single game within a match
type GameScore struct {
	Game       int `json:"game"` // 1, 2, or 3
	Team1Score int `json:"team1_score"`
	Team2Score int `json:"team2_score"`
}

// MatchHistory represents a match from a player's perspective
type MatchHistory struct {
	Match Match `json:"match"`
	Won   bool  `json:"won"`
}

// CreateMatchRequest is the payload for creating a new match
type CreateMatchRequest struct {
	Court string  `json:"court"`
	Team1 []int64 `json:"team1"`
	Team2 []int64 `json:"team2"`
}

// RecordResultRequest is the payload for recording match results
type RecordResultRequest struct {
	MatchID int64       `json:"match_id"`
	Scores  []GameScore `json:"scores"`
}
