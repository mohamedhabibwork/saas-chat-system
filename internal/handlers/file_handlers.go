package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
)

// FileHandlers handles file-related HTTP requests
type FileHandlers struct {
	storageService *services.StorageService
}

// NewFileHandlers creates a new file handlers instance
func NewFileHandlers(storageService *services.StorageService) *FileHandlers {
	return &FileHandlers{
		storageService: storageService,
	}
}

// HandleUpload handles file upload requests
func (h *FileHandlers) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		sendResponse(w, http.StatusBadRequest, "Failed to parse form", nil)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "Failed to get file", nil)
		return
	}
	defer file.Close()

	// Get user ID from context
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// Upload file
	fileRecord, err := h.storageService.UploadFile(userID, header)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusOK, "File uploaded successfully", fileRecord)
}

// HandleDownload handles file download requests
func (h *FileHandlers) HandleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Get file ID from query parameters
	fileID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid file ID", nil)
		return
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// Get file
	file, reader, err := h.storageService.GetFile(fileID, userID)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	defer reader.Close()

	// Set response headers
	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Disposition", "attachment; filename="+file.Filename)

	// Copy file content to response
	if _, err := io.Copy(w, reader); err != nil {
		sendResponse(w, http.StatusInternalServerError, "Failed to send file", nil)
		return
	}
}

// HandleDelete handles file deletion requests
func (h *FileHandlers) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Get file ID from query parameters
	fileID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid file ID", nil)
		return
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// Delete file
	if err := h.storageService.DeleteFile(fileID, userID); err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusOK, "File deleted successfully", nil)
}

// HandleList handles file listing requests
func (h *FileHandlers) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// List files
	files, err := h.storageService.ListFiles(userID)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusOK, "Files retrieved successfully", files)
}

// Helper functions

func getUserIDFromContext(ctx context.Context) (int, error) {
	// TODO: Implement user ID retrieval from context
	// This should be set by authentication middleware
	return 0, fmt.Errorf("not implemented")
} 