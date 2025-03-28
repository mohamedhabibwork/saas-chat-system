package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"saas-chat-system/internal/models"
	"saas-chat-system/internal/services"
)

// TrackingHandler handles tracking-related HTTP requests
type TrackingHandler struct {
	trackingService *services.TrackingService
}

// NewTrackingHandler creates a new tracking handler
func NewTrackingHandler(trackingService *services.TrackingService) *TrackingHandler {
	return &TrackingHandler{
		trackingService: trackingService,
	}
}

// @Summary      Track user events
// @Description  Record user events for analytics and tracking
// @Tags         Tracking
// @Accept       json
// @Produce      json
// @Param        events body []models.TrackingEvent true "Array of events to track"
// @Success      200 {object} map[string]interface{} "Events tracked successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/tracking/events [post]
func (h *TrackingHandler) TrackEvents(c *gin.Context) {
	var events []models.TrackingEvent
	if err := c.ShouldBindJSON(&events); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID := c.GetString("tenant_id")
	for i := range events {
		events[i].TenantID = tenantID
	}

	for _, event := range events {
		if err := h.trackingService.TrackEvent(c.Request.Context(), &event); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Events tracked successfully"})
}

// @Summary      Track system metrics
// @Description  Record system metrics for monitoring
// @Tags         Tracking
// @Accept       json
// @Produce      json
// @Param        metrics body []models.TrackingMetric true "Array of metrics to track"
// @Success      200 {object} map[string]interface{} "Metrics tracked successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/tracking/metrics [post]
func (h *TrackingHandler) TrackMetrics(c *gin.Context) {
	var metrics []models.TrackingMetric
	if err := c.ShouldBindJSON(&metrics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID := c.GetString("tenant_id")
	for i := range metrics {
		metrics[i].TenantID = tenantID
	}

	for _, metric := range metrics {
		if err := h.trackingService.TrackMetric(c.Request.Context(), &metric); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Metrics tracked successfully"})
}

// @Summary      Track system errors
// @Description  Record system errors for monitoring and debugging
// @Tags         Tracking
// @Accept       json
// @Produce      json
// @Param        errors body []models.TrackingError true "Array of errors to track"
// @Success      200 {object} map[string]interface{} "Errors tracked successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/tracking/errors [post]
func (h *TrackingHandler) TrackErrors(c *gin.Context) {
	var errors []models.TrackingError
	if err := c.ShouldBindJSON(&errors); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tenantID := c.GetString("tenant_id")
	for i := range errors {
		errors[i].TenantID = tenantID
	}

	for _, err := range errors {
		if err := h.trackingService.TrackError(c.Request.Context(), &err); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Errors tracked successfully"})
}

// @Summary      Get tracking statistics
// @Description  Retrieve statistics about tracked events, metrics, and errors
// @Tags         Tracking
// @Accept       json
// @Produce      json
// @Success      200 {object} models.TrackingStats "Tracking statistics"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/tracking/stats [get]
func (h *TrackingHandler) GetTrackingStats(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	stats, err := h.trackingService.GetTrackingStats(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// @Summary      Get tracked events
// @Description  Retrieve tracked events with optional filtering
// @Tags         Tracking
// @Accept       json
// @Produce      json
// @Param        limit query int false "Maximum number of events to return"
// @Param        offset query int false "Number of events to skip"
// @Success      200 {array} models.TrackingEvent "List of events"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/tracking/events [get]
func (h *TrackingHandler) GetEvents(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	events, err := h.trackingService.GetEvents(c.Request.Context(), tenantID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

// @Summary      Get tracked metrics
// @Description  Retrieve tracked metrics with optional filtering
// @Tags         Tracking
// @Accept       json
// @Produce      json
// @Param        limit query int false "Maximum number of metrics to return"
// @Param        offset query int false "Number of metrics to skip"
// @Success      200 {array} models.TrackingMetric "List of metrics"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/tracking/metrics [get]
func (h *TrackingHandler) GetMetrics(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	metrics, err := h.trackingService.GetMetrics(c.Request.Context(), tenantID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// @Summary      Get tracked errors
// @Description  Retrieve tracked errors with optional filtering
// @Tags         Tracking
// @Accept       json
// @Produce      json
// @Param        limit query int false "Maximum number of errors to return"
// @Param        offset query int false "Number of errors to skip"
// @Success      200 {array} models.TrackingError "List of errors"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/tracking/errors [get]
func (h *TrackingHandler) GetErrors(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	errors, err := h.trackingService.GetErrors(c.Request.Context(), tenantID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, errors)
}

// @Summary      Clean up old tracking data
// @Description  Remove old tracking data based on retention policy
// @Tags         Tracking
// @Accept       json
// @Produce      json
// @Param        older_than query string false "Duration of data to keep (e.g., '30d')"
// @Success      200 {object} map[string]interface{} "Cleanup completed successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/tracking/cleanup [post]
func (h *TrackingHandler) CleanupOldData(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	olderThan := c.DefaultQuery("older_than", "30d")

	duration, err := parseDuration(olderThan)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid duration format"})
		return
	}

	if err := h.trackingService.CleanupOldData(c.Request.Context(), tenantID, duration); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Old data cleaned up successfully"})
}

// parseDuration parses a duration string into a time.Duration
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}
