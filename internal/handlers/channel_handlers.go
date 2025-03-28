package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"saas-chat-system/internal/models"
	"saas-chat-system/internal/services"
	"github.com/pion/webrtc/v3"
)

// ChannelHandlers handles channel-related HTTP requests
type ChannelHandlers struct {
	channelService *services.ChannelService
	webrtcService  *services.WebRTCService
}

// NewChannelHandlers creates a new channel handlers instance
func NewChannelHandlers(channelService *services.ChannelService, webrtcService *services.WebRTCService) *ChannelHandlers {
	return &ChannelHandlers{
		channelService: channelService,
		webrtcService:  webrtcService,
	}
}

// @Summary      Create channel
// @Description  Create a new channel
// @Tags         Channels
// @Accept       json
// @Produce      json
// @Param        channel body models.Channel true "Channel details"
// @Success      201 {object} models.Channel "Channel created successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels [post]
func (h *ChannelHandlers) HandleCreateChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var channel models.Channel
	if err := json.NewDecoder(r.Body).Decode(&channel); err != nil {
		sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	channel.CreatedBy = userID

	if err := h.channelService.CreateChannel(&channel); err != nil {
		sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, channel, "Channel created successfully", http.StatusCreated)
}

// @Summary      List channels
// @Description  Get a list of channels
// @Tags         Channels
// @Accept       json
// @Produce      json
// @Success      200 {array} models.Channel "List of channels"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels [get]
func (h *ChannelHandlers) HandleListChannels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	channels, err := h.channelService.ListChannels(userID)
	if err != nil {
		sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, channels, "Channels retrieved successfully", http.StatusOK)
}

// @Summary      Update channel
// @Description  Update channel settings
// @Tags         Channels
// @Accept       json
// @Produce      json
// @Param        channel_id query integer true "Channel ID"
// @Param        settings body models.ChannelSettings true "Channel settings"
// @Success      200 {object} map[string]interface{} "Channel updated successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels/{channel_id} [put]
func (h *ChannelHandlers) HandleUpdateChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	var settings models.ChannelSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.channelService.UpdateChannel(channelID, settings); err != nil {
		sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, nil, "Channel updated successfully", http.StatusOK)
}

// @Summary      Add member
// @Description  Add a member to a channel
// @Tags         Channels
// @Accept       json
// @Produce      json
// @Param        channel_id query integer true "Channel ID"
// @Param        member body struct{UserID int,Role string} true "Member details"
// @Success      200 {object} map[string]interface{} "Member added successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels/{channel_id}/members [post]
func (h *ChannelHandlers) HandleAddMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	var request struct {
		UserID int    `json:"user_id"`
		Role   string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.channelService.AddMember(channelID, request.UserID, request.Role); err != nil {
		sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, nil, "Member added successfully", http.StatusOK)
}

// @Summary      Remove member
// @Description  Remove a member from a channel
// @Tags         Channels
// @Accept       json
// @Produce      json
// @Param        channel_id query integer true "Channel ID"
// @Param        user_id query integer true "User ID"
// @Success      200 {object} map[string]interface{} "Member removed successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels/{channel_id}/members/{user_id} [delete]
func (h *ChannelHandlers) HandleRemoveMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := h.channelService.RemoveMember(channelID, userID); err != nil {
		sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, nil, "Member removed successfully", http.StatusOK)
}

// @Summary      Get channel members
// @Description  Get a list of members in a channel
// @Tags         Channels
// @Accept       json
// @Produce      json
// @Param        id query integer true "Channel ID"
// @Success      200 {array} models.User "List of channel members"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels/{id}/members [get]
func (h *ChannelHandlers) HandleGetMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	members, err := h.channelService.GetMembers(channelID)
	if err != nil {
		sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, members, "Members retrieved successfully", http.StatusOK)
}

// @Summary      Add message
// @Description  Add a message to a channel
// @Tags         Channels
// @Accept       json
// @Produce      json
// @Param        channel_id query integer true "Channel ID"
// @Param        message body struct{Content string,Type string,Metadata map[string]interface{}} true "Message details"
// @Success      200 {object} map[string]interface{} "Message added successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels/{channel_id}/messages [post]
func (h *ChannelHandlers) HandleAddMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		Content  string                 `json:"content"`
		Type     string                 `json:"type"`
		Metadata map[string]interface{} `json:"metadata"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.channelService.AddMessage(channelID, userID, request.Content, request.Type, request.Metadata); err != nil {
		sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, nil, "Message added successfully", http.StatusOK)
}

// @Summary      Get messages
// @Description  Get messages from a channel
// @Tags         Channels
// @Accept       json
// @Produce      json
// @Param        channel_id query integer true "Channel ID"
// @Param        limit query integer false "Number of messages to return (default: 50)"
// @Param        offset query integer false "Number of messages to skip (default: 0)"
// @Success      200 {array} models.Message "List of messages"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels/{channel_id}/messages [get]
func (h *ChannelHandlers) HandleGetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	// Get pagination parameters
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50 // Default limit
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	messages, err := h.channelService.GetMessages(channelID, limit, offset)
	if err != nil {
		sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, messages, "Messages retrieved successfully", http.StatusOK)
}

// @Summary      Create WebRTC connection
// @Description  Create a WebRTC connection for a channel
// @Tags         WebRTC
// @Accept       json
// @Produce      json
// @Param        channel_id query integer true "Channel ID"
// @Param        request body struct{StreamType string} true "Stream type"
// @Success      200 {object} webrtc.SessionDescription "WebRTC connection created successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels/{channel_id}/webrtc [post]
func (h *ChannelHandlers) HandleWebRTCConnection(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		StreamType string `json:"stream_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	peerConnection, err := h.webrtcService.CreatePeerConnection(channelID, userID, request.StreamType)
	if err != nil {
		sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	// Set local description
	if err := peerConnection.SetLocalDescription(offer); err != nil {
		sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, offer, "WebRTC connection created successfully", http.StatusOK)
}

// @Summary      Handle WebRTC answer
// @Description  Handle WebRTC answer for a channel
// @Tags         WebRTC
// @Accept       json
// @Produce      json
// @Param        channel_id query integer true "Channel ID"
// @Param        answer body struct{Answer webrtc.SessionDescription} true "WebRTC answer"
// @Success      200 {object} map[string]interface{} "WebRTC answer handled successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels/{channel_id}/webrtc/answer [post]
func (h *ChannelHandlers) HandleWebRTCAnswer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		Answer webrtc.SessionDescription `json:"answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.webrtcService.HandleAnswer(channelID, userID, request.Answer); err != nil {
		sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, nil, "WebRTC answer handled successfully", http.StatusOK)
}

// @Summary      Handle WebRTC ICE candidate
// @Description  Handle WebRTC ICE candidate for a channel
// @Tags         WebRTC
// @Accept       json
// @Produce      json
// @Param        channel_id query integer true "Channel ID"
// @Param        candidate body webrtc.ICECandidateInit true "ICE candidate"
// @Success      200 {object} map[string]interface{} "ICE candidate handled successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels/{channel_id}/webrtc/ice [post]
func (h *ChannelHandlers) HandleWebRTCICECandidate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID, err := strconv.Atoi(r.URL.Query().Get("channel_id"))
	if err != nil {
		sendResponse(w, false, nil, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request struct {
		Candidate webrtc.ICECandidateInit `json:"candidate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.webrtcService.HandleICECandidate(channelID, userID, request.Candidate); err != nil {
		sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, nil, "WebRTC ICE candidate handled successfully", http.StatusOK)
}
