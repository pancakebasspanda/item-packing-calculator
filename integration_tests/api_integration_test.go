//go:build integration

package integration_tests

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"item-packing-calculator/internal/adapters/api"
	"item-packing-calculator/internal/adapters/storage"
	"item-packing-calculator/internal/core/service"
)

// setupTestServer wires up the real core logic exactly like main.go
func setupTestServer(t *testing.T) *httptest.Server {
	// initialize
	initialSizes := []int{5000, 2000, 1000, 500, 250}
	store := storage.NewMemoryStorage(initialSizes)
	calculatorService := service.NewCalculatorService(store)
	handler := api.NewHandler(calculatorService)

	// routing
	mux := http.NewServeMux()
	api.RegisterRoutes(mux, handler)

	return httptest.NewServer(mux)
}

func TestIntegration_CalculateEndpoint(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedJSON   string
	}{
		{
			name:           "valid items - exactly matches one pack",
			requestBody:    `{"orderQuantity": 250}`,
			expectedStatus: http.StatusOK,
			expectedJSON:   `{"orderQuantity":250, "packs":[{"packSize":250, "quantity":1}], "totalItems":250, "totalPacks":1}`,
		},
		{
			name:           "valid items - uses multiple packs",
			requestBody:    `{"orderQuantity": 501}`,
			expectedStatus: http.StatusOK,
			expectedJSON:   `{"orderQuantity":501, "packs":[{"packSize":500, "quantity":1}, {"packSize":250, "quantity":1}], "totalItems":750, "totalPacks":2}`,
		},
		{
			name:           "empty payload",
			requestBody:    `{}`,
			expectedStatus: http.StatusUnprocessableEntity,
			expectedJSON:   `{"error": "order quantity must be greater than zero"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Post(ts.URL+"/api/v1/packing/calculate", "application/json", bytes.NewBuffer([]byte(tt.requestBody)))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tt.expectedJSON != "" {
				assert.JSONEq(t, tt.expectedJSON, string(bodyBytes))
			}
		})
	}
}

func TestIntegration_AddPackSizeEndpoint(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedJSON   string
	}{
		{
			name:           "Valid addition",
			requestBody:    `{"size": 999}`,
			expectedStatus: http.StatusCreated, // 201
			expectedJSON:   "",
		},
		{
			name:           "Invalid addition - zero size",
			requestBody:    `{"size": 0}`,
			expectedStatus: http.StatusConflict, // 409
			expectedJSON:   `{"error": "pack size must be greater than zero"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Post(ts.URL+"/api/v1/packs", "application/json", bytes.NewBuffer([]byte(tt.requestBody)))
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tt.expectedJSON != "" {
				assert.JSONEq(t, tt.expectedJSON, string(bodyBytes))
			}
		})
	}
}

func TestIntegration_RemovePackSizeEndpoint(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedJSON   string
	}{
		{
			name:           "Valid removal - existing size",
			requestBody:    `{"size": 250}`,
			expectedStatus: http.StatusNoContent, // 204
			expectedJSON:   "",
		},
		{
			name:           "Remove non-existent size",
			requestBody:    `{"size": 999999}`,
			expectedStatus: http.StatusNotFound,
			expectedJSON:   `{"error": "pack size 999999 not found"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodDelete, ts.URL+"/api/v1/packs", bytes.NewBuffer([]byte(tt.requestBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			bodyBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tt.expectedJSON != "" {
				assert.JSONEq(t, tt.expectedJSON, string(bodyBytes))
			}
		})
	}
}
