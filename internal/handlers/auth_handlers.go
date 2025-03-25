package handlers

import (
	"encoding/json"
	"net/http"

	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
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

// HandleRegister handles user registration
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
			Username:  req.Username,
			Email:     req.Email,
			PasswordHash: req.Password,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			IsActive:  true,
		}

		if err := authService.Register(user); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, user, "", http.StatusCreated)
	}
}

// HandleLogin handles user login
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

// HandleLogout handles user logout
func HandleLogout(authService *services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get session ID from request
		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			sendResponse(w, false, nil, "Missing session ID", http.StatusBadRequest)
			return
		}

		// Logout user
		if err := authService.Logout(sessionID); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, nil, "", http.StatusOK)
	}
}

// HandleResetPassword handles password reset requests
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

// HandleNewPassword handles setting a new password
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