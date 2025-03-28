package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"saas-chat-system/internal/models"
	"saas-chat-system/internal/websocket"
)

// MessageController handles message-related operations
type MessageController struct {
	*BaseController
	hub *websocket.Hub
}

// NewMessageController creates a new message controller
func NewMessageController(hub *websocket.Hub) *MessageController {
	return &MessageController{
		BaseController: NewBaseController(),
		hub: hub,
	}
}

// SendMessage handles sending a new message
func (c *MessageController) SendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	var msg models.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "Invalid request body"))
		return
	}

	// Validate message type
	if !isValidMessageType(msg.Type) {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_TYPE", "Invalid message type"))
		return
	}

	// Store message in database
	query := `
		INSERT INTO messages (type, sender_id, content, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`
	err := c.db.QueryRow(query, msg.Type, msg.SenderID, msg.Content).Scan(&msg.SenderID, &msg.Timestamp)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "DATABASE_ERROR", "Failed to store message"))
		return
	}

	// Broadcast message through WebSocket
	messageJSON, _ := json.Marshal(msg)
	// Convert to the correct Message type for the Hub
	var hubMsg websocket.Message
	hubMsg.Type = msg.Type
	hubMsg.Channel = "global" // Default channel
	hubMsg.Payload = messageJSON
	c.hub.BroadcastMessage(&hubMsg)

	c.RespondWithJSON(w, http.StatusCreated, msg)
}

// GetMessage retrieves a specific message by ID
func (c *MessageController) GetMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_ID", "Invalid message ID"))
		return
	}

	var msg models.Message
	query := `
		SELECT id, type, sender_id, content, timestamp
		FROM messages
		WHERE id = $1
	`
	err = c.db.QueryRow(query, id).Scan(&msg.SenderID, &msg.Type, &msg.SenderID, &msg.Content, &msg.Timestamp)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusNotFound, "NOT_FOUND", "Message not found"))
		return
	}

	c.RespondWithJSON(w, http.StatusOK, msg)
}

// UpdateMessage updates an existing message
func (c *MessageController) UpdateMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_ID", "Invalid message ID"))
		return
	}

	var msg models.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_BODY", "Invalid request body"))
		return
	}

	query := `
		UPDATE messages
		SET content = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, type, sender_id, content, timestamp
	`
	err = c.db.QueryRow(query, msg.Content, id).Scan(
		&msg.SenderID, &msg.Type, &msg.SenderID, &msg.Content, &msg.Timestamp,
	)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusNotFound, "NOT_FOUND", "Message not found"))
		return
	}

	// Broadcast update through WebSocket
	messageJSON, _ := json.Marshal(msg)
	// Convert to the correct Message type for the Hub
	var hubMsg websocket.Message
	hubMsg.Type = "update"
	hubMsg.Channel = "global" // Default channel
	hubMsg.Payload = messageJSON
	c.hub.BroadcastMessage(&hubMsg)

	c.RespondWithJSON(w, http.StatusOK, msg)
}

// DeleteMessage deletes a message
func (c *MessageController) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_ID", "Invalid message ID"))
		return
	}

	query := `DELETE FROM messages WHERE id = $1`
	result, err := c.db.Exec(query, id)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete message"))
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "DATABASE_ERROR", "Failed to get affected rows"))
		return
	}

	if rowsAffected == 0 {
		c.RespondWithError(w, c.NewAPIError(http.StatusNotFound, "NOT_FOUND", "Message not found"))
		return
	}

	// Broadcast deletion through WebSocket
	deleteMsg := map[string]interface{}{
		"type": "delete",
		"id":   id,
	}
	messageJSON, _ := json.Marshal(deleteMsg)
	
	// Convert to the correct Message type for the Hub
	var hubMsg websocket.Message
	hubMsg.Type = "delete"
	hubMsg.Channel = "global" // Default channel
	hubMsg.Payload = messageJSON
	c.hub.BroadcastMessage(&hubMsg)

	c.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Message deleted successfully"})
}

// GetMessageHistory retrieves message history with pagination
func (c *MessageController) GetMessageHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	// Parse query parameters
	msgType := r.URL.Query().Get("type")
	senderID, _ := strconv.Atoi(r.URL.Query().Get("sender_id"))
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
		SELECT id, type, sender_id, content, timestamp
		FROM messages
		WHERE ($1 = '' OR type = $1)
		AND ($2 = 0 OR sender_id = $2)
		ORDER BY timestamp DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := c.db.Query(query, msgType, senderID, limit, offset)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "DATABASE_ERROR", "Failed to retrieve messages"))
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(&msg.SenderID, &msg.Type, &msg.SenderID, &msg.Content, &msg.Timestamp)
		if err != nil {
			c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "DATABASE_ERROR", "Failed to scan message"))
			return
		}
		messages = append(messages, msg)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM messages
		WHERE ($1 = '' OR type = $1)
		AND ($2 = 0 OR sender_id = $2)
	`
	var total int
	err = c.db.QueryRow(countQuery, msgType, senderID).Scan(&total)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "DATABASE_ERROR", "Failed to get total count"))
		return
	}

	response := map[string]interface{}{
		"messages": messages,
		"total":    total,
		"page":     page,
		"limit":    limit,
	}

	c.RespondWithJSON(w, http.StatusOK, response)
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
