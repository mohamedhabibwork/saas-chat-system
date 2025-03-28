package models

import (
	"time"
)

// User is defined in model_resolver.go

// Role represents a user role
type Role struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Permission represents a specific permission
type Permission struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Session represents a user session
type Session struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Token        string    `json:"token"`
	DeviceInfo   string    `json:"device_info"`
	IPAddress    string    `json:"ip_address"`
	LastActivity time.Time `json:"last_activity"`
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuthToken represents an authentication token
type AuthToken struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	Type      string    `json:"type"` // "access", "refresh", "reset"
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// LoginAttempt represents a login attempt
type LoginAttempt struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	IPAddress string    `json:"ip_address"`
	Success   bool      `json:"success"`
	CreatedAt time.Time `json:"created_at"`
}

// PasswordReset represents a password reset request
type PasswordReset struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
} 