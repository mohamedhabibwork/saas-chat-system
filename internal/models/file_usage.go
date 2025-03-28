package models

import "time"

// FileUsage represents the current file usage statistics for a user
type FileUsage struct {
	TotalSize   int64 `json:"total_size"`    // Total storage used in bytes
	TotalFiles  int   `json:"total_files"`   // Total number of files
	FilesToday  int   `json:"files_today"`   // Number of files uploaded today
}

// FileLimits represents the file usage limits for a subscription plan
type FileLimits struct {
	Plan              string   `json:"plan"`               // Subscription plan name
	MaxStorage        int64    `json:"max_storage"`        // Maximum storage in bytes
	MaxFiles          int      `json:"max_files"`          // Maximum number of files
	MaxDailyUploads   int      `json:"max_daily_uploads"`  // Maximum daily uploads
	MaxFileSize       int64    `json:"max_file_size"`      // Maximum file size in bytes
	AllowedExtensions []string `json:"allowed_extensions"` // List of allowed file extensions
	AllowedMimeTypes  []string `json:"allowed_mime_types"` // List of allowed MIME types
}

// FileUsageHistory represents historical file usage data
type FileUsageHistory struct {
	Date           time.Time `json:"date"`            // Date of the usage record
	FilesUploaded  int       `json:"files_uploaded"`  // Number of files uploaded
	TotalSize      int64     `json:"total_size"`      // Total storage used
}

// File is defined in model_resolver.go

// FileShare represents a file sharing record
type FileShare struct {
	ID          int       `json:"id"`           // Share record ID
	FileID      int       `json:"file_id"`      // Shared file ID
	SharedBy    int       `json:"shared_by"`    // User who shared the file
	SharedWith  int       `json:"shared_with"`  // User with whom the file is shared
	Permissions string    `json:"permissions"`  // Share permissions (read, write, etc.)
	CreatedAt   time.Time `json:"created_at"`   // Share creation timestamp
	ExpiresAt   time.Time `json:"expires_at"`   // Share expiration timestamp
}

// SubscriptionPlan is defined in model_resolver.go as SubscriptionPlanType

// JSON is a custom type for handling JSON data
type JSON map[string]interface{} 