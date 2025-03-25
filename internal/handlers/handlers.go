package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"awesomeProject/internal/models"
	"awesomeProject/internal/database"
	"awesomeProject/internal/websocket"
)

// Error codes
const (
	ErrBadRequest          = "BAD_REQUEST"
	ErrDuplicateEntry      = "DUPLICATE_ENTRY"
	ErrInternalServerError = "INTERNAL_SERVER_ERROR"
)

// NewAPIError creates a new API error
func NewAPIError(status int, code, message string) *models.APIError {
	return &models.APIError{
		Status:  status,
		Code:    code,
		Message: message,
	}
}

// RespondWithError sends an error response
func RespondWithError(w http.ResponseWriter, err *models.APIError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)
	_ = json.NewEncoder(w).Encode(err)
}

// RespondWithJSON sends a JSON response
func RespondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(hub *models.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading connection:", err)
		return
	}

	// Get auth info from query params
	username := r.URL.Query().Get("username")
	tenantID := 0
	userID := 0

	// Fetch user details including tenant
	err = database.DB.QueryRow("SELECT id, tenant_id FROM users WHERE username = $1", username).Scan(&userID, &tenantID)
	if err != nil {
		log.Println("Error authenticating user:", err)
		_ = conn.Close()
		return
	}

	// Create a new client
	client := &models.Client{
		ID:       userID,
		Username: username,
		TenantID: tenantID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Groups:   make(map[int]bool),
		Topics:   make(map[string]bool),
	}

	// Load user's group memberships
	rows, err := database.DB.Query(`
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
	rows, err = database.DB.Query(`
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
	hub.Register <- client

	// Start goroutines for reading and writing
	go websocket.ReadPump(hub, client)
	go websocket.WritePump(client)
}

// HandleTenants handles tenant-related HTTP requests
func HandleTenants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	var tenant struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&tenant); err != nil {
		RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid request body"))
		return
	}

	var tenantID int
	err := database.DB.QueryRow("INSERT INTO tenants (name) VALUES ($1) RETURNING id", tenant.Name).Scan(&tenantID)
	if err != nil {
		RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrDuplicateEntry, "Tenant name already exists"))
		return
	}

	resp := map[string]interface{}{"id": tenantID, "name": tenant.Name}
	RespondWithJSON(w, http.StatusCreated, resp)
}

// HandleRegister handles user registration
func HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RespondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	var user struct {
		Username string `json:"username"`
		Password string `json:"password"`
		TenantID int    `json:"tenant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid request body"))
		return
	}

	var userID int
	err := database.DB.QueryRow("INSERT INTO users (username, password, tenant_id) VALUES ($1, $2, $3) RETURNING id",
		user.Username, user.Password, user.TenantID).Scan(&userID)
	if err != nil {
		RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrDuplicateEntry, "Username already exists in this tenant"))
		return
	}

	resp := map[string]interface{}{"id": userID, "username": user.Username, "tenant_id": user.TenantID}
	RespondWithJSON(w, http.StatusCreated, resp)
}

// HandleGroups handles group-related HTTP requests
func HandleGroups(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var group struct {
			Name     string `json:"name"`
			TenantID int    `json:"tenant_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid request body"))
			return
		}

		var groupID int
		err := database.DB.QueryRow("INSERT INTO groups (name, tenant_id) VALUES ($1, $2) RETURNING id",
			group.Name, group.TenantID).Scan(&groupID)
		if err != nil {
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrDuplicateEntry, "Group creation failed"))
			return
		}

		resp := map[string]interface{}{"id": groupID, "name": group.Name, "tenant_id": group.TenantID}
		RespondWithJSON(w, http.StatusCreated, resp)
	} else if r.Method == http.MethodGet {
		tenantID := r.URL.Query().Get("tenant_id")
		if tenantID == "" {
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "tenant_id is required"))
			return
		}

		rows, err := database.DB.Query("SELECT id, name FROM groups WHERE tenant_id = $1", tenantID)
		if err != nil {
			RespondWithError(w, NewAPIError(http.StatusInternalServerError, ErrInternalServerError, "Failed to fetch groups"))
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

		RespondWithJSON(w, http.StatusOK, groups)
	} else {
		RespondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
	}
}

// HandleGroupMembers handles group membership-related HTTP requests
func HandleGroupMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var membership struct {
			GroupID int `json:"group_id"`
			UserID  int `json:"user_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&membership); err != nil {
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid request body"))
			return
		}

		_, err := database.DB.Exec("INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)",
			membership.GroupID, membership.UserID)
		if err != nil {
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Failed to add member to group"))
			return
		}

		RespondWithJSON(w, http.StatusCreated, map[string]string{"status": "success"})
	} else {
		RespondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
	}
}

// HandleTopics handles topic-related HTTP requests
func HandleTopics(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var topic struct {
			Name     string `json:"name"`
			TenantID int    `json:"tenant_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&topic); err != nil {
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid request body"))
			return
		}

		var topicID int
		err := database.DB.QueryRow("INSERT INTO topics (name, tenant_id) VALUES ($1, $2) RETURNING id",
			topic.Name, topic.TenantID).Scan(&topicID)
		if err != nil {
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrDuplicateEntry, "Topic creation failed"))
			return
		}

		resp := map[string]interface{}{"id": topicID, "name": topic.Name, "tenant_id": topic.TenantID}
		RespondWithJSON(w, http.StatusCreated, resp)
	} else if r.Method == http.MethodGet {
		tenantID := r.URL.Query().Get("tenant_id")
		if tenantID == "" {
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "tenant_id is required"))
			return
		}

		rows, err := database.DB.Query("SELECT id, name FROM topics WHERE tenant_id = $1", tenantID)
		if err != nil {
			RespondWithError(w, NewAPIError(http.StatusInternalServerError, ErrInternalServerError, "Failed to fetch topics"))
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

		RespondWithJSON(w, http.StatusOK, topics)
	} else {
		RespondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
	}
}

// HandleTopicSubscriptions handles topic subscription-related HTTP requests
func HandleTopicSubscriptions(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var subscription struct {
			TopicID int `json:"topic_id"`
			UserID  int `json:"user_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&subscription); err != nil {
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid request body"))
			return
		}

		_, err := database.DB.Exec("INSERT INTO topic_subscriptions (topic_id, user_id) VALUES ($1, $2)",
			subscription.TopicID, subscription.UserID)
		if err != nil {
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Failed to add subscription"))
			return
		}

		RespondWithJSON(w, http.StatusCreated, map[string]string{"status": "success"})
	} else {
		RespondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
	}
}

// HandleMessages handles message-related HTTP requests
func HandleMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondWithError(w, NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
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
		RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "tenant_id, type, and user_id are required"))
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
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "other_user_id is required for private messages"))
			return
		}
		// Get private messages between two users
		rows, err = database.DB.Query(`
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
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "group_id is required for group messages"))
			return
		}
		// Verify user is a member of this group
		var isMember bool
		err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)",
			groupID, userID).Scan(&isMember)
		if err != nil || !isMember {
			RespondWithError(w, NewAPIError(http.StatusForbidden, "FORBIDDEN", "User is not a member of this group"))
			return
		}

		// Get group messages
		rows, err = database.DB.Query(`
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
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "topic_name is required for notifications"))
			return
		}

		// Get topic ID
		var topicID int
		err = database.DB.QueryRow("SELECT id FROM topics WHERE name = $1 AND tenant_id = $2",
			topicName, tenantID).Scan(&topicID)
		if err != nil {
			RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Topic not found"))
			return
		}

		// Verify user is subscribed to this topic
		var isSubscribed bool
		err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM topic_subscriptions WHERE topic_id = $1 AND user_id = $2)",
			topicID, userID).Scan(&isSubscribed)
		if err != nil || !isSubscribed {
			RespondWithError(w, NewAPIError(http.StatusForbidden, "FORBIDDEN", "User is not subscribed to this topic"))
			return
		}

		// Get topic messages
		rows, err = database.DB.Query(`
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
		RespondWithError(w, NewAPIError(http.StatusBadRequest, ErrBadRequest, "Invalid message type"))
		return
	}

	if err != nil {
		log.Println("Database error:", err)
		RespondWithError(w, NewAPIError(http.StatusInternalServerError, ErrInternalServerError, "Failed to fetch messages"))
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

	RespondWithJSON(w, http.StatusOK, messages)
} 