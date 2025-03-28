package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"saas-chat-system/internal/models"
)

// TicketService handles ticket-related operations
type TicketService struct {
	db                  *sql.DB
	notificationService *NotificationService
}

// NewTicketService creates a new ticket service
func NewTicketService(db *sql.DB, notificationService *NotificationService) *TicketService {
	return &TicketService{
		db:                  db,
		notificationService: notificationService,
	}
}

// CreateTicket creates a new support ticket
func (s *TicketService) CreateTicket(ctx context.Context, ticket *models.Ticket) error {
	ticket.ID = uuid.New().String()
	ticket.Status = models.TicketStatusOpen
	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()

	query := `
		INSERT INTO tickets (
			id, tenant_id, user_id, title, description,
			priority, status, category, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := s.db.ExecContext(ctx, query,
		ticket.ID, ticket.TenantID, ticket.UserID,
		ticket.Title, ticket.Description, ticket.Priority,
		ticket.Status, ticket.Category,
		ticket.CreatedAt, ticket.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creating ticket: %v", err)
	}

	return nil
}

// GetTicket retrieves a ticket by ID
func (s *TicketService) GetTicket(ctx context.Context, ticketID string) (*models.Ticket, error) {
	var ticket models.Ticket
	query := `
		SELECT id, tenant_id, user_id, title, description,
			   priority, status, category, created_at, updated_at
		FROM tickets
		WHERE id = $1
	`
	err := s.db.QueryRowContext(ctx, query, ticketID).Scan(
		&ticket.ID, &ticket.TenantID, &ticket.UserID,
		&ticket.Title, &ticket.Description, &ticket.Priority,
		&ticket.Status, &ticket.Category,
		&ticket.CreatedAt, &ticket.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ticket not found")
		}
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
	query := `
		UPDATE tickets
		SET title = $1, description = $2, priority = $3,
			status = $4, category = $5, updated_at = $6
		WHERE id = $7
	`
	result, err := s.db.ExecContext(ctx, query,
		ticket.Title, ticket.Description, ticket.Priority,
		ticket.Status, ticket.Category, ticket.UpdatedAt,
		ticket.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating ticket: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("ticket not found")
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
	query := `
		UPDATE tickets
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	result, err := s.db.ExecContext(ctx, query,
		ticket.Status, ticket.UpdatedAt, ticket.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating ticket status: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("ticket not found")
	}

	return nil
}

// ListTickets retrieves tickets with optional filtering
func (s *TicketService) ListTickets(ctx context.Context, tenantID string, filters map[string]interface{}) ([]*models.Ticket, error) {
	query := `
		SELECT id, tenant_id, user_id, title, description,
			   priority, status, category, created_at, updated_at
		FROM tickets
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	i := 2

	// Add filters to the query
	for key, value := range filters {
		query += fmt.Sprintf(" AND %s = $%d", key, i)
		args = append(args, value)
		i++
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error retrieving tickets: %v", err)
	}
	defer rows.Close()

	var tickets []*models.Ticket
	for rows.Next() {
		var ticket models.Ticket
		err := rows.Scan(
			&ticket.ID, &ticket.TenantID, &ticket.UserID,
			&ticket.Title, &ticket.Description, &ticket.Priority,
			&ticket.Status, &ticket.Category,
			&ticket.CreatedAt, &ticket.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning ticket row: %v", err)
		}
		tickets = append(tickets, &ticket)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ticket rows: %v", err)
	}

	return tickets, nil
}

// AddComment adds a comment to a ticket
func (s *TicketService) AddComment(ctx context.Context, comment *models.TicketComment) error {
	comment.ID = uuid.New().String()
	comment.CreatedAt = time.Now()
	comment.UpdatedAt = time.Now()

	query := `
		INSERT INTO ticket_comments (
			id, ticket_id, user_id, content,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := s.db.ExecContext(ctx, query,
		comment.ID, comment.TicketID, comment.UserID,
		comment.Content, comment.CreatedAt, comment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error adding comment: %v", err)
	}

	return nil
}

// GetComments retrieves all comments for a ticket
func (s *TicketService) GetComments(ctx context.Context, ticketID string) ([]*models.TicketComment, error) {
	query := `
		SELECT id, ticket_id, user_id, content,
			   created_at, updated_at
		FROM ticket_comments
		WHERE ticket_id = $1
		ORDER BY created_at ASC
	`
	rows, err := s.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving comments: %v", err)
	}
	defer rows.Close()

	var comments []*models.TicketComment
	for rows.Next() {
		var comment models.TicketComment
		err := rows.Scan(
			&comment.ID, &comment.TicketID, &comment.UserID,
			&comment.Content, &comment.CreatedAt, &comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning comment row: %v", err)
		}
		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating comment rows: %v", err)
	}

	return comments, nil
}

// AddAttachment adds an attachment to a ticket or comment
func (s *TicketService) AddAttachment(ctx context.Context, attachment *models.TicketAttachment) error {
	attachment.ID = uuid.New().String()
	attachment.CreatedAt = time.Now()
	attachment.UpdatedAt = time.Now()

	query := `
		INSERT INTO ticket_attachments (
			id, ticket_id, comment_id, file_name,
			file_type, file_size, file_url,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := s.db.ExecContext(ctx, query,
		attachment.ID, attachment.TicketID, attachment.CommentID,
		attachment.FileName, attachment.FileType, attachment.FileSize,
		attachment.FileURL, attachment.CreatedAt, attachment.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error adding attachment: %v", err)
	}

	return nil
}

// GetAttachments retrieves all attachments for a ticket
func (s *TicketService) GetAttachments(ctx context.Context, ticketID string) ([]*models.TicketAttachment, error) {
	query := `
		SELECT id, ticket_id, comment_id, file_name,
			   file_type, file_size, file_url,
			   created_at, updated_at
		FROM ticket_attachments
		WHERE ticket_id = $1
	`
	rows, err := s.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving attachments: %v", err)
	}
	defer rows.Close()

	var attachments []*models.TicketAttachment
	for rows.Next() {
		var attachment models.TicketAttachment
		err := rows.Scan(
			&attachment.ID, &attachment.TicketID, &attachment.CommentID,
			&attachment.FileName, &attachment.FileType, &attachment.FileSize,
			&attachment.FileURL, &attachment.CreatedAt, &attachment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning attachment row: %v", err)
		}
		attachments = append(attachments, &attachment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attachment rows: %v", err)
	}

	return attachments, nil
}

// GetCommentAttachments retrieves all attachments for a comment
func (s *TicketService) GetCommentAttachments(ctx context.Context, commentID string) ([]*models.TicketAttachment, error) {
	query := `
		SELECT id, ticket_id, comment_id, file_name,
			   file_type, file_size, file_url,
			   created_at, updated_at
		FROM ticket_attachments
		WHERE comment_id = $1
	`
	rows, err := s.db.QueryContext(ctx, query, commentID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving comment attachments: %v", err)
	}
	defer rows.Close()

	var attachments []*models.TicketAttachment
	for rows.Next() {
		var attachment models.TicketAttachment
		err := rows.Scan(
			&attachment.ID, &attachment.TicketID, &attachment.CommentID,
			&attachment.FileName, &attachment.FileType, &attachment.FileSize,
			&attachment.FileURL, &attachment.CreatedAt, &attachment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning attachment row: %v", err)
		}
		attachments = append(attachments, &attachment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attachment rows: %v", err)
	}

	return attachments, nil
}

// DeleteAttachment deletes an attachment
func (s *TicketService) DeleteAttachment(ctx context.Context, attachmentID string) error {
	query := `DELETE FROM ticket_attachments WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, attachmentID)
	if err != nil {
		return fmt.Errorf("error deleting attachment: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking delete result: %v", err)
	}
	if rows == 0 {
		return fmt.Errorf("attachment not found")
	}

	return nil
}
