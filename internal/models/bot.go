package models

import (
	"encoding/json"
	"time"
)

// Bot represents a chat bot in the system
type Bot struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	TenantID    int             `json:"tenant_id"`
	Token       string          `json:"token"`
	Config      BotConfig       `json:"config"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// BotConfig represents the configuration for a bot
type BotConfig struct {
	ModelType    string          `json:"model_type"`    // e.g., "gpt-4", "claude", "custom"
	ModelConfig  json.RawMessage `json:"model_config"`  // Specific configuration for the AI model
	ResponseType string          `json:"response_type"` // "sse" or "websocket"
	Settings     BotSettings     `json:"settings"`
}

// BotSettings represents the bot's behavior settings
type BotSettings struct {
	MaxTokens        int      `json:"max_tokens"`
	Temperature      float64  `json:"temperature"`
	StopSequences    []string `json:"stop_sequences"`
	AllowedUsers     []int    `json:"allowed_users"`     // User IDs that can interact with this bot
	RestrictedTopics []string `json:"restricted_topics"` // Topics the bot should avoid
	CustomPrompts    []string `json:"custom_prompts"`    // Custom system prompts
}

// BotResponse represents a response from the bot
type BotResponse struct {
	ID        int       `json:"id"`
	BotID     int       `json:"bot_id"`
	UserID    int       `json:"user_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// BotInteraction represents an interaction between a user and a bot
type BotInteraction struct {
	ID        int       `json:"id"`
	BotID     int       `json:"bot_id"`
	UserID    int       `json:"user_id"`
	Message   string    `json:"message"`
	Response  string    `json:"response"`
	CreatedAt time.Time `json:"created_at"`
}

// BotStats represents statistics for a bot
type BotStats struct {
	TotalInteractions int     `json:"total_interactions"`
	AverageResponseTime float64 `json:"average_response_time"`
	SuccessRate       float64 `json:"success_rate"`
	LastActive        time.Time `json:"last_active"`
} 