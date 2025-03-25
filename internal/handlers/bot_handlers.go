package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
)

// CreateBotRequest represents the bot creation request body
type CreateBotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ModelType   string `json:"model_type"`
	ModelConfig struct {
		APIKey     string   `json:"api_key"`
		ModelName  string   `json:"model_name"`
		MaxTokens  int      `json:"max_tokens"`
		Temperature float64 `json:"temperature"`
		CustomPrompt string `json:"custom_prompt"`
		RestrictedTopics []string `json:"restricted_topics"`
	} `json:"model_config"`
}

// UpdateBotRequest represents the bot update request body
type UpdateBotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ModelConfig struct {
		MaxTokens  int      `json:"max_tokens"`
		Temperature float64 `json:"temperature"`
		CustomPrompt string `json:"custom_prompt"`
		RestrictedTopics []string `json:"restricted_topics"`
	} `json:"model_config"`
}

// HandleListBots handles listing all bots
func HandleListBots(botService *services.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := getUserIDFromContext(r.Context())
		if userID == 0 {
			sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
			return
		}

		bots, err := botService.ListBots(userID)
		if err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, bots, "", http.StatusOK)
	}
}

// HandleCreateBot handles creating a new bot
func HandleCreateBot(botService *services.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := getUserIDFromContext(r.Context())
		if userID == 0 {
			sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req CreateBotRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Name == "" {
			sendResponse(w, false, nil, "Missing bot name", http.StatusBadRequest)
			return
		}
		if req.ModelType == "" {
			sendResponse(w, false, nil, "Missing model type", http.StatusBadRequest)
			return
		}
		if req.ModelConfig.APIKey == "" {
			sendResponse(w, false, nil, "Missing API key", http.StatusBadRequest)
			return
		}
		if req.ModelConfig.ModelName == "" {
			sendResponse(w, false, nil, "Missing model name", http.StatusBadRequest)
			return
		}

		// Create bot
		bot := &models.Bot{
			UserID:      userID,
			Name:        req.Name,
			Description: req.Description,
			ModelType:   req.ModelType,
			ModelConfig: models.ModelConfig{
				APIKey:          req.ModelConfig.APIKey,
				ModelName:       req.ModelConfig.ModelName,
				MaxTokens:       req.ModelConfig.MaxTokens,
				Temperature:     req.ModelConfig.Temperature,
				CustomPrompt:    req.ModelConfig.CustomPrompt,
				RestrictedTopics: req.ModelConfig.RestrictedTopics,
			},
		}

		if err := botService.CreateBot(bot); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, bot, "", http.StatusCreated)
	}
}

// HandleUpdateBot handles updating an existing bot
func HandleUpdateBot(botService *services.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := getUserIDFromContext(r.Context())
		if userID == 0 {
			sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get bot ID from query parameters
		botID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			sendResponse(w, false, nil, "Invalid bot ID", http.StatusBadRequest)
			return
		}

		var req UpdateBotRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.Name == "" {
			sendResponse(w, false, nil, "Missing bot name", http.StatusBadRequest)
			return
		}

		// Update bot
		bot := &models.Bot{
			ID:          botID,
			UserID:      userID,
			Name:        req.Name,
			Description: req.Description,
			ModelConfig: models.ModelConfig{
				MaxTokens:       req.ModelConfig.MaxTokens,
				Temperature:     req.ModelConfig.Temperature,
				CustomPrompt:    req.ModelConfig.CustomPrompt,
				RestrictedTopics: req.ModelConfig.RestrictedTopics,
			},
		}

		if err := botService.UpdateBot(bot); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, bot, "", http.StatusOK)
	}
}

// HandleDeleteBot handles deleting a bot
func HandleDeleteBot(botService *services.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := getUserIDFromContext(r.Context())
		if userID == 0 {
			sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get bot ID from query parameters
		botID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			sendResponse(w, false, nil, "Invalid bot ID", http.StatusBadRequest)
			return
		}

		if err := botService.DeleteBot(botID, userID); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, nil, "", http.StatusOK)
	}
}

// HandleGetBot handles retrieving a bot by ID
func HandleGetBot(botService *services.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := getUserIDFromContext(r.Context())
		if userID == 0 {
			sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get bot ID from query parameters
		botID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			sendResponse(w, false, nil, "Invalid bot ID", http.StatusBadRequest)
			return
		}

		bot, err := botService.GetBot(botID)
		if err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		// Check if user has permission to access this bot
		if !botService.IsUserAllowed(userID, botID) {
			sendResponse(w, false, nil, "Access denied", http.StatusForbidden)
			return
		}

		sendResponse(w, true, bot, "", http.StatusOK)
	}
}

// HandleGetBotStats handles retrieving bot statistics
func HandleGetBotStats(botService *services.BotService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := getUserIDFromContext(r.Context())
		if userID == 0 {
			sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get bot ID from query parameters
		botID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			sendResponse(w, false, nil, "Invalid bot ID", http.StatusBadRequest)
			return
		}

		// Check if user has permission to access this bot
		if !botService.IsUserAllowed(userID, botID) {
			sendResponse(w, false, nil, "Access denied", http.StatusForbidden)
			return
		}

		stats, err := botService.GetBotStats(botID)
		if err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, stats, "", http.StatusOK)
	}
} 