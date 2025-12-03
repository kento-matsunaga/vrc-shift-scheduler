package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EventRepository implements event.EventRepository for PostgreSQL
type EventRepository struct {
	db *pgxpool.Pool
}

// NewEventRepository creates a new EventRepository
func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

// Save saves an event (insert or update)
func (r *EventRepository) Save(ctx context.Context, e *event.Event) error {
	query := `
		INSERT INTO events (
			event_id, tenant_id, event_name, event_type, description,
			is_active, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (event_id) DO UPDATE SET
			event_name = EXCLUDED.event_name,
			event_type = EXCLUDED.event_type,
			description = EXCLUDED.description,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Exec(ctx, query,
		e.EventID().String(),
		e.TenantID().String(),
		e.EventName(),
		string(e.EventType()),
		e.Description(),
		e.IsActive(),
		e.CreatedAt(),
		e.UpdatedAt(),
		e.DeletedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save event: %w", err)
	}

	return nil
}

// FindByID finds an event by ID within a tenant
func (r *EventRepository) FindByID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) (*event.Event, error) {
	query := `
		SELECT
			event_id, tenant_id, event_name, event_type, description,
			is_active, created_at, updated_at, deleted_at
		FROM events
		WHERE tenant_id = $1 AND event_id = $2 AND deleted_at IS NULL
	`

	var (
		eventIDStr    string
		tenantIDStr   string
		eventName     string
		eventTypeStr  string
		description   string
		isActive      bool
		createdAt     time.Time
		updatedAt     time.Time
		deletedAt     sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), eventID.String()).Scan(
		&eventIDStr,
		&tenantIDStr,
		&eventName,
		&eventTypeStr,
		&description,
		&isActive,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Event", eventID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find event: %w", err)
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return event.ReconstructEvent(
		common.EventID(eventIDStr),
		common.TenantID(tenantIDStr),
		eventName,
		event.EventType(eventTypeStr),
		description,
		isActive,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}

// FindByTenantID finds all events within a tenant
func (r *EventRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*event.Event, error) {
	query := `
		SELECT
			event_id, tenant_id, event_name, event_type, description,
			is_active, created_at, updated_at, deleted_at
		FROM events
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find events by tenant: %w", err)
	}
	defer rows.Close()

	var events []*event.Event
	for rows.Next() {
		var (
			eventIDStr    string
			tenantIDStr   string
			eventName     string
			eventTypeStr  string
			description   string
			isActive      bool
			createdAt     time.Time
			updatedAt     time.Time
			deletedAt     sql.NullTime
		)

		err := rows.Scan(
			&eventIDStr,
			&tenantIDStr,
			&eventName,
			&eventTypeStr,
			&description,
			&isActive,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		var deletedAtPtr *time.Time
		if deletedAt.Valid {
			deletedAtPtr = &deletedAt.Time
		}

		e, err := event.ReconstructEvent(
			common.EventID(eventIDStr),
			common.TenantID(tenantIDStr),
			eventName,
			event.EventType(eventTypeStr),
			description,
			isActive,
			createdAt,
			updatedAt,
			deletedAtPtr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct event: %w", err)
		}

		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %w", err)
	}

	return events, nil
}

// FindActiveByTenantID finds all active events within a tenant
func (r *EventRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*event.Event, error) {
	query := `
		SELECT
			event_id, tenant_id, event_name, event_type, description,
			is_active, created_at, updated_at, deleted_at
		FROM events
		WHERE tenant_id = $1 AND is_active = true AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find active events: %w", err)
	}
	defer rows.Close()

	var events []*event.Event
	for rows.Next() {
		var (
			eventIDStr    string
			tenantIDStr   string
			eventName     string
			eventTypeStr  string
			description   string
			isActive      bool
			createdAt     time.Time
			updatedAt     time.Time
			deletedAt     sql.NullTime
		)

		err := rows.Scan(
			&eventIDStr,
			&tenantIDStr,
			&eventName,
			&eventTypeStr,
			&description,
			&isActive,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event row: %w", err)
		}

		var deletedAtPtr *time.Time
		if deletedAt.Valid {
			deletedAtPtr = &deletedAt.Time
		}

		e, err := event.ReconstructEvent(
			common.EventID(eventIDStr),
			common.TenantID(tenantIDStr),
			eventName,
			event.EventType(eventTypeStr),
			description,
			isActive,
			createdAt,
			updatedAt,
			deletedAtPtr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct event: %w", err)
		}

		events = append(events, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating event rows: %w", err)
	}

	return events, nil
}

// Delete deletes an event (physical delete)
func (r *EventRepository) Delete(ctx context.Context, tenantID common.TenantID, eventID common.EventID) error {
	query := `
		DELETE FROM events
		WHERE tenant_id = $1 AND event_id = $2
	`

	result, err := r.db.Exec(ctx, query, tenantID.String(), eventID.String())
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return common.NewNotFoundError("Event", eventID.String())
	}

	return nil
}

// ExistsByName checks if an event with the given name exists within a tenant
func (r *EventRepository) ExistsByName(ctx context.Context, tenantID common.TenantID, eventName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM events
			WHERE tenant_id = $1 AND event_name = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, tenantID.String(), eventName).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check event existence: %w", err)
	}

	return exists, nil
}

