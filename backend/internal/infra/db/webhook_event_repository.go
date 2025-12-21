package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WebhookEventRepository implements billing.WebhookEventRepository for PostgreSQL
type WebhookEventRepository struct {
	db *pgxpool.Pool
}

// NewWebhookEventRepository creates a new WebhookEventRepository
func NewWebhookEventRepository(db *pgxpool.Pool) *WebhookEventRepository {
	return &WebhookEventRepository{db: db}
}

// TryInsert attempts to insert a webhook event, returns false if already exists
func (r *WebhookEventRepository) TryInsert(ctx context.Context, provider string, eventID string, payloadJSON *string) (bool, error) {
	query := `
		INSERT INTO webhook_events (provider, event_id, payload_json)
		VALUES ($1, $2, $3)
		ON CONFLICT (provider, event_id) DO NOTHING
	`

	result, err := r.db.Exec(ctx, query, provider, eventID, payloadJSON)
	if err != nil {
		return false, fmt.Errorf("failed to insert webhook event: %w", err)
	}

	// If rows affected is 0, the event already existed
	return result.RowsAffected() > 0, nil
}

// DeleteOlderThan deletes webhook events older than the specified number of days
func (r *WebhookEventRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	query := `
		DELETE FROM webhook_events
		WHERE received_at < NOW() - INTERVAL '1 day' * $1
	`

	result, err := r.db.Exec(ctx, query, days)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old webhook events: %w", err)
	}

	return result.RowsAffected(), nil
}
