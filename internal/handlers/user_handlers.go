package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"saas-chat-system/internal/database"
	"saas-chat-system/internal/services"
)

// UserPreferences represents user configurable settings
type UserPreferences struct {
	Timezone string `json:"timezone"`
}

// @Summary      Get user preferences
// @Description  Get the current user's preferences including timezone
// @Tags         Users
// @Accept       json
// @Produce      json
// @Success      200 {object} UserPreferences "User preferences"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/users/preferences [get]
func GetUserPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		RespondWithError(w, NewAPIError(http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated"))
		return
	}

	// Get user's timezone from database
	var timezone string
	err := database.DB.QueryRow("SELECT timezone FROM users WHERE id = $1", userID).Scan(&timezone)
	if err != nil {
		timezone = "UTC" // Default to UTC if not found
	}

	// Return user preferences
	prefs := UserPreferences{
		Timezone: timezone,
	}

	RespondWithJSON(w, http.StatusOK, prefs)
}

// @Summary      Update user preferences
// @Description  Update the current user's preferences
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        preferences body UserPreferences true "User preferences"
// @Success      200 {object} UserPreferences "Updated preferences"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/users/preferences [put]
func UpdateUserPreferencesHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		RespondWithError(w, NewAPIError(http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated"))
		return
	}

	// Parse request body
	var prefs UserPreferences
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		RespondWithError(w, NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "Invalid request body"))
		return
	}

	// Validate timezone
	if prefs.Timezone != "" {
		_, err := time.LoadLocation(prefs.Timezone)
		if err != nil {
			RespondWithError(w, NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "Invalid timezone"))
			return
		}
	}

	// Update user's timezone in database
	_, err := database.DB.Exec("UPDATE users SET timezone = $1, updated_at = NOW() WHERE id = $2",
		prefs.Timezone, userID)
	if err != nil {
		RespondWithError(w, NewAPIError(http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update preferences"))
		return
	}

	// Return updated preferences
	RespondWithJSON(w, http.StatusOK, prefs)
}

// @Summary      Get timezone list
// @Description  Get a list of supported IANA timezone names
// @Tags         Users
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string][]string "List of timezones"
// @Router       /api/v1/users/timezones [get]
func GetTimezoneListHandler(w http.ResponseWriter, r *http.Request) {
	// This is a simplified list of timezones
	// In a real app, you might want to generate this dynamically or include more information
	timezones := []string{
		"UTC",
		"Europe/London",
		"Europe/Paris",
		"Europe/Berlin",
		"America/New_York",
		"America/Chicago",
		"America/Denver",
		"America/Los_Angeles",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Asia/Kolkata",
		"Australia/Sydney",
		"Pacific/Auckland",
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"timezones": timezones,
	})
}

// UserHandlers handles user-related HTTP requests
type UserHandlers struct {
	userService *services.UserService
}

// NewUserHandlers creates a new UserHandlers instance
func NewUserHandlers(userService *services.UserService) *UserHandlers {
	return &UserHandlers{
		userService: userService,
	}
}

// @Summary      Create a new user
// @Description  Create a new user account
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        user body models.User true "User details"
// @Success      201 {object} models.User "User created successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      409 {object} map[string]interface{} "User already exists"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/users [post]
func (h *UserHandlers) HandleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user services.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		SendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.userService.Create(r.Context(), &user)
	if err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	SendResponse(w, true, user, "User created successfully", http.StatusCreated)
}

// @Summary      Get user details
// @Description  Get details of a specific user
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Success      200 {object} services.User "User details"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "User not found"
// @Router       /api/v1/users/{id} [get]
func (h *UserHandlers) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		SendResponse(w, false, nil, "User ID is required", http.StatusBadRequest)
		return
	}

	user, err := h.userService.Get(r.Context(), userID)
	if err != nil {
		SendResponse(w, false, nil, "User not found", http.StatusNotFound)
		return
	}

	SendResponse(w, true, user, "User retrieved successfully", http.StatusOK)
}

// @Summary      Update user details
// @Description  Update an existing user's information
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Param        user body services.User true "Updated user details"
// @Success      200 {object} services.User "User updated successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "User not found"
// @Router       /api/v1/users/{id} [put]
func (h *UserHandlers) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		SendResponse(w, false, nil, "User ID is required", http.StatusBadRequest)
		return
	}

	var user services.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		SendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	user.ID = userID
	err := h.userService.Update(r.Context(), &user)
	if err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	SendResponse(w, true, user, "User updated successfully", http.StatusOK)
}

// @Summary      Delete a user
// @Description  Delete an existing user account
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        id path string true "User ID"
// @Success      204 "User deleted successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "User not found"
// @Router       /api/v1/users/{id} [delete]
func (h *UserHandlers) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		SendResponse(w, false, nil, "User ID is required", http.StatusBadRequest)
		return
	}

	if err := h.userService.Delete(r.Context(), userID); err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	SendResponse(w, true, nil, "User deleted successfully", http.StatusNoContent)
}

// @Summary      List all users
// @Description  Get a list of all users with optional filtering
// @Tags         Users
// @Accept       json
// @Produce      json
// @Param        page query integer false "Page number"
// @Param        limit query integer false "Number of items per page"
// @Success      200 {array} services.User "List of users"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/users [get]
func (h *UserHandlers) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}

	users, err := h.userService.List(r.Context(), page, limit)
	if err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	SendResponse(w, true, users, "Users retrieved successfully", http.StatusOK)
}
