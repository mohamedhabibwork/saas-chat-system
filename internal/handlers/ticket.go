package handlers

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/mohamedhabibwork/saas-chat-system/internal/models"
	"github.com/mohamedhabibwork/saas-chat-system/internal/services"
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

// CreateTicket handles requests to create a new ticket
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

// GetTicket handles requests to retrieve a ticket
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

// UpdateTicket handles requests to update a ticket
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

// UpdateTicketStatus handles requests to update a ticket's status
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

// ListTickets handles requests to list tickets
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

// AddComment handles requests to add a comment to a ticket
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

// GetComments handles requests to retrieve comments for a ticket
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

// UploadAttachment handles requests to upload an attachment
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

// GetAttachments handles requests to retrieve attachments for a ticket
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

// DeleteAttachment handles requests to delete an attachment
func (h *TicketHandler) DeleteAttachment(c *gin.Context) {
	attachmentID := c.Param("attachment_id")
	// TODO: Implement attachment ownership verification
	if err := h.ticketService.DeleteAttachment(c.Request.Context(), attachmentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "attachment deleted successfully"})
} 