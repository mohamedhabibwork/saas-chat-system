package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"saas-chat-system/internal/services"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, this should be more restrictive
	},
}

// ChatHandlers handles chat-related HTTP requests
type ChatHandlers struct {
	chatService *services.ChatService
}

// NewChatHandlers creates a new ChatHandlers instance
func NewChatHandlers(chatService *services.ChatService) *ChatHandlers {
	return &ChatHandlers{
		chatService: chatService,
	}
}

// @Summary      Create a new channel
// @Description  Create a new chat channel
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Param        channel body services.Channel true "Channel details"
// @Success      201 {object} services.Channel "Channel created successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/channels [post]
func (h *ChatHandlers) HandleCreateChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var channel services.Channel
	if err := json.NewDecoder(r.Body).Decode(&channel); err != nil {
		SendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.chatService.CreateChannel(r.Context(), &channel)
	if err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	SendResponse(w, true, channel, "Channel created successfully", http.StatusCreated)
}

// @Summary      Join a channel
// @Description  Join a chat channel
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Param        channel_id path string true "Channel ID"
// @Success      200 {object} map[string]interface{} "Joined channel successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Channel not found"
// @Router       /api/v1/channels/{channel_id}/join [post]
func (h *ChatHandlers) HandleJoinChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		SendResponse(w, false, nil, "Channel ID is required", http.StatusBadRequest)
		return
	}

	// Upgrade to WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.chatService.JoinChannel(channelID, conn)
	if err != nil {
		conn.Close()
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}
}

// @Summary      Leave a channel
// @Description  Leave a chat channel
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Param        channel_id path string true "Channel ID"
// @Success      200 {object} map[string]interface{} "Left channel successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Channel not found"
// @Router       /api/v1/channels/{channel_id}/leave [post]
func (h *ChatHandlers) HandleLeaveChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		SendResponse(w, false, nil, "Channel ID is required", http.StatusBadRequest)
		return
	}

	// Get WebSocket connection from context
	conn := r.Context().Value("ws_conn").(*websocket.Conn)

	err := h.chatService.LeaveChannel(channelID, conn)
	if err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	SendResponse(w, true, nil, "Left channel successfully", http.StatusOK)
}

// @Summary      Get channel messages
// @Description  Get messages from a chat channel
// @Tags         Chat
// @Accept       json
// @Produce      json
// @Param        channel_id path string true "Channel ID"
// @Param        page query integer false "Page number"
// @Param        limit query integer false "Number of messages per page"
// @Success      200 {array} services.Message "List of messages"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Channel not found"
// @Router       /api/v1/channels/{channel_id}/messages [get]
func (h *ChatHandlers) HandleGetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID := r.URL.Query().Get("channel_id")
	if channelID == "" {
		SendResponse(w, false, nil, "Channel ID is required", http.StatusBadRequest)
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 50
	}

	messages, err := h.chatService.GetMessages(r.Context(), channelID, page, limit)
	if err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	SendResponse(w, true, messages, "Messages retrieved successfully", http.StatusOK)
} 