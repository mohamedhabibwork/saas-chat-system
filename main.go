package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"github.com/your-project/handlers"
	"github.com/your-project/middleware"
	"github.com/your-project/services"
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
	if err != nil {
		return err
	}

	// Create locations table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS locations (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			latitude DOUBLE PRECISION NOT NULL,
			longitude DOUBLE PRECISION NOT NULL,
			accuracy DOUBLE PRECISION,
			altitude DOUBLE PRECISION,
			speed DOUBLE PRECISION,
			heading DOUBLE PRECISION,
			timestamp TIMESTAMP NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create location_history table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS location_history (
			id VARCHAR(36) PRIMARY KEY,
			user_id VARCHAR(36) NOT NULL,
			tenant_id VARCHAR(36) NOT NULL,
			locations JSONB NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		intValue, err := strconv.Atoi(value)
		if err == nil {
			return intValue
		}
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

	// Initialize tracking service
	trackingService := services.NewTrackingService(db)
	trackingHandler := handlers.NewTrackingHandler(trackingService)

	// Register tracking routes
	tracking := r.Group("/api/v1/tracking")
	tracking.Use(middleware.AuthRequired())
	{
		// Event tracking endpoints
		tracking.POST("/events", trackingHandler.TrackEvents)
		tracking.GET("/events", trackingHandler.GetEvents)
		
		// Metric tracking endpoints
		tracking.POST("/metrics", trackingHandler.TrackMetrics)
		tracking.GET("/metrics", trackingHandler.GetMetrics)
		
		// Error tracking endpoints
		tracking.POST("/errors", trackingHandler.TrackErrors)
		tracking.GET("/errors", trackingHandler.GetErrors)
		
		// Statistics endpoint
		tracking.GET("/stats", trackingHandler.GetTrackingStats)
		
		// Maintenance endpoint
		tracking.POST("/cleanup", trackingHandler.CleanupOldData)
	}

	// Initialize location service
	locationService := services.NewLocationService(db)
	locationHandler := handlers.NewLocationHandler(locationService)

	// Register location routes
	location := r.Group("/api/v1/location")
	location.Use(middleware.AuthRequired())
	{
		// Current location endpoints
		location.POST("/current", locationHandler.UpdateLocation)
		location.GET("/current", locationHandler.GetCurrentLocation)
		
		// History endpoints
		location.GET("/history", locationHandler.GetLocationHistory)
		location.POST("/history", locationHandler.SaveLocationHistory)
		
		// Statistics endpoint
		location.GET("/stats", locationHandler.GetLocationStats)
	}

	// Initialize reporting service
	reportingService := services.NewReportingService(db)
	reportingHandler := handlers.NewReportingHandler(reportingService)

	// Register reporting routes
	api.POST("/reports/user-activity", reportingHandler.GenerateUserActivityReport)
	api.POST("/reports/location", reportingHandler.GenerateLocationReport)
	api.POST("/reports/system-health", reportingHandler.GenerateSystemHealthReport)

	// Initialize email service
	emailService := services.NewEmailService(
		getEnv("SMTP_HOST", "smtp.gmail.com"),
		getEnvAsInt("SMTP_PORT", 587),
		getEnv("SMTP_USERNAME", ""),
		getEnv("SMTP_PASSWORD", ""),
		getEnv("SMTP_FROM_EMAIL", "noreply@yourdomain.com"),
		getEnv("SMTP_FROM_NAME", "Chat System Reports"),
	)

	// Initialize Firebase Cloud Messaging client
	fcmClient := services.NewFCMClient(
		getEnv("FIREBASE_PROJECT_ID", ""),
		getEnv("FIREBASE_PRIVATE_KEY", ""),
		getEnv("FIREBASE_CLIENT_EMAIL", ""),
	)

	// Initialize notification service
	notificationService := services.NewNotificationService(emailService, fcmClient)

	// Initialize scheduler service
	schedulerService := services.NewSchedulerService(db, reportingService, emailService)
	schedulerHandler := handlers.NewSchedulerHandler(schedulerService)

	// Start scheduler service
	if err := schedulerService.Start(context.Background()); err != nil {
		log.Fatal("Failed to start scheduler service:", err)
	}
	defer schedulerService.Stop()

	// Register scheduler routes
	scheduler := r.Group("/api/v1/scheduler")
	scheduler.Use(middleware.AuthRequired())
	{
		// Report schedule management endpoints
		scheduler.POST("/schedules", schedulerHandler.CreateSchedule)
		scheduler.GET("/schedules", schedulerHandler.ListSchedules)
		scheduler.GET("/schedules/:id", schedulerHandler.GetSchedule)
		scheduler.PUT("/schedules/:id", schedulerHandler.UpdateSchedule)
		scheduler.DELETE("/schedules/:id", schedulerHandler.DeleteSchedule)
	}

	// Initialize ticket service with notification service
	ticketService := services.NewTicketService(db, notificationService)
	ticketHandler := handlers.NewTicketHandler(ticketService)

	// Register ticket routes
	tickets := r.Group("/api/v1/tickets")
	tickets.Use(middleware.AuthRequired())
	{
		// Ticket management endpoints
		tickets.POST("", ticketHandler.CreateTicket)
		tickets.GET("", ticketHandler.ListTickets)
		tickets.GET("/:id", ticketHandler.GetTicket)
		tickets.PUT("/:id", ticketHandler.UpdateTicket)
		tickets.PUT("/:id/status", ticketHandler.UpdateTicketStatus)

		// Comment endpoints
		tickets.POST("/:id/comments", ticketHandler.AddComment)
		tickets.GET("/:id/comments", ticketHandler.GetComments)

		// Attachment endpoints
		tickets.POST("/:id/attachments", ticketHandler.UploadAttachment)
		tickets.GET("/:id/attachments", ticketHandler.GetAttachments)
		tickets.DELETE("/:id/attachments/:attachment_id", ticketHandler.DeleteAttachment)
	}

	// Start cleanup job for old tracking data
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		for range ticker.C {
			err := trackingHandler.CleanupOldData(context.Background())
			if err != nil {
				log.Println("Error cleaning up old tracking data:", err)
			}
		}
	}()

	// Initialize forum service
	forumService := services.NewForumService(db)
	forumHandler := handlers.NewForumHandler(forumService)

	// Register forum routes
	forum := r.Group("/api/v1/forum")
	{
		// Categories
		forum.POST("/categories", forumHandler.CreateCategory)
		forum.GET("/categories", forumHandler.GetCategories)

		// Topics
		forum.POST("/topics", forumHandler.CreateTopic)
		forum.GET("/categories/:id/topics", forumHandler.GetTopics)
		forum.PUT("/topics/:id", forumHandler.UpdateTopic)
		forum.DELETE("/topics/:id", forumHandler.DeleteTopic)

		// Posts
		forum.POST("/topics/:id/posts", forumHandler.CreatePost)
		forum.GET("/topics/:id/posts", forumHandler.GetPosts)
		forum.PUT("/posts/:id", forumHandler.UpdatePost)
		forum.DELETE("/posts/:id", forumHandler.DeletePost)

		// Subscriptions
		forum.POST("/topics/:id/subscribe", forumHandler.SubscribeToTopic)
		forum.DELETE("/topics/:id/subscribe", forumHandler.UnsubscribeFromTopic)
	}

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
