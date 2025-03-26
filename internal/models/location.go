package models

import (
	"time"
)

// Location represents a user's location
type Location struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Accuracy  float64   `json:"accuracy,omitempty"` // Accuracy in meters
	Altitude  float64   `json:"altitude,omitempty"` // Altitude in meters
	Speed     float64   `json:"speed,omitempty"`    // Speed in meters per second
	Heading   float64   `json:"heading,omitempty"`  // Heading in degrees
	Timestamp time.Time `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LocationHistory represents a user's location history
type LocationHistory struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Locations []Location `json:"locations" gorm:"type:jsonb"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LocationStats represents aggregated location statistics
type LocationStats struct {
	TotalLocations int64     `json:"total_locations"`
	TotalHistory   int64     `json:"total_history"`
	LastLocation   Location  `json:"last_location"`
	LastUpdate     time.Time `json:"last_update"`
	AverageSpeed   float64   `json:"average_speed"`
	MaxSpeed       float64   `json:"max_speed"`
	TotalDistance  float64   `json:"total_distance"` // Total distance traveled in meters
} 