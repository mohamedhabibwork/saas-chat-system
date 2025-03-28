package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"saas-chat-system/internal/database"
	"saas-chat-system/internal/models"
)

// Upgrader for WebSocket connections
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections (for development)
	},
}

// SimpleClient represents a WebSocket client connection (simplified version)
type SimpleClient struct {
	ID   int
	Conn *websocket.Conn
	Send chan []byte
	Hub  *SimpleHub
	mu   sync.Mutex
}

// SimpleHub maintains the set of active clients and broadcasts messages (simplified version)
type SimpleHub struct {
	clients    map[int]*SimpleClient
	broadcast  chan []byte
	register   chan *SimpleClient
	unregister chan *SimpleClient
	mu         sync.RWMutex
}

// SimpleMessage represents a WebSocket message (simplified version)
type SimpleMessage struct {
	Type    string          `json:"type"`
	UserID  int             `json:"user_id"`
	Content json.RawMessage `json:"content"`
}

// NewSimpleHub creates a new SimpleHub instance
func NewSimpleHub() *SimpleHub {
	return &SimpleHub{
		clients:    make(map[int]*SimpleClient),
		broadcast:  make(chan []byte),
		register:   make(chan *SimpleClient),
		unregister: make(chan *SimpleClient),
	}
}

// Run starts the simple hub
func (h *SimpleHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
			}
			h.mu.Unlock()
			close(client.Send)

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *SimpleHub) Broadcast(message []byte) {
	h.broadcast <- message
}

// SendToUser sends a message to a specific user
func (h *SimpleHub) SendToUser(userID int, message []byte) {
	h.mu.RLock()
	if client, ok := h.clients[userID]; ok {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.clients, userID)
		}
	}
	h.mu.RUnlock()
}

// HandleSimpleWebSocket handles WebSocket connections for the simplified version
func (h *SimpleHub) HandleSimpleWebSocket(conn *websocket.Conn, userID int) {
	client := &SimpleClient{
		ID:   userID,
		Conn: conn,
		Send: make(chan []byte, 256),
		Hub:  h,
	}

	h.register <- client

	// Start goroutines for reading and writing
	go client.simplReadPump()
	go client.simpleWritePump()
}

// simplReadPump pumps messages from the WebSocket connection to the hub
func (c *SimpleClient) simplReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg SimpleMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("error unmarshaling message: %v", err)
			continue
		}

		// Handle different message types
		switch msg.Type {
		case "chat":
			// Broadcast chat message
			c.Hub.broadcast <- message
		case "private":
			// Send private message to specific user
			c.Hub.SendToUser(msg.UserID, message)
		case "group":
			// Handle group message
			c.Hub.broadcast <- message
		case "channel":
			// Handle channel message
			c.Hub.broadcast <- message
		}
	}
}

// simpleWritePump pumps messages from the hub to the WebSocket connection
func (c *SimpleClient) simpleWritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.mu.Lock()
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				c.mu.Unlock()
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.mu.Unlock()
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				c.mu.Unlock()
				return
			}
			c.mu.Unlock()
		}
	}
}

// ReadPump handles reading messages from the websocket connection
func ReadPump(hub *models.Hub, client *models.Client) {
	defer func() {
		hub.Unregister <- client
		_ = client.Conn.(*websocket.Conn).Close()
	}()

	conn := client.Conn.(*websocket.Conn)
	conn.SetReadLimit(1024)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, rawMessage, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error: %v", err)
			}
			break
		}

		// Process message
		var msg models.Message
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			log.Println("Error parsing message:", err)
			continue
		}

		// Set sender details
		msg.SenderID = client.ID
		msg.Sender = client.Username
		msg.TenantID = client.TenantID
		msg.Timestamp = time.Now()

		// Route message based on type
		switch msg.Type {
		case "private":
			savePrivateMessage(client, msg)
			hub.PrivateMessages <- msg

		case "group":
			if _, err := saveGroupMessage(client, msg); err == nil {
				hub.GroupMessages <- msg
			}

		case "topic":
			if _, err := saveTopicMessage(client, msg); err == nil {
				hub.TopicMessages <- msg
			}
		}
	}
}

// WritePump handles writing messages to the websocket connection
func WritePump(client *models.Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		_ = client.Conn.(*websocket.Conn).Close()
	}()

	conn := client.Conn.(*websocket.Conn)

	for {
		select {
		case message, ok := <-client.Send:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			// Add queued messages
			n := len(client.Send)
			for i := 0; i < n; i++ {
				_, _ = w.Write([]byte{'\n'})
				_, _ = w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func savePrivateMessage(client *models.Client, msg models.Message) {
	// Get receiver ID
	var receiverID int
	err := database.DB.QueryRow("SELECT id FROM users WHERE username = $1 AND tenant_id = $2",
		msg.Receiver, client.TenantID).Scan(&receiverID)
	if err != nil {
		log.Println("Error finding receiver:", err)
		return
	}

	msg.ReceiverID = receiverID

	// Insert message
	_, err = database.DB.Exec(
		"INSERT INTO messages (sender_id, receiver_id, tenant_id, content, message_type) VALUES ($1, $2, $3, $4, $5)",
		client.ID, receiverID, client.TenantID, msg.Content, "private")
	if err != nil {
		log.Println("Error saving private message:", err)
	}
}

func saveGroupMessage(client *models.Client, msg models.Message) (int, error) {
	// Verify group exists and get ID
	var groupID int
	err := database.DB.QueryRow("SELECT id FROM groups WHERE name = $1 AND tenant_id = $2",
		msg.GroupName, client.TenantID).Scan(&groupID)
	if err != nil {
		log.Println("Error finding group:", err)
		return 0, err
	}

	// Verify user is a member of this group
	var isMember bool
	err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)",
		groupID, client.ID).Scan(&isMember)
	if err != nil || !isMember {
		log.Println("User not authorized to post to this group")
		return 0, fmt.Errorf("not a group member")
	}

	// Insert message
	_, err = database.DB.Exec(
		"INSERT INTO messages (sender_id, group_id, tenant_id, content, message_type) VALUES ($1, $2, $3, $4, $5)",
		client.ID, groupID, client.TenantID, msg.Content, "group")
	if err != nil {
		log.Println("Error saving group message:", err)
		return 0, err
	}

	return groupID, nil
}

func saveTopicMessage(client *models.Client, msg models.Message) (int, error) {
	// Verify topic exists and get ID
	var topicID int
	err := database.DB.QueryRow("SELECT id FROM topics WHERE name = $1 AND tenant_id = $2",
		msg.TopicName, client.TenantID).Scan(&topicID)
	if err != nil {
		log.Println("Error finding topic:", err)
		return 0, err
	}

	// Insert message
	_, err = database.DB.Exec(
		"INSERT INTO messages (sender_id, topic_id, tenant_id, content, message_type) VALUES ($1, $2, $3, $4, $5)",
		client.ID, topicID, client.TenantID, msg.Content, "notification")
	if err != nil {
		log.Println("Error saving topic message:", err)
		return 0, err
	}

	return topicID, nil
}
