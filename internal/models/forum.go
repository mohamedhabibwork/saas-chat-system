package models

import (
	"time"
)

// ForumCategory represents a category in the forum
type ForumCategory struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TenantID    string    `json:"tenant_id" gorm:"not null"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Order       int       `json:"order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ForumTopic represents a topic in the forum
type ForumTopic struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	TenantID    string    `json:"tenant_id" gorm:"not null"`
	CategoryID  string    `json:"category_id" gorm:"not null"`
	UserID      string    `json:"user_id" gorm:"not null"`
	Title       string    `json:"title" gorm:"not null"`
	Content     string    `json:"content" gorm:"not null"`
	Views       int       `json:"views" gorm:"default:0"`
	IsPinned    bool      `json:"is_pinned" gorm:"default:false"`
	IsLocked    bool      `json:"is_locked" gorm:"default:false"`
	LastPostAt  time.Time `json:"last_post_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ForumPost represents a post in a forum topic
type ForumPost struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	TopicID   string    `json:"topic_id" gorm:"not null"`
	UserID    string    `json:"user_id" gorm:"not null"`
	Content   string    `json:"content" gorm:"not null"`
	IsEdited  bool      `json:"is_edited" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ForumSubscription represents a user's subscription to a topic
type ForumSubscription struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	TopicID   string    `json:"topic_id" gorm:"not null"`
	UserID    string    `json:"user_id" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
}

// ForumNotification represents a notification for forum activity
type ForumNotification struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id" gorm:"not null"`
	TopicID   string    `json:"topic_id"`
	PostID    string    `json:"post_id"`
	Type      string    `json:"type" gorm:"not null"` // new_post, mention, etc.
	Read      bool      `json:"read" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
} 