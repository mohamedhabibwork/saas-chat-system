package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"awesomeProject/internal/models"
	"awesomeProject/internal/services"
)

// SubscribeRequest represents the subscription request body
type SubscribeRequest struct {
	PlanID         int    `json:"plan_id"`
	PaymentMethod  string `json:"payment_method"`
}

// UsageResponse represents the usage statistics response
type UsageResponse struct {
	MessagesSent int `json:"messages_sent"`
	TokensUsed   int `json:"tokens_used"`
	StorageUsed  int `json:"storage_used"`
	PeriodStart  string `json:"period_start"`
	PeriodEnd    string `json:"period_end"`
}

// HandleListPlans handles listing available subscription plans
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

// HandleSubscribe handles creating a new subscription
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
		userID, err := getUserIDFromContext(r)
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

// HandleCancelSubscription handles cancelling a subscription
func HandleCancelSubscription(subscriptionService *services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get subscription ID from query parameters
		subscriptionID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			sendResponse(w, false, nil, "Invalid subscription ID", http.StatusBadRequest)
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, err := getUserIDFromContext(r)
		if err != nil {
			sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Verify subscription belongs to user
		if err := verifySubscriptionOwnership(subscriptionService, subscriptionID, userID); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusForbidden)
			return
		}

		// Cancel subscription
		if err := subscriptionService.CancelSubscription(subscriptionID); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, nil, "", http.StatusOK)
	}
}

// HandleRenewSubscription handles renewing a subscription
func HandleRenewSubscription(subscriptionService *services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get subscription ID from query parameters
		subscriptionID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			sendResponse(w, false, nil, "Invalid subscription ID", http.StatusBadRequest)
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, err := getUserIDFromContext(r)
		if err != nil {
			sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Verify subscription belongs to user
		if err := verifySubscriptionOwnership(subscriptionService, subscriptionID, userID); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusForbidden)
			return
		}

		// Renew subscription
		if err := subscriptionService.RenewSubscription(subscriptionID); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		sendResponse(w, true, nil, "", http.StatusOK)
	}
}

// HandleGetUsage handles retrieving subscription usage statistics
func HandleGetUsage(subscriptionService *services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			sendResponse(w, false, nil, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get subscription ID from query parameters
		subscriptionID, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			sendResponse(w, false, nil, "Invalid subscription ID", http.StatusBadRequest)
			return
		}

		// Get user ID from context (set by auth middleware)
		userID, err := getUserIDFromContext(r)
		if err != nil {
			sendResponse(w, false, nil, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Verify subscription belongs to user
		if err := verifySubscriptionOwnership(subscriptionService, subscriptionID, userID); err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusForbidden)
			return
		}

		// Get usage statistics
		usage, err := subscriptionService.GetUsage(subscriptionID)
		if err != nil {
			sendResponse(w, false, nil, err.Error(), http.StatusInternalServerError)
			return
		}

		response := UsageResponse{
			MessagesSent: usage.MessagesSent,
			TokensUsed:   usage.TokensUsed,
			StorageUsed:  usage.StorageUsed,
			PeriodStart:  usage.PeriodStart.Format("2006-01-02T15:04:05Z"),
			PeriodEnd:    usage.PeriodEnd.Format("2006-01-02T15:04:05Z"),
		}

		sendResponse(w, true, response, "", http.StatusOK)
	}
}

// Helper functions

func getUserIDFromContext(r *http.Request) (int, error) {
	// TODO: Implement getting user ID from context
	// This should be set by the authentication middleware
	return 0, nil
}

func verifySubscriptionOwnership(subscriptionService *services.SubscriptionService, subscriptionID, userID int) error {
	// TODO: Implement subscription ownership verification
	// This should check if the subscription belongs to the user
	return nil
} 