package models

import (
	"fmt"
	"time"
)

// Plan represents a subscription plan
type Plan struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Interval    string    `json:"interval"` // monthly, yearly
	Features    string    `json:"features"` // JSON array of features
	Limits      PlanLimits `json:"limits"`  // JSON object of limits
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PlanLimits represents the limits for a subscription plan
type PlanLimits struct {
	MessagesPerDay  int   `json:"messages_per_day"`
	TokensPerMonth  int   `json:"tokens_per_month"`
	StorageGB       int64 `json:"storage_gb"`
	MaxFileSize     int64 `json:"max_file_size"`
	MaxChannels     int   `json:"max_channels"`
	MaxBots         int   `json:"max_bots"`
	MaxTabs         int   `json:"max_tabs"`
	MaxDailyUploads int   `json:"max_daily_uploads"`
}

// Usage represents usage statistics for a subscription
type Usage struct {
	ID              int       `json:"id"`
	SubscriptionID  int       `json:"subscription_id"`
	MessagesSent    int       `json:"messages_sent"`
	TokensUsed      int       `json:"tokens_used"`
	StorageUsed     int64     `json:"storage_used"`
	FilesUploaded   int       `json:"files_uploaded"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	CreatedAt       time.Time `json:"created_at"`
	LastUpdated     time.Time `json:"last_updated"`
}

// Use SubscriptionPlanType from model_resolver.go
// SubscriptionPlanType and its constants are defined in model_resolver.go

// Subscription struct and related methods are defined in model_resolver.go

// ReportSchedule represents a scheduled report configuration
type ReportSchedule struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	TenantID      string    `json:"tenant_id" gorm:"not null"`
	Name          string    `json:"name" gorm:"not null"`
	Type          string    `json:"type" gorm:"not null"` // user_activity, location, system_health
	Frequency     string    `json:"frequency" gorm:"not null"` // daily, weekly, monthly
	DayOfWeek     int       `json:"day_of_week"` // 0-6 for weekly reports
	DayOfMonth    int       `json:"day_of_month"` // 1-31 for monthly reports
	TimeOfDay     string    `json:"time_of_day" gorm:"not null"` // HH:MM format
	Recipients    []string  `json:"recipients" gorm:"type:text[]"`
	Format        string    `json:"format" gorm:"not null"` // json, csv, pdf
	Options       string    `json:"options" gorm:"type:jsonb"`
	LastRun       time.Time `json:"last_run"`
	NextRun       time.Time `json:"next_run"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CalculateNextRun determines the next run time for a report schedule
func (rs *ReportSchedule) CalculateNextRun() time.Time {
	now := time.Now()
	next := now

	// Parse time of day
	hour, minute := 0, 0
	_, err := fmt.Sscanf(rs.TimeOfDay, "%d:%d", &hour, &minute)
	if err != nil {
		return next
	}

	// Set the time of day
	next = time.Date(next.Year(), next.Month(), next.Day(), hour, minute, 0, 0, next.Location())

	// If the time has already passed today, move to the next occurrence
	if next.Before(now) {
		switch rs.Frequency {
		case "daily":
			next = next.Add(24 * time.Hour)
		case "weekly":
			daysUntilNext := (rs.DayOfWeek - int(next.Weekday()) + 7) % 7
			next = next.AddDate(0, 0, daysUntilNext)
		case "monthly":
			next = time.Date(next.Year(), next.Month()+1, rs.DayOfMonth, hour, minute, 0, 0, next.Location())
		}
	}

	return next
} 