package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mohamedhabibwork/saas-chat-system/internal/database"
	"github.com/mohamedhabibwork/saas-chat-system/internal/models"
)

// TrackingService handles tracking operations
type TrackingService struct {
	db *database.DB
}

// NewTrackingService creates a new tracking service
func NewTrackingService(db *database.DB) *TrackingService {
	return &TrackingService{db: db}
}

// TrackEvent records a new tracking event
func (s *TrackingService) TrackEvent(ctx context.Context, event *models.TrackingEvent) error {
	event.ID = uuid.New().String()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Create(event).Error
}

// TrackMetric records a new tracking metric
func (s *TrackingService) TrackMetric(ctx context.Context, metric *models.TrackingMetric) error {
	metric.ID = uuid.New().String()
	metric.CreatedAt = time.Now()
	metric.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Create(metric).Error
}

// TrackError records a new tracking error
func (s *TrackingService) TrackError(ctx context.Context, err *models.TrackingError) error {
	err.ID = uuid.New().String()
	err.CreatedAt = time.Now()
	err.UpdatedAt = time.Now()

	return s.db.WithContext(ctx).Create(err).Error
}

// GetTrackingStats retrieves tracking statistics
func (s *TrackingService) GetTrackingStats(ctx context.Context, tenantID string) (*models.TrackingStats, error) {
	var stats models.TrackingStats

	// Get total counts
	if err := s.db.WithContext(ctx).Model(&models.TrackingEvent{}).Where("tenant_id = ?", tenantID).Count(&stats.TotalEvents).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&models.TrackingMetric{}).Where("tenant_id = ?", tenantID).Count(&stats.TotalMetrics).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&models.TrackingError{}).Where("tenant_id = ?", tenantID).Count(&stats.TotalErrors).Error; err != nil {
		return nil, err
	}

	// Get unique event types
	if err := s.db.WithContext(ctx).Model(&models.TrackingEvent{}).
		Where("tenant_id = ?", tenantID).
		Distinct().
		Pluck("event_type", &stats.EventTypes).Error; err != nil {
		return nil, err
	}

	// Get unique metric names
	if err := s.db.WithContext(ctx).Model(&models.TrackingMetric{}).
		Where("tenant_id = ?", tenantID).
		Distinct().
		Pluck("name", &stats.MetricNames).Error; err != nil {
		return nil, err
	}

	// Get unique error messages
	if err := s.db.WithContext(ctx).Model(&models.TrackingError{}).
		Where("tenant_id = ?", tenantID).
		Distinct().
		Pluck("message", &stats.ErrorMessages).Error; err != nil {
		return nil, err
	}

	// Get last timestamps
	var lastEvent models.TrackingEvent
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("timestamp DESC").First(&lastEvent).Error; err == nil {
		stats.LastEventTime = lastEvent.Timestamp
	}

	var lastMetric models.TrackingMetric
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("timestamp DESC").First(&lastMetric).Error; err == nil {
		stats.LastMetricTime = lastMetric.Timestamp
	}

	var lastError models.TrackingError
	if err := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("timestamp DESC").First(&lastError).Error; err == nil {
		stats.LastErrorTime = lastError.Timestamp
	}

	return &stats, nil
}

// GetEvents retrieves tracking events with pagination
func (s *TrackingService) GetEvents(ctx context.Context, tenantID string, limit, offset int) ([]models.TrackingEvent, error) {
	var events []models.TrackingEvent
	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&events).Error
	return events, err
}

// GetMetrics retrieves tracking metrics with pagination
func (s *TrackingService) GetMetrics(ctx context.Context, tenantID string, limit, offset int) ([]models.TrackingMetric, error) {
	var metrics []models.TrackingMetric
	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&metrics).Error
	return metrics, err
}

// GetErrors retrieves tracking errors with pagination
func (s *TrackingService) GetErrors(ctx context.Context, tenantID string, limit, offset int) ([]models.TrackingError, error) {
	var errors []models.TrackingError
	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("timestamp DESC").
		Limit(limit).
		Offset(offset).
		Find(&errors).Error
	return errors, err
}

// CleanupOldData removes tracking data older than the specified duration
func (s *TrackingService) CleanupOldData(ctx context.Context, tenantID string, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)

	// Delete old events
	if err := s.db.WithContext(ctx).Where("tenant_id = ? AND timestamp < ?", tenantID, cutoff).Delete(&models.TrackingEvent{}).Error; err != nil {
		return err
	}

	// Delete old metrics
	if err := s.db.WithContext(ctx).Where("tenant_id = ? AND timestamp < ?", tenantID, cutoff).Delete(&models.TrackingMetric{}).Error; err != nil {
		return err
	}

	// Delete old errors
	if err := s.db.WithContext(ctx).Where("tenant_id = ? AND timestamp < ?", tenantID, cutoff).Delete(&models.TrackingError{}).Error; err != nil {
		return err
	}

	return nil
} 