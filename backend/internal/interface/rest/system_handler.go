package rest

import (
	"encoding/json"
	"log"
	"net/http"

	appSystem "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/system"
)

// SystemHandler handles system-related HTTP requests
type SystemHandler struct {
	systemUsecase *appSystem.Usecase
}

// NewSystemHandler creates a new SystemHandler
func NewSystemHandler(systemUsecase *appSystem.Usecase) *SystemHandler {
	return &SystemHandler{
		systemUsecase: systemUsecase,
	}
}

// ReleaseStatusResponse represents the release status response
type ReleaseStatusResponse struct {
	Released bool `json:"released"`
}

// UpdateReleaseStatusRequest represents the request to update release status
type UpdateReleaseStatusRequest struct {
	Released bool `json:"released"`
}

// GetReleaseStatus handles GET /api/v1/public/system/release-status
// This endpoint is public (no authentication required)
func (h *SystemHandler) GetReleaseStatus(w http.ResponseWriter, r *http.Request) {
	output, err := h.systemUsecase.GetReleaseStatus(r.Context())
	if err != nil {
		log.Printf("[ERROR] GetReleaseStatus failed: %v", err)
		RespondDomainError(w, err)
		return
	}

	RespondSuccess(w, ReleaseStatusResponse{
		Released: output.Released,
	})
}

// GetReleaseStatusAdmin handles GET /api/v1/admin/system/release-status
// This endpoint requires admin authentication
func (h *SystemHandler) GetReleaseStatusAdmin(w http.ResponseWriter, r *http.Request) {
	output, err := h.systemUsecase.GetReleaseStatus(r.Context())
	if err != nil {
		log.Printf("[ERROR] GetReleaseStatusAdmin failed: %v", err)
		RespondDomainError(w, err)
		return
	}

	RespondSuccess(w, ReleaseStatusResponse{
		Released: output.Released,
	})
}

// UpdateReleaseStatus handles PUT /api/v1/admin/system/release-status
// This endpoint requires admin authentication
func (h *SystemHandler) UpdateReleaseStatus(w http.ResponseWriter, r *http.Request) {
	var req UpdateReleaseStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "invalid request body")
		return
	}

	err := h.systemUsecase.UpdateReleaseStatus(r.Context(), appSystem.UpdateReleaseStatusInput{
		Released: req.Released,
	})
	if err != nil {
		log.Printf("[ERROR] UpdateReleaseStatus failed: %v", err)
		RespondDomainError(w, err)
		return
	}

	RespondSuccess(w, ReleaseStatusResponse{
		Released: req.Released,
	})
}
