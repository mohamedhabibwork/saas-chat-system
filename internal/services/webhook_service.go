package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WebhookEvent struct {
	Type      string                 `json:"type"`
	Timestamp time.Time             `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

type WebhookSubscription struct {
	ID        int       `json:"id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	Secret    string    `json:"secret"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WebhookService struct {
	db *sql.DB
}

func NewWebhookService(db *sql.DB) *WebhookService {
	return &WebhookService{db: db}
}

// Subscribe registers a new webhook subscription
func (s *WebhookService) Subscribe(subscription *WebhookSubscription) error {
	query := `
		INSERT INTO webhook_subscriptions (
			url, events, secret, active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`
	
	eventsJSON, err := json.Marshal(subscription.Events)
	if err != nil {
		return fmt.Errorf("error marshaling events: %v", err)
	}

	return s.db.QueryRow(
		query,
		subscription.URL,
		eventsJSON,
		subscription.Secret,
		subscription.Active,
	).Scan(&subscription.ID)
}

// Unsubscribe removes a webhook subscription
func (s *WebhookService) Unsubscribe(id int) error {
	_, err := s.db.Exec(`
		DELETE FROM webhook_subscriptions WHERE id = $1
	`, id)
	return err
}

// TriggerEvent sends webhook notifications for an event
func (s *WebhookService) TriggerEvent(eventType string, data map[string]interface{}) error {
	event := WebhookEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	// Get all active subscriptions for this event type
	rows, err := s.db.Query(`
		SELECT id, url, events, secret
		FROM webhook_subscriptions
		WHERE active = true
		AND events @> $1::jsonb
	`, fmt.Sprintf(`["%s"]`, eventType))
	if err != nil {
		return fmt.Errorf("error querying subscriptions: %v", err)
	}
	defer rows.Close()

	// Send webhook to each subscriber
	for rows.Next() {
		var sub WebhookSubscription
		var eventsJSON []byte
		err := rows.Scan(&sub.ID, &sub.URL, &eventsJSON, &sub.Secret)
		if err != nil {
			return fmt.Errorf("error scanning subscription: %v", err)
		}

		if err := json.Unmarshal(eventsJSON, &sub.Events); err != nil {
			return fmt.Errorf("error unmarshaling events: %v", err)
		}

		if err := s.sendWebhook(sub, event); err != nil {
			// Log error but continue with other subscribers
			fmt.Printf("Error sending webhook to %s: %v\n", sub.URL, err)
		}
	}

	return nil
}

// sendWebhook sends a webhook to a specific subscription
func (s *WebhookService) sendWebhook(sub WebhookSubscription, event WebhookEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshaling event: %v", err)
	}

	req, err := http.NewRequest("POST", sub.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Webhook-Event", event.Type)
	req.Header.Set("X-Webhook-Timestamp", event.Timestamp.Format(time.RFC3339))
	
	// Add signature for security
	signature := generateSignature(payload, sub.Secret)
	req.Header.Set("X-Webhook-Signature", signature)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending webhook: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

// generateSignature creates a HMAC signature for webhook security
func generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// Common webhook event types
const (
	EventUserRegistered     = "user.registered"
	EventUserLoggedIn       = "user.logged_in"
	EventUserLoggedOut      = "user.logged_out"
	EventPasswordReset      = "user.password_reset"
	EventSubscriptionStart  = "subscription.started"
	EventSubscriptionEnd    = "subscription.ended"
	EventSubscriptionUpdate = "subscription.updated"
	EventStorageLimit      = "storage.limit_reached"
	EventStorageWarning    = "storage.warning"
	EventFileUploaded      = "file.uploaded"
	EventFileDeleted       = "file.deleted"
	EventChannelCreated    = "channel.created"
	EventChannelDeleted    = "channel.deleted"
	EventMessageSent       = "message.sent"
	EventBotCreated        = "bot.created"
	EventBotDeleted        = "bot.deleted"
) 