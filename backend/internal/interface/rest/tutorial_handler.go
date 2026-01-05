package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	apptutorial "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/tutorial"
	tutorialdomain "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tutorial"
	"github.com/go-chi/chi/v5"
)

type TutorialHandler struct {
	listUC *apptutorial.ListTutorialsUsecase
	getUC  *apptutorial.GetTutorialUsecase
}

func NewTutorialHandler(
	listUC *apptutorial.ListTutorialsUsecase,
	getUC *apptutorial.GetTutorialUsecase,
) *TutorialHandler {
	return &TutorialHandler{
		listUC: listUC,
		getUC:  getUC,
	}
}

// List handles GET /api/v1/tutorials
func (h *TutorialHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tutorials, err := h.listUC.Execute(ctx)
	if err != nil {
		RespondInternalError(w)
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"tutorials": tutorials,
	}})
}

// Get handles GET /api/v1/tutorials/:id
func (h *TutorialHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	tutorial, err := h.getUC.Execute(ctx, id)
	if err != nil {
		if errors.Is(err, tutorialdomain.ErrTutorialNotFound) {
			RespondNotFound(w, "Tutorial not found")
			return
		}
		RespondInternalError(w)
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"tutorial": tutorial,
	}})
}

// AdminTutorialHandler handles admin operations
type AdminTutorialHandler struct {
	listAllUC *apptutorial.ListAllTutorialsUsecase
	createUC  *apptutorial.CreateTutorialUsecase
	updateUC  *apptutorial.UpdateTutorialUsecase
	deleteUC  *apptutorial.DeleteTutorialUsecase
}

func NewAdminTutorialHandler(
	listAllUC *apptutorial.ListAllTutorialsUsecase,
	createUC *apptutorial.CreateTutorialUsecase,
	updateUC *apptutorial.UpdateTutorialUsecase,
	deleteUC *apptutorial.DeleteTutorialUsecase,
) *AdminTutorialHandler {
	return &AdminTutorialHandler{
		listAllUC: listAllUC,
		createUC:  createUC,
		updateUC:  updateUC,
		deleteUC:  deleteUC,
	}
}

// List handles GET /api/v1/admin/tutorials
func (h *AdminTutorialHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tutorials, err := h.listAllUC.Execute(ctx)
	if err != nil {
		RespondInternalError(w)
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"tutorials": tutorials,
	}})
}

// Create handles POST /api/v1/admin/tutorials
func (h *AdminTutorialHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req apptutorial.CreateTutorialInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	result, err := h.createUC.Execute(ctx, req)
	if err != nil {
		if errors.Is(err, tutorialdomain.ErrCategoryRequired) ||
			errors.Is(err, tutorialdomain.ErrCategoryTooLong) ||
			errors.Is(err, tutorialdomain.ErrTitleRequired) ||
			errors.Is(err, tutorialdomain.ErrTitleTooLong) ||
			errors.Is(err, tutorialdomain.ErrBodyRequired) {
			RespondBadRequest(w, err.Error())
			return
		}
		RespondInternalError(w)
		return
	}

	RespondJSON(w, http.StatusCreated, SuccessResponse{Data: map[string]interface{}{
		"tutorial": result,
	}})
}

// Update handles PUT /api/v1/admin/tutorials/:id
func (h *AdminTutorialHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	var req apptutorial.UpdateTutorialInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}
	req.ID = id

	result, err := h.updateUC.Execute(ctx, req)
	if err != nil {
		if errors.Is(err, tutorialdomain.ErrTutorialNotFound) {
			RespondNotFound(w, "Tutorial not found")
			return
		}
		if errors.Is(err, tutorialdomain.ErrCategoryRequired) ||
			errors.Is(err, tutorialdomain.ErrCategoryTooLong) ||
			errors.Is(err, tutorialdomain.ErrTitleRequired) ||
			errors.Is(err, tutorialdomain.ErrTitleTooLong) ||
			errors.Is(err, tutorialdomain.ErrBodyRequired) {
			RespondBadRequest(w, err.Error())
			return
		}
		RespondInternalError(w)
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"tutorial": result,
	}})
}

// Delete handles DELETE /api/v1/admin/tutorials/:id
func (h *AdminTutorialHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if err := h.deleteUC.Execute(ctx, id); err != nil {
		if errors.Is(err, tutorialdomain.ErrTutorialNotFound) {
			RespondNotFound(w, "Tutorial not found")
			return
		}
		RespondInternalError(w)
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"success": true,
	}})
}
