package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"

	appshift "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/shift"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/go-chi/chi/v5"
)

// InstanceHandler handles instance-related HTTP requests
type InstanceHandler struct {
	createInstanceUC *appshift.CreateInstanceUsecase
	listInstancesUC  *appshift.ListInstancesUsecase
	getInstanceUC    *appshift.GetInstanceUsecase
	updateInstanceUC *appshift.UpdateInstanceUsecase
	deleteInstanceUC *appshift.DeleteInstanceUsecase
}

// NewInstanceHandler creates a new InstanceHandler with injected usecases
func NewInstanceHandler(
	createInstanceUC *appshift.CreateInstanceUsecase,
	listInstancesUC *appshift.ListInstancesUsecase,
	getInstanceUC *appshift.GetInstanceUsecase,
	updateInstanceUC *appshift.UpdateInstanceUsecase,
	deleteInstanceUC *appshift.DeleteInstanceUsecase,
) *InstanceHandler {
	return &InstanceHandler{
		createInstanceUC: createInstanceUC,
		listInstancesUC:  listInstancesUC,
		getInstanceUC:    getInstanceUC,
		updateInstanceUC: updateInstanceUC,
		deleteInstanceUC: deleteInstanceUC,
	}
}

// CreateInstanceRequest represents the request body for creating an instance
type CreateInstanceRequest struct {
	Name         string `json:"name"`
	DisplayOrder int    `json:"display_order"`
	MaxMembers   *int   `json:"max_members"`
}

// UpdateInstanceRequest represents the request body for updating an instance
type UpdateInstanceRequest struct {
	Name         *string `json:"name"`
	DisplayOrder *int    `json:"display_order"`
	MaxMembers   *int    `json:"max_members"`
}

// InstanceResponse represents an instance in API responses
type InstanceResponse struct {
	InstanceID   string `json:"instance_id"`
	TenantID     string `json:"tenant_id"`
	EventID      string `json:"event_id"`
	Name         string `json:"name"`
	DisplayOrder int    `json:"display_order"`
	MaxMembers   *int   `json:"max_members"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// toInstanceResponse converts a domain instance to an API response
func toInstanceResponse(instance *shift.Instance) InstanceResponse {
	return InstanceResponse{
		InstanceID:   instance.InstanceID().String(),
		TenantID:     instance.TenantID().String(),
		EventID:      instance.EventID().String(),
		Name:         instance.Name(),
		DisplayOrder: instance.DisplayOrder(),
		MaxMembers:   instance.MaxMembers(),
		CreatedAt:    instance.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    instance.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}

// CreateInstance handles POST /api/v1/events/:event_id/instances
func (h *InstanceHandler) CreateInstance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// event_id の取得
	eventIDStr := chi.URLParam(r, "event_id")
	if eventIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "event_id is required", nil)
		return
	}

	eventID, err := common.ParseEventID(eventIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid event_id format", nil)
		return
	}

	// リクエストボディのパース
	var req CreateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// バリデーション
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "name is required", nil)
		return
	}

	// Usecaseの実行
	input := appshift.CreateInstanceInput{
		TenantID:     tenantID,
		EventID:      eventID,
		Name:         req.Name,
		DisplayOrder: req.DisplayOrder,
		MaxMembers:   req.MaxMembers,
	}

	newInstance, err := h.createInstanceUC.Execute(ctx, input)
	if err != nil {
		slog.Error("CreateInstance failed",
			"error", err,
			"tenant_id", tenantID.String(),
			"event_id", eventID.String(),
		)
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	writeSuccess(w, http.StatusCreated, toInstanceResponse(newInstance))
}

// GetInstances handles GET /api/v1/events/:event_id/instances
func (h *InstanceHandler) GetInstances(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// event_id の取得
	eventIDStr := chi.URLParam(r, "event_id")
	if eventIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "event_id is required", nil)
		return
	}

	eventID, err := common.ParseEventID(eventIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid event_id format", nil)
		return
	}

	// Usecaseの実行
	input := appshift.ListInstancesInput{
		TenantID: tenantID,
		EventID:  eventID,
	}

	instances, err := h.listInstancesUC.Execute(ctx, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch instances", nil)
		return
	}

	// レスポンス構築
	instanceResponses := make([]InstanceResponse, 0, len(instances))
	for _, instance := range instances {
		instanceResponses = append(instanceResponses, toInstanceResponse(instance))
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"instances": instanceResponses,
		"count":     len(instanceResponses),
	})
}

// GetInstance handles GET /api/v1/instances/:instance_id
func (h *InstanceHandler) GetInstance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// instance_id の取得
	instanceIDStr := chi.URLParam(r, "instance_id")
	if instanceIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "instance_id is required", nil)
		return
	}

	instanceID, err := shift.ParseInstanceID(instanceIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid instance_id format", nil)
		return
	}

	// Usecaseの実行
	input := appshift.GetInstanceInput{
		TenantID:   tenantID,
		InstanceID: instanceID,
	}

	instance, err := h.getInstanceUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	writeSuccess(w, http.StatusOK, toInstanceResponse(instance))
}

// UpdateInstance handles PUT /api/v1/instances/:instance_id
func (h *InstanceHandler) UpdateInstance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// instance_id の取得
	instanceIDStr := chi.URLParam(r, "instance_id")
	if instanceIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "instance_id is required", nil)
		return
	}

	instanceID, err := shift.ParseInstanceID(instanceIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid instance_id format", nil)
		return
	}

	// リクエストボディのパース
	var req UpdateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// Usecaseの実行
	input := appshift.UpdateInstanceInput{
		TenantID:     tenantID,
		InstanceID:   instanceID,
		Name:         req.Name,
		DisplayOrder: req.DisplayOrder,
		MaxMembers:   req.MaxMembers,
	}

	updatedInstance, err := h.updateInstanceUC.Execute(ctx, input)
	if err != nil {
		slog.Error("UpdateInstance failed",
			"error", err,
			"tenant_id", tenantID.String(),
			"instance_id", instanceID.String(),
		)
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	writeSuccess(w, http.StatusOK, toInstanceResponse(updatedInstance))
}

// CheckInstanceDeletableResponse represents the response for checking if an instance can be deleted
type CheckInstanceDeletableResponse struct {
	CanDelete      bool   `json:"can_delete"`
	SlotCount      int    `json:"slot_count"`
	AssignedSlots  int    `json:"assigned_slots"`
	BlockingReason string `json:"blocking_reason,omitempty"`
}

// CheckInstanceDeletable handles GET /api/v1/instances/:instance_id/deletable
func (h *InstanceHandler) CheckInstanceDeletable(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// instance_id の取得
	instanceIDStr := chi.URLParam(r, "instance_id")
	if instanceIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "instance_id is required", nil)
		return
	}

	instanceID, err := shift.ParseInstanceID(instanceIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid instance_id format", nil)
		return
	}

	// Usecaseの実行
	input := appshift.DeleteInstanceInput{
		TenantID:   tenantID,
		InstanceID: instanceID,
	}

	result, err := h.deleteInstanceUC.CheckDeletable(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	writeSuccess(w, http.StatusOK, CheckInstanceDeletableResponse{
		CanDelete:      result.CanDelete,
		SlotCount:      result.SlotCount,
		AssignedSlots:  result.AssignedSlots,
		BlockingReason: result.BlockingReason,
	})
}

// DeleteInstance handles DELETE /api/v1/instances/:instance_id
func (h *InstanceHandler) DeleteInstance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// instance_id の取得
	instanceIDStr := chi.URLParam(r, "instance_id")
	if instanceIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "instance_id is required", nil)
		return
	}

	instanceID, err := shift.ParseInstanceID(instanceIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid instance_id format", nil)
		return
	}

	// Usecaseの実行
	input := appshift.DeleteInstanceInput{
		TenantID:   tenantID,
		InstanceID: instanceID,
	}

	if err := h.deleteInstanceUC.Execute(ctx, input); err != nil {
		slog.Error("DeleteInstance failed",
			"error", err,
			"tenant_id", tenantID.String(),
			"instance_id", instanceID.String(),
		)
		RespondDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
