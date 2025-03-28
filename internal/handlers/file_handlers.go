package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"saas-chat-system/internal/services"
)

// FileHandlers handles file-related HTTP requests
type FileHandlers struct {
	fileService    *services.FileService
	channelService *services.ChannelService
}

// NewFileHandlers creates a new FileHandlers instance
func NewFileHandlers(fileService *services.FileService, channelService *services.ChannelService) *FileHandlers {
	return &FileHandlers{
		fileService:    fileService,
		channelService: channelService,
	}
}

// @Summary      Upload file
// @Description  Upload a file to a channel
// @Tags         Files
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "File to upload"
// @Param        channel_id formData integer true "Channel ID"
// @Success      200 {object} map[string]interface{} "File uploaded successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Access denied"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/files/upload [post]
func (h *FileHandlers) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		sendResponse(w, false, nil, "Failed to get file from request", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get channel ID from form data
	channelID, err := strconv.Atoi(r.FormValue("channel_id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	// Check if user has access to the channel
	if !h.channelService.HasAccess(channelID, userID) {
		sendResponse(w, false, nil, "Access denied", http.StatusForbidden)
		return
	}

	uploadedFile, err := h.fileService.Upload(file, header, channelID, userID)
	if err != nil {
		sendResponse(w, false, nil, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, uploadedFile, "File uploaded successfully", http.StatusOK)
}

// @Summary      Download a file
// @Description  Download a file from a channel
// @Tags         Files
// @Accept       json
// @Produce      octet-stream
// @Param        file_id query integer true "File ID"
// @Success      200 {file} binary "File content"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Access denied"
// @Failure      404 {object} map[string]interface{} "File not found"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/files/download [get]
func (h *FileHandlers) HandleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fileID, err := strconv.Atoi(r.URL.Query().Get("file_id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid file ID", http.StatusBadRequest)
		return
	}

	// Get the file metadata and check permissions
	file, err := h.fileService.GetFile(fileID)
	if err != nil {
		sendResponse(w, false, nil, "File not found", http.StatusNotFound)
		return
	}

	// Get channel ID from metadata
	channelID, ok := file.Metadata["channel_id"].(int)
	if !ok {
		sendResponse(w, false, nil, "Invalid file metadata", http.StatusInternalServerError)
		return
	}

	// Check if user has access to the channel containing the file
	if !h.channelService.HasAccess(channelID, userID) {
		sendResponse(w, false, nil, "Access denied", http.StatusForbidden)
		return
	}

	// Set appropriate headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file.Filename))
	w.Header().Set("Content-Type", file.MimeType)

	if err := h.fileService.Download(w, fileID); err != nil {
		sendResponse(w, false, nil, "Failed to download file", http.StatusInternalServerError)
		return
	}
}

// @Summary      Delete a file
// @Description  Delete a file from a channel
// @Tags         Files
// @Accept       json
// @Produce      json
// @Param        file_id query integer true "File ID"
// @Success      200 {object} map[string]interface{} "File deleted successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Access denied"
// @Failure      404 {object} map[string]interface{} "File not found"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/files/delete [delete]
func (h *FileHandlers) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	fileID, err := strconv.Atoi(r.URL.Query().Get("file_id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid file ID", http.StatusBadRequest)
		return
	}

	// Get the file metadata and check permissions
	file, err := h.fileService.GetFile(fileID)
	if err != nil {
		sendResponse(w, false, nil, "File not found", http.StatusNotFound)
		return
	}

	// Get channel ID from metadata
	channelID, ok := file.Metadata["channel_id"].(int)
	if !ok {
		sendResponse(w, false, nil, "Invalid file metadata", http.StatusInternalServerError)
		return
	}

	// Check if user has access to the channel containing the file
	if !h.channelService.HasAccess(channelID, userID) {
		sendResponse(w, false, nil, "Access denied", http.StatusForbidden)
		return
	}

	if err := h.fileService.Delete(fileID); err != nil {
		sendResponse(w, false, nil, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, nil, "File deleted successfully", http.StatusOK)
}

// @Summary      List files in a channel
// @Description  Get a list of all files in a specific channel
// @Tags         Files
// @Accept       json
// @Produce      json
// @Param        channel_id query integer true "Channel ID"
// @Success      200 {array} models.File "List of files"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      403 {object} map[string]interface{} "Access denied"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/files/list [get]
func (h *FileHandlers) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	// Check if user has access to the channel
	if !h.channelService.HasAccess(channelID, userID) {
		sendResponse(w, false, nil, "Access denied", http.StatusForbidden)
		return
	}

	files, err := h.fileService.List(channelID)
	if err != nil {
		sendResponse(w, false, nil, "Failed to list files", http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, files, "Files retrieved successfully", http.StatusOK)
} 