package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
)

// Config represents the application configuration
type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
}

func main() {
	// コマンドライン引数のパース
	envFlag := flag.String("env", "development", "Environment (development, staging, production)")
	tenantCount := flag.Int("tenants", 1, "Number of tenants to create")
	flag.Parse()

	log.Printf("🌱 VRC Shift Scheduler - Seed Data Generator")
	log.Printf("Environment: %s", *envFlag)
	log.Printf("Tenant Count: %d", *tenantCount)

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

	// シードデータを生成
	if err := seedData(ctx, pool, *tenantCount); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}

	log.Println("🎉 Seed data generation completed!")
}

func seedData(ctx context.Context, pool *pgxpool.Pool, tenantCount int) error {
	// リポジトリの初期化
	eventRepo := db.NewEventRepository(pool)
	businessDayRepo := db.NewEventBusinessDayRepository(pool)
	memberRepo := db.NewMemberRepository(pool)
	slotRepo := db.NewShiftSlotRepository(pool)

	for i := 0; i < tenantCount; i++ {
		tenantID := common.NewTenantID()
		log.Printf("\n📦 Creating tenant %d/%d: %s", i+1, tenantCount, tenantID)

		// 0. テナントを作成
		if err := createTenant(ctx, pool, tenantID, fmt.Sprintf("テストテナント #%d", i+1)); err != nil {
			return fmt.Errorf("failed to create tenant: %w", err)
		}
		log.Printf("   ✅ Tenant created: %s", tenantID)

		// 1. イベントを作成
		eventID, err := createEvent(ctx, eventRepo, tenantID, fmt.Sprintf("テストイベント #%d", i+1))
		if err != nil {
			return fmt.Errorf("failed to create event: %w", err)
		}
		log.Printf("   ✅ Event created: %s", eventID)

		// 2. 営業日を作成（7日間）
		businessDayIDs, err := createBusinessDays(ctx, businessDayRepo, tenantID, eventID, 7)
		if err != nil {
			return fmt.Errorf("failed to create business days: %w", err)
		}
		log.Printf("   ✅ Business days created: %d", len(businessDayIDs))

		// 3. メンバーを作成（5人）
		memberIDs, err := createMembers(ctx, memberRepo, tenantID, 5)
		if err != nil {
			return fmt.Errorf("failed to create members: %w", err)
		}
		log.Printf("   ✅ Members created: %d", len(memberIDs))

		// 4. ポジションを作成
		positionIDs, err := createPositions(ctx, pool, tenantID)
		if err != nil {
			return fmt.Errorf("failed to create positions: %w", err)
		}
		log.Printf("   ✅ Positions created: %d", len(positionIDs))

		// 5. シフト枠を作成（各営業日に2〜3枠）
		totalSlots := 0
		for _, bdID := range businessDayIDs {
			slots, err := createShiftSlots(ctx, slotRepo, tenantID, bdID, positionIDs)
			if err != nil {
				return fmt.Errorf("failed to create shift slots: %w", err)
			}
			totalSlots += len(slots)
		}
		log.Printf("   ✅ Shift slots created: %d", totalSlots)
	}

	return nil
}

func createTenant(ctx context.Context, pool *pgxpool.Pool, tenantID common.TenantID, name string) error {
	query := `
		INSERT INTO tenants (tenant_id, tenant_name, timezone, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (tenant_id) DO NOTHING
	`
	now := time.Now()
	_, err := pool.Exec(ctx, query, string(tenantID), name, "Asia/Tokyo", true, now)
	return err
}

func createPositions(ctx context.Context, pool *pgxpool.Pool, tenantID common.TenantID) ([]shift.PositionID, error) {
	positions := []struct {
		name        string
		description string
	}{
		{"受付", "来場者の受付業務"},
		{"案内", "イベント会場の案内業務"},
		{"配信", "イベントの配信サポート業務"},
	}

	ids := make([]shift.PositionID, 0, len(positions))
	now := time.Now()

	for i, pos := range positions {
		positionID := shift.NewPositionID()
		query := `
			INSERT INTO positions (position_id, tenant_id, position_name, description, display_order, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
			ON CONFLICT (position_id) DO NOTHING
		`
		_, err := pool.Exec(ctx, query, string(positionID), string(tenantID), pos.name, pos.description, i+1, true, now)
		if err != nil {
			return nil, err
		}
		ids = append(ids, positionID)
	}

	return ids, nil
}

func createEvent(ctx context.Context, repo *db.EventRepository, tenantID common.TenantID, name string) (common.EventID, error) {
	ev, err := event.NewEvent(
		tenantID,
		name,
		event.EventTypeNormal,
		"テスト用イベントです。Alpha版での動作確認用に作成されました。",
	)
	if err != nil {
		return "", err
	}

	if err := repo.Save(ctx, ev); err != nil {
		return "", err
	}

	return ev.EventID(), nil
}

func createBusinessDays(ctx context.Context, repo *db.EventBusinessDayRepository, tenantID common.TenantID, eventID common.EventID, count int) ([]event.BusinessDayID, error) {
	ids := make([]event.BusinessDayID, 0, count)
	now := time.Now()

	for i := 0; i < count; i++ {
		targetDate := now.AddDate(0, 0, i)

		// 21:00 - 23:30 の営業時間
		startTime := time.Date(2000, 1, 1, 21, 0, 0, 0, time.UTC)
		endTime := time.Date(2000, 1, 1, 23, 30, 0, 0, time.UTC)

		bd, err := event.NewEventBusinessDay(
			tenantID,
			eventID,
			targetDate,
			startTime,
			endTime,
			event.OccurrenceTypeSpecial,
			nil, // recurring_pattern_id
		)
		if err != nil {
			return nil, err
		}

		if err := repo.Save(ctx, bd); err != nil {
			return nil, err
		}

		ids = append(ids, bd.BusinessDayID())
	}

	return ids, nil
}

func createMembers(ctx context.Context, repo *db.MemberRepository, tenantID common.TenantID, count int) ([]common.MemberID, error) {
	ids := make([]common.MemberID, 0, count)

	names := []string{"田中太郎", "佐藤花子", "鈴木一郎", "高橋美咲", "伊藤翔太", "渡辺さくら", "山本健太", "中村愛", "小林大輔", "加藤結衣"}

	for i := 0; i < count && i < len(names); i++ {
		m, err := member.NewMember(
			tenantID,
			names[i],
			fmt.Sprintf("test_user_%d@example.com", i+1),
			fmt.Sprintf("1234567890123456%02d", i+1), // Discord User ID
		)
		if err != nil {
			return nil, err
		}

		if err := repo.Save(ctx, m); err != nil {
			return nil, err
		}

		ids = append(ids, m.MemberID())
	}

	return ids, nil
}

func createShiftSlots(ctx context.Context, repo *db.ShiftSlotRepository, tenantID common.TenantID, businessDayID event.BusinessDayID, positionIDs []shift.PositionID) ([]shift.SlotID, error) {
	ids := make([]shift.SlotID, 0, len(positionIDs))

	slotConfigs := []struct {
		name          string
		instanceName  string
		startHour     int
		startMinute   int
		endHour       int
		endMinute     int
		requiredCount int
	}{
		{"受付", "受付1", 21, 0, 22, 0, 2},
		{"案内", "案内1", 21, 30, 23, 0, 1},
		{"配信", "配信1", 21, 0, 23, 30, 1},
	}

	for i, positionID := range positionIDs {
		if i >= len(slotConfigs) {
			break
		}
		cfg := slotConfigs[i]

		startTime := time.Date(2000, 1, 1, cfg.startHour, cfg.startMinute, 0, 0, time.UTC)
		endTime := time.Date(2000, 1, 1, cfg.endHour, cfg.endMinute, 0, 0, time.UTC)

		slot, err := shift.NewShiftSlot(
			tenantID,
			businessDayID,
			positionID,
			cfg.name,
			cfg.instanceName,
			startTime,
			endTime,
			cfg.requiredCount,
			i+1, // priority
		)
		if err != nil {
			return nil, err
		}

		if err := repo.Save(ctx, slot); err != nil {
			return nil, err
		}

		ids = append(ids, slot.SlotID())
	}

	return ids, nil
}
