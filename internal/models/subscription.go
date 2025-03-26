package models

import (
	"fmt"
	"time"
)

// SubscriptionPlan represents available subscription plans
type SubscriptionPlan string

const (
	PlanFree     SubscriptionPlan = "free"
	PlanBasic    SubscriptionPlan = "basic"
	PlanPro      SubscriptionPlan = "pro"
	PlanEnterprise SubscriptionPlan = "enterprise"
)

// SubscriptionFeatures represents available features for each plan
var SubscriptionFeatures = map[SubscriptionPlan][]string{
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

// Subscription represents a tenant's subscription
type Subscription struct {
	ID            string         `json:"id" gorm:"primaryKey"`
	TenantID      string         `json:"tenant_id" gorm:"not null"`
	Plan          SubscriptionPlan `json:"plan" gorm:"not null"`
	Status        string         `json:"status" gorm:"not null"` // active, suspended, cancelled
	StartDate     time.Time      `json:"start_date" gorm:"not null"`
	EndDate       time.Time      `json:"end_date" gorm:"not null"`
	AutoRenew     bool           `json:"auto_renew" gorm:"not null"`
	PaymentMethod string         `json:"payment_method" gorm:"not null"`
	BillingEmail  string         `json:"billing_email" gorm:"not null"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

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