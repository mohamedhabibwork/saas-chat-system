package models

import (
	"time"
)

// TrackingEvent represents a tracked event in the system
type TrackingEvent struct {
	ID        string                 `json:"id" gorm:"primaryKey"`
	EventType string                 `json:"event_type"`
	Timestamp time.Time             `json:"timestamp"`
	UserID    string                `json:"user_id"`
	TenantID  string                `json:"tenant_id"`
	Metadata  map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// TrackingMetric represents a tracked metric in the system
type TrackingMetric struct {
	ID        string                 `json:"id" gorm:"primaryKey"`
	Name      string                 `json:"name"`
	Value     float64               `json:"value"`
	Timestamp time.Time             `json:"timestamp"`
	UserID    string                `json:"user_id"`
	TenantID  string                `json:"tenant_id"`
	Metadata  map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// TrackingError represents a tracked error in the system
type TrackingError struct {
	ID        string                 `json:"id" gorm:"primaryKey"`
	Message   string                 `json:"message"`
	Stack     string                 `json:"stack,omitempty"`
	Timestamp time.Time             `json:"timestamp"`
	UserID    string                `json:"user_id"`
	TenantID  string                `json:"tenant_id"`
	Metadata  map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// TrackingStats represents aggregated tracking statistics
type TrackingStats struct {
	TotalEvents    int64   `json:"total_events"`
	TotalMetrics   int64   `json:"total_metrics"`
	TotalErrors    int64   `json:"total_errors"`
	EventTypes     []string `json:"event_types"`
	MetricNames    []string `json:"metric_names"`
	ErrorMessages  []string `json:"error_messages"`
	LastEventTime  time.Time `json:"last_event_time"`
	LastMetricTime time.Time `json:"last_metric_time"`
	LastErrorTime  time.Time `json:"last_error_time"`
} 