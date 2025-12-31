package handler

import (
	"backend/model"
	"backend/service"
	"encoding/json"
	"fmt"
	"net/http"
)

// MatchHandler handles match endpoints
type MatchHandler struct {
	matchService *service.MatchService
}

// NewMatchHandler creates a new match handler
func NewMatchHandler(matchSvc *service.MatchService) *MatchHandler {
	return &MatchHandler{
		matchService: matchSvc,
	}
}

// GetHistory handles GET /api/matches
func (h *MatchHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	payload := getUserFromContext(r.Context())
	if payload == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	history, err := h.matchService.GetHistory(payload.UserID, 20)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get match history", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, history)
}

// Create handles POST /api/matches (Organizer only)
func (h *MatchHandler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req model.CreateMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	match, err := h.matchService.Create(req.Court, req.Team1, req.Team2)
	if err != nil {
		switch err {
		case service.ErrInvalidTeam:
			respondError(w, http.StatusBadRequest, "Invalid team configuration",
				"Each team must have 1-2 players")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to create match", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusCreated, match)
}

// RecordResult handles PUT /api/matches/result (Organizer only)
func (h *MatchHandler) RecordResult(w http.ResponseWriter, r *http.Request) {
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

	var req model.RecordResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	match, err := h.matchService.RecordResult(req.MatchID, req.Scores)
	if err != nil {
		switch err {
		case service.ErrMatchNotFound:
			respondError(w, http.StatusNotFound, "Match not found", "")
		default:
			respondError(w, http.StatusInternalServerError, "Failed to record result", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, match)
}

// GetActive handles GET /api/matches/active
func (h *MatchHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	matches, err := h.matchService.GetActive()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get active matches", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, matches)
}

// GetUserHistory handles GET /api/users/{userID}/matches (Admin only)
func (h *MatchHandler) GetUserHistory(w http.ResponseWriter, r *http.Request) {
	payload := getUserFromContext(r.Context())
	if payload == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	// Check if user is admin or organizer
	if payload.Role != "admin" && payload.Role != "organizer" {
		respondError(w, http.StatusForbidden, "Admin access required", "")
		return
	}

	// Get user ID from URL query
	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		respondError(w, http.StatusBadRequest, "User ID is required", "")
		return
	}

	var userID int64
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID", "")
		return
	}

	history, err := h.matchService.GetHistory(userID, 50)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get match history", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, history)
}
