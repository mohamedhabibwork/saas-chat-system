package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"saas-chat-system/internal/models"
)

// ChannelController handles channel-related operations
type ChannelController struct {
	*BaseController
}

// NewChannelController creates a new channel controller
func NewChannelController() *ChannelController {
	return &ChannelController{
		BaseController: NewBaseController(),
	}
}

// CreateChannel creates a new channel
func (c *ChannelController) CreateChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	var channel struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		TenantID    int    `json:"tenant_id"`
		CreatedBy   int    `json:"created_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&channel); err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "Invalid request body"))
		return
	}

	// Validate tenant and creator
	if !c.ValidateTenant(channel.TenantID) {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_TENANT", "Invalid tenant ID"))
		return
	}

	if !c.ValidateUser(channel.CreatedBy, channel.TenantID) {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_USER", "Invalid user ID"))
		return
	}

	// Insert channel
	var channelID int
	err := c.db.QueryRow(
		"INSERT INTO channels (name, description, tenant_id, created_by) VALUES ($1, $2, $3, $4) RETURNING id",
		channel.Name, channel.Description, channel.TenantID, channel.CreatedBy,
	).Scan(&channelID)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "DUPLICATE_ENTRY", "Channel name already exists in this tenant"))
		return
	}

	// Add creator as admin
	_, err = c.db.Exec(
		"INSERT INTO channel_members (channel_id, user_id, role) VALUES ($1, $2, 'admin')",
		channelID, channel.CreatedBy,
	)
	if err != nil {
		// Log error but don't fail the request
	}

	resp := map[string]interface{}{
		"id":          channelID,
		"name":        channel.Name,
		"description": channel.Description,
		"tenant_id":   channel.TenantID,
		"created_by":  channel.CreatedBy,
	}
	c.RespondWithJSON(w, http.StatusCreated, resp)
}

// ListChannels lists all channels for a tenant
func (c *ChannelController) ListChannels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "tenant_id is required"))
		return
	}

	rows, err := c.db.Query(`
		SELECT c.id, c.name, c.description, c.created_by, c.created_at,
			   u.username as creator_username
		FROM channels c
		JOIN users u ON c.created_by = u.id
		WHERE c.tenant_id = $1
		ORDER BY c.created_at DESC
	`, tenantID)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to fetch channels"))
		return
	}
	defer rows.Close()

	var channels []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var description string
		var createdBy int
		var createdAt time.Time
		var creatorUsername string
		if err := rows.Scan(&id, &name, &description, &createdBy, &createdAt, &creatorUsername); err != nil {
			continue
		}
		channels = append(channels, map[string]interface{}{
			"id":               id,
			"name":             name,
			"description":      description,
			"created_by":       createdBy,
			"creator_username": creatorUsername,
			"created_at":       createdAt,
		})
	}

	c.RespondWithJSON(w, http.StatusOK, channels)
}

// JoinChannel adds a user to a channel
func (c *ChannelController) JoinChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	var membership struct {
		ChannelID int `json:"channel_id"`
		UserID    int `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&membership); err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "Invalid request body"))
		return
	}

	// Get channel tenant ID
	var tenantID int
	err := c.db.QueryRow("SELECT tenant_id FROM channels WHERE id = $1", membership.ChannelID).Scan(&tenantID)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_CHANNEL", "Invalid channel ID"))
		return
	}

	// Validate user
	if !c.ValidateUser(membership.UserID, tenantID) {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_USER", "Invalid user ID"))
		return
	}

	// Add user to channel
	_, err = c.db.Exec(
		"INSERT INTO channel_members (channel_id, user_id, role) VALUES ($1, $2, 'member')",
		membership.ChannelID, membership.UserID,
	)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "DUPLICATE_ENTRY", "User is already a member of this channel"))
		return
	}

	c.RespondWithJSON(w, http.StatusCreated, map[string]string{"status": "success"})
}

// LeaveChannel removes a user from a channel
func (c *ChannelController) LeaveChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	channelID := r.URL.Query().Get("channel_id")
	userID := r.URL.Query().Get("user_id")

	if channelID == "" || userID == "" {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "channel_id and user_id are required"))
		return
	}

	result, err := c.db.Exec(
		"DELETE FROM channel_members WHERE channel_id = $1 AND user_id = $2",
		channelID, userID,
	)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to remove user from channel"))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.RespondWithError(w, c.NewAPIError(http.StatusNotFound, "NOT_FOUND", "User is not a member of this channel"))
		return
	}

	c.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
