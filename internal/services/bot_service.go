package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"saas-chat-system/internal/models"
	"saas-chat-system/internal/websocket"
)

// Bot represents a chat bot
type Bot struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	TenantID    string         `json:"tenant_id"`
	Token       string         `json:"token"`
	Config      json.RawMessage `json:"config"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// BotService handles bot-related operations
type BotService struct {
	hub *websocket.Hub
	db  *sql.DB
}

// NewBotService creates a new bot service
func NewBotService(hub *websocket.Hub, db *sql.DB) *BotService {
	return &BotService{
		hub: hub,
		db:  db,
	}
}

// Create creates a new bot
func (s *BotService) Create(ctx context.Context, bot *Bot) error {
	query := `
		INSERT INTO bots (name, description, tenant_id, token, config)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := s.db.QueryRowContext(ctx, query,
		bot.Name,
		bot.Description,
		bot.TenantID,
		bot.Token,
		bot.Config,
	).Scan(&bot.ID, &bot.CreatedAt, &bot.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create bot: %v", err)
	}

	return nil
}

// Get retrieves a bot by ID
func (s *BotService) Get(ctx context.Context, id string) (*Bot, error) {
	query := `
		SELECT id, name, description, tenant_id, token, config, created_at, updated_at
		FROM bots
		WHERE id = $1`

	bot := &Bot{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&bot.ID,
		&bot.Name,
		&bot.Description,
		&bot.TenantID,
		&bot.Token,
		&bot.Config,
		&bot.CreatedAt,
		&bot.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bot not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bot: %v", err)
	}

	return bot, nil
}

// Update updates an existing bot
func (s *BotService) Update(ctx context.Context, bot *Bot) error {
	query := `
		UPDATE bots
		SET name = $1, description = $2, token = $3, config = $4, updated_at = NOW()
		WHERE id = $5 AND tenant_id = $6
		RETURNING updated_at`

	err := s.db.QueryRowContext(ctx, query,
		bot.Name,
		bot.Description,
		bot.Token,
		bot.Config,
		bot.ID,
		bot.TenantID,
	).Scan(&bot.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("bot not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update bot: %v", err)
	}

	return nil
}

// Delete deletes a bot
func (s *BotService) Delete(ctx context.Context, id string, tenantID string) error {
	query := `DELETE FROM bots WHERE id = $1 AND tenant_id = $2`

	result, err := s.db.ExecContext(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete bot: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bot not found")
	}

	return nil
}

// List retrieves a list of bots for a tenant
func (s *BotService) List(ctx context.Context, tenantID string, page, limit int) ([]*Bot, error) {
	query := `
		SELECT id, name, description, tenant_id, token, config, created_at, updated_at
		FROM bots
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	offset := (page - 1) * limit
	rows, err := s.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list bots: %v", err)
	}
	defer rows.Close()

	var bots []*Bot
	for rows.Next() {
		bot := &Bot{}
		err := rows.Scan(
			&bot.ID,
			&bot.Name,
			&bot.Description,
			&bot.TenantID,
			&bot.Token,
			&bot.Config,
			&bot.CreatedAt,
			&bot.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bot: %v", err)
		}
		bots = append(bots, bot)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bot rows: %v", err)
	}

	return bots, nil
}

// ProcessMessage processes a message and generates a bot response
func (s *BotService) ProcessMessage(ctx context.Context, botID int, userID int, message string) error {
	// Get bot configuration
	bot, err := s.getBot(botID)
	if err != nil {
		return fmt.Errorf("failed to get bot: %v", err)
	}

	// Check if user is allowed to interact with the bot
	if !s.isUserAllowed(bot, userID) {
		return fmt.Errorf("user not allowed to interact with this bot")
	}

	// Check subscription limits
	if err := s.checkSubscriptionLimits(userID); err != nil {
		return err
	}

	// Generate response based on bot configuration
	response, err := s.generateResponse(ctx, bot, message)
	if err != nil {
		return fmt.Errorf("failed to generate response: %v", err)
	}

	// Save interaction
	if err := s.saveInteraction(botID, userID, message, response); err != nil {
		return fmt.Errorf("failed to save interaction: %v", err)
	}

	// Send response based on configured response type
	if err := s.sendResponse(bot, userID, response); err != nil {
		return fmt.Errorf("failed to send response: %v", err)
	}

	return nil
}

// getBot retrieves a bot by ID
func (s *BotService) getBot(botID int) (*models.Bot, error) {
	var bot models.Bot
	query := `
		SELECT id, name, tenant_id, token, config, created_at, updated_at
		FROM bots
		WHERE id = $1
	`
	err := s.db.QueryRow(query, botID).Scan(
		&bot.ID, &bot.Name, &bot.TenantID, &bot.Token,
		&bot.Config, &bot.CreatedAt, &bot.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &bot, nil
}

// isUserAllowed checks if a user is allowed to interact with a bot
func (s *BotService) isUserAllowed(bot *models.Bot, userID int) bool {
	for _, allowedID := range bot.Config.Settings.AllowedUsers {
		if allowedID == userID {
			return true
		}
	}
	return false
}

// checkSubscriptionLimits verifies if the user has exceeded their subscription limits
func (s *BotService) checkSubscriptionLimits(userID int) error {
	// Get user's subscription and usage
	subscription, err := s.getUserSubscription(userID)
	if err != nil {
		return fmt.Errorf("failed to get subscription: %v", err)
	}

	// Convert string ID to int
	var subscriptionID int
	_, err = fmt.Sscanf(subscription.ID, "%d", &subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to parse subscription ID: %v", err)
	}

	usage, err := s.getUserUsage(subscriptionID)
	if err != nil {
		return fmt.Errorf("failed to get usage: %v", err)
	}

	// Check limits using Plan2 field instead
	if subscription.Plan2 != nil && subscription.Plan2.Limits.MessagesPerDay > 0 {
		if usage.MessagesSent >= subscription.Plan2.Limits.MessagesPerDay {
			return fmt.Errorf("daily message limit exceeded")
		}
	}

	if subscription.Plan2 != nil && subscription.Plan2.Limits.TokensPerMonth > 0 {
		if usage.TokensUsed >= subscription.Plan2.Limits.TokensPerMonth {
			return fmt.Errorf("monthly token limit exceeded")
		}
	}

	return nil
}

// generateResponse generates a response using the configured AI model
func (s *BotService) generateResponse(ctx context.Context, bot *models.Bot, message string) (string, error) {
	// Create AI model client based on configuration
	client, err := s.createAIClient(bot.Config.ModelType, bot.Config.ModelConfig)
	if err != nil {
		return "", err
	}

	// Generate response
	response, err := client.Generate(ctx, message, bot.Config.Settings)
	if err != nil {
		return "", err
	}

	return response, nil
}

// sendResponse sends the response to the user based on the configured response type
func (s *BotService) sendResponse(bot *models.Bot, userID int, response string) error {
	switch bot.Config.ResponseType {
	case "sse":
		return s.sendSSEResponse(userID, response)
	case "websocket":
		return s.sendWebSocketResponse(userID, response)
	default:
		return fmt.Errorf("unsupported response type: %s", bot.Config.ResponseType)
	}
}

// sendSSEResponse sends a response using Server-Sent Events
func (s *BotService) sendSSEResponse(userID int, response string) error {
	// Implementation for SSE response
	// This would typically involve writing to an SSE connection
	return nil
}

// sendWebSocketResponse sends a response using WebSocket
func (s *BotService) sendWebSocketResponse(userID int, response string) error {
	// Create a websocket message structure
	type WSMessage struct {
		Type    string          `json:"type"`
		UserID  int             `json:"user_id"`
		Content json.RawMessage `json:"content"`
	}
	
	message := WSMessage{
		Type:    "bot_response",
		UserID:  userID,
		Content: json.RawMessage(`{"text": "` + response + `"}`),
	}
	messageJSON, _ := json.Marshal(message)
	s.hub.SendToUser(userID, messageJSON)
	return nil
}

// saveInteraction saves a bot interaction to the database
func (s *BotService) saveInteraction(botID, userID int, message, response string) error {
	query := `
		INSERT INTO bot_interactions (bot_id, user_id, message, response, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`
	_, err := s.db.Exec(query, botID, userID, message, response)
	return err
}

// getUserSubscription retrieves a user's subscription
func (s *BotService) getUserSubscription(userID int) (*models.Subscription, error) {
	var subscription models.Subscription
	subscription.Plan2 = &models.Plan{}
	
	query := `
		SELECT s.id, s.user_id, s.plan_id, s.status, s.start_date, s.end_date,
			   s.auto_renew, s.payment_method, s.created_at, s.updated_at,
			   p.id, p.name, p.description, p.price, p.interval, p.features,
			   p.limits, p.created_at, p.updated_at
		FROM subscriptions s
		JOIN plans p ON s.plan_id = p.id
		WHERE s.user_id = $1 AND s.status = 'active'
		ORDER BY s.created_at DESC
		LIMIT 1
	`
	err := s.db.QueryRow(query, userID).Scan(
		&subscription.ID, &subscription.UserID, &subscription.PlanID,
		&subscription.Status, &subscription.StartDate, &subscription.EndDate,
		&subscription.AutoRenew, &subscription.PaymentMethod,
		&subscription.CreatedAt, &subscription.UpdatedAt,
		&subscription.Plan2.ID, &subscription.Plan2.Name,
		&subscription.Plan2.Description, &subscription.Plan2.Price,
		&subscription.Plan2.Interval, &subscription.Plan2.Features,
		&subscription.Plan2.Limits, &subscription.Plan2.CreatedAt,
		&subscription.Plan2.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

// getUserUsage retrieves a user's usage statistics
func (s *BotService) getUserUsage(subscriptionID int) (*models.Usage, error) {
	var usage models.Usage
	query := `
		SELECT id, subscription_id, messages_sent, tokens_used, storage_used,
			   period_start, period_end, last_updated
		FROM usage
		WHERE subscription_id = $1
		ORDER BY period_start DESC
		LIMIT 1
	`
	err := s.db.QueryRow(query, subscriptionID).Scan(
		&usage.ID, &usage.SubscriptionID, &usage.MessagesSent,
		&usage.TokensUsed, &usage.StorageUsed, &usage.PeriodStart,
		&usage.PeriodEnd, &usage.LastUpdated,
	)
	if err != nil {
		return nil, err
	}
	return &usage, nil
}

// createAIClient creates an AI model client based on the configuration
func (s *BotService) createAIClient(modelType string, config json.RawMessage) (AIClient, error) {
	switch modelType {
	case "gpt-4":
		return NewGPT4Client(config)
	case "claude":
		return NewClaudeClient(config)
	default:
		return nil, fmt.Errorf("unsupported AI model type: %s", modelType)
	}
}

// GetBot retrieves a bot by ID
func (s *BotService) GetBot(botID int) (*models.Bot, error) {
	return s.getBot(botID)
}

// IsUserAllowed checks if a user is allowed to interact with a bot
func (s *BotService) IsUserAllowed(bot *models.Bot, userID int) bool {
	return s.isUserAllowed(bot, userID)
}

// ListBots retrieves all bots for a user
func (s *BotService) ListBots(userID int) ([]*models.Bot, error) {
	query := `
		SELECT id, name, tenant_id, token, config, created_at, updated_at
		FROM bots
		WHERE tenant_id = $1
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bots []*models.Bot
	for rows.Next() {
		var bot models.Bot
		err := rows.Scan(
			&bot.ID, &bot.Name, &bot.TenantID, &bot.Token,
			&bot.Config, &bot.CreatedAt, &bot.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bots = append(bots, &bot)
	}
	return bots, nil
}

// CreateBot creates a new bot
func (s *BotService) CreateBot(bot *models.Bot) error {
	query := `
		INSERT INTO bots (name, tenant_id, token, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`
	err := s.db.QueryRow(
		query,
		bot.Name, bot.TenantID, bot.Token, bot.Config,
	).Scan(&bot.ID)
	return err
}

// UpdateBot updates an existing bot
func (s *BotService) UpdateBot(bot *models.Bot) error {
	query := `
		UPDATE bots
		SET name = $1, token = $2, config = $3, updated_at = NOW()
		WHERE id = $4 AND tenant_id = $5
	`
	result, err := s.db.Exec(
		query,
		bot.Name, bot.Token, bot.Config, bot.ID, bot.TenantID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("bot not found or not owned by tenant")
	}
	return nil
}

// DeleteBot deletes a bot
func (s *BotService) DeleteBot(botID int, tenantID int) error {
	query := "DELETE FROM bots WHERE id = $1 AND tenant_id = $2"
	result, err := s.db.Exec(query, botID, tenantID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("bot not found or not owned by tenant")
	}
	return nil
}

// AddToChannel adds a bot to a channel
func (s *BotService) AddToChannel(ctx context.Context, botID, channelID string) error {
	query := `
		INSERT INTO channel_bots (channel_id, bot_id)
		VALUES ($1, $2)`

	_, err := s.db.ExecContext(ctx, query, channelID, botID)
	if err != nil {
		return fmt.Errorf("failed to add bot to channel: %v", err)
	}

	return nil
}

// RemoveFromChannel removes a bot from a channel
func (s *BotService) RemoveFromChannel(ctx context.Context, botID, channelID string) error {
	query := `
		DELETE FROM channel_bots
		WHERE channel_id = $1 AND bot_id = $2`

	result, err := s.db.ExecContext(ctx, query, channelID, botID)
	if err != nil {
		return fmt.Errorf("failed to remove bot from channel: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bot not found in channel")
	}

	return nil
}
