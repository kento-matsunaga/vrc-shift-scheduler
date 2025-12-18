package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EventBusinessDayRepository implements event.EventBusinessDayRepository for PostgreSQL
type EventBusinessDayRepository struct {
	db *pgxpool.Pool
}

// NewEventBusinessDayRepository creates a new EventBusinessDayRepository
func NewEventBusinessDayRepository(db *pgxpool.Pool) *EventBusinessDayRepository {
	return &EventBusinessDayRepository{db: db}
}

// pgtypeTimeToTime converts pgtype.Time to time.Time
func pgtypeTimeToTime(pt pgtype.Time) time.Time {
	if !pt.Valid {
		return time.Time{}
	}
	return time.Date(0, 1, 1, int(pt.Microseconds/3600000000), int((pt.Microseconds%3600000000)/60000000), int((pt.Microseconds%60000000)/1000000), 0, time.UTC)
}

// Save saves an event business day (insert or update)
func (r *EventBusinessDayRepository) Save(ctx context.Context, bd *event.EventBusinessDay) error {
	query := `
		INSERT INTO event_business_days (
			business_day_id, tenant_id, event_id, target_date, start_time, end_time,
			occurrence_type, recurring_pattern_id, is_active, valid_from, valid_to,
			created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (business_day_id) DO UPDATE SET
			target_date = EXCLUDED.target_date,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			occurrence_type = EXCLUDED.occurrence_type,
			recurring_pattern_id = EXCLUDED.recurring_pattern_id,
			is_active = EXCLUDED.is_active,
			valid_from = EXCLUDED.valid_from,
			valid_to = EXCLUDED.valid_to,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	var recurringPatternID *string
	if bd.RecurringPatternID() != nil {
		id := bd.RecurringPatternID().String()
		recurringPatternID = &id
	}

	_, err := r.db.Exec(ctx, query,
		bd.BusinessDayID().String(),
		bd.TenantID().String(),
		bd.EventID().String(),
		bd.TargetDate(),
		bd.StartTime(),
		bd.EndTime(),
		string(bd.OccurrenceType()),
		recurringPatternID,
		bd.IsActive(),
		bd.ValidFrom(),
		bd.ValidTo(),
		bd.CreatedAt(),
		bd.UpdatedAt(),
		bd.DeletedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save event business day: %w", err)
	}

	return nil
}

// FindByID finds a business day by ID within a tenant
func (r *EventBusinessDayRepository) FindByID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) (*event.EventBusinessDay, error) {
	query := `
		SELECT
			business_day_id, tenant_id, event_id, target_date, start_time, end_time,
			occurrence_type, recurring_pattern_id, is_active, valid_from, valid_to,
			created_at, updated_at, deleted_at
		FROM event_business_days
		WHERE tenant_id = $1 AND business_day_id = $2 AND deleted_at IS NULL
	`

	var (
		businessDayIDStr    string
		tenantIDStr         string
		eventIDStr          string
		targetDate          time.Time
		startTime           pgtype.Time
		endTime             pgtype.Time
		occurrenceTypeStr   string
		recurringPatternID  sql.NullString
		isActive            bool
		validFrom           sql.NullTime
		validTo             sql.NullTime
		createdAt           time.Time
		updatedAt           time.Time
		deletedAt           sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), businessDayID.String()).Scan(
		&businessDayIDStr,
		&tenantIDStr,
		&eventIDStr,
		&targetDate,
		&startTime,
		&endTime,
		&occurrenceTypeStr,
		&recurringPatternID,
		&isActive,
		&validFrom,
		&validTo,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("EventBusinessDay", businessDayID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find event business day: %w", err)
	}

	return r.scanToBusinessDay(
		businessDayIDStr, tenantIDStr, eventIDStr, targetDate, pgtypeTimeToTime(startTime), pgtypeTimeToTime(endTime),
		occurrenceTypeStr, recurringPatternID, isActive, validFrom, validTo,
		createdAt, updatedAt, deletedAt,
	)
}

// FindByEventID finds all business days for an event
func (r *EventBusinessDayRepository) FindByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*event.EventBusinessDay, error) {
	query := `
		SELECT
			business_day_id, tenant_id, event_id, target_date, start_time, end_time,
			occurrence_type, recurring_pattern_id, is_active, valid_from, valid_to,
			created_at, updated_at, deleted_at
		FROM event_business_days
		WHERE tenant_id = $1 AND event_id = $2 AND deleted_at IS NULL
		ORDER BY target_date ASC, start_time ASC
	`

	return r.queryBusinessDays(ctx, query, tenantID.String(), eventID.String())
}

// FindByEventIDAndDateRange finds business days within a date range for an event
func (r *EventBusinessDayRepository) FindByEventIDAndDateRange(ctx context.Context, tenantID common.TenantID, eventID common.EventID, startDate, endDate time.Time) ([]*event.EventBusinessDay, error) {
	query := `
		SELECT
			business_day_id, tenant_id, event_id, target_date, start_time, end_time,
			occurrence_type, recurring_pattern_id, is_active, valid_from, valid_to,
			created_at, updated_at, deleted_at
		FROM event_business_days
		WHERE tenant_id = $1 AND event_id = $2
			AND target_date >= $3 AND target_date <= $4
			AND deleted_at IS NULL
		ORDER BY target_date ASC, start_time ASC
	`

	return r.queryBusinessDays(ctx, query, tenantID.String(), eventID.String(), startDate, endDate)
}

// FindActiveByEventID finds all active business days for an event
func (r *EventBusinessDayRepository) FindActiveByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*event.EventBusinessDay, error) {
	query := `
		SELECT
			business_day_id, tenant_id, event_id, target_date, start_time, end_time,
			occurrence_type, recurring_pattern_id, is_active, valid_from, valid_to,
			created_at, updated_at, deleted_at
		FROM event_business_days
		WHERE tenant_id = $1 AND event_id = $2 AND is_active = true AND deleted_at IS NULL
		ORDER BY target_date ASC, start_time ASC
	`

	return r.queryBusinessDays(ctx, query, tenantID.String(), eventID.String())
}

// FindByTenantIDAndDate finds all business days on a specific date within a tenant
func (r *EventBusinessDayRepository) FindByTenantIDAndDate(ctx context.Context, tenantID common.TenantID, date time.Time) ([]*event.EventBusinessDay, error) {
	query := `
		SELECT
			business_day_id, tenant_id, event_id, target_date, start_time, end_time,
			occurrence_type, recurring_pattern_id, is_active, valid_from, valid_to,
			created_at, updated_at, deleted_at
		FROM event_business_days
		WHERE tenant_id = $1 AND target_date = $2 AND deleted_at IS NULL
		ORDER BY start_time ASC
	`

	return r.queryBusinessDays(ctx, query, tenantID.String(), date)
}

// Delete deletes a business day (physical delete)
func (r *EventBusinessDayRepository) Delete(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) error {
	query := `
		DELETE FROM event_business_days
		WHERE tenant_id = $1 AND business_day_id = $2
	`

	result, err := r.db.Exec(ctx, query, tenantID.String(), businessDayID.String())
	if err != nil {
		return fmt.Errorf("failed to delete event business day: %w", err)
	}

	if result.RowsAffected() == 0 {
		return common.NewNotFoundError("EventBusinessDay", businessDayID.String())
	}

	return nil
}

// ExistsByEventIDAndDate checks if a business day exists for the given event and date
func (r *EventBusinessDayRepository) ExistsByEventIDAndDate(ctx context.Context, tenantID common.TenantID, eventID common.EventID, date time.Time, startTime time.Time) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM event_business_days
			WHERE tenant_id = $1 AND event_id = $2 AND target_date = $3 AND start_time = $4 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, tenantID.String(), eventID.String(), date, startTime).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check event business day existence: %w", err)
	}

	return exists, nil
}

// queryBusinessDays executes a query and returns a list of business days
func (r *EventBusinessDayRepository) queryBusinessDays(ctx context.Context, query string, args ...interface{}) ([]*event.EventBusinessDay, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query event business days: %w", err)
	}
	defer rows.Close()

	var businessDays []*event.EventBusinessDay
	for rows.Next() {
		var (
			businessDayIDStr   string
			tenantIDStr        string
			eventIDStr         string
			targetDate         time.Time
			startTime          pgtype.Time
			endTime            pgtype.Time
			occurrenceTypeStr  string
			recurringPatternID sql.NullString
			isActive           bool
			validFrom          sql.NullTime
			validTo            sql.NullTime
			createdAt          time.Time
			updatedAt          time.Time
			deletedAt          sql.NullTime
		)

		err := rows.Scan(
			&businessDayIDStr,
			&tenantIDStr,
			&eventIDStr,
			&targetDate,
			&startTime,
			&endTime,
			&occurrenceTypeStr,
			&recurringPatternID,
			&isActive,
			&validFrom,
			&validTo,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan business day row: %w", err)
		}

		// pgtype.Time を time.Time に変換
		startTimeVal := pgtypeTimeToTime(startTime)
		endTimeVal := pgtypeTimeToTime(endTime)

		bd, err := r.scanToBusinessDay(
			businessDayIDStr, tenantIDStr, eventIDStr, targetDate, startTimeVal, endTimeVal,
			occurrenceTypeStr, recurringPatternID, isActive, validFrom, validTo,
			createdAt, updatedAt, deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct business day: %w", err)
		}

		businessDays = append(businessDays, bd)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating business day rows: %w", err)
	}

	return businessDays, nil
}

// scanToBusinessDay converts scanned row data to EventBusinessDay entity
func (r *EventBusinessDayRepository) scanToBusinessDay(
	businessDayIDStr, tenantIDStr, eventIDStr string,
	targetDate, startTime, endTime time.Time,
	occurrenceTypeStr string,
	recurringPatternID sql.NullString,
	isActive bool,
	validFrom, validTo sql.NullTime,
	createdAt, updatedAt time.Time,
	deletedAt sql.NullTime,
) (*event.EventBusinessDay, error) {
	var recurringPatternIDPtr *common.EventID
	if recurringPatternID.Valid {
		id := common.EventID(recurringPatternID.String)
		recurringPatternIDPtr = &id
	}

	var validFromPtr, validToPtr *time.Time
	if validFrom.Valid {
		validFromPtr = &validFrom.Time
	}
	if validTo.Valid {
		validToPtr = &validTo.Time
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return event.ReconstructEventBusinessDay(
		event.BusinessDayID(businessDayIDStr),
		common.TenantID(tenantIDStr),
		common.EventID(eventIDStr),
		targetDate,
		startTime,
		endTime,
		event.OccurrenceType(occurrenceTypeStr),
		recurringPatternIDPtr,
		isActive,
		validFromPtr,
		validToPtr,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}

