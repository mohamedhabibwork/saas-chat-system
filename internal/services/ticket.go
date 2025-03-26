package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedhabibwork/saas-chat-system/internal/database"
	"github.com/mohamedhabibwork/saas-chat-system/internal/models"
)

// TicketService handles ticket-related operations
type TicketService struct {
	db               *database.DB
	notificationService *NotificationService
}

// NewTicketService creates a new ticket service
func NewTicketService(db *database.DB, notificationService *NotificationService) *TicketService {
	return &TicketService{
		db: db,
		notificationService: notificationService,
	}
}

// CreateTicket creates a new support ticket
func (s *TicketService) CreateTicket(ctx context.Context, ticket *models.Ticket) error {
	ticket.ID = uuid.New().String()
	ticket.Status = models.TicketStatusOpen
	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Create(ticket).Error; err != nil {
		return fmt.Errorf("error creating ticket: %v", err)
	}

	return nil
}

// GetTicket retrieves a ticket by ID
func (s *TicketService) GetTicket(ctx context.Context, ticketID string) (*models.Ticket, error) {
	var ticket models.Ticket
	if err := s.db.WithContext(ctx).First(&ticket, "id = ?", ticketID).Error; err != nil {
		return nil, fmt.Errorf("error retrieving ticket: %v", err)
	}
	return &ticket, nil
}

// UpdateTicket updates an existing ticket
func (s *TicketService) UpdateTicket(ctx context.Context, ticket *models.Ticket) error {
	if !ticket.CanBeUpdated() {
		return fmt.Errorf("cannot update closed ticket")
	}

	ticket.UpdatedAt = time.Now()
	if err := s.db.WithContext(ctx).Save(ticket).Error; err != nil {
		return fmt.Errorf("error updating ticket: %v", err)
	}

	return nil
}

// UpdateTicketStatus updates the status of a ticket
func (s *TicketService) UpdateTicketStatus(ctx context.Context, ticketID string, status models.TicketStatus) error {
	ticket, err := s.GetTicket(ctx, ticketID)
	if err != nil {
		return err
	}

	if !ticket.CanBeUpdated() {
		return fmt.Errorf("cannot update closed ticket")
	}

	ticket.UpdateStatus(status)
	if err := s.db.WithContext(ctx).Save(ticket).Error; err != nil {
		return fmt.Errorf("error updating ticket status: %v", err)
	}

	return nil
}

// ListTickets retrieves tickets with optional filtering
func (s *TicketService) ListTickets(ctx context.Context, tenantID string, filters map[string]interface{}) ([]*models.Ticket, error) {
	var tickets []*models.Ticket
	query := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	// Apply filters
	for key, value := range filters {
		query = query.Where(key+" = ?", value)
	}

	if err := query.Order("created_at DESC").Find(&tickets).Error; err != nil {
		return nil, fmt.Errorf("error retrieving tickets: %v", err)
	}

	return tickets, nil
}

// AddComment adds a comment to a ticket
func (s *TicketService) AddComment(ctx context.Context, comment *models.TicketComment) error {
	comment.ID = uuid.New().String()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Create(comment).Error; err != nil {
		return fmt.Errorf("error adding comment: %v", err)
	}

	return nil
}

// GetComments retrieves all comments for a ticket
func (s *TicketService) GetComments(ctx context.Context, ticketID string) ([]*models.TicketComment, error) {
	var comments []*models.TicketComment
	if err := s.db.WithContext(ctx).Where("ticket_id = ?", ticketID).Order("created_at ASC").Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("error retrieving comments: %v", err)
	}
	return comments, nil
}

// AddAttachment adds an attachment to a ticket or comment
func (s *TicketService) AddAttachment(ctx context.Context, attachment *models.TicketAttachment) error {
	attachment.ID = uuid.New().String()
	attachment.CreatedAt = time.Now()
	attachment.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Create(attachment).Error; err != nil {
		return fmt.Errorf("error adding attachment: %v", err)
	}

	return nil
}

// GetAttachments retrieves all attachments for a ticket
func (s *TicketService) GetAttachments(ctx context.Context, ticketID string) ([]*models.TicketAttachment, error) {
	var attachments []*models.TicketAttachment
	if err := s.db.WithContext(ctx).Where("ticket_id = ?", ticketID).Find(&attachments).Error; err != nil {
		return nil, fmt.Errorf("error retrieving attachments: %v", err)
	}
	return attachments, nil
}

// GetCommentAttachments retrieves all attachments for a comment
func (s *TicketService) GetCommentAttachments(ctx context.Context, commentID string) ([]*models.TicketAttachment, error) {
	var attachments []*models.TicketAttachment
	if err := s.db.WithContext(ctx).Where("comment_id = ?", commentID).Find(&attachments).Error; err != nil {
		return nil, fmt.Errorf("error retrieving comment attachments: %v", err)
	}
	return attachments, nil
}

// DeleteAttachment deletes an attachment
func (s *TicketService) DeleteAttachment(ctx context.Context, attachmentID string) error {
	if err := s.db.WithContext(ctx).Delete(&models.TicketAttachment{}, "id = ?", attachmentID).Error; err != nil {
		return fmt.Errorf("error deleting attachment: %v", err)
	}
	return nil
} 