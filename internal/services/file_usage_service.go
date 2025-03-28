package services

import (
	"database/sql"
	"errors"
	"time"

	"saas-chat-system/internal/models"
)

// FileUsageService handles file usage tracking and limits
type FileUsageService struct {
	db *sql.DB
}

// NewFileUsageService creates a new file usage service
func NewFileUsageService(db *sql.DB) *FileUsageService {
	return &FileUsageService{db: db}
}

// GetUserFileUsage retrieves the current file usage for a user
func (s *FileUsageService) GetUserFileUsage(userID int) (*models.FileUsage, error) {
	var usage models.FileUsage
	err := s.db.QueryRow(`
		SELECT 
			COALESCE(SUM(size), 0) as total_size,
			COUNT(*) as total_files,
			COUNT(*) FILTER (WHERE created_at >= NOW() - INTERVAL '24 hours') as files_today
		FROM files
		WHERE user_id = $1
	`, userID).Scan(&usage.TotalSize, &usage.TotalFiles, &usage.FilesToday)

	if err != nil {
		return nil, err
	}

	return &usage, nil
}

// GetSubscriptionLimits retrieves the file usage limits for a user's subscription
func (s *FileUsageService) GetSubscriptionLimits(userID int) (*models.FileLimits, error) {
	var limits models.FileLimits
	err := s.db.QueryRow(`
		SELECT 
			s.plan,
			pl.max_storage,
			pl.max_files,
			pl.max_daily_uploads,
			pl.max_file_size,
			pl.allowed_extensions,
			pl.allowed_mime_types
		FROM subscriptions s
		JOIN subscription_plans pl ON s.plan = pl.name
		WHERE s.user_id = $1 AND s.status = 'active'
		ORDER BY s.created_at DESC
		LIMIT 1
	`, userID).Scan(
		&limits.Plan,
		&limits.MaxStorage,
		&limits.MaxFiles,
		&limits.MaxDailyUploads,
		&limits.MaxFileSize,
		&limits.AllowedExtensions,
		&limits.AllowedMimeTypes,
	)

	if err != nil {
		return nil, err
	}

	return &limits, nil
}

// CheckFileUploadAllowed checks if a file upload is allowed based on user's subscription limits
func (s *FileUsageService) CheckFileUploadAllowed(userID int, fileSize int64, fileType string) error {
	// Get current usage
	usage, err := s.GetUserFileUsage(userID)
	if err != nil {
		return err
	}

	// Get subscription limits
	limits, err := s.GetSubscriptionLimits(userID)
	if err != nil {
		return err
	}

	// Check total storage limit
	if usage.TotalSize+fileSize > limits.MaxStorage {
		return errors.New("storage limit exceeded")
	}

	// Check total files limit
	if usage.TotalFiles >= limits.MaxFiles {
		return errors.New("maximum number of files reached")
	}

	// Check daily upload limit
	if usage.FilesToday >= limits.MaxDailyUploads {
		return errors.New("daily upload limit reached")
	}

	// Check file size limit
	if fileSize > limits.MaxFileSize {
		return errors.New("file size exceeds limit")
	}

	// Check file type
	if !isAllowedFileType(fileType, limits.AllowedExtensions, limits.AllowedMimeTypes) {
		return errors.New("file type not allowed")
	}

	return nil
}

// UpdateFileUsage updates the file usage statistics for a user
func (s *FileUsageService) UpdateFileUsage(userID int, fileSize int64) error {
	_, err := s.db.Exec(`
		UPDATE subscriptions
		SET 
			storage_used = storage_used + $1,
			files_uploaded = files_uploaded + 1,
			daily_uploads = daily_uploads + 1,
			updated_at = NOW()
		WHERE user_id = $2 AND status = 'active'
	`, fileSize, userID)

	return err
}

// ResetDailyUploads resets the daily upload counter for all users
func (s *FileUsageService) ResetDailyUploads() error {
	_, err := s.db.Exec(`
		UPDATE subscriptions
		SET daily_uploads = 0, updated_at = NOW()
		WHERE status = 'active'
	`)

	return err
}

// GetFileUsageHistory retrieves the file usage history for a user
func (s *FileUsageService) GetFileUsageHistory(userID int, startDate, endDate time.Time) ([]models.FileUsageHistory, error) {
	rows, err := s.db.Query(`
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as files_uploaded,
			SUM(size) as total_size
		FROM files
		WHERE user_id = $1 AND created_at BETWEEN $2 AND $3
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`, userID, startDate, endDate)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.FileUsageHistory
	for rows.Next() {
		var h models.FileUsageHistory
		err := rows.Scan(&h.Date, &h.FilesUploaded, &h.TotalSize)
		if err != nil {
			return nil, err
		}
		history = append(history, h)
	}

	return history, nil
}

// GetStorageUsageByType retrieves storage usage breakdown by file type
func (s *FileUsageService) GetStorageUsageByType(userID int) (map[string]int64, error) {
	rows, err := s.db.Query(`
		SELECT 
			mime_type,
			SUM(size) as total_size
		FROM files
		WHERE user_id = $1
		GROUP BY mime_type
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usage := make(map[string]int64)
	for rows.Next() {
		var mimeType string
		var size int64
		err := rows.Scan(&mimeType, &size)
		if err != nil {
			return nil, err
		}
		usage[mimeType] = size
	}

	return usage, nil
}

// isAllowedFileType checks if a file type is allowed based on subscription limits
func isAllowedFileType(fileType string, allowedExtensions, allowedMimeTypes []string) bool {
	// Check file extension
	for _, ext := range allowedExtensions {
		if fileType == ext {
			return true
		}
	}

	// Check MIME type
	for _, mime := range allowedMimeTypes {
		if fileType == mime {
			return true
		}
	}

	return false
}
