package api

import (
	"go.opentelemetry.io/otel"

	"item-packing-calculator/internal/core/ports"
)

var tracer = otel.Tracer("api-handler")

// Handler handles http requests and responses
type Handler struct {
	service ports.PackService
}

// NewHandler creates a new http handler with the required dependencies
func NewHandler(service ports.PackService) *Handler {
	return &Handler{
		service: service,
	}
}
