package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"saas-chat-system/internal/services"
)

// CreateBotRequest represents the bot creation request body
type CreateBotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ModelType   string `json:"model_type"`
	ModelConfig struct {
		APIKey           string   `json:"api_key"`
		ModelName        string   `json:"model_name"`
		MaxTokens        int      `json:"max_tokens"`
		Temperature      float64  `json:"temperature"`
		CustomPrompt     string   `json:"custom_prompt"`
		RestrictedTopics []string `json:"restricted_topics"`
	} `json:"model_config"`
}

// UpdateBotRequest represents the bot update request body
type UpdateBotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ModelConfig struct {
		MaxTokens        int      `json:"max_tokens"`
		Temperature      float64  `json:"temperature"`
		CustomPrompt     string   `json:"custom_prompt"`
		RestrictedTopics []string `json:"restricted_topics"`
	} `json:"model_config"`
}

// BotHandlers handles bot-related HTTP requests
type BotHandlers struct {
	botService *services.BotService
}

// NewBotHandlers creates a new BotHandlers instance
func NewBotHandlers(botService *services.BotService) *BotHandlers {
	return &BotHandlers{
		botService: botService,
	}
}

// @Summary      Create a new bot
// @Description  Create a new chat bot
// @Tags         Bots
// @Accept       json
// @Produce      json
// @Param        bot body services.Bot true "Bot details"
// @Success      201 {object} services.Bot "Bot created successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      500 {object} map[string]interface{} "Internal Server Error"
// @Router       /api/v1/bots [post]
func (h *BotHandlers) HandleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var bot services.Bot
	if err := json.NewDecoder(r.Body).Decode(&bot); err != nil {
		SendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get tenant ID from context
	tenantID := r.Context().Value("tenant_id").(string)
	bot.TenantID = tenantID

	err := h.botService.Create(r.Context(), &bot)
	if err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	SendResponse(w, true, bot, "Bot created successfully", http.StatusCreated)
}

// @Summary      Get bot details
// @Description  Get details of a specific bot
// @Tags         Bots
// @Accept       json
// @Produce      json
// @Param        id path string true "Bot ID"
// @Success      200 {object} services.Bot "Bot details"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Bot not found"
// @Router       /api/v1/bots/{id} [get]
func (h *BotHandlers) HandleGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	botID := r.URL.Query().Get("id")
	if botID == "" {
		SendResponse(w, false, nil, "Bot ID is required", http.StatusBadRequest)
		return
	}

	bot, err := h.botService.Get(r.Context(), botID)
	if err != nil {
		SendResponse(w, false, nil, "Bot not found", http.StatusNotFound)
		return
	}

	SendResponse(w, true, bot, "Bot retrieved successfully", http.StatusOK)
}

// @Summary      Update bot details
// @Description  Update an existing bot's configuration
// @Tags         Bots
// @Accept       json
// @Produce      json
// @Param        id path string true "Bot ID"
// @Param        bot body services.Bot true "Updated bot details"
// @Success      200 {object} services.Bot "Bot updated successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Bot not found"
// @Router       /api/v1/bots/{id} [put]
func (h *BotHandlers) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	botID := r.URL.Query().Get("id")
	if botID == "" {
		SendResponse(w, false, nil, "Bot ID is required", http.StatusBadRequest)
		return
	}

	var bot services.Bot
	if err := json.NewDecoder(r.Body).Decode(&bot); err != nil {
		SendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
		return
	}

	bot.ID = botID
	// Get tenant ID from context
	bot.TenantID = r.Context().Value("tenant_id").(string)

	err := h.botService.Update(r.Context(), &bot)
	if err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	SendResponse(w, true, bot, "Bot updated successfully", http.StatusOK)
}

// @Summary      Delete a bot
// @Description  Delete an existing bot
// @Tags         Bots
// @Accept       json
// @Produce      json
// @Param        id path string true "Bot ID"
// @Success      204 "Bot deleted successfully"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Bot not found"
// @Router       /api/v1/bots/{id} [delete]
func (h *BotHandlers) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	botID := r.URL.Query().Get("id")
	if botID == "" {
		SendResponse(w, false, nil, "Bot ID is required", http.StatusBadRequest)
		return
	}

	// Get tenant ID from context
	tenantID := r.Context().Value("tenant_id").(string)

	if err := h.botService.Delete(r.Context(), botID, tenantID); err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	SendResponse(w, true, nil, "Bot deleted successfully", http.StatusNoContent)
}

// @Summary      List all bots
// @Description  Get a list of all bots with optional filtering
// @Tags         Bots
// @Accept       json
// @Produce      json
// @Param        page query integer false "Page number"
// @Param        limit query integer false "Number of items per page"
// @Success      200 {array} services.Bot "List of bots"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Router       /api/v1/bots [get]
func (h *BotHandlers) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 {
		limit = 10
	}

	// Get tenant ID from context
	tenantID := r.Context().Value("tenant_id").(string)

	bots, err := h.botService.List(r.Context(), tenantID, page, limit)
	if err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	SendResponse(w, true, bots, "Bots retrieved successfully", http.StatusOK)
}

// @Summary      Add bot to channel
// @Description  Add a bot to a chat channel
// @Tags         Bots
// @Accept       json
// @Produce      json
// @Param        channel_id path string true "Channel ID"
// @Param        bot_id path string true "Bot ID"
// @Success      200 {object} map[string]interface{} "Bot added to channel successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Channel or bot not found"
// @Router       /api/v1/channels/{channel_id}/bots/{bot_id} [post]
func (h *BotHandlers) HandleAddToChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID := r.URL.Query().Get("channel_id")
	botID := r.URL.Query().Get("bot_id")

	if err := h.botService.AddToChannel(r.Context(), botID, channelID); err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	SendResponse(w, true, nil, "Bot added to channel successfully", http.StatusOK)
}

// @Summary      Remove bot from channel
// @Description  Remove a bot from a chat channel
// @Tags         Bots
// @Accept       json
// @Produce      json
// @Param        channel_id path string true "Channel ID"
// @Param        bot_id path string true "Bot ID"
// @Success      200 {object} map[string]interface{} "Bot removed from channel successfully"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Channel or bot not found"
// @Router       /api/v1/channels/{channel_id}/bots/{bot_id} [delete]
func (h *BotHandlers) HandleRemoveFromChannel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	channelID := r.URL.Query().Get("channel_id")
	botID := r.URL.Query().Get("bot_id")

	if err := h.botService.RemoveFromChannel(r.Context(), botID, channelID); err != nil {
		SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
		return
	}

	SendResponse(w, true, nil, "Bot removed from channel successfully", http.StatusOK)
}

// @Summary      Get bot statistics
// @Description  Get usage and performance statistics for a specific bot
// @Tags         Bots
// @Accept       json
// @Produce      json
// @Param        id query integer true "Bot ID"
// @Success      200 {object} models.BotStats "Bot statistics"
// @Failure      400 {object} map[string]interface{} "Bad Request"
// @Failure      401 {object} map[string]interface{} "Unauthorized"
// @Failure      404 {object} map[string]interface{} "Bot not found"
// @Router       /api/v1/bots/{id}/stats [get]
func HandleGetBotStats(botService *services.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			SendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			SendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get bot ID from query parameters
		botID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			SendResponse(w, false, nil, "Invalid bot ID", http.StatusBadRequest)
			return
		}

		bot, err := botService.GetBot(botID)
		if err != nil {
			SendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		if !botService.IsUserAllowed(bot, userID) {
			SendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// TODO: Implement bot statistics
		SendResponse(w, true, nil, "Bot statistics not implemented", http.StatusNotImplemented)
	}
} 