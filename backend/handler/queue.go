package handler

import (
	"backend/model"
	"backend/service"
	"net/http"
)

// QueueHandler handles queue endpoints
type QueueHandler struct {
	queueService *service.QueueService
}

// NewQueueHandler creates a new queue handler
func NewQueueHandler(queueSvc *service.QueueService) *QueueHandler {
	return &QueueHandler{
		queueService: queueSvc,
	}
}

// GetStatus handles GET /api/queue
func (h *QueueHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	payload := getUserFromContext(r.Context())

	var userID int64 = 0
	if payload != nil {
		userID = payload.UserID
	}

	info, err := h.queueService.GetStatus(userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get queue status", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, info)
}

// Join handles POST /api/queue/join
func (h *QueueHandler) Join(w http.ResponseWriter, r *http.Request) {
	payload := getUserFromContext(r.Context())
	if payload == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	entry, err := h.queueService.Join(payload.UserID)
	if err != nil {
		switch err {
		case service.ErrAlreadyInQueue:
			respondError(w, http.StatusConflict, "Already in queue", "")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to join queue", err.Error())
		}
		return
	}

	// Get updated status
	info, _ := h.queueService.GetStatus(payload.UserID)

	respondJSON(w, http.StatusOK, model.QueueResponse{
		Entry:   entry,
		Info:    *info,
		Message: "Joined queue successfully",
	})
}

// Leave handles POST /api/queue/leave
func (h *QueueHandler) Leave(w http.ResponseWriter, r *http.Request) {
	payload := getUserFromContext(r.Context())
	if payload == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	err := h.queueService.Leave(payload.UserID)
	if err != nil {
		switch err {
		case service.ErrNotInQueue:
			respondError(w, http.StatusNotFound, "Not in queue", "")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to leave queue", err.Error())
		}
		return
	}

	// Get updated status
	info, _ := h.queueService.GetStatus(payload.UserID)

	respondJSON(w, http.StatusOK, model.QueueResponse{
		Info:    *info,
		Message: "Left queue successfully",
	})
}

// CallNext handles POST /api/queue/call (Organizer only)
func (h *QueueHandler) CallNext(w http.ResponseWriter, r *http.Request) {
	payload := getUserFromContext(r.Context())
	if payload == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Check if user is organizer or admin
	if payload.Role != "organizer" && payload.Role != "admin" {
		respondError(w, http.StatusForbidden, "Organizer access required", "")
		return
	}

	// Call 4 players (for a doubles match)
	called, err := h.queueService.CallNext(4)
	if err != nil {
		if err == service.ErrQueueEmpty {
			respondJSON(w, http.StatusOK, map[string]interface{}{
				"called":  []model.QueueEntry{},
				"count":   0,
				"message": "Queue is empty",
			})
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to call players", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"called": called,
		"count":  len(called),
	})
}
