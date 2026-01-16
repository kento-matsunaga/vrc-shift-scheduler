package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ShiftSlotTemplateRepository implements shift.ShiftSlotTemplateRepository for PostgreSQL
type ShiftSlotTemplateRepository struct {
	db *pgxpool.Pool
}

// NewShiftSlotTemplateRepository creates a new ShiftSlotTemplateRepository
func NewShiftSlotTemplateRepository(db *pgxpool.Pool) *ShiftSlotTemplateRepository {
	return &ShiftSlotTemplateRepository{db: db}
}

// Save saves a shift slot template with its items (insert or update)
func (r *ShiftSlotTemplateRepository) Save(ctx context.Context, template *shift.ShiftSlotTemplate) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Save template first (upsert)
	templateQuery := `
		INSERT INTO shift_slot_templates (
			template_id, tenant_id, event_id, template_name, description,
			created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (template_id) DO UPDATE SET
			template_name = EXCLUDED.template_name,
			description = EXCLUDED.description,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	var deletedAtValue interface{}
	if template.DeletedAt() != nil {
		deletedAtValue = *template.DeletedAt()
	} else {
		deletedAtValue = nil
	}

	result, err := tx.Exec(ctx, templateQuery,
		template.TemplateID().String(),
		template.TenantID().String(),
		template.EventID().String(),
		template.TemplateName(),
		template.Description(),
		template.CreatedAt(),
		template.UpdatedAt(),
		deletedAtValue,
	)
	if err != nil {
		return fmt.Errorf("failed to save template: %w", err)
	}

	// Debug: Check if template was saved
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("template save returned 0 rows affected")
	}

	// Delete existing items (CASCADE will handle this on template updates via ON CONFLICT)
	// Only delete if this is an update (template already exists)
	deleteItemsQuery := `
		DELETE FROM shift_slot_template_items
		WHERE template_id = $1
	`
	// Ignore errors here as the template might be new
	_, _ = tx.Exec(ctx, deleteItemsQuery, template.TemplateID().String())

	// Save items
	itemQuery := `
		INSERT INTO shift_slot_template_items (
			item_id, template_id, slot_name, instance_name,
			start_time, end_time, required_count, priority,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	for _, item := range template.Items() {
		// Debug: Verify template_id matches
		if item.TemplateID() != template.TemplateID() {
			return fmt.Errorf("template_id mismatch: item has %s, template has %s", item.TemplateID().String(), template.TemplateID().String())
		}

		_, err = tx.Exec(ctx, itemQuery,
			item.ItemID().String(),
			item.TemplateID().String(),
			item.SlotName(),
			item.InstanceName(),
			item.StartTime(),
			item.EndTime(),
			item.RequiredCount(),
			item.Priority(),
			item.CreatedAt(),
			item.UpdatedAt(),
		)
		if err != nil {
			return fmt.Errorf("failed to save template item (template_id=%s, item_id=%s): %w", item.TemplateID().String(), item.ItemID().String(), err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// FindByID finds a shift slot template by ID within a tenant
func (r *ShiftSlotTemplateRepository) FindByID(ctx context.Context, tenantID common.TenantID, templateID common.ShiftSlotTemplateID) (*shift.ShiftSlotTemplate, error) {
	// Find template
	templateQuery := `
		SELECT
			template_id, tenant_id, event_id, template_name, description,
			created_at, updated_at, deleted_at
		FROM shift_slot_templates
		WHERE tenant_id = $1 AND template_id = $2 AND deleted_at IS NULL
	`

	var (
		templateIDStr   string
		tenantIDStr     string
		eventIDStr      string
		templateName    string
		description     string
		createdAt       time.Time
		updatedAt       time.Time
		deletedAt       sql.NullTime
	)

	err := r.db.QueryRow(ctx, templateQuery, tenantID.String(), templateID.String()).Scan(
		&templateIDStr,
		&tenantIDStr,
		&eventIDStr,
		&templateName,
		&description,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("ShiftSlotTemplate", templateID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find template: %w", err)
	}

	// Find template items
	items, err := r.findItemsByTemplateID(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to find template items: %w", err)
	}

	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return shift.ReconstituteShiftSlotTemplate(
		common.ShiftSlotTemplateID(templateIDStr),
		common.TenantID(tenantIDStr),
		common.EventID(eventIDStr),
		templateName,
		description,
		items,
		createdAt,
		updatedAt,
		deletedAtPtr,
	), nil
}

// FindByEventID finds all shift slot templates for an event (excluding soft-deleted)
func (r *ShiftSlotTemplateRepository) FindByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*shift.ShiftSlotTemplate, error) {
	templateQuery := `
		SELECT
			template_id, tenant_id, event_id, template_name, description,
			created_at, updated_at, deleted_at
		FROM shift_slot_templates
		WHERE tenant_id = $1 AND event_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, templateQuery, tenantID.String(), eventID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []*shift.ShiftSlotTemplate
	for rows.Next() {
		var (
			templateIDStr string
			tenantIDStr   string
			eventIDStr    string
			templateName  string
			description   string
			createdAt     time.Time
			updatedAt     time.Time
			deletedAt     sql.NullTime
		)

		err := rows.Scan(
			&templateIDStr,
			&tenantIDStr,
			&eventIDStr,
			&templateName,
			&description,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template row: %w", err)
		}

		// Find items for this template
		items, err := r.findItemsByTemplateID(ctx, common.ShiftSlotTemplateID(templateIDStr))
		if err != nil {
			return nil, fmt.Errorf("failed to find template items: %w", err)
		}

		var deletedAtPtr *time.Time
		if deletedAt.Valid {
			deletedAtPtr = &deletedAt.Time
		}

		template := shift.ReconstituteShiftSlotTemplate(
			common.ShiftSlotTemplateID(templateIDStr),
			common.TenantID(tenantIDStr),
			common.EventID(eventIDStr),
			templateName,
			description,
			items,
			createdAt,
			updatedAt,
			deletedAtPtr,
		)

		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating template rows: %w", err)
	}

	return templates, nil
}

// Delete deletes a shift slot template (physical delete)
func (r *ShiftSlotTemplateRepository) Delete(ctx context.Context, tenantID common.TenantID, templateID common.ShiftSlotTemplateID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Delete items first (due to foreign key constraint)
	deleteItemsQuery := `
		DELETE FROM shift_slot_template_items
		WHERE template_id = $1
	`
	_, err = tx.Exec(ctx, deleteItemsQuery, templateID.String())
	if err != nil {
		return fmt.Errorf("failed to delete template items: %w", err)
	}

	// Delete template
	deleteTemplateQuery := `
		DELETE FROM shift_slot_templates
		WHERE tenant_id = $1 AND template_id = $2
	`
	result, err := tx.Exec(ctx, deleteTemplateQuery, tenantID.String(), templateID.String())
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	if result.RowsAffected() == 0 {
		return common.NewNotFoundError("ShiftSlotTemplate", templateID.String())
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// findItemsByTemplateID finds all items for a template
func (r *ShiftSlotTemplateRepository) findItemsByTemplateID(ctx context.Context, templateID common.ShiftSlotTemplateID) ([]*shift.ShiftSlotTemplateItem, error) {
	itemQuery := `
		SELECT
			item_id, template_id, slot_name, instance_name,
			start_time, end_time, required_count, priority,
			created_at, updated_at
		FROM shift_slot_template_items
		WHERE template_id = $1
		ORDER BY start_time ASC, priority ASC
	`

	rows, err := r.db.Query(ctx, itemQuery, templateID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query template items: %w", err)
	}
	defer rows.Close()

	var items []*shift.ShiftSlotTemplateItem
	for rows.Next() {
		var (
			itemIDStr     string
			templateIDStr string
			slotName      string
			instanceName  string
			startTime     time.Time
			endTime       time.Time
			requiredCount int
			priority      int
			createdAt     time.Time
			updatedAt     time.Time
		)

		err := rows.Scan(
			&itemIDStr,
			&templateIDStr,
			&slotName,
			&instanceName,
			&startTime,
			&endTime,
			&requiredCount,
			&priority,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template item row: %w", err)
		}

		item := shift.ReconstituteShiftSlotTemplateItem(
			common.ShiftSlotTemplateItemID(itemIDStr),
			common.ShiftSlotTemplateID(templateIDStr),
			slotName,
			instanceName,
			startTime,
			endTime,
			requiredCount,
			priority,
			createdAt,
			updatedAt,
		)

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating template item rows: %w", err)
	}

	return items, nil
}
