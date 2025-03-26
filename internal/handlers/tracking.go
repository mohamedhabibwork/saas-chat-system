package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mohamedhabibwork/saas-chat-system/internal/models"
	"github.com/mohamedhabibwork/saas-chat-system/internal/services"
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

// TrackEvents handles batch event tracking
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

// TrackMetrics handles batch metric tracking
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

// TrackErrors handles batch error tracking
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

// GetTrackingStats retrieves tracking statistics
func (h *TrackingHandler) GetTrackingStats(c *gin.Context) {
	tenantID := c.GetString("tenant_id")

	stats, err := h.trackingService.GetTrackingStats(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetEvents retrieves tracking events
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

// GetMetrics retrieves tracking metrics
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

// GetErrors retrieves tracking errors
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

// CleanupOldData removes old tracking data
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