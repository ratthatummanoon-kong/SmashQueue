package service

import (
	"backend/model"
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
)

var (
	ErrMatchNotFound = errors.New("match not found")
	ErrInvalidTeam   = errors.New("invalid team composition")
)

// MatchService handles match operations
type MatchService struct {
	userService *UserService
	db          *sql.DB
}

// NewMatchService creates a new match service
func NewMatchService(userSvc *UserService, db *sql.DB) *MatchService {
	return &MatchService{
		userService: userSvc,
		db:          db,
	}
}

// Create creates a new match
func (s *MatchService) Create(court string, team1, team2 []int64) (*model.Match, error) {
	if len(team1) == 0 || len(team2) == 0 {
		return nil, ErrInvalidTeam
	}
	if len(team1) > 2 || len(team2) > 2 {
		return nil, ErrInvalidTeam
	}

	ctx := context.Background()

	var match model.Match
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO matches (court, team1, team2, result, started_at, created_at)
		VALUES ($1, $2, $3, 'pending', NOW(), NOW())
		RETURNING id, court, team1, team2, result, started_at, ended_at, created_at
	`, court, pq.Array(team1), pq.Array(team2)).Scan(
		&match.ID, &match.Court, pq.Array(&match.Team1), pq.Array(&match.Team2),
		&match.Result, &match.StartedAt, &match.EndedAt, &match.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Update queue entries to 'playing' status
	allPlayers := append(team1, team2...)
	for _, playerID := range allPlayers {
		s.db.ExecContext(ctx, `
			UPDATE queue_entries SET status = 'playing' 
			WHERE user_id = $1 AND status IN ('waiting', 'called')
		`, playerID)
	}

	return &match, nil
}

// RecordResult records the result of a match
func (s *MatchService) RecordResult(matchID int64, scores []model.GameScore) (*model.Match, error) {
	ctx := context.Background()

	// Get match
	var match model.Match
	err := s.db.QueryRowContext(ctx, `
		SELECT id, court, team1, team2, result, started_at
		FROM matches WHERE id = $1
	`, matchID).Scan(&match.ID, &match.Court, pq.Array(&match.Team1), pq.Array(&match.Team2), &match.Result, &match.StartedAt)

	if err == sql.ErrNoRows {
		return nil, ErrMatchNotFound
	}
	if err != nil {
		return nil, err
	}

	// Calculate winner
	team1Wins := 0
	team2Wins := 0
	for _, score := range scores {
		if score.Team1Score > score.Team2Score {
			team1Wins++
		} else if score.Team2Score > score.Team1Score {
			team2Wins++
		}
	}

	result := "draw"
	if team1Wins > team2Wins {
		result = "team1"
	} else if team2Wins > team1Wins {
		result = "team2"
	}

	// Update match
	now := time.Now()
	_, err = s.db.ExecContext(ctx, `
		UPDATE matches SET result = $2, ended_at = $3 WHERE id = $1
	`, matchID, result, now)
	if err != nil {
		return nil, err
	}

	// Insert scores
	for i, score := range scores {
		s.db.ExecContext(ctx, `
			INSERT INTO match_scores (match_id, game_number, team1_score, team2_score)
			VALUES ($1, $2, $3, $4)
		`, matchID, i+1, score.Team1Score, score.Team2Score)
	}

	// Update player stats
	for _, playerID := range match.Team1 {
		s.userService.UpdateStats(playerID, result == "team1")
	}
	for _, playerID := range match.Team2 {
		s.userService.UpdateStats(playerID, result == "team2")
	}

	// Remove from queue
	allPlayers := append(match.Team1, match.Team2...)
	for _, playerID := range allPlayers {
		s.db.ExecContext(ctx, `
			DELETE FROM queue_entries WHERE user_id = $1 AND status = 'playing'
		`, playerID)
	}

	match.Result = result
	match.EndedAt = &now
	match.Scores = scores

	return &match, nil
}

// GetHistory returns match history for a user
func (s *MatchService) GetHistory(userID int64, limit int) ([]model.MatchHistory, error) {
	if limit <= 0 {
		limit = 20
	}

	ctx := context.Background()

	rows, err := s.db.QueryContext(ctx, `
		SELECT m.id, m.court, m.team1, m.team2, m.result, m.started_at, m.ended_at
		FROM matches m
		WHERE $1 = ANY(m.team1) OR $1 = ANY(m.team2)
		ORDER BY m.started_at DESC
		LIMIT $2
	`, userID, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []model.MatchHistory
	for rows.Next() {
		var match model.Match
		rows.Scan(&match.ID, &match.Court, pq.Array(&match.Team1), pq.Array(&match.Team2),
			&match.Result, &match.StartedAt, &match.EndedAt)

		// Get scores for this match
		match.Scores = s.getMatchScores(ctx, match.ID)

		// Determine if user won
		inTeam1 := contains(match.Team1, userID)
		won := (inTeam1 && match.Result == "team1") || (!inTeam1 && match.Result == "team2")

		history = append(history, model.MatchHistory{
			Match: match,
			Won:   won,
		})
	}

	return history, nil
}

// GetActive returns active matches
func (s *MatchService) GetActive() ([]model.Match, error) {
	ctx := context.Background()

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, court, team1, team2, result, started_at, ended_at, created_at
		FROM matches WHERE result = 'pending'
		ORDER BY started_at DESC
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []model.Match
	for rows.Next() {
		var match model.Match
		rows.Scan(&match.ID, &match.Court, pq.Array(&match.Team1), pq.Array(&match.Team2),
			&match.Result, &match.StartedAt, &match.EndedAt, &match.CreatedAt)
		match.Scores = s.getMatchScores(ctx, match.ID)
		matches = append(matches, match)
	}

	return matches, nil
}

// GetAllCompleted returns all completed matches with player names (for admin)
func (s *MatchService) GetAllCompleted(limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 50
	}

	ctx := context.Background()

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, court, team1, team2, result, started_at, ended_at
		FROM matches 
		WHERE result != 'pending'
		ORDER BY ended_at DESC
		LIMIT $1
	`, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var match model.Match
		rows.Scan(&match.ID, &match.Court, pq.Array(&match.Team1), pq.Array(&match.Team2),
			&match.Result, &match.StartedAt, &match.EndedAt)

		// Get scores
		match.Scores = s.getMatchScores(ctx, match.ID)

		// Get player names
		team1Names := s.getPlayerNames(ctx, match.Team1)
		team2Names := s.getPlayerNames(ctx, match.Team2)

		results = append(results, map[string]interface{}{
			"id":          match.ID,
			"court":       match.Court,
			"team1":       match.Team1,
			"team2":       match.Team2,
			"team1_names": team1Names,
			"team2_names": team2Names,
			"scores":      match.Scores,
			"result":      match.Result,
			"started_at":  match.StartedAt,
			"ended_at":    match.EndedAt,
		})
	}

	return results, nil
}

func (s *MatchService) getPlayerNames(ctx context.Context, playerIDs []int64) []string {
	if len(playerIDs) == 0 {
		return []string{}
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT name FROM users WHERE id = ANY($1)
	`, pq.Array(playerIDs))
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}
	return names
}

func (s *MatchService) getMatchScores(ctx context.Context, matchID int64) []model.GameScore {
	rows, err := s.db.QueryContext(ctx, `
		SELECT game_number, team1_score, team2_score
		FROM match_scores WHERE match_id = $1 ORDER BY game_number
	`, matchID)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var scores []model.GameScore
	for rows.Next() {
		var score model.GameScore
		rows.Scan(&score.Game, &score.Team1Score, &score.Team2Score)
		scores = append(scores, score)
	}
	return scores
}

func contains(slice []int64, val int64) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
