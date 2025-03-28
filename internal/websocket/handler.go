package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"saas-chat-system/internal/encryption"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 8192
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Add origin validation for production
		return true
	},
}

// Client represents a WebSocket client
type Client struct {
	hub             *Hub
	conn            *websocket.Conn
	send            chan []byte
	userID          string
	encryptionKeys  map[string]string // Map of channel to encryption key
	encryptionMutex sync.RWMutex
}

// Message represents a WebSocket message
type Message struct {
	Type      string          `json:"type"`
	Channel   string          `json:"channel,omitempty"`
	UserID    string          `json:"userId,omitempty"`
	Timestamp int64           `json:"timestamp,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	// For E2E encryption
	EncryptedData string `json:"encryptedData,omitempty"`
	PublicKey     string `json:"publicKey,omitempty"`
}

// Hub maintains WebSocket clients and broadcasts messages
type Hub struct {
	clients           map[*Client]bool
	channels          map[string]map[*Client]bool
	userChannels      map[string]map[string]bool
	register          chan *Client
	unregister        chan *Client
	broadcast         chan *Message
	encryptionService *encryption.Service
	sync.RWMutex
}

// NewHub creates a new hub
func NewHub(encryptionService *encryption.Service) *Hub {
	return &Hub{
		clients:           make(map[*Client]bool),
		channels:          make(map[string]map[*Client]bool),
		userChannels:      make(map[string]map[string]bool),
		register:          make(chan *Client),
		unregister:        make(chan *Client),
		broadcast:         make(chan *Message),
		encryptionService: encryptionService,
	}
}

// RegisterClient registers a client with the hub
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient unregisters a client from the hub
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// BroadcastMessage broadcasts a message to all clients in a channel
func (h *Hub) BroadcastMessage(msg *Message) {
	h.broadcast <- msg
}

// Run runs the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.Lock()
			h.clients[client] = true
			h.Unlock()
		case client := <-h.unregister:
			h.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// Remove client from all channels
				for channel, clients := range h.channels {
					if _, ok := clients[client]; ok {
						delete(clients, client)

						// Send user left notification to channel
						h.sendUserStatusUpdate(client.userID, channel, false)
					}
				}

				// Remove user from user channels
				if channels, ok := h.userChannels[client.userID]; ok {
					delete(h.userChannels, client.userID)

					// Remove user from all channels
					for channel := range channels {
						if clientsInChannel, ok := h.channels[channel]; ok {
							for c := range clientsInChannel {
								if c.userID == client.userID {
									delete(clientsInChannel, c)
								}
							}
						}
					}
				}
			}
			h.Unlock()
		case message := <-h.broadcast:
			h.RLock()
			// Get clients in the channel
			clients, ok := h.channels[message.Channel]
			if !ok {
				h.RUnlock()
				continue
			}

			// Marshal the message to bytes once
			messageBytes, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				h.RUnlock()
				continue
			}

			// Send to all clients in the channel
			for client := range clients {
				// For messages with encrypted data, we don't need to encrypt again
				if message.EncryptedData != "" {
					select {
					case client.send <- messageBytes:
					default:
						h.Unlock()
						h.UnregisterClient(client)
						h.Lock()
					}
					continue
				}

				// For regular messages, encrypt with the client's channel key
				encryptedMsg := *message

				client.encryptionMutex.RLock()
				_, hasKey := client.encryptionKeys[message.Channel]
				client.encryptionMutex.RUnlock()

				// If client has a channel key, encrypt the message
				if hasKey && message.Payload != nil {
					// Encrypt the payload
					encryptedData, err := h.encryptionService.EncryptString(string(message.Payload))
					if err != nil {
						log.Printf("Error encrypting message: %v", err)
						continue
					}

					// Set encrypted data and clear payload
					encryptedMsg.EncryptedData = encryptedData
					encryptedMsg.Payload = nil

					// Marshal the encrypted message
					encryptedBytes, err := json.Marshal(encryptedMsg)
					if err != nil {
						log.Printf("Error marshaling encrypted message: %v", err)
						continue
					}

					select {
					case client.send <- encryptedBytes:
					default:
						h.Unlock()
						h.UnregisterClient(client)
						h.Lock()
					}
				} else {
					// No channel key, send as is (for system messages)
					select {
					case client.send <- messageBytes:
					default:
						h.Unlock()
						h.UnregisterClient(client)
						h.Lock()
					}
				}
			}
			h.RUnlock()
		}
	}
}

// JoinChannel adds a client to a channel
func (h *Hub) JoinChannel(client *Client, channel string) {
	h.Lock()
	defer h.Unlock()

	// Initialize channel if it doesn't exist
	if _, ok := h.channels[channel]; !ok {
		h.channels[channel] = make(map[*Client]bool)
	}

	// Add client to channel
	h.channels[channel][client] = true

	// Initialize user channels if they don't exist
	if _, ok := h.userChannels[client.userID]; !ok {
		h.userChannels[client.userID] = make(map[string]bool)
	}

	// Add channel to user channels
	h.userChannels[client.userID][channel] = true

	// Generate a channel key for the client if they don't have one
	client.encryptionMutex.RLock()
	_, hasKey := client.encryptionKeys[channel]
	client.encryptionMutex.RUnlock()

	if !hasKey {
		// Initialize the user's channel key map
		client.encryptionMutex.Lock()
		if client.encryptionKeys == nil {
			client.encryptionKeys = make(map[string]string)
		}
		
		// Generate and store a random key for this channel
		client.encryptionKeys[channel] = generateRandomKey(32)
		client.encryptionMutex.Unlock()
		
		// Send channel join success message
		joinMsg := &Message{
			Type:      "system",
			Channel:   channel,
			Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
			Payload:   json.RawMessage(`{"action":"joined"}`),
		}

		msgBytes, _ := json.Marshal(joinMsg)
		client.send <- msgBytes
	}

	// Send user joined notification to channel
	h.sendUserStatusUpdate(client.userID, channel, true)
}

// LeaveChannel removes a client from a channel
func (h *Hub) LeaveChannel(client *Client, channel string) {
	h.Lock()
	defer h.Unlock()

	// Remove client from channel
	if clients, ok := h.channels[channel]; ok {
		delete(clients, client)
	}

	// Remove channel from user channels
	if channels, ok := h.userChannels[client.userID]; ok {
		delete(channels, channel)
	}

	// Remove channel key
	client.encryptionMutex.Lock()
	delete(client.encryptionKeys, channel)
	client.encryptionMutex.Unlock()

	// Send user left notification to channel
	h.sendUserStatusUpdate(client.userID, channel, false)
}

// sendUserStatusUpdate sends a user status update to a channel
func (h *Hub) sendUserStatusUpdate(userID, channel string, joined bool) {
	statusType := "user_left"
	if joined {
		statusType = "user_joined"
	}

	statusMsg := &Message{
		Type:      statusType,
		Channel:   channel,
		UserID:    userID,
		Timestamp: time.Now().Unix(),
	}

	h.broadcast <- statusMsg
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(userID int, message []byte) {
	h.RLock()
	defer h.RUnlock()
	
	// Convert int userID to string to match our structure
	userIDStr := fmt.Sprintf("%d", userID)
	
	// Find all clients with this userID
	for client := range h.clients {
		if client.userID == userIDStr {
			select {
			case client.send <- message:
			default:
				go h.UnregisterClient(client)
			}
		}
	}
}

// Handler handles WebSocket connections
func (h *Hub) Handler(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		http.Error(w, "Missing userId parameter", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	client := &Client{
		hub:            h,
		conn:           conn,
		send:           make(chan []byte, 256),
		userID:         userID,
		encryptionKeys: make(map[string]string),
	}

	h.RegisterClient(client)

	// Handle WebSocket connection
	go client.writePump()
	go client.readPump()
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			log.Printf("Error decoding message: %v", err)
			continue
		}

		// Set user ID from the connection
		msg.UserID = c.userID

		// Set timestamp if not set
		if msg.Timestamp == 0 {
			msg.Timestamp = time.Now().Unix()
		}

		// Handle message based on type
		switch msg.Type {
		case "join_channel":
			var payload struct {
				Channel string `json:"channel"`
			}
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("Error parsing join channel payload: %v", err)
				continue
			}
			c.hub.JoinChannel(c, payload.Channel)

		case "leave_channel":
			var payload struct {
				Channel string `json:"channel"`
			}
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("Error parsing leave channel payload: %v", err)
				continue
			}
			c.hub.LeaveChannel(c, payload.Channel)

		case "key_exchange_response":
			var payload struct {
				Channel    string `json:"channel"`
				ChannelKey string `json:"channelKey"`
			}
			if err := json.Unmarshal(msg.Payload, &payload); err != nil {
				log.Printf("Error parsing key exchange payload: %v", err)
				continue
			}

			c.encryptionMutex.Lock()
			c.encryptionKeys[payload.Channel] = payload.ChannelKey
			c.encryptionMutex.Unlock()

		case "chat_message":
			// If message has encrypted data, broadcast as is
			if msg.EncryptedData != "" {
				c.hub.BroadcastMessage(&msg)
				continue
			}

			// Get channel key
			c.encryptionMutex.RLock()
			_, hasKey := c.encryptionKeys[msg.Channel]
			c.encryptionMutex.RUnlock()

			// If client has a channel key, encrypt the message
			if hasKey && msg.Payload != nil {
				// Encrypt the payload
				encryptedData, err := c.hub.encryptionService.EncryptString(string(msg.Payload))
				if err != nil {
					log.Printf("Error encrypting message: %v", err)
					continue
				}

				// Set encrypted data and clear payload
				msg.EncryptedData = encryptedData
				msg.Payload = nil
			}

			// Broadcast the message
			c.hub.BroadcastMessage(&msg)

		default:
			// For other message types, broadcast as is
			c.hub.BroadcastMessage(&msg)
		}
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// generateRandomKey generates a random key of specified length
func generateRandomKey(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
