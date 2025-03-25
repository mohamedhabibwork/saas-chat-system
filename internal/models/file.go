package models

import "time"

// File represents a file record in the database
type File struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Filename    string    `json:"filename"`
	Filepath    string    `json:"filepath"`
	URL         string    `json:"url"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
}

// FileResponse represents the response structure for file operations
type FileResponse struct {
	ID          int       `json:"id"`
	Filename    string    `json:"filename"`
	URL         string    `json:"url"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	CreatedAt   time.Time `json:"created_at"`
}

// ToResponse converts a File to a FileResponse
func (f *File) ToResponse() *FileResponse {
	return &FileResponse{
		ID:          f.ID,
		Filename:    f.Filename,
		URL:         f.URL,
		Size:        f.Size,
		ContentType: f.ContentType,
		CreatedAt:   f.CreatedAt,
	}
} 