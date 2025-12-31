package handler

import (
	"backend/model"
	"backend/service"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authSvc,
	}
}

// Register handles POST /api/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Username and password are required", "")
		return
	}

	resp, err := h.authService.Register(req)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			respondError(w, http.StatusConflict, "Username already exists", "")
		case service.ErrPhoneExists:
			respondError(w, http.StatusConflict, "Phone number already registered", "")
		case service.ErrWeakPassword:
			respondError(w, http.StatusBadRequest, "Password does not meet requirements",
				"Password must be at least 8 characters with uppercase, lowercase, number, and special character")
		case service.ErrPasswordMatch:
			respondError(w, http.StatusBadRequest, "Passwords do not match", "")
		case service.ErrPasswordSameAsUser:
			respondError(w, http.StatusBadRequest, "Password cannot be the same as username", "")
		case service.ErrInvalidPhone:
			respondError(w, http.StatusBadRequest, "Invalid phone number format", "Phone must be 10 digits starting with 08 or 09")
		default:
			respondError(w, http.StatusInternalServerError, "Registration failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusCreated, resp)
}

// Login handles POST /api/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "Username and password are required", "")
		return
	}

	resp, refreshToken, err := h.authService.Login(req)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid credentials", "")
		return
	}

	// Set refresh token as HttpOnly cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(7 * 24 * time.Hour / time.Second),
	})

	respondJSON(w, http.StatusOK, resp)
}

// Refresh handles POST /api/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		respondError(w, http.StatusUnauthorized, "No refresh token provided", "")
		return
	}

	resp, err := h.authService.RefreshAccessToken(cookie.Value)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "Invalid or expired refresh token", "")
		return
	}

	respondJSON(w, http.StatusOK, resp)
}

// Logout handles POST /api/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err == nil {
		h.authService.Logout(cookie.Value)
	}

	// Clear the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	respondJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

// ForgotPassword handles POST /api/forgot-password (verify username + phone)
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req model.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.Username == "" || req.Phone == "" {
		respondError(w, http.StatusBadRequest, "Username and phone are required", "")
		return
	}

	err := h.authService.VerifyUserForReset(req)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			respondError(w, http.StatusNotFound, "User not found", "")
		case service.ErrPhoneMismatch:
			respondError(w, http.StatusBadRequest, "Phone number does not match", "")
		default:
			respondError(w, http.StatusInternalServerError, "Verification failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Verification successful. You can now reset your password."})
}

// ResetPassword handles POST /api/reset-password
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req model.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if req.Username == "" || req.Phone == "" || req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "All fields are required", "")
		return
	}

	err := h.authService.ResetPassword(req)
	if err != nil {
		switch err {
		case service.ErrUserNotFound:
			respondError(w, http.StatusNotFound, "User not found", "")
		case service.ErrPhoneMismatch:
			respondError(w, http.StatusBadRequest, "Phone number does not match", "")
		case service.ErrWeakPassword:
			respondError(w, http.StatusBadRequest, "Password does not meet requirements", "")
		default:
			respondError(w, http.StatusInternalServerError, "Password reset failed", err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Password reset successful. You can now login with your new password."})
}

// UpdatePlayerAdmin handles PUT /api/admin/users/:id (admin only)
func (h *AuthHandler) UpdatePlayerAdmin(w http.ResponseWriter, r *http.Request) {
	// Get user ID from URL
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		// Try to get from path
		// Expected path: /api/admin/users/123
		parts := splitPath(r.URL.Path)
		if len(parts) >= 4 {
			idStr = parts[3]
		}
	}

	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || userID <= 0 {
		respondError(w, http.StatusBadRequest, "Invalid user ID", "")
		return
	}

	var req model.UpdatePlayerAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	err = h.authService.UpdatePlayerAdmin(userID, req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error(), "")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Player updated successfully"})
}

func splitPath(path string) []string {
	var parts []string
	for _, p := range splitString(path, '/') {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

func splitString(s string, sep rune) []string {
	var parts []string
	var current string
	for _, c := range s {
		if c == sep {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.NewSuccessResponse(data))
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.NewErrorResponse(status, message, details))
}
