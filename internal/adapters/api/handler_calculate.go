package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// HandleCalculate processes the packing calculation request
func (h *Handler) HandleCalculate(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.HandleCalculate")
	defer span.End()

	logger := zerolog.Ctx(ctx)

	var req CalculatePackingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid json payload")
		logger.Warn().Err(err).Msg("failed to decode request body")

		// 400 Bad Request: The JSON was malformed and couldn't be parsed
		respondError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	span.SetAttributes(attribute.Int("http.request.order_quantity", req.OrderQuantity))

	if req.OrderQuantity <= 0 {
		err := errors.New("order quantity must be greater than zero")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Warn().Int("quantity", req.OrderQuantity).Msg("client requested invalid order quantity")

		// 422 Unprocessable Entity: The JSON was valid, but the value violates business rules
		respondError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	result, err := h.service.Calculate(ctx, req.OrderQuantity)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "internal calculation error")
		logger.Error().Err(err).Msg("failed to calculate packs")

		// 500 Internal Server Error: The service/storage layer failed
		respondError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// log friendly result
	packSummary := make(map[int]int)
	for _, p := range result.Packs {
		packSummary[p.PackSize] = p.Quantity
	}

	span.SetStatus(codes.Ok, "calculation successful")
	logger.Info().
		Int("requested", result.OrderQuantity).
		Int("shipped", result.TotalItems).
		Int("total_packs", result.TotalPacks).
		Interface("pack_breakdown", packSummary).
		Msg("successfully calculated order packing")

	// 200 OK: Calculation was completely successful
	respondJSON(w, http.StatusOK, result)
}
