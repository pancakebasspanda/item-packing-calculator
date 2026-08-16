package api

import (
	"net/http"
)

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	// expose api endpoints
	mux.HandleFunc("POST /api/v1/packing/calculate", h.HandleCalculate)
	mux.HandleFunc("POST /api/v1/packs", h.HandleAddSize)
	mux.HandleFunc("DELETE /api/v1/packs", h.HandleRemoveSize)

	// expose the openapi spec to the frontend for the contract
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		http.ServeFile(w, r, "openapi.yaml")
	})

	// expose interactive swagger ui documentation
	mux.HandleFunc("GET /docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "swagger.html")
	})

}
