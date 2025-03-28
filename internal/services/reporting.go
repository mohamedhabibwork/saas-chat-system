package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"saas-chat-system/internal/models"
)

// ReportingService handles report generation for tracking data
type ReportingService struct {
	db *sql.DB
}

// NewReportingService creates a new reporting service
func NewReportingService(db *sql.DB) *ReportingService {
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
	ID        string        `json:"id"`
	Type      string        `json:"type"`
	Options   ReportOptions `json:"options"`
	Data      interface{}   `json:"data"`
	CreatedAt time.Time     `json:"created_at"`
	Status    string        `json:"status"` // "pending", "completed", "failed"
	Error     string        `json:"error,omitempty"`
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
		Events  []models.TrackingEvent  `json:"events"`
		Metrics []models.TrackingMetric `json:"metrics"`
		Errors  []models.TrackingError  `json:"errors"`
		Summary map[string]interface{}  `json:"summary"`
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
		Locations []models.Location      `json:"locations"`
		Stats     *models.LocationStats  `json:"stats"`
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
	
	query := `
		SELECT id, user_id, tenant_id, event_type, metadata, 
		       timestamp, created_at, updated_at
		FROM tracking_events
		WHERE user_id = $1 AND tenant_id = $2 AND timestamp BETWEEN $3 AND $4
		ORDER BY timestamp DESC
	`
	
	rows, err := s.db.QueryContext(ctx, query, opts.UserID, opts.TenantID, opts.StartTime, opts.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var event models.TrackingEvent
		var metadataRaw []byte
		
		err := rows.Scan(
			&event.ID, &event.UserID, &event.TenantID, &event.EventType, 
			&metadataRaw, &event.Timestamp, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &event.Metadata); err != nil {
				return nil, err
			}
		}
		
		events = append(events, event)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return events, nil
}

func (s *ReportingService) getUserMetrics(ctx context.Context, opts ReportOptions) ([]models.TrackingMetric, error) {
	var metrics []models.TrackingMetric
	
	query := `
		SELECT id, user_id, tenant_id, name, value, 
		       metadata, timestamp, created_at, updated_at
		FROM tracking_metrics
		WHERE user_id = $1 AND tenant_id = $2 AND timestamp BETWEEN $3 AND $4
		ORDER BY timestamp DESC
	`
	
	rows, err := s.db.QueryContext(ctx, query, opts.UserID, opts.TenantID, opts.StartTime, opts.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var metric models.TrackingMetric
		var metadataRaw []byte
		
		err := rows.Scan(
			&metric.ID, &metric.UserID, &metric.TenantID, &metric.Name, 
			&metric.Value, &metadataRaw, &metric.Timestamp, 
			&metric.CreatedAt, &metric.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &metric.Metadata); err != nil {
				return nil, err
			}
		}
		
		metrics = append(metrics, metric)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return metrics, nil
}

func (s *ReportingService) getUserErrors(ctx context.Context, opts ReportOptions) ([]models.TrackingError, error) {
	var errors []models.TrackingError
	
	query := `
		SELECT id, user_id, tenant_id, message, stack, 
		       metadata, timestamp, created_at, updated_at
		FROM tracking_errors
		WHERE user_id = $1 AND tenant_id = $2 AND timestamp BETWEEN $3 AND $4
		ORDER BY timestamp DESC
	`
	
	rows, err := s.db.QueryContext(ctx, query, opts.UserID, opts.TenantID, opts.StartTime, opts.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var trackingError models.TrackingError
		var metadataRaw []byte
		
		err := rows.Scan(
			&trackingError.ID, &trackingError.UserID, &trackingError.TenantID, 
			&trackingError.Message, &trackingError.Stack, 
			&metadataRaw, &trackingError.Timestamp, 
			&trackingError.CreatedAt, &trackingError.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &trackingError.Metadata); err != nil {
				return nil, err
			}
		}
		
		errors = append(errors, trackingError)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return errors, nil
}

func (s *ReportingService) getLocationHistory(ctx context.Context, opts ReportOptions) ([]models.Location, error) {
	var locations []models.Location
	
	query := `
		SELECT id, user_id, tenant_id, latitude, longitude, accuracy, 
		       altitude, speed, heading, timestamp, metadata, created_at, updated_at
		FROM locations
		WHERE user_id = $1 AND tenant_id = $2 AND timestamp BETWEEN $3 AND $4
		ORDER BY timestamp ASC
	`
	
	rows, err := s.db.QueryContext(ctx, query, opts.UserID, opts.TenantID, opts.StartTime, opts.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var location models.Location
		var metadataRaw []byte
		
		err := rows.Scan(
			&location.ID, &location.UserID, &location.TenantID, 
			&location.Latitude, &location.Longitude, &location.Accuracy, 
			&location.Altitude, &location.Speed, &location.Heading, 
			&location.Timestamp, &metadataRaw, 
			&location.CreatedAt, &location.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &location.Metadata); err != nil {
				return nil, err
			}
		}
		
		locations = append(locations, location)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return locations, nil
}

func (s *ReportingService) getLocationStats(ctx context.Context, opts ReportOptions) (*models.LocationStats, error) {
	var stats models.LocationStats
	
	query := `
		SELECT COUNT(*) as total_locations, 
		       (SELECT COUNT(*) FROM location_history WHERE user_id = $1 AND tenant_id = $2) as total_history,
		       (SELECT AVG(speed) FROM locations WHERE user_id = $1 AND tenant_id = $2 AND speed > 0) as average_speed,
		       (SELECT MAX(speed) FROM locations WHERE user_id = $1 AND tenant_id = $2) as max_speed
		FROM locations
		WHERE user_id = $1 AND tenant_id = $2
	`
	
	err := s.db.QueryRowContext(ctx, query, opts.UserID, opts.TenantID).Scan(
		&stats.TotalLocations, &stats.TotalHistory, &stats.AverageSpeed, &stats.MaxSpeed,
	)
	if err != nil {
		return nil, err
	}
	
	// Get last location
	lastLocationQuery := `
		SELECT id, user_id, tenant_id, latitude, longitude, accuracy, 
		       altitude, speed, heading, timestamp, created_at, updated_at
		FROM locations
		WHERE user_id = $1 AND tenant_id = $2
		ORDER BY timestamp DESC
		LIMIT 1
	`
	
	var lastLocation models.Location
	err = s.db.QueryRowContext(ctx, lastLocationQuery, opts.UserID, opts.TenantID).Scan(
		&lastLocation.ID, &lastLocation.UserID, &lastLocation.TenantID,
		&lastLocation.Latitude, &lastLocation.Longitude, &lastLocation.Accuracy,
		&lastLocation.Altitude, &lastLocation.Speed, &lastLocation.Heading,
		&lastLocation.Timestamp, &lastLocation.CreatedAt, &lastLocation.UpdatedAt,
	)
	if err == nil {
		stats.LastLocation = lastLocation
		stats.LastUpdate = lastLocation.Timestamp
	}
	
	return &stats, nil
}

func (s *ReportingService) getSystemMetrics(ctx context.Context, opts ReportOptions) ([]models.TrackingMetric, error) {
	var metrics []models.TrackingMetric
	
	query := `
		SELECT id, user_id, tenant_id, name, value, 
		       metadata, timestamp, created_at, updated_at
		FROM tracking_metrics
		WHERE tenant_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC
	`
	
	rows, err := s.db.QueryContext(ctx, query, opts.TenantID, opts.StartTime, opts.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var metric models.TrackingMetric
		var metadataRaw []byte
		
		err := rows.Scan(
			&metric.ID, &metric.UserID, &metric.TenantID, &metric.Name, 
			&metric.Value, &metadataRaw, &metric.Timestamp, 
			&metric.CreatedAt, &metric.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &metric.Metadata); err != nil {
				return nil, err
			}
		}
		
		metrics = append(metrics, metric)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return metrics, nil
}

func (s *ReportingService) getSystemErrors(ctx context.Context, opts ReportOptions) ([]models.TrackingError, error) {
	var errors []models.TrackingError
	
	query := `
		SELECT id, user_id, tenant_id, message, stack, 
		       metadata, timestamp, created_at, updated_at
		FROM tracking_errors
		WHERE tenant_id = $1 AND timestamp BETWEEN $2 AND $3
		ORDER BY timestamp DESC
	`
	
	rows, err := s.db.QueryContext(ctx, query, opts.TenantID, opts.StartTime, opts.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var trackingError models.TrackingError
		var metadataRaw []byte
		
		err := rows.Scan(
			&trackingError.ID, &trackingError.UserID, &trackingError.TenantID, 
			&trackingError.Message, &trackingError.Stack, 
			&metadataRaw, &trackingError.Timestamp, 
			&trackingError.CreatedAt, &trackingError.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		if len(metadataRaw) > 0 {
			if err := json.Unmarshal(metadataRaw, &trackingError.Metadata); err != nil {
				return nil, err
			}
		}
		
		errors = append(errors, trackingError)
	}
	
	if err = rows.Err(); err != nil {
		return nil, err
	}
	
	return errors, nil
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
