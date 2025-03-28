package services

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"saas-chat-system/internal/models"
)

// SubscriptionService handles subscription-related operations
type SubscriptionService struct {
	db *sql.DB
}

// NewSubscriptionService creates a new subscription service
func NewSubscriptionService(db *sql.DB) *SubscriptionService {
	return &SubscriptionService{
		db: db,
	}
}

// GetPlan retrieves a subscription plan by ID
func (s *SubscriptionService) GetPlan(planID int) (*models.Plan, error) {
	var plan models.Plan
	query := `
		SELECT id, name, description, price, interval,
			   features, limits, created_at, updated_at
		FROM plans
		WHERE id = $1
	`
	err := s.db.QueryRow(query, planID).Scan(
		&plan.ID, &plan.Name, &plan.Description,
		&plan.Price, &plan.Interval, &plan.Features,
		&plan.Limits, &plan.CreatedAt, &plan.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &plan, nil
}

// ListPlans retrieves all available subscription plans
func (s *SubscriptionService) ListPlans() ([]models.Plan, error) {
	var plans []models.Plan
	query := `
		SELECT id, name, description, price, interval,
			   features, limits, created_at, updated_at
		FROM plans
		ORDER BY price ASC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var plan models.Plan
		err := rows.Scan(
			&plan.ID, &plan.Name, &plan.Description,
			&plan.Price, &plan.Interval, &plan.Features,
			&plan.Limits, &plan.CreatedAt, &plan.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, plan)
	}

	return plans, nil
}

// Subscribe creates a new subscription for a user
func (s *SubscriptionService) Subscribe(userID, planID int, paymentMethod string) (*models.Subscription, error) {
	// Get plan
	plan, err := s.GetPlan(planID)
	if err != nil {
		return nil, err
	}

	// Calculate subscription dates
	now := time.Now()
	var endDate time.Time
	if plan.Interval == "monthly" {
		endDate = now.AddDate(0, 1, 0)
	} else {
		endDate = now.AddDate(1, 0, 0)
	}

	// Create subscription
	subscription := &models.Subscription{
		UserID:        userID,
		PlanID:        planID,
		Status:        "active",
		StartDate:     now,
		EndDate:       endDate,
		AutoRenew:     true,
		PaymentMethod: paymentMethod,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	query := `
		INSERT INTO subscriptions (
			user_id, plan_id, status, start_date, end_date,
			auto_renew, payment_method, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	err = s.db.QueryRow(
		query,
		subscription.UserID, subscription.PlanID,
		subscription.Status, subscription.StartDate,
		subscription.EndDate, subscription.AutoRenew,
		subscription.PaymentMethod, subscription.CreatedAt,
		subscription.UpdatedAt,
	).Scan(&subscription.ID)

	if err != nil {
		return nil, err
	}

	// Create initial usage record
	err = s.createUsageRecord(subscription.ID)
	if err != nil {
		return nil, err
	}

	return subscription, nil
}

// CancelSubscription cancels a user's subscription
func (s *SubscriptionService) CancelSubscription(subscriptionID string) error {
	// Convert string ID to int
	var id int
	_, err := fmt.Sscanf(subscriptionID, "%d", &id)
	if err != nil {
		return fmt.Errorf("invalid subscription ID format: %v", err)
	}

	query := `
		UPDATE subscriptions
		SET status = 'cancelled',
			auto_renew = false,
			updated_at = NOW()
		WHERE id = $1
	`
	_, err = s.db.Exec(query, id)
	return err
}

// RenewSubscription renews a user's subscription
func (s *SubscriptionService) RenewSubscription(subscriptionID string) error {
	// Convert string ID to int
	var id int
	_, err := fmt.Sscanf(subscriptionID, "%d", &id)
	if err != nil {
		return fmt.Errorf("invalid subscription ID format: %v", err)
	}

	// Get subscription
	var subscription models.Subscription
	query := `
		SELECT id, plan_id, end_date
		FROM subscriptions
		WHERE id = $1
	`
	err = s.db.QueryRow(query, id).Scan(
		&subscription.ID, &subscription.PlanID,
		&subscription.EndDate,
	)
	if err != nil {
		return err
	}

	// Get plan
	plan, err := s.GetPlan(subscription.PlanID)
	if err != nil {
		return err
	}

	// Calculate new end date
	var newEndDate time.Time
	if plan.Interval == "monthly" {
		newEndDate = subscription.EndDate.AddDate(0, 1, 0)
	} else {
		newEndDate = subscription.EndDate.AddDate(1, 0, 0)
	}

	// Update subscription
	query = `
		UPDATE subscriptions
		SET status = 'active',
			end_date = $1,
			updated_at = NOW()
		WHERE id = $2
	`
	_, err = s.db.Exec(query, newEndDate, id)
	if err != nil {
		return err
	}

	// Create new usage record
	err = s.createUsageRecord(subscriptionID)
	if err != nil {
		return err
	}

	return nil
}

// GetUsage retrieves the current usage for a subscription
func (s *SubscriptionService) GetUsage(subscriptionID string) (*models.Usage, error) {
	var usage models.Usage
	query := `
		SELECT id, subscription_id, messages_sent,
			   tokens_used, storage_used, period_start,
			   period_end, created_at
		FROM usage
		WHERE subscription_id = $1
		ORDER BY period_start DESC
		LIMIT 1
	`
	err := s.db.QueryRow(query, subscriptionID).Scan(
		&usage.ID, &usage.SubscriptionID,
		&usage.MessagesSent, &usage.TokensUsed,
		&usage.StorageUsed, &usage.PeriodStart,
		&usage.PeriodEnd, &usage.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &usage, nil
}

// UpdateUsage updates the usage statistics for a subscription
func (s *SubscriptionService) UpdateUsage(subscriptionID int, messagesSent, tokensUsed int) error {
	// Convert subscription ID to string for GetUsage
	subscriptionIDStr := fmt.Sprintf("%d", subscriptionID)
	_, err := s.GetUsage(subscriptionIDStr)
	if err != nil {
		return fmt.Errorf("failed to get usage: %v", err)
	}

	query := `
		UPDATE usage
		SET messages_sent = messages_sent + $1,
			tokens_used = tokens_used + $2,
			updated_at = NOW()
		WHERE subscription_id = $3
		AND period_end > NOW()
	`
	_, err = s.db.Exec(query, messagesSent, tokensUsed, subscriptionID)
	return err
}

// CheckLimits checks if a user has exceeded their subscription limits
func (s *SubscriptionService) CheckLimits(subscriptionID int) (*models.PlanLimits, error) {
	// Convert subscription ID to string for GetUsage
	subscriptionIDStr := fmt.Sprintf("%d", subscriptionID)
	usage, err := s.GetUsage(subscriptionIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %v", err)
	}

	// Get plan
	plan, err := s.GetPlan(subscriptionID)
	if err != nil {
		return nil, err
	}

	// Check limits
	if usage.MessagesSent >= plan.Limits.MessagesPerDay {
		return nil, errors.New("daily message limit exceeded")
	}
	if usage.TokensUsed >= plan.Limits.TokensPerMonth {
		return nil, errors.New("monthly token limit exceeded")
	}
	if usage.StorageUsed >= plan.Limits.StorageGB {
		return nil, errors.New("storage limit exceeded")
	}

	return &plan.Limits, nil
}

// UpdateFileUsage updates the file-related usage statistics for a subscription
func (s *SubscriptionService) UpdateFileUsage(subscriptionID string, fileSize int64) error {
	// Convert string ID to int
	var id int
	_, err := fmt.Sscanf(subscriptionID, "%d", &id)
	if err != nil {
		return fmt.Errorf("invalid subscription ID format: %v", err)
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %v", err)
	}
	defer tx.Rollback()

	// Get current usage
	usage, err := s.GetUsage(subscriptionID)
	if err != nil {
		return err
	}

	// Update storage used
	usage.StorageUsed += fileSize
	usage.FilesUploaded++

	// Update usage record
	query := `
		UPDATE usage
		SET storage_used = $1,
			files_uploaded = $2,
			updated_at = NOW()
		WHERE subscription_id = $3
	`
	_, err = tx.Exec(query, usage.StorageUsed, usage.FilesUploaded, id)
	if err != nil {
		return fmt.Errorf("failed to update usage: %v", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// UpdateStorageUsage updates the storage usage after file deletion
func (s *SubscriptionService) UpdateStorageUsage(subscriptionID int, storageUsed int64) error {
	// Convert subscription ID to string for GetUsage
	subscriptionIDStr := fmt.Sprintf("%d", subscriptionID)
	_, err := s.GetUsage(subscriptionIDStr)
	if err != nil {
		return fmt.Errorf("failed to get usage: %v", err)
	}

	query := `
		UPDATE usage
		SET storage_used = storage_used + $1,
			updated_at = NOW()
		WHERE subscription_id = $2
		AND period_end > NOW()
	`
	_, err = s.db.Exec(query, storageUsed, subscriptionID)
	return err
}

// GetActiveSubscription gets the active subscription for a user
func (s *SubscriptionService) GetActiveSubscription(userID int) (*models.Subscription, error) {
	var subscription models.Subscription
	var plan models.Plan

	query := `
		SELECT s.id, s.user_id, s.plan_id, s.status, s.start_date, s.end_date,
			   s.auto_renew, s.payment_method, s.created_at, s.updated_at,
			   p.id, p.name, p.description, p.price, p.interval, p.features,
			   p.limits
		FROM subscriptions s
		JOIN plans p ON s.plan_id = p.id
		WHERE s.user_id = $1 AND s.status = 'active'
		ORDER BY s.created_at DESC
		LIMIT 1
	`
	err := s.db.QueryRow(query, userID).Scan(
		&subscription.ID, &subscription.UserID, &subscription.PlanID,
		&subscription.Status, &subscription.StartDate, &subscription.EndDate,
		&subscription.AutoRenew, &subscription.PaymentMethod,
		&subscription.CreatedAt, &subscription.UpdatedAt,
		&plan.ID, &plan.Name, &plan.Description, &plan.Price,
		&plan.Interval, &plan.Features, &plan.Limits,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	subscription.Plan2 = &plan
	return &subscription, nil
}

// GetSubscription retrieves a subscription by ID
func (s *SubscriptionService) GetSubscription(subscriptionID string) (*models.Subscription, error) {
	var subscription models.Subscription
	var plan models.Plan

	query := `
		SELECT s.id, s.user_id, s.plan_id, s.status, s.start_date,
			   s.end_date, s.auto_renew, s.payment_method, s.billing_email,
			   s.created_at, s.updated_at, p.id, p.name, p.description,
			   p.price, p.interval, p.features, p.limits, p.created_at,
			   p.updated_at
		FROM subscriptions s
		JOIN plans p ON s.plan_id = p.id
		WHERE s.id = $1
	`
	err := s.db.QueryRow(query, subscriptionID).Scan(
		&subscription.ID, &subscription.UserID, &subscription.PlanID,
		&subscription.Status, &subscription.StartDate, &subscription.EndDate,
		&subscription.AutoRenew, &subscription.PaymentMethod, &subscription.BillingEmail,
		&subscription.CreatedAt, &subscription.UpdatedAt,
		&plan.ID, &plan.Name, &plan.Description, &plan.Price,
		&plan.Interval, &plan.Features, &plan.Limits, &plan.CreatedAt,
		&plan.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	subscription.Plan2 = &plan
	return &subscription, nil
}

// CreateUsageRecord creates a new usage record for a subscription
func (s *SubscriptionService) CreateUsageRecord(subscriptionID string, usage *models.Usage) error {
	// Convert string ID to int
	var id int
	_, err := fmt.Sscanf(subscriptionID, "%d", &id)
	if err != nil {
		return fmt.Errorf("invalid subscription ID format: %v", err)
	}

	query := `
		INSERT INTO usage (subscription_id, messages_sent, tokens_used,
						  storage_used, files_uploaded, period_start,
						  period_end, created_at, last_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id
	`
	return s.db.QueryRow(
		query,
		id,
		usage.MessagesSent,
		usage.TokensUsed,
		usage.StorageUsed,
		usage.FilesUploaded,
		usage.PeriodStart,
		usage.PeriodEnd,
	).Scan(&usage.ID)
}

// UpdateSubscription updates a subscription's details
func (s *SubscriptionService) UpdateSubscription(subscription *models.Subscription) error {
	query := `
		UPDATE subscriptions
		SET status = $1, start_date = $2, end_date = $3,
			auto_renew = $4, payment_method = $5, billing_email = $6,
			updated_at = NOW()
		WHERE id = $7
	`
	_, err := s.db.Exec(
		query,
		subscription.Status,
		subscription.StartDate,
		subscription.EndDate,
		subscription.AutoRenew,
		subscription.PaymentMethod,
		subscription.BillingEmail,
		subscription.ID,
	)
	return err
}

// Helper functions

func (s *SubscriptionService) createUsageRecord(subscriptionID string) error {
	// Convert string ID to int
	var id int
	_, err := fmt.Sscanf(subscriptionID, "%d", &id)
	if err != nil {
		return fmt.Errorf("invalid subscription ID format: %v", err)
	}

	now := time.Now()
	query := `
		INSERT INTO usage (subscription_id, messages_sent, tokens_used, storage_used, period_start, period_end, created_at)
		VALUES ($1, 0, 0, 0, $2, $3, NOW())
	`
	_, err = s.db.Exec(query, id, now, now.AddDate(0, 1, 0))
	return err
}
