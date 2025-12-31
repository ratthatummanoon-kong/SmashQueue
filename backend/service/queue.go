package service

import (
	"backend/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var (
	ErrAlreadyInQueue = errors.New("user is already in queue")
	ErrNotInQueue     = errors.New("user is not in queue")
	ErrQueueEmpty     = errors.New("queue is empty")
)

// QueueService handles queue operations
type QueueService struct {
	db *sql.DB
}

// NewQueueService creates a new queue service
func NewQueueService(db *sql.DB) *QueueService {
	return &QueueService{db: db}
}

// GetStatus returns the current queue status
func (s *QueueService) GetStatus(userID int64) (*model.QueueInfo, error) {
	ctx := context.Background()

	info := &model.QueueInfo{}

	// Get total in queue
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM queue_entries WHERE status = 'waiting'
	`).Scan(&info.TotalInQueue)

	// Get user's position if in queue
	if userID > 0 {
		var position int
		err := s.db.QueryRowContext(ctx, `
			SELECT position FROM queue_entries WHERE user_id = $1 AND status = 'waiting'
		`, userID).Scan(&position)
		if err == nil {
			info.YourPosition = &position

			// Estimate wait time (5 minutes per person ahead)
			waitMinutes := (position - 1) * 5
			if waitMinutes <= 0 {
				info.EstimatedWait = stringPtr("Next up!")
			} else {
				info.EstimatedWait = stringPtr(fmt.Sprintf("~%d min", waitMinutes))
			}

			// Next available court
			nextCourt := s.getNextAvailableCourt(ctx)
			info.NextCourt = &nextCourt
		}
	}

	// Get currently playing (status = 'playing')
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, position, status, joined_at
		FROM queue_entries WHERE status = 'playing'
		ORDER BY called_at DESC LIMIT 8
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var entry model.QueueEntry
			var joinedAt time.Time
			rows.Scan(&entry.ID, &entry.UserID, &entry.Position, &entry.Status, &joinedAt)
			entry.JoinedAt = joinedAt
			info.CurrentlyPlaying = append(info.CurrentlyPlaying, entry)
		}
	}

	return info, nil
}

func (s *QueueService) getNextAvailableCourt(ctx context.Context) string {
	courts := []string{"Court 1", "Court 2", "Court 3", "Court 4"}

	// Find court with fewest active matches
	for _, court := range courts {
		var count int
		s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM matches WHERE court = $1 AND result = 'pending'
		`, court).Scan(&count)
		if count == 0 {
			return court
		}
	}
	return courts[0]
}

// Join adds a user to the queue
func (s *QueueService) Join(userID int64) (*model.QueueEntry, error) {
	ctx := context.Background()

	// Check if already in queue
	var existingCount int
	s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM queue_entries WHERE user_id = $1 AND status = 'waiting'
	`, userID).Scan(&existingCount)

	if existingCount > 0 {
		return nil, ErrAlreadyInQueue
	}

	// Get next position
	var nextPosition int
	s.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(position), 0) + 1 FROM queue_entries WHERE status = 'waiting'
	`).Scan(&nextPosition)

	// Insert entry
	var entry model.QueueEntry
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO queue_entries (user_id, position, status, joined_at)
		VALUES ($1, $2, 'waiting', NOW())
		RETURNING id, user_id, position, status, joined_at
	`, userID, nextPosition).Scan(&entry.ID, &entry.UserID, &entry.Position, &entry.Status, &entry.JoinedAt)

	if err != nil {
		return nil, err
	}

	return &entry, nil
}

// Leave removes a user from the queue
func (s *QueueService) Leave(userID int64) error {
	ctx := context.Background()

	result, err := s.db.ExecContext(ctx, `
		DELETE FROM queue_entries WHERE user_id = $1 AND status = 'waiting'
	`, userID)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrNotInQueue
	}

	// Reorder positions
	s.db.ExecContext(ctx, `
		WITH ordered AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY position) as new_pos
			FROM queue_entries WHERE status = 'waiting'
		)
		UPDATE queue_entries SET position = ordered.new_pos
		FROM ordered WHERE queue_entries.id = ordered.id
	`)

	return nil
}

// CallNext calls the next N players from the queue
func (s *QueueService) CallNext(count int) ([]model.QueueEntry, error) {
	if count <= 0 {
		count = 4 // Default for doubles
	}

	ctx := context.Background()

	// Get and update next entries
	rows, err := s.db.QueryContext(ctx, `
		UPDATE queue_entries 
		SET status = 'called', called_at = NOW()
		WHERE id IN (
			SELECT id FROM queue_entries WHERE status = 'waiting'
			ORDER BY position LIMIT $1
		)
		RETURNING id, user_id, position, status, joined_at
	`, count)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var called []model.QueueEntry
	for rows.Next() {
		var entry model.QueueEntry
		rows.Scan(&entry.ID, &entry.UserID, &entry.Position, &entry.Status, &entry.JoinedAt)
		called = append(called, entry)
	}

	if len(called) == 0 {
		return nil, ErrQueueEmpty
	}

	// Reorder remaining positions
	s.db.ExecContext(ctx, `
		WITH ordered AS (
			SELECT id, ROW_NUMBER() OVER (ORDER BY position) as new_pos
			FROM queue_entries WHERE status = 'waiting'
		)
		UPDATE queue_entries SET position = ordered.new_pos
		FROM ordered WHERE queue_entries.id = ordered.id
	`)

	return called, nil
}

func stringPtr(s string) *string {
	return &s
}
