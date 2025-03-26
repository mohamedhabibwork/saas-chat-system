package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mohamedhabibwork/saas-chat-system/internal/services"
)

// ReportingHandler handles report generation requests
type ReportingHandler struct {
	reportingService *services.ReportingService
}

// NewReportingHandler creates a new reporting handler
func NewReportingHandler(reportingService *services.ReportingService) *ReportingHandler {
	return &ReportingHandler{
		reportingService: reportingService,
	}
}

// GenerateUserActivityReport handles requests to generate user activity reports
func (h *ReportingHandler) GenerateUserActivityReport(c *gin.Context) {
	var opts services.ReportOptions
	if err := c.ShouldBindJSON(&opts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default time range if not provided
	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now().Add(-24 * time.Hour)
	}
	if opts.EndTime.IsZero() {
		opts.EndTime = time.Now()
	}

	// Get user and tenant IDs from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	opts.UserID = userID.(string)

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant not found"})
		return
	}
	opts.TenantID = tenantID.(string)

	report, err := h.reportingService.GenerateUserActivityReport(c.Request.Context(), opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateLocationReport handles requests to generate location reports
func (h *ReportingHandler) GenerateLocationReport(c *gin.Context) {
	var opts services.ReportOptions
	if err := c.ShouldBindJSON(&opts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default time range if not provided
	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now().Add(-24 * time.Hour)
	}
	if opts.EndTime.IsZero() {
		opts.EndTime = time.Now()
	}

	// Get user and tenant IDs from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	opts.UserID = userID.(string)

	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant not found"})
		return
	}
	opts.TenantID = tenantID.(string)

	report, err := h.reportingService.GenerateLocationReport(c.Request.Context(), opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GenerateSystemHealthReport handles requests to generate system health reports
func (h *ReportingHandler) GenerateSystemHealthReport(c *gin.Context) {
	var opts services.ReportOptions
	if err := c.ShouldBindJSON(&opts); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default time range if not provided
	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now().Add(-24 * time.Hour)
	}
	if opts.EndTime.IsZero() {
		opts.EndTime = time.Now()
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant not found"})
		return
	}
	opts.TenantID = tenantID.(string)

	report, err := h.reportingService.GenerateSystemHealthReport(c.Request.Context(), opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
} 