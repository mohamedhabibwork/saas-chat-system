package handlers

import (
	"encoding/json"
	"net/http"

	"saas-chat-system/internal/models"
	"saas-chat-system/internal/services"
)

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ResetPasswordRequest represents the password reset request body
type ResetPasswordRequest struct {
	Email string `json:"email"`
}

// NewPasswordRequest represents the new password request body
type NewPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// Response represents a generic API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// @Summary      Register user
// @Description  Create a new user account
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "Registration details"
// @Success      201 {object} models.User "User created successfully"
// @Failure      400 {object} Response "Bad Request"
// @Failure      500 {object} Response "Internal Server Error"
// @Router       /api/auth/register [post]
func HandleRegister(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Username == "" || req.Email == "" || req.Password == "" {
			sendResponse(w, false, nil, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Create user
		user := &models.User{
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: req.Password,
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			IsActive:     true,
		}

		if err := authService.Register(user); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, user, "", http.StatusCreated)
	}
}

// @Summary      User login
// @Description  Authenticate user and get session token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body LoginRequest true "Login credentials"
// @Success      200 {object} models.Session "Login successful"
// @Failure      400 {object} Response "Bad Request"
// @Failure      401 {object} Response "Invalid credentials"
// @Router       /api/auth/login [post]
func HandleLogin(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Username == "" || req.Password == "" {
			sendResponse(w, false, nil, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Get device info and IP address
		deviceInfo := r.Header.Get("User-Agent")
		ipAddress := r.RemoteAddr

		// Login user
		session, err := authService.Login(req.Username, req.Password, deviceInfo, ipAddress)
		if err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusUnauthorized)
			return
		}

		sendResponse(w, true, session, "", http.StatusOK)
	}
}

// @Summary      Logout user
// @Description  Invalidate user session
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        X-Session-ID header string true "Session ID"
// @Success      200 {object} Response "Logout successful"
// @Failure      400 {object} Response "Bad Request"
// @Failure      401 {object} Response "Invalid session"
// @Router       /api/auth/logout [post]
func HandleLogout(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get session ID from request
		sessionToken := r.Header.Get("X-Session-ID")
		if sessionToken == "" {
			sendResponse(w, false, nil, "Missing session ID", http.StatusBadRequest)
			return
		}

		// Validate session and get session ID
		session, err := authService.ValidateSession(sessionToken)
		if err != nil {
			sendResponse(w, false, nil, "Invalid or expired session", http.StatusUnauthorized)
			return
		}

		// Logout user
		if err := authService.Logout(session.ID); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, nil, "", http.StatusOK)
	}
}

// @Summary      Request password reset
// @Description  Send password reset email to user
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body ResetPasswordRequest true "Email address"
// @Success      200 {object} Response "Reset email sent"
// @Failure      400 {object} Response "Bad Request"
// @Failure      500 {object} Response "Internal Server Error"
// @Router       /api/auth/reset-password [post]
func HandleResetPassword(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ResetPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate email
		if req.Email == "" {
			sendResponse(w, false, nil, "Missing email", http.StatusBadRequest)
			return
		}

		// Request password reset
		if err := authService.RequestPasswordReset(req.Email); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, nil, "", http.StatusOK)
	}
}

// @Summary      Set new password
// @Description  Set a new password using a reset token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request body NewPasswordRequest true "New password details"
// @Success      200 {object} Response "Password updated successfully"
// @Failure      400 {object} Response "Bad Request"
// @Failure      500 {object} Response "Internal Server Error"
// @Router       /api/auth/new-password [post]
func HandleNewPassword(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req NewPasswordRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Token == "" || req.NewPassword == "" {
			sendResponse(w, false, nil, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Reset password
		if err := authService.ResetPassword(req.Token, req.NewPassword); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, nil, "", http.StatusOK)
	}
}

// Helper function to send JSON responses
func sendResponse(w http.ResponseWriter, success bool, data interface{}, errMsg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success: success,
		Data:    data,
		Error:   errMsg,
	}

	json.NewEncoder(w).Encode(response)
}
