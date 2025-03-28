package middleware

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"saas-chat-system/internal/encryption"
)

// EncryptedRequest holds the encrypted data
type EncryptedRequest struct {
	EncryptedData string `json:"encryptedData"`
	// Public key used for handshake or key exchange
	PublicKey string `json:"publicKey,omitempty"`
}

// EncryptedResponse holds the encrypted response data
type EncryptedResponse struct {
	EncryptedData string `json:"encryptedData"`
	// Public key used for handshake or key exchange
	PublicKey string `json:"publicKey,omitempty"`
}

// EncryptionMiddleware handles request/response encryption
type EncryptionMiddleware struct {
	encryptionService *encryption.Service
}

// NewEncryptionMiddleware creates a new encryption middleware
func NewEncryptionMiddleware(encryptionService *encryption.Service) *EncryptionMiddleware {
	return &EncryptionMiddleware{
		encryptionService: encryptionService,
	}
}

// Middleware implements end-to-end encryption for API requests and responses
func (m *EncryptionMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip encryption for endpoints that handle their own encryption
		if shouldSkipEncryption(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Handle encrypted request
		if r.Method != http.MethodGet && r.Body != nil {
			// Read the encrypted request body
			body, err := ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			// Parse the encrypted request
			var encReq EncryptedRequest
			if err := json.Unmarshal(body, &encReq); err != nil {
				http.Error(w, "Invalid encrypted request format", http.StatusBadRequest)
				return
			}

			// Decrypt the data
			decrypted, err := m.encryptionService.DecryptString(encReq.EncryptedData)
			if err != nil {
				http.Error(w, "Failed to decrypt request data", http.StatusBadRequest)
				return
			}

			// Replace the request body with decrypted data
			r.Body = ioutil.NopCloser(bytes.NewReader([]byte(decrypted)))
			r.ContentLength = int64(len(decrypted))
		}

		// Capture the response
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		// Handle encrypted response
		// Only encrypt responses with status 200-299
		if rw.Status() >= 200 && rw.Status() < 300 && rw.Body() != nil {
			// Encrypt the response body
			encrypted, err := m.encryptionService.EncryptString(string(rw.Body()))
			if err != nil {
				http.Error(w, "Failed to encrypt response", http.StatusInternalServerError)
				return
			}

			// Create encrypted response
			encRes := EncryptedResponse{
				EncryptedData: encrypted,
			}

			// Write the encrypted response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(rw.Status())
			json.NewEncoder(w).Encode(encRes)
			return
		}

		// Return the original response if not encrypted
		w.WriteHeader(rw.Status())
		w.Write(rw.Body())
	})
}

// Custom response writer to capture response
type responseWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, status: http.StatusOK}
}

func (rw *responseWriter) Status() int {
	return rw.status
}

func (rw *responseWriter) Body() []byte {
	return rw.body
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	// Don't write the header yet, we'll do it after potential encryption
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	// Store the body for potential encryption
	rw.body = append(rw.body, b...)
	return len(b), nil
}

// shouldSkipEncryption returns true if the path should skip encryption
func shouldSkipEncryption(path string) bool {
	// Skip paths that handle their own encryption
	skipPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
		"/api/v1/key-exchange",
		"/health",
		"/metrics",
	}

	for _, p := range skipPaths {
		if path == p {
			return true
		}
	}

	return false
}
