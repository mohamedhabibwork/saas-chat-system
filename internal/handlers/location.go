package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohamedhabibwork/saas-chat-system/internal/models"
	"github.com/mohamedhabibwork/saas-chat-system/internal/services"
)

// LocationHandler handles location-related HTTP requests
type LocationHandler struct {
	locationService *services.LocationService
}

// NewLocationHandler creates a new location handler
func NewLocationHandler(locationService *services.LocationService) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
	}
}

// UpdateLocation handles location updates
func (h *LocationHandler) UpdateLocation(c *gin.Context) {
	var location models.Location
	if err := c.ShouldBindJSON(&location); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set tenant ID from context
	location.TenantID = c.GetString("tenant_id")
	location.UserID = c.GetString("user_id")

	if err := h.locationService.UpdateLocation(c.Request.Context(), &location); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, location)
}

// GetCurrentLocation retrieves the current location
func (h *LocationHandler) GetCurrentLocation(c *gin.Context) {
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	location, err := h.locationService.GetCurrentLocation(c.Request.Context(), userID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, location)
}

// GetLocationHistory retrieves location history
func (h *LocationHandler) GetLocationHistory(c *gin.Context) {
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	// Parse time range from query parameters
	startTime := time.Now().Add(-24 * time.Hour) // Default to last 24 hours
	endTime := time.Now()

	if startStr := c.Query("start_time"); startStr != "" {
		if t, err := time.Parse(time.RFC3339, startStr); err == nil {
			startTime = t
		}
	}

	if endStr := c.Query("end_time"); endStr != "" {
		if t, err := time.Parse(time.RFC3339, endStr); err == nil {
			endTime = t
		}
	}

	locations, err := h.locationService.GetLocationHistory(c.Request.Context(), userID, tenantID, startTime, endTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, locations)
}

// SaveLocationHistory saves location history
func (h *LocationHandler) SaveLocationHistory(c *gin.Context) {
	var history models.LocationHistory
	if err := c.ShouldBindJSON(&history); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set tenant and user IDs from context
	history.TenantID = c.GetString("tenant_id")
	history.UserID = c.GetString("user_id")

	if err := h.locationService.SaveLocationHistory(c.Request.Context(), &history); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// GetLocationStats retrieves location statistics
func (h *LocationHandler) GetLocationStats(c *gin.Context) {
	userID := c.GetString("user_id")
	tenantID := c.GetString("tenant_id")

	stats, err := h.locationService.GetLocationStats(c.Request.Context(), userID, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
} 