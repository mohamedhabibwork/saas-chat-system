package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// JobFunc is a function that is executed by the scheduler
type JobFunc func() error

// SchedulerService provides cron-based task scheduling functionality
type SchedulerService struct {
	cron     *cron.Cron
	jobs     map[string]cron.EntryID
	jobMutex sync.RWMutex
}

// NewSchedulerService creates a new scheduler service
func NewSchedulerService() *SchedulerService {
	return &SchedulerService{
		cron: cron.New(cron.WithSeconds()),
		jobs: make(map[string]cron.EntryID),
	}
}

// Start starts the scheduler
func (s *SchedulerService) Start() {
	s.cron.Start()
}

// Stop stops the scheduler
func (s *SchedulerService) Stop() {
	s.cron.Stop()
}

// AddJob adds a new job to the scheduler
func (s *SchedulerService) AddJob(name, schedule string, job JobFunc) error {
	s.jobMutex.Lock()
	defer s.jobMutex.Unlock()

	// Check if job already exists
	if _, exists := s.jobs[name]; exists {
		return fmt.Errorf("job already exists: %s", name)
	}

	// Add the job with error handling wrapper
	id, err := s.cron.AddFunc(schedule, func() {
		if err := job(); err != nil {
			fmt.Printf("Error executing job %s: %v\n", name, err)
		}
	})

	if err != nil {
		return fmt.Errorf("failed to add job: %v", err)
	}

	s.jobs[name] = id
	return nil
}

// RemoveJob removes a job from the scheduler
func (s *SchedulerService) RemoveJob(name string) error {
	s.jobMutex.Lock()
	defer s.jobMutex.Unlock()

	if id, exists := s.jobs[name]; exists {
		s.cron.Remove(id)
		delete(s.jobs, name)
		return nil
	}

	return fmt.Errorf("job not found: %s", name)
}

// ListJobs returns a list of all scheduled jobs
func (s *SchedulerService) ListJobs() []string {
	s.jobMutex.RLock()
	defer s.jobMutex.RUnlock()

	jobs := make([]string, 0, len(s.jobs))
	for name := range s.jobs {
		jobs = append(jobs, name)
	}

	return jobs
}

// GetNextRun returns the next scheduled run time for a job
func (s *SchedulerService) GetNextRun(name string) (time.Time, error) {
	s.jobMutex.RLock()
	defer s.jobMutex.RUnlock()

	if id, exists := s.jobs[name]; exists {
		entry := s.cron.Entry(id)
		return entry.Next, nil
	}

	return time.Time{}, fmt.Errorf("job not found: %s", name)
}

// Common job schedules
const (
	ScheduleEveryMinute     = "0 * * * * *"
	ScheduleEvery5Minutes   = "0 */5 * * * *"
	ScheduleEvery15Minutes  = "0 */15 * * * *"
	ScheduleEvery30Minutes  = "0 */30 * * * *"
	ScheduleHourly          = "0 0 * * * *"
	ScheduleDaily           = "0 0 0 * * *"
	ScheduleWeekly          = "0 0 0 * * 0"
	ScheduleMonthly         = "0 0 0 1 * *"
	ScheduleQuarterly       = "0 0 0 1 */3 *"
	ScheduleYearly          = "0 0 0 1 1 *"
) 