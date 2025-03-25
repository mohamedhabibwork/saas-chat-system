package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"awesomeProject/internal/models"
	"awesomeProject/internal/websocket"
)

// MessageController handles message-related operations
type MessageController struct {
	BaseController
	hub *websocket.Hub
}

// NewMessageController creates a new message controller
func NewMessageController(hub *websocket.Hub) *MessageController {
	return &MessageController{
		hub: hub,
	}
}

// SendMessage handles sending a new message
func (c *MessageController) SendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var msg models.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate message type
	if !isValidMessageType(msg.Type) {
		c.respondWithError(w, http.StatusBadRequest, "Invalid message type")
		return
	}

	// Store message in database
	query := `
		INSERT INTO messages (type, user_id, content, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`
	err := c.db.QueryRow(query, msg.Type, msg.UserID, msg.Content).Scan(&msg.ID, &msg.CreatedAt)
	if err != nil {
		c.respondWithError(w, http.StatusInternalServerError, "Failed to store message")
		return
	}

	// Broadcast message through WebSocket
	messageJSON, _ := json.Marshal(msg)
	c.hub.Broadcast(messageJSON)

	c.respondWithJSON(w, http.StatusCreated, msg)
}

// GetMessage retrieves a specific message by ID
func (c *MessageController) GetMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	var msg models.Message
	query := `
		SELECT id, type, user_id, content, created_at
		FROM messages
		WHERE id = $1
	`
	err = c.db.QueryRow(query, id).Scan(&msg.ID, &msg.Type, &msg.UserID, &msg.Content, &msg.CreatedAt)
	if err != nil {
		c.respondWithError(w, http.StatusNotFound, "Message not found")
		return
	}

	c.respondWithJSON(w, http.StatusOK, msg)
}

// UpdateMessage updates an existing message
func (c *MessageController) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		c.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	var msg models.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	query := `
		UPDATE messages
		SET content = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, type, user_id, content, created_at, updated_at
	`
	err = c.db.QueryRow(query, msg.Content, id).Scan(
		&msg.ID, &msg.Type, &msg.UserID, &msg.Content, &msg.CreatedAt, &msg.UpdatedAt,
	)
	if err != nil {
		c.respondWithError(w, http.StatusNotFound, "Message not found")
		return
	}

	// Broadcast update through WebSocket
	messageJSON, _ := json.Marshal(msg)
	c.hub.Broadcast(messageJSON)

	c.respondWithJSON(w, http.StatusOK, msg)
}

// DeleteMessage deletes a message
func (c *MessageController) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		c.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		c.respondWithError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	query := `DELETE FROM messages WHERE id = $1`
	result, err := c.db.Exec(query, id)
	if err != nil {
		c.respondWithError(w, http.StatusInternalServerError, "Failed to delete message")
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.respondWithError(w, http.StatusInternalServerError, "Failed to get affected rows")
		return
	}

	if rowsAffected == 0 {
		c.respondWithError(w, http.StatusNotFound, "Message not found")
		return
	}

	// Broadcast deletion through WebSocket
	deleteMsg := map[string]interface{}{
		"type": "delete",
		"id":   id,
	}
	messageJSON, _ := json.Marshal(deleteMsg)
	c.hub.Broadcast(messageJSON)

	c.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Message deleted successfully"})
}

// GetMessageHistory retrieves message history with pagination
func (c *MessageController) GetMessageHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse query parameters
	msgType := r.URL.Query().Get("type")
	userID, _ := strconv.Atoi(r.URL.Query().Get("user_id"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	// Set default values
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Build query
	query := `
		SELECT id, type, user_id, content, created_at
		FROM messages
		WHERE ($1 = '' OR type = $1)
		AND ($2 = 0 OR user_id = $2)
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := c.db.Query(query, msgType, userID, limit, offset)
	if err != nil {
		c.respondWithError(w, http.StatusInternalServerError, "Failed to retrieve messages")
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.ID, &msg.Type, &msg.UserID, &msg.Content, &msg.CreatedAt)
		if err != nil {
			c.respondWithError(w, http.StatusInternalServerError, "Failed to scan message")
			return
		}
		messages = append(messages, msg)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM messages
		WHERE ($1 = '' OR type = $1)
		AND ($2 = 0 OR user_id = $2)
	`
	var total int
	err = c.db.QueryRow(countQuery, msgType, userID).Scan(&total)
	if err != nil {
		c.respondWithError(w, http.StatusInternalServerError, "Failed to get total count")
		return
	}

	response := map[string]interface{}{
		"messages": messages,
		"total":    total,
		"page":     page,
		"limit":    limit,
	}

	c.respondWithJSON(w, http.StatusOK, response)
}

// Helper function to validate message type
func isValidMessageType(msgType string) bool {
	validTypes := map[string]bool{
		"chat":    true,
		"private": true,
		"group":   true,
		"channel": true,
	}
	return validTypes[msgType]
} 