package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"saas-chat-system/internal/services"
)

// SubscribeRequest represents the subscription request body
type SubscribeRequest struct {
	PlanID        int    `json:"plan_id"`
	PaymentMethod string `json:"payment_method"`
}

// UsageResponse represents the usage statistics response
type UsageResponse struct {
	MessagesSent int    `json:"messages_sent"`
	TokensUsed   int    `json:"tokens_used"`
	StorageUsed  int    `json:"storage_used"`
	PeriodStart  string `json:"period_start"`
	PeriodEnd    string `json:"period_end"`
}

// SubscriptionHandlers handles subscription-related HTTP requests
type SubscriptionHandlers struct {
	subscriptionService *services.SubscriptionService
}

// NewSubscriptionHandlers creates a new subscription handlers instance
func NewSubscriptionHandlers(subscriptionService *services.SubscriptionService) *SubscriptionHandlers {
	return &SubscriptionHandlers{
		subscriptionService: subscriptionService,
	}
}

// @Summary      List subscription plans
// @Description  Get a list of all available subscription plans
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Success      200 {array} models.Plan "List of subscription plans"
// @Failure      500 {object} APIError "Internal Server Error"
// @Router       /api/subscriptions/plans [get]
func HandleListPlans(subscriptionService *services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		plans, err := subscriptionService.ListPlans()
		if err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, plans, "", http.StatusOK)
	}
}

// @Summary      Subscribe to a plan
// @Description  Create a new subscription for the authenticated user
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Param        request body SubscribeRequest true "Subscription details"
// @Success      201 {object} models.Subscription "Subscription created successfully"
// @Failure      400 {object} APIError "Bad Request"
// @Failure      401 {object} APIError "Unauthorized"
// @Failure      500 {object} APIError "Internal Server Error"
// @Router       /api/subscriptions [post]
func HandleSubscribe(subscriptionService *services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SubscribeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			sendResponse(w, false, nil, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if req.PlanID == 0 || req.PaymentMethod == "" {
			sendResponse(w, false, nil, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, err := getUserIDFromContext(r.Context())
		if err != nil {
			sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Create subscription
		subscription, err := subscriptionService.Subscribe(userID, req.PlanID, req.PaymentMethod)
		if err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, subscription, "", http.StatusCreated)
	}
}

// @Summary      Cancel subscription
// @Description  Cancel an active subscription
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription_id query string true "Subscription ID"
// @Success      200 {object} map[string]string "Subscription cancelled successfully"
// @Failure      400 {object} APIError "Bad Request"
// @Failure      401 {object} APIError "Unauthorized"
// @Failure      404 {object} APIError "Subscription not found"
// @Router       /api/subscriptions/cancel [post]
func (h *SubscriptionHandlers) HandleCancelSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	subscriptionID := r.URL.Query().Get("subscription_id")
	if subscriptionID == "" {
		sendResponse(w, false, nil, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	if err := h.subscriptionService.CancelSubscription(subscriptionID); err != nil {
		sendResponse(w, false, nil, fmt.Sprintf("Failed to cancel subscription: %v", err), http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, nil, "Subscription cancelled successfully", http.StatusOK)
}

// @Summary      Renew subscription
// @Description  Renew an expired or cancelled subscription
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Param        subscription_id query string true "Subscription ID"
// @Success      200 {object} map[string]string "Subscription renewed successfully"
// @Failure      400 {object} APIError "Bad Request"
// @Failure      401 {object} APIError "Unauthorized"
// @Failure      404 {object} APIError "Subscription not found"
// @Router       /api/subscriptions/renew [post]
func (h *SubscriptionHandlers) HandleRenewSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	subscriptionID := r.URL.Query().Get("subscription_id")
	if subscriptionID == "" {
		sendResponse(w, false, nil, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	if err := h.subscriptionService.RenewSubscription(subscriptionID); err != nil {
		sendResponse(w, false, nil, fmt.Sprintf("Failed to renew subscription: %v", err), http.StatusInternalServerError)
		return
	}

	sendResponse(w, true, nil, "Subscription renewed successfully", http.StatusOK)
}

// @Summary      Get subscription usage
// @Description  Get usage statistics for the current subscription period
// @Tags         Subscriptions
// @Accept       json
// @Produce      json
// @Success      200 {object} UsageResponse "Subscription usage statistics"
// @Failure      401 {object} APIError "Unauthorized"
// @Failure      404 {object} APIError "Subscription not found"
// @Router       /api/subscriptions/usage [get]
func (h *SubscriptionHandlers) HandleGetUsage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, err := getUserIDFromContext(r.Context())
	if err != nil {
		sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
		return
	}

	subscriptionID := r.URL.Query().Get("subscription_id")
	if subscriptionID == "" {
		sendResponse(w, false, nil, "Invalid subscription ID", http.StatusBadRequest)
		return
	}

	usage, err := h.subscriptionService.GetUsage(subscriptionID)
	if err != nil {
		sendResponse(w, false, nil, fmt.Sprintf("Failed to get usage: %v", err), http.StatusInternalServerError)
		return
	}

	response := struct {
		MessagesSent  int       `json:"messages_sent"`
		TokensUsed    int       `json:"tokens_used"`
		StorageUsed   int64     `json:"storage_used"`
		PeriodStart   string    `json:"period_start"`
		PeriodEnd     string    `json:"period_end"`
	}{
		MessagesSent:  usage.MessagesSent,
		TokensUsed:    usage.TokensUsed,
		StorageUsed:   usage.StorageUsed,
		PeriodStart:   usage.PeriodStart.Format("2006-01-02T15:04:05Z"),
		PeriodEnd:     usage.PeriodEnd.Format("2006-01-02T15:04:05Z"),
	}

	sendResponse(w, true, response, "Usage retrieved successfully", http.StatusOK)
}

// Helper functions

func verifySubscriptionOwnership(subscriptionService *services.SubscriptionService, subscriptionID, userID int) error {
	// TODO: Implement subscription ownership verification
	// This should check if the subscription belongs to the user
	return nil
}
