package models

import (
	"time"
)

// ChannelType represents the type of channel
type ChannelType string

const (
	ChannelTypePublic  ChannelType = "public"
	ChannelTypePrivate ChannelType = "private"
	ChannelTypeDirect  ChannelType = "direct"
)

// Channel represents a chat channel
type Channel struct {
	ID          int         `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        ChannelType `json:"type"`
	CreatedBy   int         `json:"created_by"`
	TenantID    int         `json:"tenant_id"`
	Settings    ChannelSettings `json:"settings"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

// ChannelSettings represents channel-specific settings
type ChannelSettings struct {
	AllowVideo     bool `json:"allow_video"`
	AllowAudio     bool `json:"allow_audio"`
	AllowScreen    bool `json:"allow_screen"`
	MaxParticipants int `json:"max_participants"`
	Moderated      bool `json:"moderated"`
	Recording      bool `json:"recording"`
}

// ChannelMember represents a user's membership in a channel
type ChannelMember struct {
	ID        int       `json:"id"`
	ChannelID int       `json:"channel_id"`
	UserID    int       `json:"user_id"`
	Role      string    `json:"role"` // "admin", "moderator", "member"
	JoinedAt  time.Time `json:"joined_at"`
}

// ChannelMessage represents a message in a channel
type ChannelMessage struct {
	ID        int       `json:"id"`
	ChannelID int       `json:"channel_id"`
	UserID    int       `json:"user_id"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // "text", "file", "system"
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time `json:"created_at"`
}

// WebRTCConnection represents a WebRTC connection for a user in a channel
type WebRTCConnection struct {
	ID        int       `json:"id"`
	ChannelID int       `json:"channel_id"`
	UserID    int       `json:"user_id"`
	PeerID    string    `json:"peer_id"`
	StreamType string   `json:"stream_type"` // "video", "audio", "screen"
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
} 