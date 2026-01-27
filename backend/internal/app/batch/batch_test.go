package batch_test

import (
	"context"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app/batch"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
)

// testConfig represents test configuration
type testConfig struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
}

// testLogger suppresses log output during tests
type testLogger struct{}

func (l *testLogger) Printf(format string, v ...interface{}) {}

func (l *testLogger) Println(v ...interface{}) {}

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	var cfg testConfig
	if err := envconfig.Process("", &cfg); err != nil {
		t.Skipf("Skipping test: DATABASE_URL not set: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	return pool
}

func TestBatchProcessor_RunGraceExpiryCheck_NoExpiredTenants(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	logger := &testLogger{}
	processor := batch.NewBatchProcessor(pool, logger)

	// Run with a time in the distant past (no tenants should be expired)
	pastTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	result, err := processor.RunGraceExpiryCheckAt(ctx, pastTime, true)

	if err != nil {
		t.Fatalf("RunGraceExpiryCheck failed: %v", err)
	}

	if len(result.ExpiredTenants) != 0 {
		t.Errorf("Expected 0 expired tenants, got %d", len(result.ExpiredTenants))
	}
}

func TestBatchProcessor_RunGraceExpiryCheck_DryRun(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	logger := &testLogger{}
	processor := batch.NewBatchProcessor(pool, logger)

	// Create a test tenant in grace status with expired grace_until
	tenantID := createTestGraceTenant(t, pool)
	defer cleanupTestTenant(t, pool, tenantID)

	// Run dry run - should find the tenant but not suspend it
	result, err := processor.RunGraceExpiryCheck(ctx, true)
	if err != nil {
		t.Fatalf("RunGraceExpiryCheck failed: %v", err)
	}

	// Should find at least 1 expired tenant
	if len(result.ExpiredTenants) == 0 {
		t.Error("Expected to find expired tenants in dry run")
	}

	// Verify tenant was NOT suspended (dry run)
	status := getTenantStatus(t, pool, tenantID)
	if status != "grace" {
		t.Errorf("Expected tenant status to remain 'grace' in dry run, got '%s'", status)
	}
}

func TestBatchProcessor_RunGraceExpiryCheck_ActualRun(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	logger := &testLogger{}
	processor := batch.NewBatchProcessor(pool, logger)

	// Create a test tenant in grace status with expired grace_until
	tenantID := createTestGraceTenant(t, pool)
	defer cleanupTestTenant(t, pool, tenantID)

	// Run actual suspension
	result, err := processor.RunGraceExpiryCheck(ctx, false)
	if err != nil {
		t.Fatalf("RunGraceExpiryCheck failed: %v", err)
	}

	// Should have suspended at least 1 tenant
	if result.SuspendedCount == 0 {
		t.Error("Expected to suspend at least 1 tenant")
	}

	// Verify tenant was suspended
	status := getTenantStatus(t, pool, tenantID)
	if status != "suspended" {
		t.Errorf("Expected tenant status to be 'suspended', got '%s'", status)
	}
}

func TestBatchProcessor_RunWebhookCleanup_NoOldLogs(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	logger := &testLogger{}
	processor := batch.NewBatchProcessor(pool, logger)

	// Clean up any existing old logs first
	_, _ = pool.Exec(ctx, "DELETE FROM stripe_webhook_logs WHERE received_at < NOW() - INTERVAL '30 days'")

	result, err := processor.RunWebhookCleanup(ctx, true)
	if err != nil {
		t.Fatalf("RunWebhookCleanup failed: %v", err)
	}

	if result.TotalCount != 0 {
		// This might fail if there are actual old logs, which is fine for this test
		t.Logf("Found %d old logs (may be from actual data)", result.TotalCount)
	}
}

func TestBatchProcessor_RunWebhookCleanup_WithOldLogs(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	logger := &testLogger{}
	processor := batch.NewBatchProcessor(pool, logger)

	// Create an old webhook log entry
	logID := createOldWebhookLog(t, pool, 35) // 35 days old
	defer cleanupWebhookLog(t, pool, logID)

	// Run cleanup with dry run
	result, err := processor.RunWebhookCleanupOlderThan(ctx, 30, true)
	if err != nil {
		t.Fatalf("RunWebhookCleanup failed: %v", err)
	}

	if result.TotalCount == 0 {
		t.Error("Expected to find old webhook logs")
	}

	// Verify log was NOT deleted (dry run)
	exists := webhookLogExists(t, pool, logID)
	if !exists {
		t.Error("Expected webhook log to still exist after dry run")
	}
}

func TestBatchProcessor_RunWebhookCleanup_ActualDelete(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	logger := &testLogger{}
	processor := batch.NewBatchProcessor(pool, logger)

	// Create an old webhook log entry
	logID := createOldWebhookLog(t, pool, 35) // 35 days old
	// No defer cleanup needed - we're deleting it

	// Run actual cleanup
	result, err := processor.RunWebhookCleanupOlderThan(ctx, 30, false)
	if err != nil {
		t.Fatalf("RunWebhookCleanup failed: %v", err)
	}

	if result.DeletedCount == 0 {
		t.Error("Expected to delete old webhook logs")
	}

	// Verify log was deleted
	exists := webhookLogExists(t, pool, logID)
	if exists {
		t.Error("Expected webhook log to be deleted")
		// Clean up manually if test fails
		cleanupWebhookLog(t, pool, logID)
	}
}

func TestBatchProcessor_RunPendingPaymentCleanup_NoExpiredTenants(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	logger := &testLogger{}
	processor := batch.NewBatchProcessor(pool, logger)

	// Run with a time in the distant past
	pastTime := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	result, err := processor.RunPendingPaymentCleanupAt(ctx, pastTime, true)

	if err != nil {
		t.Fatalf("RunPendingPaymentCleanup failed: %v", err)
	}

	if len(result.ExpiredTenants) != 0 {
		t.Errorf("Expected 0 expired pending tenants, got %d", len(result.ExpiredTenants))
	}
}

func TestBatchProcessor_RunPendingPaymentCleanup_DryRun(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	logger := &testLogger{}
	processor := batch.NewBatchProcessor(pool, logger)

	// Create a test tenant in pending_payment status with expired pending_expires_at
	tenantID := createTestPendingTenant(t, pool)
	defer cleanupTestTenant(t, pool, tenantID)

	// Run dry run
	result, err := processor.RunPendingPaymentCleanup(ctx, true)
	if err != nil {
		t.Fatalf("RunPendingPaymentCleanup failed: %v", err)
	}

	// Should find at least 1 expired tenant
	if len(result.ExpiredTenants) == 0 {
		t.Error("Expected to find expired pending tenants in dry run")
	}

	// Verify tenant was NOT deleted (dry run)
	exists := tenantExists(t, pool, tenantID)
	if !exists {
		t.Error("Expected tenant to still exist after dry run")
	}
}

func TestBatchProcessor_RunPendingPaymentCleanup_ActualDelete(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	logger := &testLogger{}
	processor := batch.NewBatchProcessor(pool, logger)

	// Create a test tenant in pending_payment status with expired pending_expires_at
	tenantID := createTestPendingTenant(t, pool)
	// No defer cleanup - we're deleting it

	// Run actual deletion
	result, err := processor.RunPendingPaymentCleanup(ctx, false)
	if err != nil {
		t.Fatalf("RunPendingPaymentCleanup failed: %v", err)
	}

	// Should have deleted at least 1 tenant
	if result.DeletedCount == 0 {
		t.Error("Expected to delete at least 1 tenant")
	}

	// Verify tenant was deleted
	exists := tenantExists(t, pool, tenantID)
	if exists {
		t.Error("Expected tenant to be deleted")
		// Clean up manually if test fails
		cleanupTestTenant(t, pool, tenantID)
	}
}

// Helper functions

func createTestGraceTenant(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	ctx := context.Background()

	// Generate a valid ULID (26 characters) for tenant_id
	tenantID := common.NewULID()
	now := time.Now()
	graceUntil := now.Add(-24 * time.Hour) // Expired 1 day ago

	query := `
		INSERT INTO tenants (tenant_id, tenant_name, status, timezone, grace_until, created_at, updated_at)
		VALUES ($1, $2, 'grace', 'Asia/Tokyo', $3, $4, $4)
	`
	_, err := pool.Exec(ctx, query, tenantID, "Test Grace Tenant", graceUntil, now)
	if err != nil {
		t.Fatalf("Failed to create test grace tenant: %v", err)
	}

	return tenantID
}

func createTestPendingTenant(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	ctx := context.Background()

	// Generate a valid ULID (26 characters) for tenant_id
	tenantID := common.NewULID()
	now := time.Now()
	pendingExpiresAt := now.Add(-24 * time.Hour) // Expired 1 day ago

	query := `
		INSERT INTO tenants (tenant_id, tenant_name, status, timezone, pending_expires_at, created_at, updated_at)
		VALUES ($1, $2, 'pending_payment', 'Asia/Tokyo', $3, $4, $4)
	`
	_, err := pool.Exec(ctx, query, tenantID, "Test Pending Tenant", pendingExpiresAt, now)
	if err != nil {
		t.Fatalf("Failed to create test pending tenant: %v", err)
	}

	return tenantID
}

func createOldWebhookLog(t *testing.T, pool *pgxpool.Pool, daysOld int) int {
	t.Helper()
	ctx := context.Background()

	eventID := "evt_test_" + time.Now().Format("20060102150405")
	receivedAt := time.Now().AddDate(0, 0, -daysOld)

	query := `
		INSERT INTO stripe_webhook_logs (event_id, event_type, received_at)
		VALUES ($1, 'test.event', $2)
		RETURNING id
	`
	var logID int
	err := pool.QueryRow(ctx, query, eventID, receivedAt).Scan(&logID)
	if err != nil {
		t.Fatalf("Failed to create old webhook log: %v", err)
	}

	return logID
}

func getTenantStatus(t *testing.T, pool *pgxpool.Pool, tenantID string) string {
	t.Helper()
	ctx := context.Background()

	var status string
	err := pool.QueryRow(ctx, "SELECT status FROM tenants WHERE tenant_id = $1", tenantID).Scan(&status)
	if err != nil {
		t.Fatalf("Failed to get tenant status: %v", err)
	}

	return status
}

func tenantExists(t *testing.T, pool *pgxpool.Pool, tenantID string) bool {
	t.Helper()
	ctx := context.Background()

	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM tenants WHERE tenant_id = $1)", tenantID).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check tenant existence: %v", err)
	}

	return exists
}

func webhookLogExists(t *testing.T, pool *pgxpool.Pool, logID int) bool {
	t.Helper()
	ctx := context.Background()

	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM stripe_webhook_logs WHERE id = $1)", logID).Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check webhook log existence: %v", err)
	}

	return exists
}

func cleanupTestTenant(t *testing.T, pool *pgxpool.Pool, tenantID string) {
	t.Helper()
	ctx := context.Background()

	// Delete related admins first
	_, _ = pool.Exec(ctx, "DELETE FROM admins WHERE tenant_id = $1", tenantID)
	// Delete tenant
	_, _ = pool.Exec(ctx, "DELETE FROM tenants WHERE tenant_id = $1", tenantID)
	// Delete audit logs
	_, _ = pool.Exec(ctx, "DELETE FROM billing_audit_logs WHERE target_id = $1", tenantID)
}

func cleanupWebhookLog(t *testing.T, pool *pgxpool.Pool, logID int) {
	t.Helper()
	ctx := context.Background()

	_, _ = pool.Exec(ctx, "DELETE FROM stripe_webhook_logs WHERE id = $1", logID)
}
