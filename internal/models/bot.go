package models

import (
	"time"
)

// Bot is defined in model_resolver.go

// BotConfig is defined in model_resolver.go

// BotSettings is defined in model_resolver.go

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