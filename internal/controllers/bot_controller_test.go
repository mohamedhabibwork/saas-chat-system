package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"saas-chat-system/internal/models"
)

func TestBotController_CreateBot(t *testing.T) {
	// Setup
	controller := NewBotController()
	server := httptest.NewServer(http.HandlerFunc(controller.CreateBot))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedCode   string
	}{
		{
			name: "Valid bot creation",
			payload: map[string]interface{}{
				"name":      "Test Bot",
				"tenant_id": 1,
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid tenant ID",
			payload: map[string]interface{}{
				"name":      "Test Bot",
				"tenant_id": 999,
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_TENANT",
		},
		{
			name: "Missing required fields",
			payload: map[string]interface{}{
				"name": "Test Bot",
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

func TestBotController_ListBots(t *testing.T) {
	// Setup
	controller := NewBotController()
	server := httptest.NewServer(http.HandlerFunc(controller.ListBots))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Valid tenant ID",
			query:          "?tenant_id=1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing tenant ID",
			query:          "",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Invalid tenant ID",
			query:          "?tenant_id=999",
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

func TestBotController_DeleteBot(t *testing.T) {
	// Setup
	controller := NewBotController()
	server := httptest.NewServer(http.HandlerFunc(controller.DeleteBot))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Valid bot ID",
			query:          "?id=1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing bot ID",
			query:          "",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Non-existent bot ID",
			query:          "?id=999",
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
