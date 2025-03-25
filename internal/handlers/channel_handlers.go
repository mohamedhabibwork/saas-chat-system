package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
)

// ChannelHandlers handles channel-related HTTP requests
type ChannelHandlers struct {
	channelService  *services.ChannelService
	webrtcService   *services.WebRTCService
}

// NewChannelHandlers creates a new channel handlers instance
func NewChannelHandlers(channelService *services.ChannelService, webrtcService *services.WebRTCService) *ChannelHandlers {
	return &ChannelHandlers{
		channelService: channelService,
		webrtcService:  webrtcService,
	}
}

// HandleCreateChannel handles channel creation requests
func (h *ChannelHandlers) HandleCreateChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	var channel models.Channel
	if err := json.NewDecoder(r.Body).Decode(&channel); err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	channel.CreatedBy = userID

	if err := h.channelService.CreateChannel(&channel); err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusCreated, "Channel created successfully", channel)
}

// HandleListChannels handles channel listing requests
func (h *ChannelHandlers) HandleListChannels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	channels, err := h.channelService.ListChannels(userID)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusOK, "Channels retrieved successfully", channels)
}

// HandleUpdateChannel handles channel update requests
func (h *ChannelHandlers) HandleUpdateChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid channel ID", nil)
		return
	}

	var settings models.ChannelSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.channelService.UpdateChannel(channelID, settings); err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusOK, "Channel updated successfully", nil)
}

// HandleAddMember handles adding members to a channel
func (h *ChannelHandlers) HandleAddMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid channel ID", nil)
		return
	}

	var request struct {
		UserID int    `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.channelService.AddMember(channelID, request.UserID, request.Role); err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusOK, "Member added successfully", nil)
}

// HandleRemoveMember handles removing members from a channel
func (h *ChannelHandlers) HandleRemoveMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid channel ID", nil)
		return
	}

	userID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	if err := h.channelService.RemoveMember(channelID, userID); err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusOK, "Member removed successfully", nil)
}

// HandleGetMembers handles retrieving channel members
func (h *ChannelHandlers) HandleGetMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid channel ID", nil)
		return
	}

	members, err := h.channelService.GetMembers(channelID)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusOK, "Members retrieved successfully", members)
}

// HandleAddMessage handles adding messages to a channel
func (h *ChannelHandlers) HandleAddMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid channel ID", nil)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var request struct {
		Content  string                 `json:"content"`
		Type     string                 `json:"type"`
		Metadata map[string]interface{} `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.channelService.AddMessage(channelID, userID, request.Content, request.Type, request.Metadata); err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusOK, "Message added successfully", nil)
}

// HandleGetMessages handles retrieving channel messages
func (h *ChannelHandlers) HandleGetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid channel ID", nil)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	messages, err := h.channelService.GetMessages(channelID, limit, offset)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusOK, "Messages retrieved successfully", messages)
}

// HandleWebRTCConnection handles WebRTC connection requests
func (h *ChannelHandlers) HandleWebRTCConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid channel ID", nil)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var request struct {
		StreamType string `json:"stream_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Create peer connection
	peerConnection, err := h.webrtcService.CreatePeerConnection(channelID, userID, request.StreamType)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Create offer
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Set local description
	if err := peerConnection.SetLocalDescription(offer); err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusOK, "WebRTC connection created successfully", offer)
}

// HandleWebRTCAnswer handles WebRTC answer requests
func (h *ChannelHandlers) HandleWebRTCAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid channel ID", nil)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var request struct {
		Answer webrtc.SessionDescription `json:"answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.webrtcService.HandleAnswer(channelID, userID, request.Answer); err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusOK, "WebRTC answer handled successfully", nil)
}

// HandleWebRTCICECandidate handles WebRTC ICE candidate requests
func (h *ChannelHandlers) HandleWebRTCICECandidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid channel ID", nil)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var request struct {
		Candidate webrtc.ICECandidateInit `json:"candidate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.webrtcService.HandleICECandidate(channelID, userID, request.Candidate); err != nil {
		sendResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	sendResponse(w, http.StatusOK, "WebRTC ICE candidate handled successfully", nil)
} 