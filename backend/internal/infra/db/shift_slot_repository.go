package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ShiftSlotRepository implements shift.ShiftSlotRepository for PostgreSQL
type ShiftSlotRepository struct {
	db *pgxpool.Pool
}

// NewShiftSlotRepository creates a new ShiftSlotRepository
func NewShiftSlotRepository(db *pgxpool.Pool) *ShiftSlotRepository {
	return &ShiftSlotRepository{db: db}
}

// Save saves a shift slot (insert or update)
func (r *ShiftSlotRepository) Save(ctx context.Context, slot *shift.ShiftSlot) error {
	query := `
		INSERT INTO shift_slots (
			slot_id, tenant_id, business_day_id, instance_id,
			slot_name, instance_name, start_time, end_time,
			required_count, priority, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (slot_id) DO UPDATE SET
			instance_id = EXCLUDED.instance_id,
			slot_name = EXCLUDED.slot_name,
			instance_name = EXCLUDED.instance_name,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			required_count = EXCLUDED.required_count,
			priority = EXCLUDED.priority,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	var instanceIDStr *string
	if slot.InstanceID() != nil {
		s := slot.InstanceID().String()
		instanceIDStr = &s
	}

	_, err := r.db.Exec(ctx, query,
		slot.SlotID().String(),
		slot.TenantID().String(),
		slot.BusinessDayID().String(),
		instanceIDStr,
		slot.SlotName(),
		slot.InstanceName(),
		slot.StartTime(),
		slot.EndTime(),
		slot.RequiredCount(),
		slot.Priority(),
		slot.CreatedAt(),
		slot.UpdatedAt(),
		slot.DeletedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save shift slot: %w", err)
	}

	return nil
}

// FindByID finds a shift slot by ID within a tenant
func (r *ShiftSlotRepository) FindByID(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) (*shift.ShiftSlot, error) {
	query := `
		SELECT
			slot_id, tenant_id, business_day_id, instance_id,
			slot_name, instance_name, start_time, end_time,
			required_count, priority, created_at, updated_at, deleted_at
		FROM shift_slots
		WHERE tenant_id = $1 AND slot_id = $2 AND deleted_at IS NULL
	`

	var (
		slotIDStr        string
		tenantIDStr      string
		businessDayIDStr string
		instanceIDStr    sql.NullString
		slotName         string
		instanceName     string
		startTime        time.Time
		endTime          time.Time
		requiredCount    int
		priority         int
		createdAt        time.Time
		updatedAt        time.Time
		deletedAt        sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), slotID.String()).Scan(
		&slotIDStr,
		&tenantIDStr,
		&businessDayIDStr,
		&instanceIDStr,
		&slotName,
		&instanceName,
		&startTime,
		&endTime,
		&requiredCount,
		&priority,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("ShiftSlot", slotID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find shift slot: %w", err)
	}

	return r.scanToShiftSlot(
		slotIDStr, tenantIDStr, businessDayIDStr, instanceIDStr,
		slotName, instanceName, startTime, endTime,
		requiredCount, priority, createdAt, updatedAt, deletedAt,
	)
}

// FindByBusinessDayID finds all shift slots for a business day
func (r *ShiftSlotRepository) FindByBusinessDayID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) ([]*shift.ShiftSlot, error) {
	query := `
		SELECT
			slot_id, tenant_id, business_day_id, instance_id,
			slot_name, instance_name, start_time, end_time,
			required_count, priority, created_at, updated_at, deleted_at
		FROM shift_slots
		WHERE tenant_id = $1 AND business_day_id = $2 AND deleted_at IS NULL
		ORDER BY priority ASC, created_at ASC
	`

	return r.queryShiftSlots(ctx, query, tenantID.String(), businessDayID.String())
}

// FindByInstanceID finds all shift slots for an instance
func (r *ShiftSlotRepository) FindByInstanceID(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) ([]*shift.ShiftSlot, error) {
	query := `
		SELECT
			slot_id, tenant_id, business_day_id, instance_id,
			slot_name, instance_name, start_time, end_time,
			required_count, priority, created_at, updated_at, deleted_at
		FROM shift_slots
		WHERE tenant_id = $1 AND instance_id = $2 AND deleted_at IS NULL
		ORDER BY priority ASC, created_at ASC
	`

	return r.queryShiftSlots(ctx, query, tenantID.String(), instanceID.String())
}

// FindByBusinessDayIDAndInstanceID finds all shift slots for a business day and instance
func (r *ShiftSlotRepository) FindByBusinessDayIDAndInstanceID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID, instanceID shift.InstanceID) ([]*shift.ShiftSlot, error) {
	query := `
		SELECT
			slot_id, tenant_id, business_day_id, instance_id,
			slot_name, instance_name, start_time, end_time,
			required_count, priority, created_at, updated_at, deleted_at
		FROM shift_slots
		WHERE tenant_id = $1 AND business_day_id = $2 AND instance_id = $3 AND deleted_at IS NULL
		ORDER BY priority ASC, created_at ASC
	`

	return r.queryShiftSlots(ctx, query, tenantID.String(), businessDayID.String(), instanceID.String())
}

// Delete deletes a shift slot (physical delete)
func (r *ShiftSlotRepository) Delete(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) error {
	query := `
		DELETE FROM shift_slots
		WHERE tenant_id = $1 AND slot_id = $2
	`

	result, err := r.db.Exec(ctx, query, tenantID.String(), slotID.String())
	if err != nil {
		return fmt.Errorf("failed to delete shift slot: %w", err)
	}

	if result.RowsAffected() == 0 {
		return common.NewNotFoundError("ShiftSlot", slotID.String())
	}

	return nil
}

// queryShiftSlots executes a query and returns a list of shift slots
func (r *ShiftSlotRepository) queryShiftSlots(ctx context.Context, query string, args ...interface{}) ([]*shift.ShiftSlot, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query shift slots: %w", err)
	}
	defer rows.Close()

	var slots []*shift.ShiftSlot
	for rows.Next() {
		var (
			slotIDStr        string
			tenantIDStr      string
			businessDayIDStr string
			instanceIDStr    sql.NullString
			slotName         string
			instanceName     string
			startTime        time.Time
			endTime          time.Time
			requiredCount    int
			priority         int
			createdAt        time.Time
			updatedAt        time.Time
			deletedAt        sql.NullTime
		)

		err := rows.Scan(
			&slotIDStr,
			&tenantIDStr,
			&businessDayIDStr,
			&instanceIDStr,
			&slotName,
			&instanceName,
			&startTime,
			&endTime,
			&requiredCount,
			&priority,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan shift slot row: %w", err)
		}

		slot, err := r.scanToShiftSlot(
			slotIDStr, tenantIDStr, businessDayIDStr, instanceIDStr,
			slotName, instanceName, startTime, endTime,
			requiredCount, priority, createdAt, updatedAt, deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct shift slot: %w", err)
		}

		slots = append(slots, slot)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating shift slot rows: %w", err)
	}

	return slots, nil
}

// scanToShiftSlot converts scanned row data to ShiftSlot entity
func (r *ShiftSlotRepository) scanToShiftSlot(
	slotIDStr, tenantIDStr, businessDayIDStr string,
	instanceIDStr sql.NullString,
	slotName, instanceName string,
	startTime, endTime time.Time,
	requiredCount, priority int,
	createdAt, updatedAt time.Time,
	deletedAt sql.NullTime,
) (*shift.ShiftSlot, error) {
	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	var instanceIDPtr *shift.InstanceID
	if instanceIDStr.Valid {
		instanceID := shift.InstanceID(instanceIDStr.String)
		instanceIDPtr = &instanceID
	}

	return shift.ReconstructShiftSlot(
		shift.SlotID(slotIDStr),
		common.TenantID(tenantIDStr),
		event.BusinessDayID(businessDayIDStr),
		instanceIDPtr,
		slotName,
		instanceName,
		startTime,
		endTime,
		requiredCount,
		priority,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}

