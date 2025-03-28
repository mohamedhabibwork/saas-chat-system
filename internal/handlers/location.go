package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"saas-chat-system/internal/models"
	"saas-chat-system/internal/services"
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

// @Summary      Update user location
// @Description  Update the current location of a user
// @Tags         Location
// @Accept       json
// @Produce      json
// @Param        location body models.Location true "Location data"
// @Success      200 {object} models.Location "Location updated successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/location/current [post]
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

// @Summary      Get user's current location
// @Description  Retrieve the current location of a user
// @Tags         Location
// @Accept       json
// @Produce      json
// @Success      200 {object} models.Location "Current location"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Location not found"
// @Router       /api/v1/location/current [get]
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

// @Summary      Get location history
// @Description  Retrieve location history for a user
// @Tags         Location
// @Accept       json
// @Produce      json
// @Param        start_time query string false "Start time in RFC3339 format"
// @Param        end_time query string false "End time in RFC3339 format"
// @Success      200 {array} models.Location "Location history"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/location/history [get]
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

// @Summary      Save location history
// @Description  Save a batch of location history data
// @Tags         Location
// @Accept       json
// @Produce      json
// @Param        history body models.LocationHistory true "Location history data"
// @Success      200 {object} models.LocationHistory "Location history saved successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/location/history [post]
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

// @Summary      Get location statistics
// @Description  Retrieve statistics about user locations
// @Tags         Location
// @Accept       json
// @Produce      json
// @Success      200 {object} models.LocationStats "Location statistics"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/location/stats [get]
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
