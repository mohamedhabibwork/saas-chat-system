package models

import (
	"time"
)

// Plan represents a subscription plan
type Plan struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Interval    string    `json:"interval"` // "monthly", "yearly"
	Features    []string  `json:"features"`
	Limits      PlanLimits `json:"limits"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PlanLimits represents the limits for a subscription plan
type PlanLimits struct {
	MaxBots           int     `json:"max_bots"`
	MaxUsers          int     `json:"max_users"`
	MaxMessagesPerDay int     `json:"max_messages_per_day"`
	MaxTokensPerMonth int     `json:"max_tokens_per_month"`
	MaxStorageGB      int     `json:"max_storage_gb"`
	MaxFileSizeMB     int64   `json:"max_file_size_mb"`
	MaxFilesPerDay    int     `json:"max_files_per_day"`
	MaxRequests       int     `json:"max_requests"`
	CustomFeatures    []string `json:"custom_features"`
}

// Subscription represents a user's subscription to a plan
type Subscription struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	PlanID        int       `json:"plan_id"`
	Status        string    `json:"status"` // "active", "cancelled", "expired"
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	AutoRenew     bool      `json:"auto_renew"`
	PaymentMethod string    `json:"payment_method"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Usage represents the current usage statistics for a subscription
type Usage struct {
	ID                int       `json:"id"`
	SubscriptionID    int       `json:"subscription_id"`
	BotsCreated       int       `json:"bots_created"`
	MessagesSent      int       `json:"messages_sent"`
	TokensUsed        int       `json:"tokens_used"`
	StorageUsed       int64     `json:"storage_used"`
	FilesUploaded     int       `json:"files_uploaded"`
	RequestsMade      int       `json:"requests_made"`
	PeriodStart       time.Time `json:"period_start"`
	PeriodEnd         time.Time `json:"period_end"`
	LastUpdated       time.Time `json:"last_updated"`
}

// Payment represents a payment for a subscription
type Payment struct {
	ID             int       `json:"id"`
	SubscriptionID int       `json:"subscription_id"`
	Amount         float64   `json:"amount"`
	Currency       string    `json:"currency"`
	Status         string    `json:"status"` // "success", "failed", "pending"
	PaymentMethod  string    `json:"payment_method"`
	TransactionID  string    `json:"transaction_id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
} 