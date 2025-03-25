package models

import "time"

// APIError represents an API error response
type APIError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Client represents a connected user
type Client struct {
	ID       int
	Username string
	TenantID int
	Conn     interface{} // Will be set to *websocket.Conn in websocket package
	Send     chan []byte
	Groups   map[int]bool    // Group memberships
	Topics   map[string]bool // Topic subscriptions
}

// Message represents a chat message
type Message struct {
	Type       string    `json:"type"` // "private", "group", "notification"
	Content    string    `json:"content"`
	Sender     string    `json:"sender"`
	SenderID   int       `json:"sender_id"`
	TenantID   int       `json:"tenant_id"`
	Receiver   string    `json:"receiver,omitempty"`
	ReceiverID int       `json:"receiver_id,omitempty"`
	GroupID    int       `json:"group_id,omitempty"`
	GroupName  string    `json:"group_name,omitempty"`
	TopicName  string    `json:"topic_name,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	Clients         map[*Client]bool
	TenantClients   map[int]map[*Client]bool // Clients organized by tenant
	Broadcast       chan []byte
	Register        chan *Client
	Unregister      chan *Client
	PrivateMessages chan Message
	GroupMessages   chan Message
	TopicMessages   chan Message
}

// Tenant represents a tenant in the system
type Tenant struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	TenantID  int       `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Group represents a chat group
type Group struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	TenantID  int       `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Topic represents a notification topic
type Topic struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	TenantID  int       `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
}

// GroupMember represents a group membership
type GroupMember struct {
	GroupID   int       `json:"group_id"`
	UserID    int       `json:"user_id"`
	JoinedAt  time.Time `json:"joined_at"`
}

// TopicSubscription represents a topic subscription
type TopicSubscription struct {
	TopicID       int       `json:"topic_id"`
	UserID        int       `json:"user_id"`
	SubscribedAt  time.Time `json:"subscribed_at"`
}

// Bot represents a chat bot
type Bot struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	TenantID  int       `json:"tenant_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Channel represents a chat channel
type Channel struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TenantID    int       `json:"tenant_id"`
	CreatedBy   int       `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ChannelMember represents a channel membership
type ChannelMember struct {
	ChannelID int       `json:"channel_id"`
	UserID    int       `json:"user_id"`
	Role      string    `json:"role"` // "admin", "moderator", "member"
	JoinedAt  time.Time `json:"joined_at"`
}

// Tab represents a user's custom tab
type Tab struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "private", "group", "channel", "bot"
	TargetID  int       `json:"target_id"` // ID of the target (user, group, channel, or bot)
	Order     int       `json:"order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BotMessage represents a message from a bot
type BotMessage struct {
	ID        int       `json:"id"`
	BotID     int       `json:"bot_id"`
	Content   string    `json:"content"`
	TenantID  int       `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
}

// ChannelMessage represents a message in a channel
type ChannelMessage struct {
	ID            int       `json:"id"`
	ChannelID     int       `json:"channel_id"`
	SenderID      int       `json:"sender_id"`
	Content       string    `json:"content"`
	SenderTimezone string    `json:"sender_timezone,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
} 