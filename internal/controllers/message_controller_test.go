package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"saas-chat-system/internal/models"
	"saas-chat-system/internal/websocket"
)

func TestMessageController_SendMessage(t *testing.T) {
	// Setup
	hub := websocket.NewHub()
	controller := NewMessageController(hub)
	server := httptest.NewServer(http.HandlerFunc(controller.SendMessage))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
		expectedCode   string
	}{
		{
			name: "Valid chat message",
			payload: map[string]interface{}{
				"type":    "chat",
				"user_id": 1,
				"content": map[string]interface{}{
					"text": "Hello, world!",
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Valid private message",
			payload: map[string]interface{}{
				"type":    "private",
				"user_id": 1,
				"content": map[string]interface{}{
					"text": "Private message",
				},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid message type",
			payload: map[string]interface{}{
				"type":    "invalid",
				"user_id": 1,
				"content": map[string]interface{}{
					"text": "Invalid type",
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name: "Missing required fields",
			payload: map[string]interface{}{
				"type": "chat",
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

func TestMessageController_GetMessage(t *testing.T) {
	// Setup
	hub := websocket.NewHub()
	controller := NewMessageController(hub)
	server := httptest.NewServer(http.HandlerFunc(controller.GetMessage))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Valid message ID",
			query:          "?id=1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing message ID",
			query:          "",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Invalid message ID",
			query:          "?id=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Non-existent message ID",
			query:          "?id=999",
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
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

func TestMessageController_UpdateMessage(t *testing.T) {
	// Setup
	hub := websocket.NewHub()
	controller := NewMessageController(hub)
	server := httptest.NewServer(http.HandlerFunc(controller.UpdateMessage))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		query          string
		payload        map[string]interface{}
		expectedStatus int
		expectedCode   string
	}{
		{
			name:  "Valid message update",
			query: "?id=1",
			payload: map[string]interface{}{
				"content": map[string]interface{}{
					"text": "Updated message",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing message ID",
			query:          "",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Invalid message ID",
			query:          "?id=invalid",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Non-existent message ID",
			query:          "?id=999",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusNotFound,
			expectedCode:   "NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			payload, _ := json.Marshal(tt.payload)
			req, err := http.NewRequest("PUT", server.URL+tt.query, bytes.NewBuffer(payload))
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

func TestMessageController_DeleteMessage(t *testing.T) {
	// Setup
	hub := websocket.NewHub()
	controller := NewMessageController(hub)
	server := httptest.NewServer(http.HandlerFunc(controller.DeleteMessage))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Valid message deletion",
			query:          "?id=1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing message ID",
			query:          "",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Invalid message ID",
			query:          "?id=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "BAD_REQUEST",
		},
		{
			name:           "Non-existent message ID",
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

func TestMessageController_GetMessageHistory(t *testing.T) {
	// Setup
	hub := websocket.NewHub()
	controller := NewMessageController(hub)
	server := httptest.NewServer(http.HandlerFunc(controller.GetMessageHistory))
	defer server.Close()

	// Test cases
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Valid history request",
			query:          "?type=chat&user_id=1&page=1&limit=20",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Valid history request with defaults",
			query:          "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid page number",
			query:          "?page=invalid",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid limit",
			query:          "?limit=invalid",
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
