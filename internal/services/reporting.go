package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mohamedhabibwork/saas-chat-system/internal/database"
	"github.com/mohamedhabibwork/saas-chat-system/internal/models"
)

// ReportingService handles report generation for tracking data
type ReportingService struct {
	db *database.DB
}

// NewReportingService creates a new reporting service
func NewReportingService(db *database.DB) *ReportingService {
	return &ReportingService{db: db}
}

// ReportOptions defines options for report generation
type ReportOptions struct {
	StartTime time.Time
	EndTime   time.Time
	UserID    string
	TenantID  string
	Format    string // "json", "csv", "pdf"
}

// Report represents a generated report
type Report struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Options   ReportOptions `json:"options"`
	Data      interface{} `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"` // "pending", "completed", "failed"
	Error     string    `json:"error,omitempty"`
}

// GenerateUserActivityReport generates a report of user activity
func (s *ReportingService) GenerateUserActivityReport(ctx context.Context, opts ReportOptions) (*Report, error) {
	report := &Report{
		ID:        generateReportID(),
		Type:      "user_activity",
		Options:   opts,
		CreatedAt: time.Now(),
		Status:    "pending",
	}

	// Get user events
	events, err := s.getUserEvents(ctx, opts)
	if err != nil {
		report.Status = "failed"
		report.Error = err.Error()
		return report, err
	}

	// Get user metrics
	metrics, err := s.getUserMetrics(ctx, opts)
	if err != nil {
		report.Status = "failed"
		report.Error = err.Error()
		return report, err
	}

	// Get user errors
	errors, err := s.getUserErrors(ctx, opts)
	if err != nil {
		report.Status = "failed"
		report.Error = err.Error()
		return report, err
	}

	// Generate report data
	reportData := struct {
		Events     []models.TrackingEvent  `json:"events"`
		Metrics    []models.TrackingMetric `json:"metrics"`
		Errors     []models.TrackingError  `json:"errors"`
		Summary    map[string]interface{}  `json:"summary"`
	}{
		Events:  events,
		Metrics: metrics,
		Errors:  errors,
		Summary: s.generateSummary(events, metrics, errors),
	}

	report.Data = reportData
	report.Status = "completed"
	return report, nil
}

// GenerateLocationReport generates a report of user location data
func (s *ReportingService) GenerateLocationReport(ctx context.Context, opts ReportOptions) (*Report, error) {
	report := &Report{
		ID:        generateReportID(),
		Type:      "location",
		Options:   opts,
		CreatedAt: time.Now(),
		Status:    "pending",
	}

	// Get location history
	locations, err := s.getLocationHistory(ctx, opts)
	if err != nil {
		report.Status = "failed"
		report.Error = err.Error()
		return report, err
	}

	// Get location stats
	stats, err := s.getLocationStats(ctx, opts)
	if err != nil {
		report.Status = "failed"
		report.Error = err.Error()
		return report, err
	}

	// Generate report data
	reportData := struct {
		Locations []models.Location     `json:"locations"`
		Stats     *models.LocationStats `json:"stats"`
		Summary   map[string]interface{} `json:"summary"`
	}{
		Locations: locations,
		Stats:     stats,
		Summary:   s.generateLocationSummary(locations, stats),
	}

	report.Data = reportData
	report.Status = "completed"
	return report, nil
}

// GenerateSystemHealthReport generates a report of system health metrics
func (s *ReportingService) GenerateSystemHealthReport(ctx context.Context, opts ReportOptions) (*Report, error) {
	report := &Report{
		ID:        generateReportID(),
		Type:      "system_health",
		Options:   opts,
		CreatedAt: time.Now(),
		Status:    "pending",
	}

	// Get system metrics
	metrics, err := s.getSystemMetrics(ctx, opts)
	if err != nil {
		report.Status = "failed"
		report.Error = err.Error()
		return report, err
	}

	// Get system errors
	errors, err := s.getSystemErrors(ctx, opts)
	if err != nil {
		report.Status = "failed"
		report.Error = err.Error()
		return report, err
	}

	// Generate report data
	reportData := struct {
		Metrics []models.TrackingMetric `json:"metrics"`
		Errors  []models.TrackingError  `json:"errors"`
		Summary map[string]interface{}  `json:"summary"`
	}{
		Metrics: metrics,
		Errors:  errors,
		Summary: s.generateSystemSummary(metrics, errors),
	}

	report.Data = reportData
	report.Status = "completed"
	return report, nil
}

// Helper functions
func (s *ReportingService) getUserEvents(ctx context.Context, opts ReportOptions) ([]models.TrackingEvent, error) {
	var events []models.TrackingEvent
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ? AND timestamp BETWEEN ? AND ?",
			opts.UserID, opts.TenantID, opts.StartTime, opts.EndTime).
		Order("timestamp DESC").
		Find(&events).Error
	return events, err
}

func (s *ReportingService) getUserMetrics(ctx context.Context, opts ReportOptions) ([]models.TrackingMetric, error) {
	var metrics []models.TrackingMetric
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ? AND timestamp BETWEEN ? AND ?",
			opts.UserID, opts.TenantID, opts.StartTime, opts.EndTime).
		Order("timestamp DESC").
		Find(&metrics).Error
	return metrics, err
}

func (s *ReportingService) getUserErrors(ctx context.Context, opts ReportOptions) ([]models.TrackingError, error) {
	var errors []models.TrackingError
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ? AND timestamp BETWEEN ? AND ?",
			opts.UserID, opts.TenantID, opts.StartTime, opts.EndTime).
		Order("timestamp DESC").
		Find(&errors).Error
	return errors, err
}

func (s *ReportingService) getLocationHistory(ctx context.Context, opts ReportOptions) ([]models.Location, error) {
	var locations []models.Location
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ? AND timestamp BETWEEN ? AND ?",
			opts.UserID, opts.TenantID, opts.StartTime, opts.EndTime).
		Order("timestamp ASC").
		Find(&locations).Error
	return locations, err
}

func (s *ReportingService) getLocationStats(ctx context.Context, opts ReportOptions) (*models.LocationStats, error) {
	var stats models.LocationStats
	err := s.db.WithContext(ctx).
		Where("user_id = ? AND tenant_id = ?", opts.UserID, opts.TenantID).
		First(&stats).Error
	return &stats, err
}

func (s *ReportingService) getSystemMetrics(ctx context.Context, opts ReportOptions) ([]models.TrackingMetric, error) {
	var metrics []models.TrackingMetric
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND timestamp BETWEEN ? AND ?",
			opts.TenantID, opts.StartTime, opts.EndTime).
		Order("timestamp DESC").
		Find(&metrics).Error
	return metrics, err
}

func (s *ReportingService) getSystemErrors(ctx context.Context, opts ReportOptions) ([]models.TrackingError, error) {
	var errors []models.TrackingError
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND timestamp BETWEEN ? AND ?",
			opts.TenantID, opts.StartTime, opts.EndTime).
		Order("timestamp DESC").
		Find(&errors).Error
	return errors, err
}

func (s *ReportingService) generateSummary(events []models.TrackingEvent, metrics []models.TrackingMetric, errors []models.TrackingError) map[string]interface{} {
	summary := make(map[string]interface{})

	// Event summary
	eventCounts := make(map[string]int)
	for _, event := range events {
		eventCounts[event.EventType]++
	}
	summary["event_counts"] = eventCounts

	// Metric summary
	metricAverages := make(map[string]float64)
	metricCounts := make(map[string]int)
	for _, metric := range metrics {
		metricAverages[metric.Name] += metric.Value
		metricCounts[metric.Name]++
	}
	for name, count := range metricCounts {
		metricAverages[name] /= float64(count)
	}
	summary["metric_averages"] = metricAverages

	// Error summary
	errorCounts := make(map[string]int)
	for _, err := range errors {
		errorCounts[err.Message]++
	}
	summary["error_counts"] = errorCounts

	return summary
}

func (s *ReportingService) generateLocationSummary(locations []models.Location, stats *models.LocationStats) map[string]interface{} {
	summary := make(map[string]interface{})

	// Location summary
	summary["total_points"] = len(locations)
	summary["total_distance"] = stats.TotalDistance
	summary["average_speed"] = stats.AverageSpeed
	summary["max_speed"] = stats.MaxSpeed

	// Time-based summary
	if len(locations) > 0 {
		summary["start_time"] = locations[0].Timestamp
		summary["end_time"] = locations[len(locations)-1].Timestamp
		summary["duration"] = locations[len(locations)-1].Timestamp.Sub(locations[0].Timestamp)
	}

	return summary
}

func (s *ReportingService) generateSystemSummary(metrics []models.TrackingMetric, errors []models.TrackingError) map[string]interface{} {
	summary := make(map[string]interface{})

	// System metrics summary
	metricAverages := make(map[string]float64)
	metricCounts := make(map[string]int)
	for _, metric := range metrics {
		metricAverages[metric.Name] += metric.Value
		metricCounts[metric.Name]++
	}
	for name, count := range metricCounts {
		metricAverages[name] /= float64(count)
	}
	summary["metric_averages"] = metricAverages

	// Error summary
	errorCounts := make(map[string]int)
	for _, err := range errors {
		errorCounts[err.Message]++
	}
	summary["error_counts"] = errorCounts

	return summary
}

func generateReportID() string {
	return fmt.Sprintf("report_%d", time.Now().UnixNano())
} 