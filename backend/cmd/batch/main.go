package main

import (
	"context"
	"flag"
	"log"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app/batch"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
)

// Config represents the application configuration
type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
}

func main() {
	// コマンドライン引数のパース
	taskFlag := flag.String("task", "", "Task to run: grace-expiry, webhook-cleanup, pending-cleanup")
	dryRun := flag.Bool("dry-run", false, "Dry run mode (no changes)")
	flag.Parse()

	if *taskFlag == "" {
		log.Fatal("Please specify a task with -task flag. Available tasks: grace-expiry, webhook-cleanup, pending-cleanup")
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

	// BatchProcessorを使用してタスクを実行
	processor := batch.NewBatchProcessor(pool, nil)

	switch *taskFlag {
	case "grace-expiry":
		result, err := processor.RunGraceExpiryCheck(ctx, *dryRun)
		if err != nil {
			log.Fatalf("Failed to run grace-expiry task: %v", err)
		}
		if !*dryRun && result.SuspendedCount > 0 {
			log.Printf("Summary: Suspended %d tenants, Failed %d", result.SuspendedCount, result.FailedCount)
		}

	case "webhook-cleanup":
		result, err := processor.RunWebhookCleanup(ctx, *dryRun)
		if err != nil {
			log.Fatalf("Failed to run webhook-cleanup task: %v", err)
		}
		if !*dryRun && result.DeletedCount > 0 {
			log.Printf("Summary: Deleted %d webhook logs", result.DeletedCount)
		}

	case "pending-cleanup":
		result, err := processor.RunPendingPaymentCleanup(ctx, *dryRun)
		if err != nil {
			log.Fatalf("Failed to run pending-cleanup task: %v", err)
		}
		if !*dryRun && result.DeletedCount > 0 {
			log.Printf("Summary: Deleted %d tenants, Failed %d", result.DeletedCount, result.FailedCount)
		}

	default:
		log.Fatalf("Unknown task: %s", *taskFlag)
	}

	log.Println("🎉 Batch processing completed!")
}
