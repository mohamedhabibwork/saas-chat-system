package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"saas-chat-system/internal/models"
	"saas-chat-system/internal/services"
)

// SchedulerHandler handles report schedule management
type SchedulerHandler struct {
	schedulerService *services.SchedulerService
	db              *models.Database
}

// NewSchedulerHandler creates a new scheduler handler
func NewSchedulerHandler(schedulerService *services.SchedulerService, db *models.Database) *SchedulerHandler {
	return &SchedulerHandler{
		schedulerService: schedulerService,
		db:              db,
	}
}

// @Summary      Create a report schedule
// @Description  Create a new scheduled report
// @Tags         Scheduler
// @Accept       json
// @Produce      json
// @Param        schedule body models.ReportSchedule true "Schedule details"
// @Success      201 {object} models.ReportSchedule "Schedule created successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Forbidden - Feature not available"
// @Router       /api/v1/scheduler/schedules [post]
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

// @Summary      Update a report schedule
// @Description  Update an existing report schedule
// @Tags         Scheduler
// @Accept       json
// @Produce      json
// @Param        id path string true "Schedule ID"
// @Param        schedule body models.ReportSchedule true "Updated schedule details"
// @Success      200 {object} models.ReportSchedule "Schedule updated successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Schedule not found"
// @Router       /api/v1/scheduler/schedules/{id} [put]
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

// @Summary      Delete a report schedule
// @Description  Delete an existing report schedule
// @Tags         Scheduler
// @Accept       json
// @Produce      json
// @Param        id path string true "Schedule ID"
// @Success      200 {object} map[string]interface{} "Schedule deleted successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Schedule not found"
// @Router       /api/v1/scheduler/schedules/{id} [delete]
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

// @Summary      Get a report schedule
// @Description  Get details of a specific report schedule
// @Tags         Scheduler
// @Accept       json
// @Produce      json
// @Param        id path string true "Schedule ID"
// @Success      200 {object} models.ReportSchedule "Schedule details"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Schedule not found"
// @Router       /api/v1/scheduler/schedules/{id} [get]
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

// @Summary      List report schedules
// @Description  Get all report schedules for the current tenant
// @Tags         Scheduler
// @Accept       json
// @Produce      json
// @Success      200 {array} models.ReportSchedule "List of schedules"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/scheduler/schedules [get]
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
	// TODO: Implement subscription retrieval from database
	return nil, fmt.Errorf("not implemented")
}
