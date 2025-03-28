package handlers

import (
	"encoding/json"
	"net/http"
)

// SendResponse sends a JSON response with the given data and status code
func SendResponse(w http.ResponseWriter, success bool, data interface{}, message string, statusCode int) {
	response := map[string]interface{}{
		"success": success,
		"data":    data,
		"message": message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
} 