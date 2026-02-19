package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/calendar"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CalendarRepository implements calendar.CalendarRepository for PostgreSQL
type CalendarRepository struct {
	db *pgxpool.Pool
}

// NewCalendarRepository creates a new CalendarRepository
func NewCalendarRepository(db *pgxpool.Pool) *CalendarRepository {
	return &CalendarRepository{db: db}
}

// Create saves a new calendar
func (r *CalendarRepository) Create(ctx context.Context, cal *calendar.Calendar) error {
	return r.save(ctx, cal)
}

// Update updates an existing calendar
func (r *CalendarRepository) Update(ctx context.Context, cal *calendar.Calendar) error {
	return r.save(ctx, cal)
}

// save saves a calendar (insert or update) - internal method
func (r *CalendarRepository) save(ctx context.Context, cal *calendar.Calendar) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			slog.Error("failed to rollback transaction", "error", err)
		}
	}()

	// Upsert calendar
	var publicToken *string
	if cal.PublicToken() != nil {
		s := cal.PublicToken().String()
		publicToken = &s
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO calendars (calendar_id, tenant_id, title, description, is_public, public_token, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (calendar_id) DO UPDATE SET
			title = EXCLUDED.title,
			description = EXCLUDED.description,
			is_public = EXCLUDED.is_public,
			public_token = EXCLUDED.public_token,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`, cal.CalendarID().String(), cal.TenantID().String(), cal.Title(), cal.Description(),
		cal.IsPublic(), publicToken, cal.CreatedAt(), cal.UpdatedAt(), cal.DeletedAt())
	if err != nil {
		return fmt.Errorf("failed to save calendar: %w", err)
	}

	// Delete existing event associations
	_, err = tx.Exec(ctx, `DELETE FROM calendar_events WHERE calendar_id = $1`, cal.CalendarID().String())
	if err != nil {
		return fmt.Errorf("failed to delete calendar events: %w", err)
	}

	// Insert event associations
	for _, eventID := range cal.EventIDs() {
		_, err = tx.Exec(ctx, `
			INSERT INTO calendar_events (calendar_id, event_id) VALUES ($1, $2)
		`, cal.CalendarID().String(), eventID.String())
		if err != nil {
			return fmt.Errorf("failed to save calendar event: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindByID finds a calendar by ID within a tenant
func (r *CalendarRepository) FindByID(ctx context.Context, tenantID common.TenantID, calendarID common.CalendarID) (*calendar.Calendar, error) {
	var (
		calendarIDStr string
		tenantIDStr   string
		title         string
		description   sql.NullString
		isPublic      bool
		publicToken   sql.NullString
		createdAt     time.Time
		updatedAt     time.Time
		deletedAt     sql.NullTime
	)

	err := r.db.QueryRow(ctx, `
		SELECT calendar_id, tenant_id, title, description, is_public, public_token, created_at, updated_at, deleted_at
		FROM calendars
		WHERE calendar_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, calendarID.String(), tenantID.String()).Scan(
		&calendarIDStr, &tenantIDStr, &title, &description, &isPublic, &publicToken, &createdAt, &updatedAt, &deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Calendar", calendarID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find calendar: %w", err)
	}

	eventIDs, err := r.findEventIDs(ctx, calendarID)
	if err != nil {
		return nil, err
	}

	return r.toDomain(calendarIDStr, tenantIDStr, title, description, isPublic, publicToken, eventIDs, createdAt, updatedAt, deletedAt)
}

// FindByTenantID finds all calendars within a tenant
func (r *CalendarRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*calendar.Calendar, error) {
	rows, err := r.db.Query(ctx, `
		SELECT calendar_id, tenant_id, title, description, is_public, public_token, created_at, updated_at, deleted_at
		FROM calendars
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find calendars: %w", err)
	}
	defer rows.Close()

	var calendars []*calendar.Calendar
	for rows.Next() {
		var (
			calendarIDStr string
			tenantIDStr   string
			title         string
			description   sql.NullString
			isPublic      bool
			publicToken   sql.NullString
			createdAt     time.Time
			updatedAt     time.Time
			deletedAt     sql.NullTime
		)

		if err := rows.Scan(&calendarIDStr, &tenantIDStr, &title, &description, &isPublic, &publicToken, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan calendar row: %w", err)
		}

		calID, err := common.ParseCalendarID(calendarIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid calendar ID: %w", err)
		}
		eventIDs, err := r.findEventIDs(ctx, calID)
		if err != nil {
			return nil, err
		}

		cal, err := r.toDomain(calendarIDStr, tenantIDStr, title, description, isPublic, publicToken, eventIDs, createdAt, updatedAt, deletedAt)
		if err != nil {
			return nil, err
		}
		calendars = append(calendars, cal)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating calendar rows: %w", err)
	}

	return calendars, nil
}

// FindByPublicToken finds a calendar by public token
func (r *CalendarRepository) FindByPublicToken(ctx context.Context, token common.PublicToken) (*calendar.Calendar, error) {
	var (
		calendarIDStr string
		tenantIDStr   string
		title         string
		description   sql.NullString
		isPublic      bool
		publicToken   sql.NullString
		createdAt     time.Time
		updatedAt     time.Time
		deletedAt     sql.NullTime
	)

	err := r.db.QueryRow(ctx, `
		SELECT calendar_id, tenant_id, title, description, is_public, public_token, created_at, updated_at, deleted_at
		FROM calendars
		WHERE public_token = $1 AND is_public = TRUE AND deleted_at IS NULL
	`, token.String()).Scan(
		&calendarIDStr, &tenantIDStr, &title, &description, &isPublic, &publicToken, &createdAt, &updatedAt, &deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Calendar", token.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find calendar by token: %w", err)
	}

	calID, err := common.ParseCalendarID(calendarIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid calendar ID: %w", err)
	}
	eventIDs, err := r.findEventIDs(ctx, calID)
	if err != nil {
		return nil, err
	}

	return r.toDomain(calendarIDStr, tenantIDStr, title, description, isPublic, publicToken, eventIDs, createdAt, updatedAt, deletedAt)
}

// Delete soft-deletes a calendar
func (r *CalendarRepository) Delete(ctx context.Context, tenantID common.TenantID, calendarID common.CalendarID) error {
	now := time.Now()
	result, err := r.db.Exec(ctx, `
		UPDATE calendars SET deleted_at = $1, updated_at = $1
		WHERE calendar_id = $2 AND tenant_id = $3 AND deleted_at IS NULL
	`, now, calendarID.String(), tenantID.String())
	if err != nil {
		return fmt.Errorf("failed to delete calendar: %w", err)
	}

	if result.RowsAffected() == 0 {
		return common.NewNotFoundError("Calendar", calendarID.String())
	}

	return nil
}

// findEventIDs finds event IDs associated with a calendar
func (r *CalendarRepository) findEventIDs(ctx context.Context, calendarID common.CalendarID) ([]common.EventID, error) {
	rows, err := r.db.Query(ctx, `
		SELECT event_id FROM calendar_events WHERE calendar_id = $1
	`, calendarID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find event IDs: %w", err)
	}
	defer rows.Close()

	var eventIDs []common.EventID
	for rows.Next() {
		var eventIDStr string
		if err := rows.Scan(&eventIDStr); err != nil {
			return nil, fmt.Errorf("failed to scan event ID: %w", err)
		}
		eventID, err := common.ParseEventID(eventIDStr)
		if err != nil {
			return nil, err
		}
		eventIDs = append(eventIDs, eventID)
	}

	return eventIDs, rows.Err()
}

// toDomain converts database values to domain entity
func (r *CalendarRepository) toDomain(
	calendarIDStr string,
	tenantIDStr string,
	title string,
	description sql.NullString,
	isPublic bool,
	publicToken sql.NullString,
	eventIDs []common.EventID,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt sql.NullTime,
) (*calendar.Calendar, error) {
	calID, err := common.ParseCalendarID(calendarIDStr)
	if err != nil {
		return nil, err
	}

	tenID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return nil, err
	}

	var desc string
	if description.Valid {
		desc = description.String
	}

	var token *common.PublicToken
	if publicToken.Valid {
		t, err := common.ParsePublicToken(publicToken.String)
		if err != nil {
			return nil, err
		}
		token = &t
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return calendar.ReconstructCalendar(
		calID,
		tenID,
		title,
		desc,
		isPublic,
		token,
		eventIDs,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}
