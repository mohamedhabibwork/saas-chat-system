package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"saas-chat-system/internal/models"
	"saas-chat-system/internal/services"
)

// TicketHandler handles ticket-related HTTP requests
type TicketHandler struct {
	ticketService *services.TicketService
}

// NewTicketHandler creates a new ticket handler
func NewTicketHandler(ticketService *services.TicketService) *TicketHandler {
	return &TicketHandler{
		ticketService: ticketService,
	}
}

// @Summary      Create a support ticket
// @Description  Create a new support ticket
// @Tags         Tickets
// @Accept       json
// @Produce      json
// @Param        ticket body models.Ticket true "Ticket details"
// @Success      201 {object} models.Ticket "Ticket created successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/tickets [post]
func (h *TicketHandler) CreateTicket(c *gin.Context) {
	var ticket models.Ticket
	if err := c.ShouldBindJSON(&ticket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get tenant and user IDs from context
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant not found"})
		return
	}
	ticket.TenantID = tenantID.(string)

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}
	ticket.UserID = userID.(string)

	if err := h.ticketService.CreateTicket(c.Request.Context(), &ticket); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

// @Summary      Get ticket details
// @Description  Get details of a specific ticket
// @Tags         Tickets
// @Accept       json
// @Produce      json
// @Param        id path string true "Ticket ID"
// @Success      200 {object} models.Ticket "Ticket details"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Ticket not found"
// @Router       /api/v1/tickets/{id} [get]
func (h *TicketHandler) GetTicket(c *gin.Context) {
	ticketID := c.Param("id")
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	// Verify tenant ownership
	tenantID, exists := c.Get("tenant_id")
	if !exists || ticket.TenantID != tenantID.(string) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

// @Summary      Update ticket details
// @Description  Update an existing ticket
// @Tags         Tickets
// @Accept       json
// @Produce      json
// @Param        id path string true "Ticket ID"
// @Param        ticket body models.Ticket true "Updated ticket details"
// @Success      200 {object} models.Ticket "Ticket updated successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Ticket not found"
// @Router       /api/v1/tickets/{id} [put]
func (h *TicketHandler) UpdateTicket(c *gin.Context) {
	ticketID := c.Param("id")
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	// Verify tenant ownership
	tenantID, exists := c.Get("tenant_id")
	if !exists || ticket.TenantID != tenantID.(string) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var updatedTicket models.Ticket
	if err := c.ShouldBindJSON(&updatedTicket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preserve tenant ID and ID
	updatedTicket.TenantID = ticket.TenantID
	updatedTicket.ID = ticket.ID

	if err := h.ticketService.UpdateTicket(c.Request.Context(), &updatedTicket); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedTicket)
}

// @Summary      Update ticket status
// @Description  Update the status of a ticket
// @Tags         Tickets
// @Accept       json
// @Produce      json
// @Param        id path string true "Ticket ID"
// @Param        status body models.TicketStatus true "New ticket status"
// @Success      200 {object} map[string]interface{} "Ticket status updated successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Ticket not found"
// @Router       /api/v1/tickets/{id}/status [put]
func (h *TicketHandler) UpdateTicketStatus(c *gin.Context) {
	ticketID := c.Param("id")
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	// Verify tenant ownership
	tenantID, exists := c.Get("tenant_id")
	if !exists || ticket.TenantID != tenantID.(string) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var statusUpdate struct {
		Status models.TicketStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&statusUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.ticketService.UpdateTicketStatus(c.Request.Context(), ticketID, statusUpdate.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "ticket status updated successfully"})
}

// @Summary      List support tickets
// @Description  Get all support tickets with optional filtering
// @Tags         Tickets
// @Accept       json
// @Produce      json
// @Param        status query string false "Filter by status"
// @Param        category query string false "Filter by category"
// @Param        priority query string false "Filter by priority"
// @Param        assigned_to query string false "Filter by assigned user"
// @Success      200 {array} models.Ticket "List of tickets"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/tickets [get]
func (h *TicketHandler) ListTickets(c *gin.Context) {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant not found"})
		return
	}

	// Build filters from query parameters
	filters := make(map[string]interface{})
	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}
	if category := c.Query("category"); category != "" {
		filters["category"] = category
	}
	if priority := c.Query("priority"); priority != "" {
		filters["priority"] = priority
	}
	if assignedTo := c.Query("assigned_to"); assignedTo != "" {
		filters["assigned_to"] = assignedTo
	}

	tickets, err := h.ticketService.ListTickets(c.Request.Context(), tenantID.(string), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tickets)
}

// @Summary      Add comment to ticket
// @Description  Add a new comment to a ticket
// @Tags         Tickets
// @Accept       json
// @Produce      json
// @Param        id path string true "Ticket ID"
// @Param        comment body models.TicketComment true "Comment details"
// @Success      201 {object} models.TicketComment "Comment added successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Ticket not found"
// @Router       /api/v1/tickets/{id}/comments [post]
func (h *TicketHandler) AddComment(c *gin.Context) {
	ticketID := c.Param("id")
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	// Verify tenant ownership
	tenantID, exists := c.Get("tenant_id")
	if !exists || ticket.TenantID != tenantID.(string) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var comment models.TicketComment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	comment.TicketID = ticketID
	comment.UserID = c.MustGet("user_id").(string)

	if err := h.ticketService.AddComment(c.Request.Context(), &comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

// @Summary      Get ticket comments
// @Description  Get all comments for a ticket
// @Tags         Tickets
// @Accept       json
// @Produce      json
// @Param        id path string true "Ticket ID"
// @Success      200 {array} models.TicketComment "List of comments"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Ticket not found"
// @Router       /api/v1/tickets/{id}/comments [get]
func (h *TicketHandler) GetComments(c *gin.Context) {
	ticketID := c.Param("id")
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	// Verify tenant ownership
	tenantID, exists := c.Get("tenant_id")
	if !exists || ticket.TenantID != tenantID.(string) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	comments, err := h.ticketService.GetComments(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, comments)
}

// @Summary      Upload attachment to ticket
// @Description  Upload a file attachment to a ticket
// @Tags         Tickets
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path string true "Ticket ID"
// @Param        file formData file true "File to upload"
// @Success      201 {object} models.TicketAttachment "Attachment uploaded successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Ticket not found"
// @Router       /api/v1/tickets/{id}/attachments [post]
func (h *TicketHandler) UploadAttachment(c *gin.Context) {
	ticketID := c.Param("id")
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	// Verify tenant ownership
	tenantID, exists := c.Get("tenant_id")
	if !exists || ticket.TenantID != tenantID.(string) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file uploaded"})
		return
	}

	// TODO: Implement file upload to storage service
	// For now, we'll just create a placeholder attachment
	attachment := &models.TicketAttachment{
		TicketID: ticketID,
		FileName: file.Filename,
		FileType: filepath.Ext(file.Filename),
		FileSize: file.Size,
		FileURL:  "placeholder_url", // Replace with actual file URL after upload
	}

	if err := h.ticketService.AddAttachment(c.Request.Context(), attachment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, attachment)
}

// @Summary      Get ticket attachments
// @Description  Get all attachments for a ticket
// @Tags         Tickets
// @Accept       json
// @Produce      json
// @Param        id path string true "Ticket ID"
// @Success      200 {array} models.TicketAttachment "List of attachments"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Ticket not found"
// @Router       /api/v1/tickets/{id}/attachments [get]
func (h *TicketHandler) GetAttachments(c *gin.Context) {
	ticketID := c.Param("id")
	ticket, err := h.ticketService.GetTicket(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ticket not found"})
		return
	}

	// Verify tenant ownership
	tenantID, exists := c.Get("tenant_id")
	if !exists || ticket.TenantID != tenantID.(string) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	attachments, err := h.ticketService.GetAttachments(c.Request.Context(), ticketID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, attachments)
}

// @Summary      Delete ticket attachment
// @Description  Delete a specific attachment from a ticket
// @Tags         Tickets
// @Accept       json
// @Produce      json
// @Param        id path string true "Ticket ID"
// @Param        attachment_id path string true "Attachment ID"
// @Success      204 "Attachment deleted successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Attachment not found"
// @Router       /api/v1/tickets/{id}/attachments/{attachment_id} [delete]
func (h *TicketHandler) DeleteAttachment(c *gin.Context) {
	attachmentID := c.Param("attachment_id")
	// TODO: Implement attachment ownership verification
	if err := h.ticketService.DeleteAttachment(c.Request.Context(), attachmentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "attachment deleted successfully"})
}
