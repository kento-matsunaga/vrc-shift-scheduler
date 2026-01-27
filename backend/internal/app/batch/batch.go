package batch

import (
	"context"
	"log"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BatchProcessor handles batch processing tasks
type BatchProcessor struct {
	pool   *pgxpool.Pool
	logger Logger
}

// Logger interface for logging (allows mocking in tests)
type Logger interface {
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}

// DefaultLogger implements Logger using standard log package
type DefaultLogger struct{}

func (l *DefaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (l *DefaultLogger) Println(v ...interface{}) {
	log.Println(v...)
}

// NewBatchProcessor creates a new BatchProcessor
func NewBatchProcessor(pool *pgxpool.Pool, logger Logger) *BatchProcessor {
	if logger == nil {
		logger = &DefaultLogger{}
	}
	return &BatchProcessor{
		pool:   pool,
		logger: logger,
	}
}

// ExpiredGraceTenant represents a tenant with expired grace period
type ExpiredGraceTenant struct {
	TenantID   string
	TenantName string
	GraceUntil time.Time
}

// GraceExpiryResult contains the result of grace expiry check
type GraceExpiryResult struct {
	ExpiredTenants []ExpiredGraceTenant
	SuspendedCount int
	FailedCount    int
}

// RunGraceExpiryCheck checks tenants in grace period and suspends them if expired
func (b *BatchProcessor) RunGraceExpiryCheck(ctx context.Context, dryRun bool) (*GraceExpiryResult, error) {
	return b.RunGraceExpiryCheckAt(ctx, time.Now(), dryRun)
}

// RunGraceExpiryCheckAt checks tenants with grace period expired before the given time
func (b *BatchProcessor) RunGraceExpiryCheckAt(ctx context.Context, now time.Time, dryRun bool) (*GraceExpiryResult, error) {
	b.logger.Println("📋 Running grace period expiry check...")

	query := `
		SELECT tenant_id, tenant_name, grace_until
		FROM tenants
		WHERE status = 'grace'
		AND grace_until IS NOT NULL
		AND grace_until < $1
	`

	rows, err := b.pool.Query(ctx, query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := &GraceExpiryResult{}

	for rows.Next() {
		var t ExpiredGraceTenant
		if err := rows.Scan(&t.TenantID, &t.TenantName, &t.GraceUntil); err != nil {
			return nil, err
		}
		result.ExpiredTenants = append(result.ExpiredTenants, t)
	}

	if len(result.ExpiredTenants) == 0 {
		b.logger.Println("   ✅ No expired grace period tenants found")
		return result, nil
	}

	b.logger.Printf("   ⚠️ Found %d tenants with expired grace period", len(result.ExpiredTenants))

	for _, t := range result.ExpiredTenants {
		b.logger.Printf("   - %s (%s) - grace ended at %s", t.TenantName, t.TenantID, t.GraceUntil.Format(time.RFC3339))

		if !dryRun {
			if err := b.suspendTenant(ctx, t.TenantID, now); err != nil {
				b.logger.Printf("   ❌ Failed to suspend tenant %s: %v", t.TenantID, err)
				result.FailedCount++
				continue
			}
			b.logger.Printf("   ✅ Suspended tenant %s", t.TenantID)
			result.SuspendedCount++
		} else {
			b.logger.Printf("   🔍 [DRY RUN] Would suspend tenant %s", t.TenantID)
		}
	}

	return result, nil
}

func (b *BatchProcessor) suspendTenant(ctx context.Context, tenantID string, now time.Time) error {
	tx, err := b.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Update tenant status
	updateQuery := `
		UPDATE tenants
		SET status = 'suspended', updated_at = $1
		WHERE tenant_id = $2
	`
	if _, err := tx.Exec(ctx, updateQuery, now, tenantID); err != nil {
		return err
	}

	// Insert audit log
	logID := common.NewULID()
	auditQuery := `
		INSERT INTO billing_audit_logs (log_id, actor_type, action, target_type, target_id, created_at)
		VALUES ($1, 'system', 'tenant_suspended', 'tenant', $2, $3)
	`
	if _, err := tx.Exec(ctx, auditQuery, logID, tenantID, now); err != nil {
		// Log but don't fail
		b.logger.Printf("   ⚠️ Failed to log audit for tenant %s: %v", tenantID, err)
	}

	return tx.Commit(ctx)
}

// WebhookCleanupResult contains the result of webhook cleanup
type WebhookCleanupResult struct {
	TotalCount   int
	DeletedCount int64
}

// RunWebhookCleanup cleans up old webhook logs (older than 30 days)
func (b *BatchProcessor) RunWebhookCleanup(ctx context.Context, dryRun bool) (*WebhookCleanupResult, error) {
	return b.RunWebhookCleanupOlderThan(ctx, 30, dryRun)
}

// RunWebhookCleanupOlderThan cleans up webhook logs older than the specified days
func (b *BatchProcessor) RunWebhookCleanupOlderThan(ctx context.Context, days int, dryRun bool) (*WebhookCleanupResult, error) {
	b.logger.Println("🧹 Running webhook cleanup...")

	cutoffDate := time.Now().AddDate(0, 0, -days)

	// Count logs to be deleted
	countQuery := `
		SELECT COUNT(*)
		FROM stripe_webhook_logs
		WHERE received_at < $1
	`
	var count int
	if err := b.pool.QueryRow(ctx, countQuery, cutoffDate).Scan(&count); err != nil {
		return nil, err
	}

	result := &WebhookCleanupResult{TotalCount: count}

	if count == 0 {
		b.logger.Println("   ✅ No old webhook logs to clean up")
		return result, nil
	}

	b.logger.Printf("   ⚠️ Found %d webhook logs older than %s", count, cutoffDate.Format("2006-01-02"))

	if !dryRun {
		deleteQuery := `
			DELETE FROM stripe_webhook_logs
			WHERE received_at < $1
		`
		cmdTag, err := b.pool.Exec(ctx, deleteQuery, cutoffDate)
		if err != nil {
			return nil, err
		}

		result.DeletedCount = cmdTag.RowsAffected()
		b.logger.Printf("   ✅ Deleted %d old webhook logs", result.DeletedCount)
	} else {
		b.logger.Printf("   🔍 [DRY RUN] Would delete %d old webhook logs", count)
	}

	return result, nil
}

// ExpiredPendingTenant represents a tenant with expired pending payment
type ExpiredPendingTenant struct {
	TenantID         string
	TenantName       string
	PendingExpiresAt time.Time
}

// PendingCleanupResult contains the result of pending payment cleanup
type PendingCleanupResult struct {
	ExpiredTenants []ExpiredPendingTenant
	DeletedCount   int
	FailedCount    int
}

// RunPendingPaymentCleanup cleans up expired pending_payment tenants
func (b *BatchProcessor) RunPendingPaymentCleanup(ctx context.Context, dryRun bool) (*PendingCleanupResult, error) {
	return b.RunPendingPaymentCleanupAt(ctx, time.Now(), dryRun)
}

// RunPendingPaymentCleanupAt cleans up tenants with pending payment expired before the given time
func (b *BatchProcessor) RunPendingPaymentCleanupAt(ctx context.Context, now time.Time, dryRun bool) (*PendingCleanupResult, error) {
	b.logger.Println("🧹 Running pending payment cleanup...")

	query := `
		SELECT tenant_id, tenant_name, pending_expires_at
		FROM tenants
		WHERE status = 'pending_payment'
		AND pending_expires_at IS NOT NULL
		AND pending_expires_at < $1
	`

	rows, err := b.pool.Query(ctx, query, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := &PendingCleanupResult{}

	for rows.Next() {
		var t ExpiredPendingTenant
		if err := rows.Scan(&t.TenantID, &t.TenantName, &t.PendingExpiresAt); err != nil {
			return nil, err
		}
		result.ExpiredTenants = append(result.ExpiredTenants, t)
	}

	if len(result.ExpiredTenants) == 0 {
		b.logger.Println("   ✅ No expired pending payment tenants found")
		return result, nil
	}

	b.logger.Printf("   ⚠️ Found %d tenants with expired pending payment", len(result.ExpiredTenants))

	for _, t := range result.ExpiredTenants {
		b.logger.Printf("   - %s (%s) - expired at %s", t.TenantName, t.TenantID, t.PendingExpiresAt.Format(time.RFC3339))

		if !dryRun {
			if err := b.deletePendingTenant(ctx, t.TenantID, now); err != nil {
				b.logger.Printf("   ❌ Failed to delete tenant %s: %v", t.TenantID, err)
				result.FailedCount++
				continue
			}
			b.logger.Printf("   ✅ Deleted expired pending tenant %s", t.TenantID)
			result.DeletedCount++
		} else {
			b.logger.Printf("   🔍 [DRY RUN] Would delete tenant %s and associated admins", t.TenantID)
		}
	}

	return result, nil
}

func (b *BatchProcessor) deletePendingTenant(ctx context.Context, tenantID string, now time.Time) error {
	tx, err := b.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Delete associated admins
	deleteAdminsQuery := `DELETE FROM admins WHERE tenant_id = $1`
	if _, err := tx.Exec(ctx, deleteAdminsQuery, tenantID); err != nil {
		return err
	}

	// Delete tenant
	deleteTenantQuery := `DELETE FROM tenants WHERE tenant_id = $1`
	if _, err := tx.Exec(ctx, deleteTenantQuery, tenantID); err != nil {
		return err
	}

	// Insert audit log
	logID := common.NewULID()
	auditQuery := `
		INSERT INTO billing_audit_logs (log_id, actor_type, action, target_type, target_id, after_json, created_at)
		VALUES ($1, 'system', 'tenant_deleted', 'tenant', $2, $3, $4)
	`
	afterJSON := `{"reason":"pending_payment_expired"}`
	if _, err := tx.Exec(ctx, auditQuery, logID, tenantID, afterJSON, now); err != nil {
		b.logger.Printf("   ⚠️ Failed to log audit for tenant %s: %v", tenantID, err)
		// Continue - audit log failure is not critical
	}

	return tx.Commit(ctx)
}
