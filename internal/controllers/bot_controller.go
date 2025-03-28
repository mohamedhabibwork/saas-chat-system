package controllers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"saas-chat-system/internal/models"
)

// BotController handles bot-related operations
type BotController struct {
	*BaseController
}

// NewBotController creates a new bot controller
func NewBotController() *BotController {
	return &BotController{
		BaseController: NewBaseController(),
	}
}

// CreateBot creates a new bot
func (c *BotController) CreateBot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	var bot struct {
		Name     string `json:"name"`
		TenantID int    `json:"tenant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&bot); err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "Invalid request body"))
		return
	}

	// Validate tenant
	if !c.ValidateTenant(bot.TenantID) {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_TENANT", "Invalid tenant ID"))
		return
	}

	// Generate bot token
	token := generateBotToken()

	// Insert bot
	var botID int
	err := c.db.QueryRow(
		"INSERT INTO bots (name, tenant_id, token) VALUES ($1, $2, $3) RETURNING id",
		bot.Name, bot.TenantID, token,
	).Scan(&botID)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "DUPLICATE_ENTRY", "Bot name already exists in this tenant"))
		return
	}

	resp := map[string]interface{}{
		"id":    botID,
		"name":  bot.Name,
		"token": token,
	}
	c.RespondWithJSON(w, http.StatusCreated, resp)
}

// ListBots lists all bots for a tenant
func (c *BotController) ListBots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "tenant_id is required"))
		return
	}

	rows, err := c.db.Query("SELECT id, name, created_at FROM bots WHERE tenant_id = $1", tenantID)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to fetch bots"))
		return
	}
	defer rows.Close()

	var bots []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &createdAt); err != nil {
			continue
		}
		bots = append(bots, map[string]interface{}{
			"id":         id,
			"name":       name,
			"created_at": createdAt,
		})
	}

	c.RespondWithJSON(w, http.StatusOK, bots)
}

// DeleteBot deletes a bot
func (c *BotController) DeleteBot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	botID := r.URL.Query().Get("id")
	if botID == "" {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "bot ID is required"))
		return
	}

	result, err := c.db.Exec("DELETE FROM bots WHERE id = $1", botID)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to delete bot"))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.RespondWithError(w, c.NewAPIError(http.StatusNotFound, "NOT_FOUND", "Bot not found"))
		return
	}

	c.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// generateBotToken generates a random token for a bot
func generateBotToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
