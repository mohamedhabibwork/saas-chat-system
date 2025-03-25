package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Error codes
const (
	ErrBadRequest          = "BAD_REQUEST"
	ErrDuplicateEntry      = "DUPLICATE_ENTRY"
	ErrInternalServerError = "INTERNAL_SERVER_ERROR"
)

// APIError represents an API error response
type APIError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewAPIError creates a new API error
func NewAPIError(status int, code, message string) *APIError {
	return &APIError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, err *APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	_ = json.NewEncoder(w).Encode(err)
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// DB connection
var db *sql.DB

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all connections (for development)
	},
}

// Client represents a connected user
type Client struct {
	ID       int
	Username string
	TenantID int
	Conn     *websocket.Conn
	Send     chan []byte
	Groups   map[int]bool    // Group memberships
	Topics   map[string]bool // Topic subscriptions
}

// Message represents a chat message
type Message struct {
	Type           string                 `json:"type"` // "private", "group", "notification"
	Content        string                 `json:"content"`
	Sender         string                 `json:"sender"`
	SenderID       int                    `json:"sender_id"`
	TenantID       int                    `json:"tenant_id"`
	Receiver       string                 `json:"receiver,omitempty"`
	ReceiverID     int                    `json:"receiver_id,omitempty"`
	GroupID        int                    `json:"group_id,omitempty"`
	GroupName      string                 `json:"group_name,omitempty"`
	TopicName      string                 `json:"topic_name,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	SenderTimezone string                 `json:"sender_timezone,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Hub maintains active clients and broadcasts messages
type Hub struct {
	clients         map[*Client]bool
	tenantClients   map[int]map[*Client]bool // Clients organized by tenant
	broadcast       chan []byte
	register        chan *Client
	unregister      chan *Client
	privateMessages chan Message
	groupMessages   chan Message
	topicMessages   chan Message
}

func newHub() *Hub {
	return &Hub{
		clients:         make(map[*Client]bool),
		tenantClients:   make(map[int]map[*Client]bool),
		broadcast:       make(chan []byte),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		privateMessages: make(chan Message),
		groupMessages:   make(chan Message),
		topicMessages:   make(chan Message),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

			// Register client to tenant map
			if _, ok := h.tenantClients[client.TenantID]; !ok {
				h.tenantClients[client.TenantID] = make(map[*Client]bool)
			}
			h.tenantClients[client.TenantID][client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)

				// Remove from tenant map
				if tenantMap, ok := h.tenantClients[client.TenantID]; ok {
					delete(tenantMap, client)
				}

				close(client.Send)
			}

		case message := <-h.broadcast:
			// Global broadcast (system messages only)
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
					if tenantMap, ok := h.tenantClients[client.TenantID]; ok {
						delete(tenantMap, client)
					}
				}
			}

		case msg := <-h.privateMessages:
			// Find the receiver in the same tenant
			if tenantClients, ok := h.tenantClients[msg.TenantID]; ok {
				for client := range tenantClients {
					if client.ID == msg.ReceiverID || client.ID == msg.SenderID {
						messageJSON, _ := json.Marshal(msg)
						select {
						case client.Send <- messageJSON:
						default:
							close(client.Send)
							delete(h.clients, client)
							delete(tenantClients, client)
						}
					}
				}
			}

		case msg := <-h.groupMessages:
			// Send to all group members in the same tenant
			if tenantClients, ok := h.tenantClients[msg.TenantID]; ok {
				for client := range tenantClients {
					if client.Groups[msg.GroupID] {
						messageJSON, _ := json.Marshal(msg)
						select {
						case client.Send <- messageJSON:
						default:
							close(client.Send)
							delete(h.clients, client)
							delete(tenantClients, client)
						}
					}
				}
			}

		case msg := <-h.topicMessages:
			// Send to all topic subscribers in the same tenant
			if tenantClients, ok := h.tenantClients[msg.TenantID]; ok {
				for client := range tenantClients {
					if client.Topics[msg.TopicName] {
						messageJSON, _ := json.Marshal(msg)
						select {
						case client.Send <- messageJSON:
						default:
							close(client.Send)
							delete(h.clients, client)
							delete(tenantClients, client)
						}
					}
				}
			}
		}
	}
}

func setupDB() (*sql.DB, error) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get DB connection parameters
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "messenger")

	// Connect to PostgreSQL
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	if err := createTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	// Create tenants table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tenants (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create users table with tenant_id
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(50) NOT NULL,
			password VARCHAR(100) NOT NULL,
			tenant_id INTEGER REFERENCES tenants(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(username, tenant_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create groups table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS groups (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			tenant_id INTEGER REFERENCES tenants(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name, tenant_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create group_members table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS group_members (
			group_id INTEGER REFERENCES groups(id),
			user_id INTEGER REFERENCES users(id),
			joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (group_id, user_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create topics table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS topics (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			tenant_id INTEGER REFERENCES tenants(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(name, tenant_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create topic_subscriptions table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS topic_subscriptions (
			topic_id INTEGER REFERENCES topics(id),
			user_id INTEGER REFERENCES users(id),
			subscribed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (topic_id, user_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create messages table with enhanced fields
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			sender_id INTEGER REFERENCES users(id),
			receiver_id INTEGER REFERENCES users(id) NULL,
			group_id INTEGER REFERENCES groups(id) NULL,
			topic_id INTEGER REFERENCES topics(id) NULL,
			tenant_id INTEGER REFERENCES tenants(id),
			content TEXT NOT NULL,
			message_type VARCHAR(20) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}

	// Get auth info from query params
	username := r.URL.Query().Get("username")
	tenantID := 0
	userID := 0

	// Fetch user details including tenant
	err = db.QueryRow("SELECT id, tenant_id FROM users WHERE username = $1", username).Scan(&userID, &tenantID)
	if err != nil {
		log.Println("Error authenticating user:", err)
		_ = conn.Close()
		return
	}

	// Create a new client
	client := &Client{
		ID:       userID,
		Username: username,
		TenantID: tenantID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Groups:   make(map[int]bool),
		Topics:   make(map[string]bool),
	}

	// Load user's group memberships
	rows, err := db.Query(`
		SELECT g.id, g.name 
		FROM groups g 
		JOIN group_members gm ON g.id = gm.group_id 
		WHERE gm.user_id = $1 AND g.tenant_id = $2`,
		userID, tenantID)
	if err == nil {
		defer func(rows *sql.Rows) {
			_ = rows.Close()
		}(rows)
		for rows.Next() {
			var groupID int
			var groupName string
			if err := rows.Scan(&groupID, &groupName); err == nil {
				client.Groups[groupID] = true
			}
		}
	}

	// Load user's topic subscriptions
	rows, err = db.Query(`
		SELECT t.name 
		FROM topics t 
		JOIN topic_subscriptions ts ON t.id = ts.topic_id 
		WHERE ts.user_id = $1 AND t.tenant_id = $2`,
		userID, tenantID)
	if err == nil {
		defer func(rows *sql.Rows) {
			_ = rows.Close()
		}(rows)
		for rows.Next() {
			var topicName string
			if err := rows.Scan(&topicName); err == nil {
				client.Topics[topicName] = true
			}
		}
	}

	// Register client with hub
	hub.register <- client

	// Start goroutines for reading and writing
	go readPump(hub, client)
	go writePump(hub, client)
}

func readPump(hub *Hub, client *Client) {
	defer func() {
		hub.unregister <- client
		_ = client.Conn.Close()
	}()

	client.Conn.SetReadLimit(1024)
	_ = client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		_ = client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, rawMessage, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error: %v", err)
			}
			break
		}

		// Process message
		var msg Message
		if err := json.Unmarshal(rawMessage, &msg); err != nil {
			log.Println("Error parsing message:", err)
			continue
		}

		// Set sender details
		msg.SenderID = client.ID
		msg.Sender = client.Username
		msg.TenantID = client.TenantID
		msg.Timestamp = time.Now()

		// Get user's timezone
		var timezone string
		err = db.QueryRow("SELECT timezone FROM users WHERE id = $1", client.ID).Scan(&timezone)
		if err != nil {
			timezone = "UTC" // Default to UTC if not found
		}
		
		// Set sender timezone
		msg.SenderTimezone = timezone
		
		// Initialize metadata if needed
		if msg.Metadata == nil {
			msg.Metadata = make(map[string]interface{})
		}
		msg.Metadata["sender_timezone"] = timezone

		// Route message based on type
		switch msg.Type {
		case "private":
			savePrivateMessage(client, msg)
			hub.privateMessages <- msg

		case "group":
			if groupID, err := saveGroupMessage(client, msg); err == nil {
				msg.GroupID = groupID
				hub.groupMessages <- msg
			}

		case "notification":
			if _, err := saveTopicMessage(client, msg); err == nil {
				msg.TopicName = msg.TopicName // Already set
				hub.topicMessages <- msg
			}
		}
	}
}

func writePump(hub *Hub, client *Client) {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		_ = client.Conn.Close()
	}()

	// Get recipient's timezone
	var recipientTimezone string
	err := db.QueryRow("SELECT timezone FROM users WHERE id = $1", client.ID).Scan(&recipientTimezone)
	if err != nil {
		recipientTimezone = "UTC" // Default to UTC if not found
	}

	for {
		select {
		case message, ok := <-client.Send:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				_ = client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Format timestamp according to recipient's timezone
			if message.Timestamp != nil {
				loc, err := time.LoadLocation(recipientTimezone)
				if err == nil {
					message.Timestamp = message.Timestamp.In(loc)
				}
			}

			w, err := client.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_ = json.NewEncoder(w).Encode(message)

			n := len(client.Send)
			for i := 0; i < n; i++ {
				_ = w.Write([]byte{'\n'})
				_ = json.NewEncoder(w).Encode(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			_ = client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func savePrivateMessage(client *Client, msg Message) {
	// Get receiver ID
	var receiverID int
	err := db.QueryRow("SELECT id FROM users WHERE username = $1 AND tenant_id = $2",
		msg.Receiver, client.TenantID).Scan(&receiverID)
	if err != nil {
		log.Println("Error finding receiver:", err)
		return
	}

	msg.ReceiverID = receiverID

	// Insert message
	_, err = db.Exec(
		"INSERT INTO messages (sender_id, receiver_id, tenant_id, content, message_type) VALUES ($1, $2, $3, $4, $5)",
		client.ID, receiverID, client.TenantID, msg.Content, "private")
	if err != nil {
		log.Println("Error saving private message:", err)
	}
}

func saveGroupMessage(client *Client, msg Message) (int, error) {
	// Verify group exists and get ID
	var groupID int
	err := db.QueryRow("SELECT id FROM groups WHERE name = $1 AND tenant_id = $2",
		msg.GroupName, client.TenantID).Scan(&groupID)
	if err != nil {
		log.Println("Error finding group:", err)
		return 0, err
	}

	// Verify user is a member of this group
	var isMember bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)",
		groupID, client.ID).Scan(&isMember)
	if err != nil || !isMember {
		log.Println("User not authorized to post to this group")
		return 0, fmt.Errorf("not a group member")
	}

	// Insert message
	_, err = db.Exec(
		"INSERT INTO messages (sender_id, group_id, tenant_id, content, message_type) VALUES ($1, $2, $3, $4, $5)",
		client.ID, groupID, client.TenantID, msg.Content, "group")
	if err != nil {
		log.Println("Error saving group message:", err)
		return 0, err
	}

	return groupID, nil
}

func saveTopicMessage(client *Client, msg Message) (int, error) {
	// Verify topic exists and get ID
	var topicID int
	err := db.QueryRow("SELECT id FROM topics WHERE name = $1 AND tenant_id = $2",
		msg.TopicName, client.TenantID).Scan(&topicID)
	if err != nil {
		log.Println("Error finding topic:", err)
		return 0, err
	}

	// Insert message
	_, err = db.Exec(
		"INSERT INTO messages (sender_id, topic_id, tenant_id, content, message_type) VALUES ($1, $2, $3, $4, $5)",
		client.ID, topicID, client.TenantID, msg.Content, "notification")
	if err != nil {
		log.Println("Error saving topic message:", err)
		return 0, err
	}

	return topicID, nil
}

func main() {
	// Connect to database
	var err error
	db, err = setupDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	log.Println("Connected to PostgresSQL database")

	// Create a new hub
	hub := newHub()
	go hub.run()

	// HTTP routes
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		handleWebSocket(hub, w, r)
	})

	// Tenant registration endpoint
	http.HandleFunc("/tenants", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
			return
		}

		var tenant struct {
			Name string `json:"name"`
		}

		if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
			respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid request body"))
			return
		}

		var tenantID int
		err := db.QueryRow("INSERT INTO tenants (name) VALUES ($1) RETURNING id", tenant.Name).Scan(&tenantID)
		if err != nil {
			respondWithError(w, NewAPIError(http.StatusBadRequest, ErrDuplicateEntry, "Tenant name already exists"))
			return
		}

		resp := map[string]interface{}{"id": tenantID, "name": tenant.Name}
		respondWithJSON(w, http.StatusCreated, resp)
	})

	// User registration endpoint
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
			return
		}

		var user struct {
			Username string `json:"username"`
			Password string `json:"password"`
			TenantID int    `json:"tenant_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid request body"))
			return
		}

		var userID int
		err := db.QueryRow("INSERT INTO users (username, password, tenant_id) VALUES ($1, $2, $3) RETURNING id",
			user.Username, user.Password, user.TenantID).Scan(&userID)
		if err != nil {
			respondWithError(w, NewAPIError(http.StatusBadRequest, ErrDuplicateEntry, "Username already exists in this tenant"))
			return
		}

		resp := map[string]interface{}{"id": userID, "username": user.Username, "tenant_id": user.TenantID}
		respondWithJSON(w, http.StatusCreated, resp)
	})

	// User preferences endpoint
	http.HandleFunc("/user/preferences", func(w http.ResponseWriter, r *http.Request) {
		// Extract user ID from auth token or session
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr == "" {
			respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "user_id is required"))
			return
		}
		
		var userID int
		fmt.Sscanf(userIDStr, "%d", &userID)
		
		if r.Method == http.MethodGet {
			// Get user's timezone from database
			var timezone string
			err := db.QueryRow("SELECT timezone FROM users WHERE id = $1", userID).Scan(&timezone)
			if err != nil {
				timezone = "UTC" // Default to UTC if not found
			}

			// Return user preferences
			prefs := map[string]interface{}{
				"timezone": timezone,
			}

			respondWithJSON(w, http.StatusOK, prefs)
		} else if r.Method == http.MethodPut || r.Method == http.MethodPost {
			// Parse request body
			var prefs struct {
				Timezone string `json:"timezone"`
			}
			
			if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid request body"))
				return
			}

			// Validate timezone
			if prefs.Timezone != "" {
				_, err := time.LoadLocation(prefs.Timezone)
				if err != nil {
					respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid timezone"))
					return
				}
			}

			// Update user's timezone in database
			_, err := db.Exec("UPDATE users SET timezone = $1 WHERE id = $2", 
				prefs.Timezone, userID)
			if err != nil {
				respondWithError(w, NewAPIError(http.StatusInternalServerError, ErrInternalServerError, "Failed to update preferences"))
				return
			}

			// Return updated preferences
			respondWithJSON(w, http.StatusOK, prefs)
		} else {
			respondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		}
	})
	
	// Timezone list endpoint
	http.HandleFunc("/timezones", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
			return
		}
		
		// This is a simplified list of timezones
		// In a real app, you might want to generate this dynamically or include more information
		timezones := []string{
			"UTC",
			"Europe/London",
			"Europe/Paris",
			"Europe/Berlin",
			"America/New_York",
			"America/Chicago",
			"America/Denver",
			"America/Los_Angeles",
			"Asia/Tokyo",
			"Asia/Shanghai",
			"Asia/Kolkata",
			"Australia/Sydney",
			"Pacific/Auckland",
		}

		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"timezones": timezones,
		})
	})

	// Group management endpoint
	http.HandleFunc("/groups", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var group struct {
				Name     string `json:"name"`
				TenantID int    `json:"tenant_id"`
			}

			if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid request body"))
				return
			}

			var groupID int
			err := db.QueryRow("INSERT INTO groups (name, tenant_id) VALUES ($1, $2) RETURNING id",
				group.Name, group.TenantID).Scan(&groupID)
			if err != nil {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrDuplicateEntry, "Group creation failed"))
				return
			}

			resp := map[string]interface{}{"id": groupID, "name": group.Name, "tenant_id": group.TenantID}
			respondWithJSON(w, http.StatusCreated, resp)
		} else if r.Method == http.MethodGet {
			tenantID := r.URL.Query().Get("tenant_id")
			if tenantID == "" {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "tenant_id is required"))
				return
			}

			rows, err := db.Query("SELECT id, name FROM groups WHERE tenant_id = $1", tenantID)
			if err != nil {
				respondWithError(w, NewAPIError(http.StatusInternalServerError, ErrInternalServerError, "Failed to fetch groups"))
				return
			}
			defer func(rows *sql.Rows) {
				_ = rows.Close()
			}(rows)

			var groups []map[string]interface{}
			for rows.Next() {
				var id int
				var name string
				if err := rows.Scan(&id, &name); err != nil {
					continue
				}
				groups = append(groups, map[string]interface{}{"id": id, "name": name})
			}

			respondWithJSON(w, http.StatusOK, groups)
		} else {
			respondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		}
	})

	// Group membership endpoint
	http.HandleFunc("/group-members", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var membership struct {
				GroupID int `json:"group_id"`
				UserID  int `json:"user_id"`
			}

			if err := json.NewDecoder(r.Body).Decode(&membership); err != nil {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid request body"))
				return
			}

			_, err := db.Exec("INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)",
				membership.GroupID, membership.UserID)
			if err != nil {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Failed to add member to group"))
				return
			}

			respondWithJSON(w, http.StatusCreated, map[string]string{"status": "success"})
		} else {
			respondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		}
	})

	// Topic management endpoint
	http.HandleFunc("/topics", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var topic struct {
				Name     string `json:"name"`
				TenantID int    `json:"tenant_id"`
			}

			if err := json.NewDecoder(r.Body).Decode(&topic); err != nil {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid request body"))
				return
			}

			var topicID int
			err := db.QueryRow("INSERT INTO topics (name, tenant_id) VALUES ($1, $2) RETURNING id",
				topic.Name, topic.TenantID).Scan(&topicID)
			if err != nil {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrDuplicateEntry, "Topic creation failed"))
				return
			}

			resp := map[string]interface{}{"id": topicID, "name": topic.Name, "tenant_id": topic.TenantID}
			respondWithJSON(w, http.StatusCreated, resp)
		} else if r.Method == http.MethodGet {
			tenantID := r.URL.Query().Get("tenant_id")
			if tenantID == "" {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "tenant_id is required"))
				return
			}

			rows, err := db.Query("SELECT id, name FROM topics WHERE tenant_id = $1", tenantID)
			if err != nil {
				respondWithError(w, NewAPIError(http.StatusInternalServerError, ErrInternalServerError, "Failed to fetch topics"))
				return
			}
			defer func(rows *sql.Rows) {
				_ = rows.Close()
			}(rows)

			var topics []map[string]interface{}
			for rows.Next() {
				var id int
				var name string
				if err := rows.Scan(&id, &name); err != nil {
					continue
				}
				topics = append(topics, map[string]interface{}{"id": id, "name": name})
			}

			respondWithJSON(w, http.StatusOK, topics)
		} else {
			respondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		}
	})

	// Topic subscription endpoint
	http.HandleFunc("/topic-subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var subscription struct {
				TopicID int `json:"topic_id"`
				UserID  int `json:"user_id"`
			}

			if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid request body"))
				return
			}

			_, err := db.Exec("INSERT INTO topic_subscriptions (topic_id, user_id) VALUES ($1, $2)",
				subscription.TopicID, subscription.UserID)
			if err != nil {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Failed to add subscription"))
				return
			}

			respondWithJSON(w, http.StatusCreated, map[string]string{"status": "success"})
		} else {
			respondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		}
	})

	// Message history endpoint
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			respondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
			return
		}

		// Get query parameters
		tenantID := r.URL.Query().Get("tenant_id")
		messageType := r.URL.Query().Get("type") // "private", "group", "notification"
		userID := r.URL.Query().Get("user_id")
		otherUserID := r.URL.Query().Get("other_user_id") // For private messages
		groupID := r.URL.Query().Get("group_id")          // For group messages
		topicName := r.URL.Query().Get("topic_name")      // For topic messages
		limit := r.URL.Query().Get("limit")
		offset := r.URL.Query().Get("offset")

		if tenantID == "" || messageType == "" || userID == "" {
			respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "tenant_id, type, and user_id are required"))
			return
		}

		// Set defaults for pagination
		if limit == "" {
			limit = "50"
		}
		if offset == "" {
			offset = "0"
		}

		var rows *sql.Rows
		var err error

		switch messageType {
		case "private":
			if otherUserID == "" {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "other_user_id is required for private messages"))
				return
			}
			// Get private messages between two users
			rows, err = db.Query(`
                SELECT m.id, m.content, m.created_at, 
                       u1.username as sender_username, m.sender_id,
                       u2.username as receiver_username, m.receiver_id
                FROM messages m
                JOIN users u1 ON m.sender_id = u1.id
                JOIN users u2 ON m.receiver_id = u2.id
                WHERE m.tenant_id = $1 
                  AND m.message_type = 'private'
                  AND ((m.sender_id = $2 AND m.receiver_id = $3) OR (m.sender_id = $3 AND m.receiver_id = $2))
                ORDER BY m.created_at DESC
                LIMIT $4 OFFSET $5
            `, tenantID, userID, otherUserID, limit, offset)

		case "group":
			if groupID == "" {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "group_id is required for group messages"))
				return
			}
			// Verify user is a member of this group
			var isMember bool
			err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)",
				groupID, userID).Scan(&isMember)
			if err != nil || !isMember {
				respondWithError(w, NewAPIError(http.StatusForbidden, "FORBIDDEN", "User is not a member of this group"))
				return
			}

			// Get group messages
			rows, err = db.Query(`
                SELECT m.id, m.content, m.created_at, 
                       u.username as sender_username, m.sender_id,
                       g.name as group_name, g.id as group_id
                FROM messages m
                JOIN users u ON m.sender_id = u.id
                JOIN groups g ON m.group_id = g.id
                WHERE m.tenant_id = $1 
                  AND m.message_type = 'group'
                  AND m.group_id = $2
                ORDER BY m.created_at DESC
                LIMIT $3 OFFSET $4
            `, tenantID, groupID, limit, offset)

		case "notification":
			if topicName == "" {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "topic_name is required for notifications"))
				return
			}

			// Get topic ID
			var topicID int
			err = db.QueryRow("SELECT id FROM topics WHERE name = $1 AND tenant_id = $2",
				topicName, tenantID).Scan(&topicID)
			if err != nil {
				respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Topic not found"))
				return
			}

			// Verify user is subscribed to this topic
			var isSubscribed bool
			err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM topic_subscriptions WHERE topic_id = $1 AND user_id = $2)",
				topicID, userID).Scan(&isSubscribed)
			if err != nil || !isSubscribed {
				respondWithError(w, NewAPIError(http.StatusForbidden, "FORBIDDEN", "User is not subscribed to this topic"))
				return
			}

			// Get topic messages
			rows, err = db.Query(`
                SELECT m.id, m.content, m.created_at, 
                       u.username as sender_username, m.sender_id,
                       t.name as topic_name
                FROM messages m
                JOIN users u ON m.sender_id = u.id
                JOIN topics t ON m.topic_id = t.id
                WHERE m.tenant_id = $1 
                  AND m.message_type = 'notification'
                  AND t.name = $2
                ORDER BY m.created_at DESC
                LIMIT $3 OFFSET $4
            `, tenantID, topicName, limit, offset)

		default:
			respondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid message type"))
			return
		}

		if err != nil {
			log.Println("Database error:", err)
			respondWithError(w, NewAPIError(http.StatusInternalServerError, ErrInternalServerError, "Failed to fetch messages"))
			return
		}
		defer rows.Close()

		var messages []map[string]interface{}
		for rows.Next() {
			switch messageType {
			case "private":
				var id int
				var content string
				var createdAt time.Time
				var senderUsername string
				var senderID int
				var receiverUsername string
				var receiverID int

				if err := rows.Scan(&id, &content, &createdAt, &senderUsername, &senderID, &receiverUsername, &receiverID); err != nil {
					continue
				}

				messages = append(messages, map[string]interface{}{
					"id":          id,
					"content":     content,
					"timestamp":   createdAt,
					"sender":      senderUsername,
					"sender_id":   senderID,
					"receiver":    receiverUsername,
					"receiver_id": receiverID,
					"type":        "private",
				})

			case "group":
				var id int
				var content string
				var createdAt time.Time
				var senderUsername string
				var senderID int
				var groupName string
				var groupID int

				if err := rows.Scan(&id, &content, &createdAt, &senderUsername, &senderID, &groupName, &groupID); err != nil {
					continue
				}

				messages = append(messages, map[string]interface{}{
					"id":         id,
					"content":    content,
					"timestamp":  createdAt,
					"sender":     senderUsername,
					"sender_id":  senderID,
					"group_name": groupName,
					"group_id":   groupID,
					"type":       "group",
				})

			case "notification":
				var id int
				var content string
				var createdAt time.Time
				var senderUsername string
				var senderID int
				var topicName string

				if err := rows.Scan(&id, &content, &createdAt, &senderUsername, &senderID, &topicName); err != nil {
					continue
				}

				messages = append(messages, map[string]interface{}{
					"id":         id,
					"content":    content,
					"timestamp":  createdAt,
					"sender":     senderUsername,
					"sender_id":  senderID,
					"topic_name": topicName,
					"type":       "notification",
				})
			}
		}

		respondWithJSON(w, http.StatusOK, messages)
	})

	// Start the server
	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr: ":" + port,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	log.Println("Server shutting down...")
}
