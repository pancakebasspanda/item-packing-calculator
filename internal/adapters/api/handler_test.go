package api_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"item-packing-calculator/internal/adapters/api"
	"item-packing-calculator/internal/core/domain"
	"item-packing-calculator/internal/core/ports/mocks"
)

// TestNewHandler verifies the constructor correctly initializes the struct
func TestNewHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSvc := mocks.NewMockPackService(ctrl)
	handler := api.NewHandler(mockSvc)

	assert.NotNil(t, handler, "NewHandler should return a valid instance")
}

// TestHandler_HandleCalculate verifies the calculation endpoint
func TestHandler_HandleCalculate(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        string
		serviceMockOutcome func(mock *mocks.MockPackService)
		expectedStatus     int
		expectedBody       string
	}{
		{
			name:               "returns 400 Bad Request on invalid JSON body",
			requestBody:        `{"orderQuantity": "invalid"}`,
			serviceMockOutcome: func(mock *mocks.MockPackService) {},
			expectedStatus:     http.StatusBadRequest,
			expectedBody:       `{"error":"Invalid JSON payload"}`,
		},
		{
			name:               "returns 422 Unprocessable Entity on invalid order quantity",
			requestBody:        `{"orderQuantity": 0}`,
			serviceMockOutcome: func(mock *mocks.MockPackService) {},
			expectedStatus:     http.StatusUnprocessableEntity,
			expectedBody:       `{"error":"order quantity must be greater than zero"}`,
		},
		{
			name:        "returns 500 Internal Server Error when service fails",
			requestBody: `{"orderQuantity": 100}`,
			serviceMockOutcome: func(mock *mocks.MockPackService) {
				mock.EXPECT().
					Calculate(gomock.Any(), 100).
					Times(1).
					Return(domain.CalculationResult{}, errors.New("database connection lost"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"Internal server error"}`,
		},
		{
			name:        "returns 200 OK with calculation result payload",
			requestBody: `{"orderQuantity": 501}`,
			serviceMockOutcome: func(mock *mocks.MockPackService) {
				mock.EXPECT().
					Calculate(gomock.Any(), 501).
					Times(1).
					Return(domain.CalculationResult{
						OrderQuantity: 501,
						TotalItems:    750,
						TotalPacks:    2,
						Packs: []domain.PackResult{
							{PackSize: 500, Quantity: 1},
							{PackSize: 250, Quantity: 1},
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"orderQuantity":501,"totalItems":750,"totalPacks":2,"packs":[{"packSize":500,"quantity":1},{"packSize":250,"quantity":1}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockPackService(ctrl)
			tt.serviceMockOutcome(mockSvc)

			handler := api.NewHandler(mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/packing/calculate", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleCalculate(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			} else {
				assert.Empty(t, rec.Body.String())
			}
		})
	}
}

// TestHandler_HandleAddSize verifies the endpoint for adding pack sizes
func TestHandler_HandleAddSize(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        string
		serviceMockOutcome func(mock *mocks.MockPackService)
		expectedStatus     int
		expectedBody       string
	}{
		{
			name:               "returns 400 Bad Request on malformed JSON payload",
			requestBody:        `{size:}`,
			serviceMockOutcome: func(mock *mocks.MockPackService) {},
			expectedStatus:     http.StatusBadRequest,
			expectedBody:       `{"error":"Invalid JSON payload"}`,
		},
		{
			name:        "returns 409 Conflict when service rejects duplicate/invalid size",
			requestBody: `{"size": -50}`,
			serviceMockOutcome: func(mock *mocks.MockPackService) {
				mock.EXPECT().
					AddSize(gomock.Any(), -50).
					Times(1).
					Return(errors.New("pack size already exists or is invalid"))
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"pack size already exists or is invalid"}`,
		},
		{
			name:        "returns 201 Created on successful pack size addition",
			requestBody: `{"size": 3000}`,
			serviceMockOutcome: func(mock *mocks.MockPackService) {
				mock.EXPECT().
					AddSize(gomock.Any(), 3000).
					Times(1).
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "", // Empty body for successful creation responses
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockPackService(ctrl)
			tt.serviceMockOutcome(mockSvc)

			handler := api.NewHandler(mockSvc)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/packs", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleAddSize(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			} else {
				assert.Empty(t, rec.Body.String())
			}
		})
	}
}

// TestHandler_HandleRemoveSize verifies the endpoint for deleting pack sizes
func TestHandler_HandleRemoveSize(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        string
		serviceMockOutcome func(mock *mocks.MockPackService)
		expectedStatus     int
		expectedBody       string
	}{
		{
			name:        "returns 404 Not Found when size deletion fails in service",
			requestBody: `{"size": 9999}`,
			serviceMockOutcome: func(mock *mocks.MockPackService) {
				mock.EXPECT().
					RemoveSize(gomock.Any(), 9999).
					Times(1).
					Return(errors.New("pack size 9999 not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"pack size 9999 not found"}`,
		},
		{
			name:        "returns 204 No Content when pack size is successfully removed",
			requestBody: `{"size": 500}`,
			serviceMockOutcome: func(mock *mocks.MockPackService) {
				mock.EXPECT().
					RemoveSize(gomock.Any(), 500).
					Times(1).
					Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "", // Empty body for no-content responses
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSvc := mocks.NewMockPackService(ctrl)
			tt.serviceMockOutcome(mockSvc)

			handler := api.NewHandler(mockSvc)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/packs", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.HandleRemoveSize(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rec.Body.String())
			} else {
				assert.Empty(t, rec.Body.String())
			}
		})
	}
}
