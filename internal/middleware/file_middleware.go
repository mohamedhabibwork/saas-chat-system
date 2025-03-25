package middleware

import (
	"fmt"
	"net/http"

	"awesomeProject/internal/services"
)

// FileUploadMiddleware handles file upload limits based on subscription
type FileUploadMiddleware struct {
	subscriptionService *services.SubscriptionService
}

// NewFileUploadMiddleware creates a new file upload middleware
func NewFileUploadMiddleware(subscriptionService *services.SubscriptionService) *FileUploadMiddleware {
	return &FileUploadMiddleware{
		subscriptionService: subscriptionService,
	}
}

// HandleFileUpload wraps an HTTP handler with file upload limit checks
func (m *FileUploadMiddleware) HandleFileUpload(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only check limits for file uploads
		if r.Method != http.MethodPost || r.URL.Path != "/api/files/upload" {
			next.ServeHTTP(w, r)
			return
		}

		// Get user ID from context
		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get user's subscription
		subscription, err := m.subscriptionService.GetActiveSubscription(userID)
		if err != nil {
			http.Error(w, "Failed to get subscription", http.StatusInternalServerError)
			return
		}

		// Check if user has an active subscription
		if subscription == nil {
			http.Error(w, "No active subscription", http.StatusForbidden)
			return
		}

		// Get file size from request
		contentLength := r.ContentLength
		if contentLength <= 0 {
			http.Error(w, "Invalid file size", http.StatusBadRequest)
			return
		}

		// Convert subscription limits to bytes
		maxFileSize := subscription.Plan.Limits.MaxFileSizeMB * 1024 * 1024

		// Check file size against subscription limit
		if contentLength > maxFileSize {
			http.Error(w, fmt.Sprintf("File size exceeds subscription limit of %d MB", subscription.Plan.Limits.MaxFileSizeMB), http.StatusForbidden)
			return
		}

		// Check storage usage
		usage, err := m.subscriptionService.GetUsage(subscription.ID)
		if err != nil {
			http.Error(w, "Failed to get storage usage", http.StatusInternalServerError)
			return
		}

		// Convert storage limits to bytes
		maxStorage := subscription.Plan.Limits.MaxStorageGB * 1024 * 1024 * 1024
		currentUsage := usage.StorageUsed

		// Check if new file would exceed storage limit
		if currentUsage+contentLength > maxStorage {
			http.Error(w, fmt.Sprintf("Storage limit of %d GB would be exceeded", subscription.Plan.Limits.MaxStorageGB), http.StatusForbidden)
			return
		}

		// All checks passed, proceed with the request
		next.ServeHTTP(w, r)
	}
}

// Helper functions

func getUserIDFromContext(ctx context.Context) (int, error) {
	// TODO: Implement user ID retrieval from context
	// This should be set by authentication middleware
	return 0, fmt.Errorf("not implemented")
} 