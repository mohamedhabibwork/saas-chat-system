package models

import (
	"time"
)

// TicketStatus represents the current status of a ticket
type TicketStatus string

const (
	TicketStatusOpen     TicketStatus = "open"
	TicketStatusPending  TicketStatus = "pending"
	TicketStatusResolved TicketStatus = "resolved"
	TicketStatusClosed   TicketStatus = "closed"
)

// TicketPriority represents the priority level of a ticket
type TicketPriority string

const (
	TicketPriorityLow    TicketPriority = "low"
	TicketPriorityMedium TicketPriority = "medium"
	TicketPriorityHigh   TicketPriority = "high"
	TicketPriorityUrgent TicketPriority = "urgent"
)

// TicketCategory represents the category of the issue
type TicketCategory string

const (
	TicketCategoryBug        TicketCategory = "bug"
	TicketCategoryFeature    TicketCategory = "feature"
	TicketCategorySupport    TicketCategory = "support"
	TicketCategorySecurity   TicketCategory = "security"
	TicketCategoryOther      TicketCategory = "other"
)

// Ticket represents a support ticket
type Ticket struct {
	ID          string         `json:"id" gorm:"primaryKey"`
	TenantID    string         `json:"tenant_id" gorm:"not null"`
	UserID      string         `json:"user_id" gorm:"not null"`
	Title       string         `json:"title" gorm:"not null"`
	Description string         `json:"description" gorm:"not null"`
	Category    TicketCategory `json:"category" gorm:"not null"`
	Priority    TicketPriority `json:"priority" gorm:"not null"`
	Status      TicketStatus   `json:"status" gorm:"not null"`
	AssignedTo  string         `json:"assigned_to"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	ResolvedAt  *time.Time     `json:"resolved_at"`
	ClosedAt    *time.Time     `json:"closed_at"`
	NotifyOnUpdate    bool `json:"notify_on_update" gorm:"default:true"`
	NotifyOnComment   bool `json:"notify_on_comment" gorm:"default:true"`
	NotifyOnStatus    bool `json:"notify_on_status" gorm:"default:true"`
	NotifyOnAssign    bool `json:"notify_on_assign" gorm:"default:true"`
}

// TicketComment represents a comment on a ticket
type TicketComment struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	TicketID  string    `json:"ticket_id" gorm:"not null"`
	UserID    string    `json:"user_id" gorm:"not null"`
	Content   string    `json:"content" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TicketAttachment represents an attachment to a ticket or comment
type TicketAttachment struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TicketID    string    `json:"ticket_id" gorm:"not null"`
	CommentID   *string   `json:"comment_id"`
	FileName    string    `json:"file_name" gorm:"not null"`
	FileType    string    `json:"file_type" gorm:"not null"`
	FileSize    int64     `json:"file_size" gorm:"not null"`
	FileURL     string    `json:"file_url" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TicketNotification represents a notification for ticket-related events
type TicketNotification struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TicketID    string    `json:"ticket_id" gorm:"not null"`
	UserID      string    `json:"user_id" gorm:"not null"`
	Type        string    `json:"type" gorm:"not null"` // created, updated, commented, status_changed
	Title       string    `json:"title" gorm:"not null"`
	Message     string    `json:"message" gorm:"not null"`
	Read        bool      `json:"read" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at"`
}

// UpdateStatus updates the ticket status and sets appropriate timestamps
func (t *Ticket) UpdateStatus(status TicketStatus) {
	t.Status = status
	t.UpdatedAt = time.Now()

	switch status {
	case TicketStatusResolved:
		now := time.Now()
		t.ResolvedAt = &now
	case TicketStatusClosed:
		now := time.Now()
		t.ClosedAt = &now
	}
}

// IsResolved checks if the ticket is resolved or closed
func (t *Ticket) IsResolved() bool {
	return t.Status == TicketStatusResolved || t.Status == TicketStatusClosed
}

// CanBeUpdated checks if the ticket can be updated
func (t *Ticket) CanBeUpdated() bool {
	return t.Status != TicketStatusClosed
} 