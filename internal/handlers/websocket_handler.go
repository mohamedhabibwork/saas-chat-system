package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket connections and real-time communication
type WebSocketHandler struct {
	upgrader websocket.Upgrader
	clients  map[int]*Client
	mu       sync.RWMutex
}

// Client represents a WebSocket client connection
type Client struct {
	ID       int
	Conn     *websocket.Conn
	Send     chan []byte
	Channels map[int]bool // Set of channel IDs the client is subscribed to
}

// Message represents a WebSocket message
type Message struct {
	Type      string                 `json:"type"`
	ChannelID int                    `json:"channel_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
		clients: make(map[int]*Client),
	}
}

// @Summary      WebSocket connection
// @Description  Establishes a WebSocket connection for real-time communication
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Param        userId query string true "User ID"
// @Success      101 {string} string "Switching Protocols"
// @Failure      400 {object} APIError "Bad Request"
// @Failure      401 {object} APIError "Unauthorized"
// @Router       /ws [get]
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	client := &Client{
		ID:       userID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Channels: make(map[int]bool),
	}

	h.mu.Lock()
	h.clients[userID] = client
	h.mu.Unlock()

	go h.writePump(client)
	go h.readPump(client)
}

// @Summary      Join channel
// @Description  Join a WebSocket channel for real-time communication
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Param        channel body struct{Channel string} true "Channel to join"
// @Success      200 {object} Message "Joined channel successfully"
// @Failure      400 {object} APIError "Bad Request"
// @Router       /ws/join [post]
func (h *WebSocketHandler) SubscribeToChannel(userID, channelID int) {
	h.mu.RLock()
	client, exists := h.clients[userID]
	h.mu.RUnlock()

	if exists {
		client.Channels[channelID] = true
	}
}

// @Summary      Leave channel
// @Description  Leave a WebSocket channel
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Param        channel body struct{Channel string} true "Channel to leave"
// @Success      200 {object} Message "Left channel successfully"
// @Failure      400 {object} APIError "Bad Request"
// @Router       /ws/leave [post]
func (h *WebSocketHandler) UnsubscribeFromChannel(userID, channelID int) {
	h.mu.RLock()
	client, exists := h.clients[userID]
	h.mu.RUnlock()

	if exists {
		delete(client.Channels, channelID)
	}
}

// BroadcastToChannel sends a message to all clients subscribed to a channel
func (h *WebSocketHandler) BroadcastToChannel(channelID int, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.Channels[channelID] {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(h.clients, client.ID)
				client.Conn.Close()
			}
		}
	}
}

// SendToUser sends a message to a specific user
func (h *WebSocketHandler) SendToUser(userID int, message []byte) {
	h.mu.RLock()
	client, exists := h.clients[userID]
	h.mu.RUnlock()

	if exists {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.clients, userID)
			client.Conn.Close()
		}
	}
}

// @Summary      WebSocket Message Types
// @Description  Supported WebSocket message types and their payloads
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Success      200 {object} Message{type=string,channel=string,userId=string,timestamp=integer,payload=object,encryptedData=string,publicKey=string} "Message structure"
// @Router       /ws/messages [get]

// Message Types:
// @Summary      Chat Message
// @Description  Send a chat message to a channel
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Param        message body Message{type=chat_message,channel=string,payload=object} true "Chat message"
// @Success      200 {object} Message "Message sent successfully"
// @Failure      400 {object} APIError "Bad Request"
// @Router       /ws/messages/chat [post]

// @Summary      Private Message
// @Description  Send a private message to a user
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Param        message body Message{type=private,userId=string,payload=object} true "Private message"
// @Success      200 {object} Message "Message sent successfully"
// @Failure      400 {object} APIError "Bad Request"
// @Router       /ws/messages/private [post]

// @Summary      Key Exchange
// @Description  Exchange encryption keys for secure communication
// @Tags         WebSocket
// @Accept       json
// @Produce      json
// @Param        message body Message{type=key_exchange_response,channel=string,payload=object{channelKey=string}} true "Key exchange"
// @Success      200 {object} Message "Key exchange successful"
// @Failure      400 {object} APIError "Bad Request"
// @Router       /ws/messages/key-exchange [post]

// readPump pumps messages from the WebSocket connection to the hub
func (h *WebSocketHandler) readPump(client *Client) {
	defer func() {
		h.mu.Lock()
		delete(h.clients, client.ID)
		h.mu.Unlock()
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		switch msg.Type {
		case "join_channel":
			if channelID, ok := msg.Data["channel_id"].(float64); ok {
				h.SubscribeToChannel(client.ID, int(channelID))
			}
		case "leave_channel":
			if channelID, ok := msg.Data["channel_id"].(float64); ok {
				h.UnsubscribeFromChannel(client.ID, int(channelID))
			}
		case "chat_message":
			if channelID, ok := msg.Data["channel_id"].(float64); ok {
				h.BroadcastToChannel(int(channelID), message)
			}
		case "webrtc_offer":
			if targetUserID, ok := msg.Data["target_user_id"].(float64); ok {
				h.SendToUser(int(targetUserID), message)
			}
		case "webrtc_answer":
			if targetUserID, ok := msg.Data["target_user_id"].(float64); ok {
				h.SendToUser(int(targetUserID), message)
			}
		case "webrtc_ice_candidate":
			if targetUserID, ok := msg.Data["target_user_id"].(float64); ok {
				h.SendToUser(int(targetUserID), message)
			}
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (h *WebSocketHandler) writePump(client *Client) {
	defer func() {
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
