package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
)

// Config represents the application configuration
type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
}

func main() {
	// コマンドライン引数のパース
	taskFlag := flag.String("task", "", "Task to run: grace-expiry, webhook-cleanup")
	dryRun := flag.Bool("dry-run", false, "Dry run mode (no changes)")
	flag.Parse()

	if *taskFlag == "" {
		log.Fatal("Please specify a task with -task flag. Available tasks: grace-expiry, webhook-cleanup")
	}

	log.Printf("🔄 VRC Shift Scheduler - Batch Processing")
	log.Printf("Task: %s", *taskFlag)
	log.Printf("Dry Run: %v", *dryRun)

	// 環境変数の読み込み
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Failed to process env vars: %v", err)
	}

	// データベース接続
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("✅ Database connected")

	// タスクを実行
	switch *taskFlag {
	case "grace-expiry":
		if err := runGraceExpiryCheck(ctx, pool, *dryRun); err != nil {
			log.Fatalf("Failed to run grace-expiry task: %v", err)
		}
	case "webhook-cleanup":
		if err := runWebhookCleanup(ctx, pool, *dryRun); err != nil {
			log.Fatalf("Failed to run webhook-cleanup task: %v", err)
		}
	default:
		log.Fatalf("Unknown task: %s", *taskFlag)
	}

	log.Println("🎉 Batch processing completed!")
}

// runGraceExpiryCheck checks tenants in grace period and suspends them if expired
func runGraceExpiryCheck(ctx context.Context, pool *pgxpool.Pool, dryRun bool) error {
	log.Println("📋 Running grace period expiry check...")

	// tenantsテーブルからgrace状態かつgrace_untilが過去のテナントを取得
	query := `
		SELECT tenant_id, tenant_name, grace_until
		FROM tenants
		WHERE status = 'grace'
		AND grace_until IS NOT NULL
		AND grace_until < $1
	`
	now := time.Now()

	rows, err := pool.Query(ctx, query, now)
	if err != nil {
		return err
	}
	defer rows.Close()

	var expiredTenants []struct {
		tenantID   string
		tenantName string
		graceUntil time.Time
	}

	for rows.Next() {
		var t struct {
			tenantID   string
			tenantName string
			graceUntil time.Time
		}
		if err := rows.Scan(&t.tenantID, &t.tenantName, &t.graceUntil); err != nil {
			return err
		}
		expiredTenants = append(expiredTenants, t)
	}

	if len(expiredTenants) == 0 {
		log.Println("   ✅ No expired grace period tenants found")
		return nil
	}

	log.Printf("   ⚠️ Found %d tenants with expired grace period", len(expiredTenants))

	for _, t := range expiredTenants {
		log.Printf("   - %s (%s) - grace ended at %s", t.tenantName, t.tenantID, t.graceUntil.Format(time.RFC3339))

		if !dryRun {
			// ステータスをsuspendedに更新
			updateQuery := `
				UPDATE tenants
				SET status = 'suspended', updated_at = $1
				WHERE tenant_id = $2
			`
			if _, err := pool.Exec(ctx, updateQuery, now, t.tenantID); err != nil {
				log.Printf("   ❌ Failed to suspend tenant %s: %v", t.tenantID, err)
				continue
			}

			// 監査ログを記録
			logID := common.NewULID()
			auditQuery := `
				INSERT INTO billing_audit_logs (log_id, actor_type, action, target_type, target_id, created_at)
				VALUES ($1, 'system', 'tenant_suspended', 'tenant', $2, $3)
			`
			if _, err := pool.Exec(ctx, auditQuery, logID, t.tenantID, now); err != nil {
				log.Printf("   ⚠️ Failed to log audit for tenant %s: %v", t.tenantID, err)
			}

			log.Printf("   ✅ Suspended tenant %s", t.tenantID)
		} else {
			log.Printf("   🔍 [DRY RUN] Would suspend tenant %s", t.tenantID)
		}
	}

	return nil
}

// runWebhookCleanup cleans up old webhook logs
func runWebhookCleanup(ctx context.Context, pool *pgxpool.Pool, dryRun bool) error {
	log.Println("🧹 Running webhook cleanup...")

	// 30日より古いwebhookログを削除
	cutoffDate := time.Now().AddDate(0, 0, -30)

	// まず削除対象の件数を確認
	countQuery := `
		SELECT COUNT(*)
		FROM stripe_webhook_logs
		WHERE received_at < $1
	`
	var count int
	if err := pool.QueryRow(ctx, countQuery, cutoffDate).Scan(&count); err != nil {
		return err
	}

	if count == 0 {
		log.Println("   ✅ No old webhook logs to clean up")
		return nil
	}

	log.Printf("   ⚠️ Found %d webhook logs older than %s", count, cutoffDate.Format("2006-01-02"))

	if !dryRun {
		// 古いログを削除
		deleteQuery := `
			DELETE FROM stripe_webhook_logs
			WHERE received_at < $1
		`
		result, err := pool.Exec(ctx, deleteQuery, cutoffDate)
		if err != nil {
			return err
		}

		log.Printf("   ✅ Deleted %d old webhook logs", result.RowsAffected())
	} else {
		log.Printf("   🔍 [DRY RUN] Would delete %d old webhook logs", count)
	}

	return nil
}
