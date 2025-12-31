package handler

import (
	"backend/model"
	"backend/service"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// ContextKey type for context values
type ContextKey string

const (
	// UserContextKey is the key for user info in context
	UserContextKey ContextKey = "user"
)

// UserHandler handles user profile endpoints
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userSvc,
	}
}

// GetProfile handles GET /api/profile
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	payload := getUserFromContext(r.Context())
	if payload == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	profile, err := h.userService.GetProfile(payload.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found", "")
		return
	}

	respondJSON(w, http.StatusOK, profile)
}

// UpdateProfile handles PUT /api/profile
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	payload := getUserFromContext(r.Context())
	if payload == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	var req model.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(payload.UserID, req)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found", "")
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// GetStats handles GET /api/profile/stats
func (h *UserHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	payload := getUserFromContext(r.Context())
	if payload == nil {
		respondError(w, http.StatusUnauthorized, "Unauthorized", "")
		return
	}

	stats, err := h.userService.GetStats(payload.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Stats not found", "")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

// GetUserProfile handles GET /api/users/{userID}/profile (Admin only)
func (h *UserHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
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

	// Get user ID from URL
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

	profile, err := h.userService.GetProfile(userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found", "")
		return
	}

	respondJSON(w, http.StatusOK, profile)
}

// getUserFromContext extracts user payload from request context
func getUserFromContext(ctx context.Context) *model.TokenPayload {
	payload, ok := ctx.Value(UserContextKey).(*model.TokenPayload)
	if !ok {
		return nil
	}
	return payload
}
