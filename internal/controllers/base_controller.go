package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"awesomeProject/internal/models"
	"awesomeProject/internal/database"
)

// BaseController provides common functionality for all controllers
type BaseController struct {
	db *sql.DB
}

// NewBaseController creates a new base controller
func NewBaseController() *BaseController {
	return &BaseController{
		db: database.DB,
	}
}

// RespondWithError sends an error response
func (c *BaseController) RespondWithError(w http.ResponseWriter, err *models.APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	_ = json.NewEncoder(w).Encode(err)
}

// RespondWithJSON sends a JSON response
func (c *BaseController) RespondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// NewAPIError creates a new API error
func (c *BaseController) NewAPIError(status int, code, message string) *models.APIError {
	return &models.APIError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

// GetCurrentTime returns the current time
func (c *BaseController) GetCurrentTime() time.Time {
	return time.Now()
}

// ValidateTenant checks if a tenant exists
func (c *BaseController) ValidateTenant(tenantID int) bool {
	var exists bool
	err := c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1)", tenantID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

// ValidateUser checks if a user exists and belongs to a tenant
func (c *BaseController) ValidateUser(userID, tenantID int) bool {
	var exists bool
	err := c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND tenant_id = $2)", userID, tenantID).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
} 