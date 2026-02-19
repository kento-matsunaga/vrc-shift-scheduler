package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/calendar"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CalendarEntryRepository implements calendar.CalendarEntryRepository for PostgreSQL
type CalendarEntryRepository struct {
	db *pgxpool.Pool
}

// NewCalendarEntryRepository creates a new CalendarEntryRepository
func NewCalendarEntryRepository(db *pgxpool.Pool) *CalendarEntryRepository {
	return &CalendarEntryRepository{db: db}
}

// Save saves a calendar entry (insert or update)
func (r *CalendarEntryRepository) Save(ctx context.Context, entry *calendar.CalendarEntry) error {
	// Convert nullable time.Time to sql.NullTime
	var startTime, endTime sql.NullTime
	if entry.StartTime() != nil {
		startTime = sql.NullTime{Time: *entry.StartTime(), Valid: true}
	}
	if entry.EndTime() != nil {
		endTime = sql.NullTime{Time: *entry.EndTime(), Valid: true}
	}

	_, err := r.db.Exec(ctx, `
		INSERT INTO calendar_entries (entry_id, calendar_id, tenant_id, title, entry_date, start_time, end_time, note, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (entry_id) DO UPDATE SET
			title = EXCLUDED.title,
			entry_date = EXCLUDED.entry_date,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			note = EXCLUDED.note,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`,
		entry.EntryID().String(),
		entry.CalendarID().String(),
		entry.TenantID().String(),
		entry.Title(),
		entry.Date(),
		startTime,
		endTime,
		entry.Note(),
		entry.CreatedAt(),
		entry.UpdatedAt(),
		entry.DeletedAt(),
	)
	if err != nil {
		return fmt.Errorf("failed to save calendar entry: %w", err)
	}

	return nil
}

// FindByID finds a calendar entry by ID within a tenant
func (r *CalendarEntryRepository) FindByID(ctx context.Context, tenantID common.TenantID, entryID common.CalendarEntryID) (*calendar.CalendarEntry, error) {
	var (
		entryIDStr    string
		calendarIDStr string
		tenantIDStr   string
		title         string
		entryDate     time.Time
		startTime     pgtype.Time
		endTime       pgtype.Time
		note          sql.NullString
		createdAt     time.Time
		updatedAt     time.Time
		deletedAt     sql.NullTime
	)

	err := r.db.QueryRow(ctx, `
		SELECT entry_id, calendar_id, tenant_id, title, entry_date, start_time, end_time, note, created_at, updated_at, deleted_at
		FROM calendar_entries
		WHERE entry_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, entryID.String(), tenantID.String()).Scan(
		&entryIDStr, &calendarIDStr, &tenantIDStr, &title, &entryDate, &startTime, &endTime, &note, &createdAt, &updatedAt, &deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("CalendarEntry", entryID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find calendar entry: %w", err)
	}

	return r.scanEntry(entryIDStr, calendarIDStr, tenantIDStr, title, entryDate, startTime, endTime, note, createdAt, updatedAt, deletedAt)
}

// FindByCalendarID finds all entries for a calendar (ordered by date)
func (r *CalendarEntryRepository) FindByCalendarID(ctx context.Context, tenantID common.TenantID, calendarID common.CalendarID) ([]*calendar.CalendarEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT entry_id, calendar_id, tenant_id, title, entry_date, start_time, end_time, note, created_at, updated_at, deleted_at
		FROM calendar_entries
		WHERE calendar_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
		ORDER BY entry_date ASC, start_time ASC NULLS LAST
	`, calendarID.String(), tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to find calendar entries: %w", err)
	}
	defer rows.Close()

	var entries []*calendar.CalendarEntry
	for rows.Next() {
		var (
			entryIDStr    string
			calendarIDStr string
			tenantIDStr   string
			title         string
			entryDate     time.Time
			startTime     pgtype.Time
			endTime       pgtype.Time
			note          sql.NullString
			createdAt     time.Time
			updatedAt     time.Time
			deletedAt     sql.NullTime
		)

		if err := rows.Scan(&entryIDStr, &calendarIDStr, &tenantIDStr, &title, &entryDate, &startTime, &endTime, &note, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, fmt.Errorf("failed to scan calendar entry row: %w", err)
		}

		entry, err := r.scanEntry(entryIDStr, calendarIDStr, tenantIDStr, title, entryDate, startTime, endTime, note, createdAt, updatedAt, deletedAt)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating calendar entry rows: %w", err)
	}

	return entries, nil
}

// Delete soft-deletes a calendar entry
func (r *CalendarEntryRepository) Delete(ctx context.Context, tenantID common.TenantID, entryID common.CalendarEntryID) error {
	now := time.Now()
	result, err := r.db.Exec(ctx, `
		UPDATE calendar_entries SET deleted_at = $1, updated_at = $1
		WHERE entry_id = $2 AND tenant_id = $3 AND deleted_at IS NULL
	`, now, entryID.String(), tenantID.String())
	if err != nil {
		return fmt.Errorf("failed to delete calendar entry: %w", err)
	}

	if result.RowsAffected() == 0 {
		return common.NewNotFoundError("CalendarEntry", entryID.String())
	}

	return nil
}

// scanEntry converts scanned values to a CalendarEntry entity
func (r *CalendarEntryRepository) scanEntry(
	entryIDStr string,
	calendarIDStr string,
	tenantIDStr string,
	title string,
	entryDate time.Time,
	startTime pgtype.Time,
	endTime pgtype.Time,
	note sql.NullString,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt sql.NullTime,
) (*calendar.CalendarEntry, error) {
	entryID, err := common.ParseCalendarEntryID(entryIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid entry ID: %w", err)
	}

	calID, err := common.ParseCalendarID(calendarIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid calendar ID: %w", err)
	}

	tenID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID: %w", err)
	}

	var startTimePtr, endTimePtr *time.Time
	if startTime.Valid {
		// pgtype.Time stores microseconds since midnight
		t := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(startTime.Microseconds) * time.Microsecond)
		startTimePtr = &t
	}
	if endTime.Valid {
		t := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(endTime.Microseconds) * time.Microsecond)
		endTimePtr = &t
	}

	var noteStr string
	if note.Valid {
		noteStr = note.String
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return calendar.ReconstructCalendarEntry(
		entryID,
		calID,
		tenID,
		title,
		entryDate,
		startTimePtr,
		endTimePtr,
		noteStr,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}
