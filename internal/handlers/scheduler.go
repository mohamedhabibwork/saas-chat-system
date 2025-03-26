package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohamedhabibwork/saas-chat-system/internal/models"
	"github.com/mohamedhabibwork/saas-chat-system/internal/services"
)

// SchedulerHandler handles report schedule management
type SchedulerHandler struct {
	schedulerService *services.SchedulerService
}

// NewSchedulerHandler creates a new scheduler handler
func NewSchedulerHandler(schedulerService *services.SchedulerService) *SchedulerHandler {
	return &SchedulerHandler{
		schedulerService: schedulerService,
	}
}

// CreateSchedule handles requests to create a new report schedule
func (h *SchedulerHandler) CreateSchedule(c *gin.Context) {
	var schedule models.ReportSchedule
	if err := c.ShouldBindJSON(&schedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant ID from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant not found"})
		return
	}
	schedule.TenantID = tenantID.(string)

	// Check if tenant has email reports feature
	subscription, err := h.getTenantSubscription(c, schedule.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error checking subscription"})
		return
	}

	if !subscription.HasFeature("email_reports") {
		c.JSON(http.StatusForbidden, gin.H{"error": "email reports feature not available in your subscription"})
		return
	}

	if err := h.schedulerService.AddSchedule(c.Request.Context(), &schedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, schedule)
}

// UpdateSchedule handles requests to update an existing report schedule
func (h *SchedulerHandler) UpdateSchedule(c *gin.Context) {
	scheduleID := c.Param("id")
	schedule, err := h.schedulerService.GetSchedule(scheduleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		return
	}

	// Verify tenant ownership
	tenantID, exists := c.Get("tenant_id")
	if !exists || schedule.TenantID != tenantID.(string) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var updatedSchedule models.ReportSchedule
	if err := c.ShouldBindJSON(&updatedSchedule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preserve tenant ID and ID
	updatedSchedule.TenantID = schedule.TenantID
	updatedSchedule.ID = schedule.ID

	if err := h.schedulerService.UpdateSchedule(c.Request.Context(), &updatedSchedule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedSchedule)
}

// DeleteSchedule handles requests to delete a report schedule
func (h *SchedulerHandler) DeleteSchedule(c *gin.Context) {
	scheduleID := c.Param("id")
	schedule, err := h.schedulerService.GetSchedule(scheduleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		return
	}

	// Verify tenant ownership
	tenantID, exists := c.Get("tenant_id")
	if !exists || schedule.TenantID != tenantID.(string) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.schedulerService.DeleteSchedule(c.Request.Context(), scheduleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "schedule deleted successfully"})
}

// GetSchedule handles requests to retrieve a report schedule
func (h *SchedulerHandler) GetSchedule(c *gin.Context) {
	scheduleID := c.Param("id")
	schedule, err := h.schedulerService.GetSchedule(scheduleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "schedule not found"})
		return
	}

	// Verify tenant ownership
	tenantID, exists := c.Get("tenant_id")
	if !exists || schedule.TenantID != tenantID.(string) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, schedule)
}

// ListSchedules handles requests to list all report schedules for a tenant
func (h *SchedulerHandler) ListSchedules(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant not found"})
		return
	}

	schedules := h.schedulerService.ListSchedules()
	tenantSchedules := make([]*models.ReportSchedule, 0)

	for _, schedule := range schedules {
		if schedule.TenantID == tenantID.(string) {
			tenantSchedules = append(tenantSchedules, schedule)
		}
	}

	c.JSON(http.StatusOK, tenantSchedules)
}

// getTenantSubscription retrieves the subscription for a tenant
func (h *SchedulerHandler) getTenantSubscription(c *gin.Context, tenantID string) (*models.Subscription, error) {
	var subscription models.Subscription
	err := c.MustGet("db").(*database.DB).Where("tenant_id = ? AND status = ?", tenantID, "active").First(&subscription).Error
	if err != nil {
		return nil, err
	}
	return &subscription, nil
} 