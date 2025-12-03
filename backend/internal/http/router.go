package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/config"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// NewRouter creates a new chi router with all routes configured
func NewRouter(cfg *config.Config) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	// Health check endpoint
	r.Get("/health", healthHandler)

	return r
}

// healthHandler handles the health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := HealthResponse{Status: "ok"}
	json.NewEncoder(w).Encode(response)
}

