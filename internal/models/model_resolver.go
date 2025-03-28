package models

import (
	"encoding/json"
	"time"
)

// SubscriptionPlanType represents available subscription plans
type SubscriptionPlanType string

const (
	PlanFree       SubscriptionPlanType = "free"
	PlanBasic      SubscriptionPlanType = "basic"
	PlanPro        SubscriptionPlanType = "pro"
	PlanEnterprise SubscriptionPlanType = "enterprise"
)

// User represents a user in the system (extended version)
type User struct {
	ID            int               `json:"id"`
	Username      string            `json:"username"`
	Email         string            `json:"email"`
	PasswordHash  string            `json:"-"`
	FirstName     string            `json:"first_name"`
	LastName      string            `json:"last_name"`
	TenantID      int               `json:"tenant_id"`
	RoleID        int               `json:"role_id"`
	IsActive      bool              `json:"is_active"`
	Timezone      string            `json:"timezone"`
	LastLogin     time.Time         `json:"last_login"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// Bot represents a chat bot in the system (extended version)
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

// ChannelType represents the type of channel
type ChannelType string

const (
	ChannelTypePublic  ChannelType = "public"
	ChannelTypePrivate ChannelType = "private"
	ChannelTypeDirect  ChannelType = "direct"
)

// ChannelSettings represents channel-specific settings
type ChannelSettings struct {
	AllowBots       bool     `json:"allow_bots"`
	AllowFiles      bool     `json:"allow_files"`
	AllowedFileTypes []string `json:"allowed_file_types"`
	MaxFileSize     int64    `json:"max_file_size"`
	RetentionDays   int      `json:"retention_days"`
	Notifications   bool     `json:"notifications"`
}

// Channel represents a chat channel (extended version)
type Channel struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Type        ChannelType     `json:"type"`
	CreatedBy   int             `json:"created_by"`
	TenantID    int             `json:"tenant_id"`
	Settings    ChannelSettings `json:"settings"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ChannelMember represents a user's membership in a channel (extended version)
type ChannelMember struct {
	ID        int       `json:"id"`
	ChannelID int       `json:"channel_id"`
	UserID    int       `json:"user_id"`
	Role      string    `json:"role"` // "admin", "moderator", "member"
	JoinedAt  time.Time `json:"joined_at"`
}

// ChannelMessage represents a message in a channel (extended version)
type ChannelMessage struct {
	ID             int                    `json:"id"`
	ChannelID      int                    `json:"channel_id"`
	UserID         int                    `json:"user_id"`
	Content        string                 `json:"content"`
	Type           string                 `json:"type"` // "text", "file", "system"
	Metadata       map[string]interface{} `json:"metadata"`
	SenderTimezone string                 `json:"sender_timezone,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// File represents a file in the system (merged definition)
type File struct {
	ID          int                    `json:"id"`
	UserID      int                    `json:"user_id"`
	TenantID    int                    `json:"tenant_id,omitempty"`
	Filename    string                 `json:"filename"`
	Name        string                 `json:"name,omitempty"` // Original file name
	Filepath    string                 `json:"filepath,omitempty"`
	Path        string                 `json:"path,omitempty"` // Storage path
	URL         string                 `json:"url,omitempty"`
	Size        int64                  `json:"size"`
	ContentType string                 `json:"content_type,omitempty"`
	MimeType    string                 `json:"mime_type,omitempty"`  // File MIME type
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at,omitempty"`
}

// Subscription represents a tenant's subscription (extended version)
type Subscription struct {
	ID            string              `json:"id"`
	UserID        int                 `json:"user_id"`
	PlanID        int                 `json:"plan_id"`
	TenantID      string              `json:"tenant_id"`
	Plan          SubscriptionPlanType `json:"plan"`
	Status        string              `json:"status"` // active, suspended, cancelled
	StartDate     time.Time           `json:"start_date"`
	EndDate       time.Time           `json:"end_date"`
	AutoRenew     bool                `json:"auto_renew"`
	PaymentMethod string              `json:"payment_method"`
	BillingEmail  string              `json:"billing_email"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	
	// Plan information for database queries
	Plan2 *Plan `json:"-"`
	
	// Extended subscription plan fields
	MaxStorage        int64    `json:"max_storage,omitempty"`
	MaxFiles          int      `json:"max_files,omitempty"`
	MaxDailyUploads   int      `json:"max_daily_uploads,omitempty"`
	MaxFileSize       int64    `json:"max_file_size,omitempty"`
	AllowedExtensions []string `json:"allowed_extensions,omitempty"`
	AllowedMimeTypes  []string `json:"allowed_mime_types,omitempty"`
}

// HasFeature checks if a subscription plan includes a specific feature
func (s *Subscription) HasFeature(feature string) bool {
	features, exists := SubscriptionFeatures[s.Plan]
	if !exists {
		return false
	}
	for _, f := range features {
		if f == feature {
			return true
		}
	}
	return false
}

// IsActive checks if the subscription is currently active
func (s *Subscription) IsActive() bool {
	return s.Status == "active" && time.Now().Before(s.EndDate)
}

// SubscriptionFeatures represents available features for each plan
var SubscriptionFeatures = map[SubscriptionPlanType][]string{
	PlanFree: {
		"chat_basic",
		"tracking_basic",
		"reports_basic",
	},
	PlanBasic: {
		"chat_basic",
		"tracking_basic",
		"reports_basic",
		"chat_advanced",
		"tracking_advanced",
		"reports_advanced",
		"email_reports",
	},
	PlanPro: {
		"chat_basic",
		"tracking_basic",
		"reports_basic",
		"chat_advanced",
		"tracking_advanced",
		"reports_advanced",
		"email_reports",
		"chat_premium",
		"tracking_premium",
		"reports_premium",
		"custom_reports",
		"api_access",
	},
	PlanEnterprise: {
		"chat_basic",
		"tracking_basic",
		"reports_basic",
		"chat_advanced",
		"tracking_advanced",
		"reports_advanced",
		"email_reports",
		"chat_premium",
		"tracking_premium",
		"reports_premium",
		"custom_reports",
		"api_access",
		"dedicated_support",
		"white_label",
		"custom_integration",
	},
}