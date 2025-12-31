package model

import (
	"time"
)

// QueueStatus represents the state of a queue entry
type QueueStatus string

const (
	QueueStatusWaiting  QueueStatus = "waiting"
	QueueStatusCalled   QueueStatus = "called"
	QueueStatusPlaying  QueueStatus = "playing"
	QueueStatusFinished QueueStatus = "finished"
)

// QueueEntry represents a player's position in the queue
type QueueEntry struct {
	ID       int64      `json:"id"`
	UserID   int64      `json:"user_id"`
	Position int        `json:"position"`
	Status   string     `json:"status"`
	JoinedAt time.Time  `json:"joined_at"`
	CalledAt *time.Time `json:"called_at,omitempty"`
}

// QueueInfo provides queue status summary
type QueueInfo struct {
	TotalInQueue     int          `json:"total_in_queue"`
	YourPosition     *int         `json:"your_position,omitempty"`
	EstimatedWait    *string      `json:"estimated_wait,omitempty"`
	NextCourt        *string      `json:"next_court,omitempty"`
	CurrentlyPlaying []QueueEntry `json:"currently_playing"`
}

// JoinQueueRequest is the payload for joining the queue
type JoinQueueRequest struct {
	UserID int64 `json:"user_id"`
}

// QueueResponse is the response after queue operations
type QueueResponse struct {
	Entry   *QueueEntry `json:"entry,omitempty"`
	Info    QueueInfo   `json:"info"`
	Message string      `json:"message"`
}
