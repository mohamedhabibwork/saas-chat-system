package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"awesomeProject/internal/database"
	"awesomeProject/internal/models"
)

// UserPreferences represents user configurable settings
type UserPreferences struct {
	Timezone string `json:"timezone"`
}

// GetUserPreferencesHandler handles retrieving user preferences
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

// UpdateUserPreferencesHandler handles updating user preferences
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

// GetTimezoneListHandler returns a list of valid IANA timezone names
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