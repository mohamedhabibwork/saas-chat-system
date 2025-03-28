package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"saas-chat-system/internal/models"
)

func TestTabController_CreateTab(t *testing.T) {
	// Setup
	controller := NewTabController()
	server := httptest.NewServer(http.HandlerFunc(controller.CreateTab))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedCode   string
	}{
		{
			name: "Valid private tab creation",
			payload: map[string]interface{}{
				"user_id":   1,
				"name":      "Private Chat",
				"type":      "private",
				"target_id": 2,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Valid group tab creation",
			payload: map[string]interface{}{
				"user_id":   1,
				"name":      "Group Chat",
				"type":      "group",
				"target_id": 1,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Valid channel tab creation",
			payload: map[string]interface{}{
				"user_id":   1,
				"name":      "Channel Chat",
				"type":      "channel",
				"target_id": 1,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Valid bot tab creation",
			payload: map[string]interface{}{
				"user_id":   1,
				"name":      "Bot Chat",
				"type":      "bot",
				"target_id": 1,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid user ID",
			payload: map[string]interface{}{
				"user_id":   999,
				"name":      "Test Tab",
				"type":      "private",
				"target_id": 1,
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_USER",
		},
		{
			name: "Invalid tab type",
			payload: map[string]interface{}{
				"user_id":   1,
				"name":      "Test Tab",
				"type":      "invalid",
				"target_id": 1,
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_TYPE",
		},
		{
			name: "Invalid target ID",
			payload: map[string]interface{}{
				"user_id":   1,
				"name":      "Test Tab",
				"type":      "private",
				"target_id": 999,
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_TARGET",
		},
		{
			name: "Missing required fields",
			payload: map[string]interface{}{
				"name": "Test Tab",
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			payload, _ := json.Marshal(tt.payload)
			req, err := http.NewRequest("POST", server.URL, bytes.NewBuffer(payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Send request
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, resp.StatusCode)
			}

			// If we expect an error, check the error code
			if tt.expectedCode != "" {
				var errResp models.APIError
				if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
					t.Fatal(err)
				}
				if errResp.Code != tt.expectedCode {
					t.Errorf("expected error code %s; got %s", tt.expectedCode, errResp.Code)
				}
			}
		})
	}
}

func TestTabController_ListTabs(t *testing.T) {
	// Setup
	controller := NewTabController()
	server := httptest.NewServer(http.HandlerFunc(controller.ListTabs))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Valid user ID",
			query:          "?user_id=1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing user ID",
			query:          "",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Invalid user ID",
			query:          "?user_id=999",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest("GET", server.URL+tt.query, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Send request
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, resp.StatusCode)
			}

			// If we expect an error, check the error code
			if tt.expectedCode != "" {
				var errResp models.APIError
				if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
					t.Fatal(err)
				}
				if errResp.Code != tt.expectedCode {
					t.Errorf("expected error code %s; got %s", tt.expectedCode, errResp.Code)
				}
			}
		})
	}
}

func TestTabController_UpdateTabOrder(t *testing.T) {
	// Setup
	controller := NewTabController()
	server := httptest.NewServer(http.HandlerFunc(controller.UpdateTabOrder))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedCode   string
	}{
		{
			name: "Valid order update",
			payload: map[string]interface{}{
				"user_id": 1,
				"orders":  []int{1, 2, 3},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Missing user ID",
			payload: map[string]interface{}{
				"orders": []int{1, 2, 3},
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name: "Empty orders array",
			payload: map[string]interface{}{
				"user_id": 1,
				"orders":  []int{},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			payload, _ := json.Marshal(tt.payload)
			req, err := http.NewRequest("PUT", server.URL, bytes.NewBuffer(payload))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			// Send request
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, resp.StatusCode)
			}

			// If we expect an error, check the error code
			if tt.expectedCode != "" {
				var errResp models.APIError
				if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
					t.Fatal(err)
				}
				if errResp.Code != tt.expectedCode {
					t.Errorf("expected error code %s; got %s", tt.expectedCode, errResp.Code)
				}
			}
		})
	}
}

func TestTabController_DeleteTab(t *testing.T) {
	// Setup
	controller := NewTabController()
	server := httptest.NewServer(http.HandlerFunc(controller.DeleteTab))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Valid tab deletion",
			query:          "?id=1&user_id=1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing tab ID",
			query:          "?user_id=1",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Missing user ID",
			query:          "?id=1",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Non-existent tab",
			query:          "?id=999&user_id=999",
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req, err := http.NewRequest("DELETE", server.URL+tt.query, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Send request
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			// Check status code
			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, resp.StatusCode)
			}

			// If we expect an error, check the error code
			if tt.expectedCode != "" {
				var errResp models.APIError
				if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
					t.Fatal(err)
				}
				if errResp.Code != tt.expectedCode {
					t.Errorf("expected error code %s; got %s", tt.expectedCode, errResp.Code)
				}
			}
		})
	}
}
