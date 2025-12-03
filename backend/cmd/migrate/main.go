package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed ../../internal/infra/db/migrations/*.sql
var migrationsFS embed.FS

const migrationsTableSQL = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version INT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`

type migration struct {
	version   int
	name      string
	upSQL     string
	downSQL   string
	isApplied bool
}

func main() {
	var (
		databaseURL = flag.String("database-url", os.Getenv("DATABASE_URL"), "PostgreSQL connection string")
		action      = flag.String("action", "up", "Migration action: up, down, status")
		steps       = flag.Int("steps", 0, "Number of migrations to apply/rollback (0 = all)")
	)
	flag.Parse()

	if *databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx := context.Background()

	// データベース接続
	pool, err := pgxpool.New(ctx, *databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// マイグレーションテーブルの初期化
	if _, err := pool.Exec(ctx, migrationsTableSQL); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// マイグレーションファイルの読み込み
	migrations, err := loadMigrations(ctx, pool)
	if err != nil {
		log.Fatalf("Failed to load migrations: %v", err)
	}

	// アクションに応じた処理
	switch *action {
	case "up":
		if err := migrateUp(ctx, pool, migrations, *steps); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
	case "down":
		if err := migrateDown(ctx, pool, migrations, *steps); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
	case "status":
		printStatus(migrations)
	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}

func loadMigrations(ctx context.Context, pool *pgxpool.Pool) ([]migration, error) {
	// マイグレーションファイルの一覧取得
	entries, err := migrationsFS.ReadDir("internal/infra/db/migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// バージョンごとにマイグレーションをグループ化
	migrationMap := make(map[int]*migration)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		// ファイル名から情報を抽出: 001_create_table.up.sql
		var version int
		var migrationName, direction string
		parts := strings.Split(name, "_")
		if len(parts) < 2 {
			continue
		}

		fmt.Sscanf(parts[0], "%d", &version)
		lastPart := parts[len(parts)-1]
		if strings.HasSuffix(lastPart, ".up.sql") {
			direction = "up"
			migrationName = strings.Join(parts[1:len(parts)-1], "_") + "_" + strings.TrimSuffix(lastPart, ".up.sql")
		} else if strings.HasSuffix(lastPart, ".down.sql") {
			direction = "down"
			migrationName = strings.Join(parts[1:len(parts)-1], "_") + "_" + strings.TrimSuffix(lastPart, ".down.sql")
		} else {
			continue
		}

		// SQLファイルの読み込み
		content, err := migrationsFS.ReadFile(filepath.Join("internal/infra/db/migrations", name))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", name, err)
		}

		// マイグレーションオブジェクトの作成/更新
		mig, exists := migrationMap[version]
		if !exists {
			mig = &migration{
				version: version,
				name:    migrationName,
			}
			migrationMap[version] = mig
		}

		if direction == "up" {
			mig.upSQL = string(content)
		} else {
			mig.downSQL = string(content)
		}
	}

	// 適用済みマイグレーションの取得
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	appliedVersions := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan migration version: %w", err)
		}
		appliedVersions[version] = true
	}

	// マイグレーション一覧の作成（バージョン順）
	var migrations []migration
	for version, mig := range migrationMap {
		mig.isApplied = appliedVersions[version]
		migrations = append(migrations, *mig)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})

	return migrations, nil
}

func migrateUp(ctx context.Context, pool *pgxpool.Pool, migrations []migration, steps int) error {
	applied := 0
	for _, mig := range migrations {
		if mig.isApplied {
			continue
		}

		if steps > 0 && applied >= steps {
			break
		}

		log.Printf("Applying migration %03d: %s", mig.version, mig.name)

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// マイグレーションの実行
		if _, err := tx.Exec(ctx, mig.upSQL); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %d: %w", mig.version, err)
		}

		// バージョンの記録
		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", mig.version); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %d: %w", mig.version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", mig.version, err)
		}

		log.Printf("✓ Applied migration %03d", mig.version)
		applied++
	}

	if applied == 0 {
		log.Println("No migrations to apply")
	} else {
		log.Printf("Successfully applied %d migration(s)", applied)
	}

	return nil
}

func migrateDown(ctx context.Context, pool *pgxpool.Pool, migrations []migration, steps int) error {
	// 逆順に処理
	rolledBack := 0
	for i := len(migrations) - 1; i >= 0; i-- {
		mig := migrations[i]
		if !mig.isApplied {
			continue
		}

		if steps > 0 && rolledBack >= steps {
			break
		}

		log.Printf("Rolling back migration %03d: %s", mig.version, mig.name)

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// ダウンマイグレーションの実行
		if _, err := tx.Exec(ctx, mig.downSQL); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to rollback migration %d: %w", mig.version, err)
		}

		// バージョンの削除
		if _, err := tx.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", mig.version); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to delete migration record %d: %w", mig.version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit rollback %d: %w", mig.version, err)
		}

		log.Printf("✓ Rolled back migration %03d", mig.version)
		rolledBack++
	}

	if rolledBack == 0 {
		log.Println("No migrations to roll back")
	} else {
		log.Printf("Successfully rolled back %d migration(s)", rolledBack)
	}

	return nil
}

func printStatus(migrations []migration) {
	fmt.Println("Migration Status:")
	fmt.Println("================")
	for _, mig := range migrations {
		status := "[ ]"
		if mig.isApplied {
			status = "[✓]"
		}
		fmt.Printf("%s %03d: %s\n", status, mig.version, mig.name)
	}
}

