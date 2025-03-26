package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedhabibwork/saas-chat-system/internal/models"
)

// NotificationService handles sending notifications via email and FCM
type NotificationService struct {
	emailService *EmailService
	fcmClient   *FCMClient
}

// NewNotificationService creates a new notification service
func NewNotificationService(emailService *EmailService, fcmClient *FCMClient) *NotificationService {
	return &NotificationService{
		emailService: emailService,
		fcmClient:   fcmClient,
	}
}

// SendTicketNotification sends notifications for ticket-related events
func (s *NotificationService) SendTicketNotification(ctx context.Context, notification *models.TicketNotification) error {
	// Send email notification
	if err := s.emailService.SendTicketEmail(ctx, notification); err != nil {
		return fmt.Errorf("error sending email notification: %v", err)
	}

	// Send FCM notification
	if err := s.fcmClient.SendTicketNotification(ctx, notification); err != nil {
		return fmt.Errorf("error sending FCM notification: %v", err)
	}

	return nil
}

// SendForumNotification sends notifications for forum-related events
func (s *NotificationService) SendForumNotification(ctx context.Context, notification *models.ForumNotification) error {
	// Send email notification
	if err := s.emailService.SendForumEmail(ctx, notification); err != nil {
		return fmt.Errorf("error sending email notification: %v", err)
	}

	// Send FCM notification
	if err := s.fcmClient.SendForumNotification(ctx, notification); err != nil {
		return fmt.Errorf("error sending FCM notification: %v", err)
	}

	return nil
}

// FCMClient handles Firebase Cloud Messaging notifications
type FCMClient struct {
	// Add FCM client configuration here
}

// SendTicketNotification sends a ticket notification via FCM
func (c *FCMClient) SendTicketNotification(ctx context.Context, notification *models.TicketNotification) error {
	// Implement FCM notification sending
	return nil
}

// SendForumNotification sends a forum notification via FCM
func (c *FCMClient) SendForumNotification(ctx context.Context, notification *models.ForumNotification) error {
	// Implement FCM notification sending
	return nil
} 