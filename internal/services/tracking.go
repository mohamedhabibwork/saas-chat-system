package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"saas-chat-system/internal/models"
)

// TrackingService handles tracking operations
type TrackingService struct {
	db *sql.DB
}

// NewTrackingService creates a new tracking service
func NewTrackingService(db *sql.DB) *TrackingService {
	return &TrackingService{db: db}
}

// TrackEvent records a new tracking event
func (s *TrackingService) TrackEvent(ctx context.Context, event *models.TrackingEvent) error {
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	query := `
		INSERT INTO tracking_events (
			id, tenant_id, user_id, event_type,
			metadata, timestamp, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := s.db.ExecContext(ctx, query,
		event.ID, event.TenantID, event.UserID,
		event.EventType, event.Metadata, event.Timestamp,
		event.CreatedAt, event.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creating tracking event: %v", err)
	}

	return nil
}

// TrackMetric records a new tracking metric
func (s *TrackingService) TrackMetric(ctx context.Context, metric *models.TrackingMetric) error {
	metric.ID = uuid.New().String()
	metric.CreatedAt = time.Now()
	metric.UpdatedAt = time.Now()

	query := `
		INSERT INTO tracking_metrics (
			id, tenant_id, user_id, name, value,
			metadata, timestamp, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := s.db.ExecContext(ctx, query,
		metric.ID, metric.TenantID, metric.UserID,
		metric.Name, metric.Value, metric.Metadata,
		metric.Timestamp, metric.CreatedAt, metric.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("error creating tracking metric: %v", err)
	}

	return nil
}

// TrackError records a new tracking error
func (s *TrackingService) TrackError(ctx context.Context, err *models.TrackingError) error {
	err.ID = uuid.New().String()
	err.CreatedAt = time.Now()
	err.UpdatedAt = time.Now()

	query := `
		INSERT INTO tracking_errors (
			id, tenant_id, user_id, message,
			stack, metadata, timestamp, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err2 := s.db.ExecContext(ctx, query,
		err.ID, err.TenantID, err.UserID,
		err.Message, err.Stack, err.Metadata,
		err.Timestamp, err.CreatedAt, err.UpdatedAt,
	)
	if err2 != nil {
		return fmt.Errorf("error creating tracking error: %v", err2)
	}

	return nil
}

// GetTrackingStats retrieves tracking statistics
func (s *TrackingService) GetTrackingStats(ctx context.Context, tenantID string) (*models.TrackingStats, error) {
	var stats models.TrackingStats

	// Get total counts
	err := s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM tracking_events WHERE tenant_id = $1",
		tenantID,
	).Scan(&stats.TotalEvents)
	if err != nil {
		return nil, fmt.Errorf("error getting event count: %v", err)
	}

	err = s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM tracking_metrics WHERE tenant_id = $1",
		tenantID,
	).Scan(&stats.TotalMetrics)
	if err != nil {
		return nil, fmt.Errorf("error getting metric count: %v", err)
	}

	err = s.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM tracking_errors WHERE tenant_id = $1",
		tenantID,
	).Scan(&stats.TotalErrors)
	if err != nil {
		return nil, fmt.Errorf("error getting error count: %v", err)
	}

	// Get unique event types
	rows, err := s.db.QueryContext(ctx,
		"SELECT DISTINCT event_type FROM tracking_events WHERE tenant_id = $1",
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting event types: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var eventType string
		if err := rows.Scan(&eventType); err != nil {
			return nil, fmt.Errorf("error scanning event type: %v", err)
		}
		stats.EventTypes = append(stats.EventTypes, eventType)
	}

	// Get unique metric names
	rows, err = s.db.QueryContext(ctx,
		"SELECT DISTINCT name FROM tracking_metrics WHERE tenant_id = $1",
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting metric names: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("error scanning metric name: %v", err)
		}
		stats.MetricNames = append(stats.MetricNames, name)
	}

	// Get unique error messages
	rows, err = s.db.QueryContext(ctx,
		"SELECT DISTINCT message FROM tracking_errors WHERE tenant_id = $1",
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("error getting error messages: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var message string
		if err := rows.Scan(&message); err != nil {
			return nil, fmt.Errorf("error scanning error message: %v", err)
		}
		stats.ErrorMessages = append(stats.ErrorMessages, message)
	}

	// Get last timestamps
	err = s.db.QueryRowContext(ctx,
		"SELECT timestamp FROM tracking_events WHERE tenant_id = $1 ORDER BY timestamp DESC LIMIT 1",
		tenantID,
	).Scan(&stats.LastEventTime)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting last event time: %v", err)
	}

	err = s.db.QueryRowContext(ctx,
		"SELECT timestamp FROM tracking_metrics WHERE tenant_id = $1 ORDER BY timestamp DESC LIMIT 1",
		tenantID,
	).Scan(&stats.LastMetricTime)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting last metric time: %v", err)
	}

	err = s.db.QueryRowContext(ctx,
		"SELECT timestamp FROM tracking_errors WHERE tenant_id = $1 ORDER BY timestamp DESC LIMIT 1",
		tenantID,
	).Scan(&stats.LastErrorTime)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("error getting last error time: %v", err)
	}

	return &stats, nil
}

// GetEvents retrieves tracking events with pagination
func (s *TrackingService) GetEvents(ctx context.Context, tenantID string, limit, offset int) ([]models.TrackingEvent, error) {
	query := `
		SELECT id, tenant_id, user_id, event_type,
			   metadata, timestamp, created_at, updated_at
		FROM tracking_events
		WHERE tenant_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := s.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error retrieving events: %v", err)
	}
	defer rows.Close()

	var events []models.TrackingEvent
	for rows.Next() {
		var event models.TrackingEvent
		err := rows.Scan(
			&event.ID, &event.TenantID, &event.UserID,
			&event.EventType, &event.Metadata, &event.Timestamp,
			&event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning event row: %v", err)
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %v", err)
	}

	return events, nil
}

// GetMetrics retrieves tracking metrics with pagination
func (s *TrackingService) GetMetrics(ctx context.Context, tenantID string, limit, offset int) ([]models.TrackingMetric, error) {
	query := `
		SELECT id, tenant_id, user_id, name, value,
			   metadata, timestamp, created_at, updated_at
		FROM tracking_metrics
		WHERE tenant_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := s.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error retrieving metrics: %v", err)
	}
	defer rows.Close()

	var metrics []models.TrackingMetric
	for rows.Next() {
		var metric models.TrackingMetric
		err := rows.Scan(
			&metric.ID, &metric.TenantID, &metric.UserID,
			&metric.Name, &metric.Value, &metric.Metadata,
			&metric.Timestamp, &metric.CreatedAt, &metric.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning metric row: %v", err)
		}
		metrics = append(metrics, metric)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating metric rows: %v", err)
	}

	return metrics, nil
}

// GetErrors retrieves tracking errors with pagination
func (s *TrackingService) GetErrors(ctx context.Context, tenantID string, limit, offset int) ([]models.TrackingError, error) {
	query := `
		SELECT id, tenant_id, user_id, message,
			   stack, metadata, timestamp, created_at, updated_at
		FROM tracking_errors
		WHERE tenant_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := s.db.QueryContext(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error retrieving errors: %v", err)
	}
	defer rows.Close()

	var errors []models.TrackingError
	for rows.Next() {
		var trackingError models.TrackingError
		err := rows.Scan(
			&trackingError.ID, &trackingError.TenantID,
			&trackingError.UserID, &trackingError.Message,
			&trackingError.Stack, &trackingError.Metadata,
			&trackingError.Timestamp, &trackingError.CreatedAt,
			&trackingError.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning error row: %v", err)
		}
		errors = append(errors, trackingError)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating error rows: %v", err)
	}

	return errors, nil
}

// CleanupOldData removes tracking data older than the specified duration
func (s *TrackingService) CleanupOldData(ctx context.Context, tenantID string, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	// Delete old events
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM tracking_events WHERE tenant_id = $1 AND timestamp < $2",
		tenantID, cutoff,
	)
	if err != nil {
		return fmt.Errorf("error deleting old events: %v", err)
	}

	// Delete old metrics
	_, err = s.db.ExecContext(ctx,
		"DELETE FROM tracking_metrics WHERE tenant_id = $1 AND timestamp < $2",
		tenantID, cutoff,
	)
	if err != nil {
		return fmt.Errorf("error deleting old metrics: %v", err)
	}

	// Delete old errors
	_, err = s.db.ExecContext(ctx,
		"DELETE FROM tracking_errors WHERE tenant_id = $1 AND timestamp < $2",
		tenantID, cutoff,
	)
	if err != nil {
		return fmt.Errorf("error deleting old errors: %v", err)
	}

	return nil
}
