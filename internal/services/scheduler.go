package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"saas-chat-system/internal/models"
)

// SchedulerService handles scheduled report generation and delivery
type SchedulerService struct {
	db            *sql.DB
	reportingSvc  *ReportingService
	emailSvc      *EmailService
	schedules     map[string]*models.ReportSchedule
	scheduleMutex sync.RWMutex
	stopChan      chan struct{}
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService(db *sql.DB, reportingSvc *ReportingService, emailSvc *EmailService) *SchedulerService {
	return &SchedulerService{
		db:           db,
		reportingSvc: reportingSvc,
		emailSvc:     emailSvc,
		schedules:    make(map[string]*models.ReportSchedule),
		stopChan:     make(chan struct{}),
	}
}

// Start starts the scheduler service
func (s *SchedulerService) Start(ctx context.Context) error {
	// Load all active schedules
	if err := s.loadSchedules(ctx); err != nil {
		return fmt.Errorf("error loading schedules: %v", err)
	}

	// Start the scheduler loop
	go s.run(ctx)

	return nil
}

// Stop stops the scheduler service
func (s *SchedulerService) Stop() {
	close(s.stopChan)
}

// AddSchedule adds a new report schedule
func (s *SchedulerService) AddSchedule(ctx context.Context, schedule *models.ReportSchedule) error {
	// Validate schedule
	if err := s.validateSchedule(schedule); err != nil {
		return err
	}

	// Save to database
	query := `
		INSERT INTO report_schedules (
			id, tenant_id, name, type, options, frequency, 
			time_of_day, day_of_week, day_of_month, recipients,
			format, last_run, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	recipientsJSON, err := json.Marshal(schedule.Recipients)
	if err != nil {
		return fmt.Errorf("error marshaling recipients: %v", err)
	}

	optionsJSON := []byte(schedule.Options)

	_, err = s.db.ExecContext(ctx, query,
		schedule.ID, schedule.TenantID, schedule.Name, schedule.Type,
		optionsJSON, schedule.Frequency, schedule.TimeOfDay, schedule.DayOfWeek,
		schedule.DayOfMonth, recipientsJSON, schedule.Format,
		schedule.LastRun, schedule.CreatedAt, schedule.UpdatedAt)

	if err != nil {
		return fmt.Errorf("error saving schedule: %v", err)
	}

	// Add to memory
	s.scheduleMutex.Lock()
	s.schedules[schedule.ID] = schedule
	s.scheduleMutex.Unlock()

	return nil
}

// UpdateSchedule updates an existing report schedule
func (s *SchedulerService) UpdateSchedule(ctx context.Context, schedule *models.ReportSchedule) error {
	// Validate schedule
	if err := s.validateSchedule(schedule); err != nil {
		return err
	}

	// Update in database
	query := `
		UPDATE report_schedules
		SET tenant_id = $1, name = $2, type = $3, options = $4,
			frequency = $5, time_of_day = $6, day_of_week = $7, day_of_month = $8,
			recipients = $9, format = $10, last_run = $11, updated_at = $12
		WHERE id = $13
	`

	recipientsJSON, err := json.Marshal(schedule.Recipients)
	if err != nil {
		return fmt.Errorf("error marshaling recipients: %v", err)
	}

	optionsJSON := []byte(schedule.Options)

	_, err = s.db.ExecContext(ctx, query,
		schedule.TenantID, schedule.Name, schedule.Type,
		optionsJSON, schedule.Frequency, schedule.TimeOfDay, schedule.DayOfWeek,
		schedule.DayOfMonth, recipientsJSON, schedule.Format,
		schedule.LastRun, time.Now(), schedule.ID)

	if err != nil {
		return fmt.Errorf("error updating schedule: %v", err)
	}

	// Update in memory
	s.scheduleMutex.Lock()
	s.schedules[schedule.ID] = schedule
	s.scheduleMutex.Unlock()

	return nil
}

// DeleteSchedule deletes a report schedule
func (s *SchedulerService) DeleteSchedule(ctx context.Context, scheduleID string) error {
	// Delete from database
	query := `DELETE FROM report_schedules WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, scheduleID)
	if err != nil {
		return fmt.Errorf("error deleting schedule: %v", err)
	}

	// Delete from memory
	s.scheduleMutex.Lock()
	delete(s.schedules, scheduleID)
	s.scheduleMutex.Unlock()

	return nil
}

// GetSchedule retrieves a report schedule
func (s *SchedulerService) GetSchedule(scheduleID string) (*models.ReportSchedule, error) {
	s.scheduleMutex.RLock()
	defer s.scheduleMutex.RUnlock()

	schedule, exists := s.schedules[scheduleID]
	if !exists {
		return nil, fmt.Errorf("schedule not found: %s", scheduleID)
	}

	return schedule, nil
}

// ListSchedules retrieves all report schedules
func (s *SchedulerService) ListSchedules() []*models.ReportSchedule {
	s.scheduleMutex.RLock()
	defer s.scheduleMutex.RUnlock()

	schedules := make([]*models.ReportSchedule, 0, len(s.schedules))
	for _, schedule := range s.schedules {
		schedules = append(schedules, schedule)
	}

	return schedules
}

// run is the main scheduler loop
func (s *SchedulerService) run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.checkAndRunSchedules(ctx)
		}
	}
}

// checkAndRunSchedules checks for and runs any due schedules
func (s *SchedulerService) checkAndRunSchedules(ctx context.Context) {
	s.scheduleMutex.RLock()
	defer s.scheduleMutex.RUnlock()

	now := time.Now()
	for _, schedule := range s.schedules {
		if schedule.NextRun.Before(now) {
			go s.runSchedule(ctx, schedule)
		}
	}
}

// runSchedule generates and delivers a scheduled report
func (s *SchedulerService) runSchedule(ctx context.Context, schedule *models.ReportSchedule) {
	// Parse report options
	var options ReportOptions
	if err := json.Unmarshal([]byte(schedule.Options), &options); err != nil {
		fmt.Printf("Error parsing report options: %v\n", err)
		return
	}

	// Generate report based on type
	var report *Report
	var err error
	switch schedule.Type {
	case "user_activity":
		report, err = s.reportingSvc.GenerateUserActivityReport(ctx, options)
	case "location":
		report, err = s.reportingSvc.GenerateLocationReport(ctx, options)
	case "system_health":
		report, err = s.reportingSvc.GenerateSystemHealthReport(ctx, options)
	default:
		fmt.Printf("Unknown report type: %s\n", schedule.Type)
		return
	}

	if err != nil {
		fmt.Printf("Error generating report: %v\n", err)
		return
	}

	// Send report via email
	if err := s.emailSvc.SendReport(schedule.Recipients, schedule, report); err != nil {
		fmt.Printf("Error sending report: %v\n", err)
		return
	}

	// Update schedule
	schedule.LastRun = time.Now()
	schedule.NextRun = schedule.CalculateNextRun()

	query := `
		UPDATE report_schedules
		SET last_run = $1, updated_at = $2
		WHERE id = $3
	`
	_, err = s.db.ExecContext(ctx, query, schedule.LastRun, time.Now(), schedule.ID)
	if err != nil {
		fmt.Printf("Error updating schedule: %v\n", err)
		return
	}

	s.scheduleMutex.Lock()
	s.schedules[schedule.ID] = schedule
	s.scheduleMutex.Unlock()
}

// loadSchedules loads all active schedules from the database
func (s *SchedulerService) loadSchedules(ctx context.Context) error {
	query := `
		SELECT id, tenant_id, name, type, options, frequency,
		       time_of_day, day_of_week, day_of_month, recipients, format,
			   last_run, created_at, updated_at
		FROM report_schedules
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("error loading schedules: %v", err)
	}
	defer rows.Close()

	s.scheduleMutex.Lock()
	defer s.scheduleMutex.Unlock()

	for rows.Next() {
		var schedule models.ReportSchedule
		var recipientsJSON, optionsJSON []byte
		
		err := rows.Scan(
			&schedule.ID, &schedule.TenantID, &schedule.Name, 
			&schedule.Type, &optionsJSON, &schedule.Frequency, &schedule.TimeOfDay,
			&schedule.DayOfWeek, &schedule.DayOfMonth, &recipientsJSON, &schedule.Format,
			&schedule.LastRun, &schedule.CreatedAt, &schedule.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("error scanning schedule: %v", err)
		}

		// Parse JSON fields
		if err := json.Unmarshal(recipientsJSON, &schedule.Recipients); err != nil {
			return fmt.Errorf("error unmarshaling recipients: %v", err)
		}

		if len(optionsJSON) > 0 {
			schedule.Options = string(optionsJSON)
		}

		schedule.NextRun = schedule.CalculateNextRun()
		s.schedules[schedule.ID] = &schedule
	}

	if err = rows.Err(); err != nil {
		return fmt.Errorf("error iterating schedules: %v", err)
	}

	return nil
}

// validateSchedule validates a report schedule
func (s *SchedulerService) validateSchedule(schedule *models.ReportSchedule) error {
	if schedule.Name == "" {
		return fmt.Errorf("schedule name is required")
	}

	if schedule.Type == "" {
		return fmt.Errorf("schedule type is required")
	}

	if schedule.Frequency == "" {
		return fmt.Errorf("schedule frequency is required")
	}

	if schedule.TimeOfDay == "" {
		return fmt.Errorf("schedule time of day is required")
	}

	if len(schedule.Recipients) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}

	if schedule.Format == "" {
		return fmt.Errorf("schedule format is required")
	}

	// Validate frequency-specific fields
	switch schedule.Frequency {
	case "weekly":
		if schedule.DayOfWeek < 0 || schedule.DayOfWeek > 6 {
			return fmt.Errorf("invalid day of week: %d", schedule.DayOfWeek)
		}
	case "monthly":
		if schedule.DayOfMonth < 1 || schedule.DayOfMonth > 31 {
			return fmt.Errorf("invalid day of month: %d", schedule.DayOfMonth)
		}
	}

	return nil
}
