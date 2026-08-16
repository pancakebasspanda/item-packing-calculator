package api

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// HandleAddSize processes requests to add a new pack size
func (h *Handler) HandleAddSize(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.HandleAddSize")
	defer span.End()
	logger := zerolog.Ctx(ctx)

	var req SizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid json payload")
		respondError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	span.SetAttributes(attribute.Int("http.request.size", req.Size))

	if err := h.service.AddSize(ctx, req.Size); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Warn().Err(err).Int("size", req.Size).Msg("failed to add size")

		respondError(w, http.StatusConflict, err.Error())
		return
	}

	span.SetStatus(codes.Ok, "size added")
	logger.Info().Int("size", req.Size).Msg("pack size successfully added")

	w.WriteHeader(http.StatusCreated)
}

// HandleRemoveSize processes requests to remove a pack size
func (h *Handler) HandleRemoveSize(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "Handler.HandleRemoveSize")
	defer span.End()
	logger := zerolog.Ctx(ctx)

	var req SizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid json payload")
		respondError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	span.SetAttributes(attribute.Int("http.request.size", req.Size))

	if err := h.service.RemoveSize(ctx, req.Size); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		logger.Warn().Err(err).Int("size", req.Size).Msg("failed to remove size")

		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	span.SetStatus(codes.Ok, "size removed")
	logger.Info().Int("size", req.Size).Msg("pack size successfully removed")

	w.WriteHeader(http.StatusNoContent)
}
