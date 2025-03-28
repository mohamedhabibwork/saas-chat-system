package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"saas-chat-system/internal/models"
)

// TabController handles tab-related operations
type TabController struct {
	*BaseController
}

// NewTabController creates a new tab controller
func NewTabController() *TabController {
	return &TabController{
		BaseController: NewBaseController(),
	}
}

// CreateTab creates a new tab for a user
func (c *TabController) CreateTab(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	var tab struct {
		UserID   int    `json:"user_id"`
		Name     string `json:"name"`
		Type     string `json:"type"`
		TargetID int    `json:"target_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&tab); err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "Invalid request body"))
		return
	}

	// Get user's tenant ID
	var tenantID int
	err := c.db.QueryRow("SELECT tenant_id FROM users WHERE id = $1", tab.UserID).Scan(&tenantID)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_USER", "Invalid user ID"))
		return
	}

	// Validate target based on type
	switch tab.Type {
	case "private":
		if !c.ValidateUser(tab.TargetID, tenantID) {
			c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_TARGET", "Invalid target user ID"))
			return
		}
	case "group":
		var exists bool
		err = c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM groups WHERE id = $1 AND tenant_id = $2)", tab.TargetID, tenantID).Scan(&exists)
		if err != nil || !exists {
			c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_TARGET", "Invalid group ID"))
			return
		}
	case "channel":
		var exists bool
		err = c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM channels WHERE id = $1 AND tenant_id = $2)", tab.TargetID, tenantID).Scan(&exists)
		if err != nil || !exists {
			c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_TARGET", "Invalid channel ID"))
			return
		}
	case "bot":
		var exists bool
		err = c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM bots WHERE id = $1 AND tenant_id = $2)", tab.TargetID, tenantID).Scan(&exists)
		if err != nil || !exists {
			c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_TARGET", "Invalid bot ID"))
			return
		}
	default:
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "INVALID_TYPE", "Invalid tab type"))
		return
	}

	// Get the highest order number for this user
	var maxOrder int
	err = c.db.QueryRow("SELECT COALESCE(MAX(\"order\"), 0) FROM tabs WHERE user_id = $1", tab.UserID).Scan(&maxOrder)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to get tab order"))
		return
	}

	// Insert tab
	var tabID int
	err = c.db.QueryRow(
		"INSERT INTO tabs (user_id, name, type, target_id, \"order\") VALUES ($1, $2, $3, $4, $5) RETURNING id",
		tab.UserID, tab.Name, tab.Type, tab.TargetID, maxOrder+1,
	).Scan(&tabID)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to create tab"))
		return
	}

	resp := map[string]interface{}{
		"id":        tabID,
		"user_id":   tab.UserID,
		"name":      tab.Name,
		"type":      tab.Type,
		"target_id": tab.TargetID,
		"order":     maxOrder + 1,
	}
	c.RespondWithJSON(w, http.StatusCreated, resp)
}

// ListTabs lists all tabs for a user
func (c *TabController) ListTabs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "user_id is required"))
		return
	}

	rows, err := c.db.Query(`
		SELECT id, name, type, target_id, "order", created_at
		FROM tabs
		WHERE user_id = $1
		ORDER BY "order" ASC
	`, userID)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to fetch tabs"))
		return
	}
	defer rows.Close()

	var tabs []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var tabType string
		var targetID int
		var order int
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &tabType, &targetID, &order, &createdAt); err != nil {
			continue
		}
		tabs = append(tabs, map[string]interface{}{
			"id":         id,
			"name":       name,
			"type":       tabType,
			"target_id":  targetID,
			"order":      order,
			"created_at": createdAt,
		})
	}

	c.RespondWithJSON(w, http.StatusOK, tabs)
}

// UpdateTabOrder updates the order of tabs for a user
func (c *TabController) UpdateTabOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	var request struct {
		UserID int   `json:"user_id"`
		Orders []int `json:"orders"` // Array of tab IDs in their new order
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "Invalid request body"))
		return
	}

	// Start transaction
	tx, err := c.db.Begin()
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to start transaction"))
		return
	}

	// Update order for each tab
	for order, tabID := range request.Orders {
		_, err = tx.Exec("UPDATE tabs SET \"order\" = $1 WHERE id = $2 AND user_id = $3", order+1, tabID, request.UserID)
		if err != nil {
			tx.Rollback()
			c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to update tab order"))
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to commit transaction"))
		return
	}

	c.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

// DeleteTab deletes a tab
func (c *TabController) DeleteTab(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		c.RespondWithError(w, c.NewAPIError(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed"))
		return
	}

	tabID := r.URL.Query().Get("id")
	userID := r.URL.Query().Get("user_id")

	if tabID == "" || userID == "" {
		c.RespondWithError(w, c.NewAPIError(http.StatusBadRequest, "BAD_REQUEST", "id and user_id are required"))
		return
	}

	result, err := c.db.Exec("DELETE FROM tabs WHERE id = $1 AND user_id = $2", tabID, userID)
	if err != nil {
		c.RespondWithError(w, c.NewAPIError(http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Failed to delete tab"))
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.RespondWithError(w, c.NewAPIError(http.StatusNotFound, "NOT_FOUND", "Tab not found"))
		return
	}

	c.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "success"})
}
