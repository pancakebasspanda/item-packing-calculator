package api

import (
	"encoding/json"
	"net/http"
)

// CalculatePackingRequest represents the expected JSON body from the checkout UI
type CalculatePackingRequest struct {
	OrderQuantity int `json:"orderQuantity"`
}

// SizeRequest represents the expected JSON body for adding/removing sizes
type SizeRequest struct {
	Size int `json:"size"`
}

// ErrorResponse represents a standardized JSON error returned to the client
type ErrorResponse struct {
	Error string `json:"error"`
}

// respondJSON is a helper to write a standard JSON response
func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// Fallback if encoding fails (rare, but good practice)
		http.Error(w, `{"error":"Failed to encode response"}`, http.StatusInternalServerError)
	}
}

// respondError is a helper to write a standard JSON error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, ErrorResponse{Error: message})
}
