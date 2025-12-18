package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/security"
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
	assignmentRepo := db.NewShiftAssignmentRepository(pool)
	adminRepo := db.NewAdminRepository(pool)
	attendanceRepo := db.NewAttendanceRepository(pool)
	scheduleRepo := db.NewScheduleRepository(pool)

	for i := 0; i < tenantCount; i++ {
		tenantID := common.NewTenantID()
		log.Printf("\n📦 Creating tenant %d/%d: %s", i+1, tenantCount, tenantID)

		// 0. テナントを作成
		if err := createTenant(ctx, pool, tenantID, fmt.Sprintf("テストテナント #%d", i+1)); err != nil {
			return fmt.Errorf("failed to create tenant: %w", err)
		}
		log.Printf("   ✅ Tenant created: %s", tenantID)

		// 0.5. 管理者を作成
		adminEmail, err := createAdmin(ctx, adminRepo, tenantID, i+1)
		if err != nil {
			return fmt.Errorf("failed to create admin: %w", err)
		}
		log.Printf("   ✅ Admin created: %s (password: password123)", adminEmail)

		// 1. イベントを作成
		eventID, err := createEvent(ctx, eventRepo, tenantID, fmt.Sprintf("テストイベント #%d", i+1))
		if err != nil {
			return fmt.Errorf("failed to create event: %w", err)
		}
		log.Printf("   ✅ Event created: %s", eventID)

		// 2. 営業日を作成（毎週土曜日、11月を含む）
		// これにより本出席データのテストが可能
		businessDayIDs, pastBusinessDayIDs, err := createWeeklyBusinessDays(ctx, businessDayRepo, tenantID, eventID)
		if err != nil {
			return fmt.Errorf("failed to create business days: %w", err)
		}
		log.Printf("   ✅ Business days created: %d (past: %d, future: %d)", len(businessDayIDs), len(pastBusinessDayIDs), len(businessDayIDs)-len(pastBusinessDayIDs))

		// 3. メンバーを作成（10人）
		memberIDs, err := createMembers(ctx, memberRepo, tenantID, 10)
		if err != nil {
			return fmt.Errorf("failed to create members: %w", err)
		}
		log.Printf("   ✅ Members created: %d", len(memberIDs))

		// 3.5. ロールを作成してメンバーに割り当て
		roleCount, err := createRolesAndAssignToMembers(ctx, pool, tenantID, memberIDs)
		if err != nil {
			return fmt.Errorf("failed to create roles: %w", err)
		}
		log.Printf("   ✅ Roles created: %d", roleCount)

		// 4. ポジションを作成
		positionIDs, err := createPositions(ctx, pool, tenantID)
		if err != nil {
			return fmt.Errorf("failed to create positions: %w", err)
		}
		log.Printf("   ✅ Positions created: %d", len(positionIDs))

		// 5. シフト枠を作成（各営業日に2〜3枠）
		allSlotIDs := make([]shift.SlotID, 0)
		pastSlotIDs := make([]shift.SlotID, 0)
		for _, bdID := range businessDayIDs {
			slots, err := createShiftSlots(ctx, slotRepo, tenantID, bdID, positionIDs)
			if err != nil {
				return fmt.Errorf("failed to create shift slots: %w", err)
			}
			allSlotIDs = append(allSlotIDs, slots...)

			// 過去の営業日のシフト枠を記録
			for _, pastBDID := range pastBusinessDayIDs {
				if bdID == pastBDID {
					pastSlotIDs = append(pastSlotIDs, slots...)
					break
				}
			}
		}
		log.Printf("   ✅ Shift slots created: %d (past: %d)", len(allSlotIDs), len(pastSlotIDs))

		// 6. 過去のシフト枠にランダムに割り当て（本出席データのため）
		assignmentCount, err := createShiftAssignments(ctx, assignmentRepo, tenantID, pastSlotIDs, memberIDs)
		if err != nil {
			return fmt.Errorf("failed to create shift assignments: %w", err)
		}
		log.Printf("   ✅ Shift assignments created: %d", assignmentCount)

		// 7. 出欠収集を作成（過去と未来のイベント用）
		attendanceCount, err := createAttendanceCollections(ctx, attendanceRepo, tenantID, eventID, memberIDs)
		if err != nil {
			return fmt.Errorf("failed to create attendance collections: %w", err)
		}
		log.Printf("   ✅ Attendance collections created: %d", attendanceCount)

		// 8. 日程調整を作成
		scheduleCount, err := createSchedules(ctx, scheduleRepo, tenantID, eventID, memberIDs)
		if err != nil {
			return fmt.Errorf("failed to create schedules: %w", err)
		}
		log.Printf("   ✅ Schedules created: %d", scheduleCount)
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

	names := []string{
		"ERENOA",
		"【LAAT】",
		"moyashiuri",
		"Yuichi_Snadra",
		"コーヒーキメた冷蔵庫お嬢様",
		"makkun_0627",
		"2943ten",
		"みらくるみらい",
		"ELtaso",
		"Ninomae Kazuaki",
	}

	for i := 0; i < count && i < len(names); i++ {
		m, err := member.NewMember(
			tenantID,
			names[i],
			fmt.Sprintf("discord_user_%d", 100000000000000000+i), // Discord User ID
			fmt.Sprintf("test_user_%d@example.com", i+1),
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

func createAdmin(ctx context.Context, repo *db.AdminRepository, tenantID common.TenantID, index int) (string, error) {
	now := time.Now()
	email := fmt.Sprintf("admin%d@example.com", index)

	// パスワードをハッシュ化 (password123)
	hasher := security.NewBcryptHasher()
	passwordHash, err := hasher.Hash("password123")
	if err != nil {
		return "", err
	}

	role, err := auth.NewRole("owner")
	if err != nil {
		return "", err
	}

	admin, err := auth.NewAdmin(
		now,
		tenantID,
		email,
		passwordHash,
		fmt.Sprintf("管理者 #%d", index),
		role,
	)
	if err != nil {
		return "", err
	}

	if err := repo.Save(ctx, admin); err != nil {
		return "", err
	}

	return email, nil
}

func createAttendanceCollections(ctx context.Context, repo *db.AttendanceRepository, tenantID common.TenantID, eventID common.EventID, memberIDs []common.MemberID) (int, error) {
	count := 0
	now := time.Now()

	// 11月の全ての土曜日を取得
	novemberDates := getSaturdaysInMonth(2025, 11)

	// 10月の最後の2つの土曜日
	octoberDates := getSaturdaysInMonth(2025, 10)
	if len(octoberDates) >= 2 {
		octoberDates = octoberDates[len(octoberDates)-2:]
	}

	// 12月の最初の2つの土曜日
	decemberDates := getSaturdaysInMonth(2025, 12)
	if len(decemberDates) >= 2 {
		decemberDates = decemberDates[:2]
	}

	// 全ての日付をマージ
	allDates := append(octoberDates, novemberDates...)
	allDates = append(allDates, decemberDates...)

	for _, targetDate := range allDates {

		collection, err := attendance.NewAttendanceCollection(
			now,
			tenantID,
			fmt.Sprintf("イベント出欠確認 %s", targetDate.Format("1/2")),
			fmt.Sprintf("イベント日程: %s", targetDate.Format("2006年1月2日")),
			attendance.TargetTypeEvent,
			eventID.String(),
			nil, // deadline
		)
		if err != nil {
			return count, err
		}

		if err := repo.Save(ctx, collection); err != nil {
			return count, err
		}

		// 対象日を1つ作成
		targetDateEntity, err := attendance.NewTargetDate(
			now,
			collection.CollectionID(),
			targetDate,
			1,
		)
		if err != nil {
			return count, err
		}

		if err := repo.SaveTargetDates(ctx, collection.CollectionID(), []*attendance.TargetDate{targetDateEntity}); err != nil {
			return count, err
		}

		// メンバーの70%が回答（ランダムに参加/不参加）
		responseCount := int(float64(len(memberIDs)) * 0.7)
		for j := 0; j < responseCount; j++ {
			memberID := memberIDs[j]

			// ランダムに参加/不参加を決定
			responseType := attendance.ResponseTypeAttending
			if (count+j)%3 == 0 { // 約1/3の確率で不参加
				responseType = attendance.ResponseTypeAbsent
			}

			response, err := attendance.NewAttendanceResponse(
				now,
				collection.CollectionID(),
				tenantID,
				memberID,
				targetDateEntity.TargetDateID(),
				responseType,
				"",
			)
			if err != nil {
				continue
			}

			if err := repo.UpsertResponse(ctx, response); err != nil {
				continue
			}
		}

		count++
	}

	return count, nil
}

func createSchedules(ctx context.Context, repo *db.ScheduleRepository, tenantID common.TenantID, eventID common.EventID, memberIDs []common.MemberID) (int, error) {
	count := 0
	now := time.Now()

	// 11月の日程調整を作成
	novemberDates := getSaturdaysInMonth(2025, 11)

	// 11月の全土曜日を候補日とした日程調整
	scheduleID := common.NewScheduleID()

	// 11月の土曜日を候補日として作成
	candidateDates := make([]*schedule.CandidateDate, 0, len(novemberDates))
	for j, candidateDate := range novemberDates {
		candidate, err := schedule.NewCandidateDate(
			now,
			scheduleID,
			candidateDate,
			nil, // startTime
			nil, // endTime
			j+1,
		)
		if err != nil {
			return count, err
		}
		candidateDates = append(candidateDates, candidate)
	}

	eventIDPtr := eventID
	scheduleEntity, err := schedule.NewDateSchedule(
		now,
		scheduleID,
		tenantID,
		"2025年11月のイベント日程調整",
		"11月のイベント開催日を決定するための日程調整です。候補日から都合の良い日を選んでください。",
		&eventIDPtr,
		candidateDates,
		nil, // deadline
	)
	if err != nil {
		return count, err
	}

	if err := repo.Save(ctx, scheduleEntity); err != nil {
		return count, err
	}

	// メンバーの50%が回答
	responseCount := len(memberIDs) / 2
	for j := 0; j < responseCount; j++ {
		memberID := memberIDs[j]

		// 各候補日への回答（最初の半分を○、残りを×）
		for k, candidate := range candidateDates {
			availability := schedule.AvailabilityAvailable
			if k >= len(candidateDates)/2 { // 後半の候補日
				availability = schedule.AvailabilityUnavailable
			}

			response, err := schedule.NewDateScheduleResponse(
				now,
				scheduleEntity.ScheduleID(),
				tenantID,
				memberID,
				candidate.CandidateID(),
				availability,
				"",
			)
			if err != nil {
				continue
			}

			if err := repo.UpsertResponse(ctx, response); err != nil {
				continue
			}
		}
	}

	count++

	// 12月の日程調整も作成
	decemberDates := getSaturdaysInMonth(2025, 12)
	if len(decemberDates) >= 3 {
		decemberDates = decemberDates[:3]
	}

	scheduleID = common.NewScheduleID()

	// 12月の候補日を作成
	candidateDates = make([]*schedule.CandidateDate, 0, len(decemberDates))
	for j, candidateDate := range decemberDates {
		candidate, err := schedule.NewCandidateDate(
			now,
			scheduleID,
			candidateDate,
			nil, // startTime
			nil, // endTime
			j+1,
		)
		if err != nil {
			return count, err
		}
		candidateDates = append(candidateDates, candidate)
	}

	eventIDPtr = eventID
	scheduleEntity, err = schedule.NewDateSchedule(
		now,
		scheduleID,
		tenantID,
		"2025年12月のイベント日程調整",
		"12月のイベント開催日を決定するための日程調整です。候補日から都合の良い日を選んでください。",
		&eventIDPtr,
		candidateDates,
		nil, // deadline
	)
	if err != nil {
		return count, err
	}

	if err := repo.Save(ctx, scheduleEntity); err != nil {
		return count, err
	}

	// メンバーの30%が回答
	responseCount = len(memberIDs) / 3
	for j := 0; j < responseCount; j++ {
		memberID := memberIDs[j]

		// 各候補日への回答（最初の2つを○、最後を×）
		for k, candidate := range candidateDates {
			availability := schedule.AvailabilityAvailable
			if k == len(candidateDates)-1 { // 最後の候補日
				availability = schedule.AvailabilityUnavailable
			}

			response, err := schedule.NewDateScheduleResponse(
				now,
				scheduleEntity.ScheduleID(),
				tenantID,
				memberID,
				candidate.CandidateID(),
				availability,
				"",
			)
			if err != nil {
				continue
			}

			if err := repo.UpsertResponse(ctx, response); err != nil {
				continue
			}
		}
	}

	count++

	return count, nil
}
// createWeeklyBusinessDays creates business days for every Saturday
// Includes November 2025 and surrounding months
func createWeeklyBusinessDays(ctx context.Context, repo *db.EventBusinessDayRepository, tenantID common.TenantID, eventID common.EventID) ([]event.BusinessDayID, []event.BusinessDayID, error) {
	allIDs := make([]event.BusinessDayID, 0)
	pastIDs := make([]event.BusinessDayID, 0)
	now := time.Now()

	// 11月の全ての土曜日を取得
	novemberDates := getSaturdaysInMonth(2025, 11)

	// 10月の最後の2つの土曜日
	octoberDates := getSaturdaysInMonth(2025, 10)
	if len(octoberDates) >= 2 {
		octoberDates = octoberDates[len(octoberDates)-2:]
	}

	// 12月の最初の2つの土曜日
	decemberDates := getSaturdaysInMonth(2025, 12)
	if len(decemberDates) >= 2 {
		decemberDates = decemberDates[:2]
	}

	// 全ての日付をマージ
	allDates := append(octoberDates, novemberDates...)
	allDates = append(allDates, decemberDates...)

	for _, targetDate := range allDates {
		// 21:30 - 23:00 の営業時間
		startTime := time.Date(2000, 1, 1, 21, 30, 0, 0, time.UTC)
		endTime := time.Date(2000, 1, 1, 23, 0, 0, 0, time.UTC)

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
			return nil, nil, err
		}

		if err := repo.Save(ctx, bd); err != nil {
			return nil, nil, err
		}

		allIDs = append(allIDs, bd.BusinessDayID())
		// 現在より前の日付は過去として記録
		if targetDate.Before(now) {
			pastIDs = append(pastIDs, bd.BusinessDayID())
		}
	}

	return allIDs, pastIDs, nil
}

// getSaturdaysInMonth returns all Saturdays in the given month
func getSaturdaysInMonth(year, month int) []time.Time {
	saturdays := make([]time.Time, 0)

	// 月の最初の日
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	// 最初の土曜日を見つける
	daysUntilSaturday := (6 - int(firstDay.Weekday()) + 7) % 7
	if firstDay.Weekday() == time.Saturday {
		daysUntilSaturday = 0
	}
	firstSaturday := firstDay.AddDate(0, 0, daysUntilSaturday)

	// 月の全ての土曜日を取得
	current := firstSaturday
	for current.Month() == time.Month(month) {
		saturdays = append(saturdays, current)
		current = current.AddDate(0, 0, 7) // 次の週
	}

	return saturdays
}

// createBusinessDaysWithHistory creates business days for both past and future
// startOffset: negative for past days (e.g., -15 for 15 days ago)
// endOffset: positive for future days (e.g., 7 for 7 days ahead)
func createBusinessDaysWithHistory(ctx context.Context, repo *db.EventBusinessDayRepository, tenantID common.TenantID, eventID common.EventID, startOffset, endOffset int) ([]event.BusinessDayID, []event.BusinessDayID, error) {
	allIDs := make([]event.BusinessDayID, 0)
	pastIDs := make([]event.BusinessDayID, 0)
	now := time.Now()

	for i := startOffset; i <= endOffset; i++ {
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
			return nil, nil, err
		}

		if err := repo.Save(ctx, bd); err != nil {
			return nil, nil, err
		}

		allIDs = append(allIDs, bd.BusinessDayID())
		if i < 0 {
			pastIDs = append(pastIDs, bd.BusinessDayID())
		}
	}

	return allIDs, pastIDs, nil
}

// createShiftAssignments creates shift assignments for given slots
// 各シフト枠にランダムにメンバーを割り当て（本出席データのモックとして）
func createShiftAssignments(ctx context.Context, repo *db.ShiftAssignmentRepository, tenantID common.TenantID, slotIDs []shift.SlotID, memberIDs []common.MemberID) (int, error) {
	count := 0

	// 各シフト枠に対して処理
	for idx, slotID := range slotIDs {
		// 80%の確率でシフト枠を満たす
		shouldAssign := (idx % 10) < 8

		if !shouldAssign {
			continue
		}

		// 1〜2人を割り当て（シフト枠によって変える）
		assignCount := 1
		if (idx % 3) == 0 {
			assignCount = 2
		}

		// メンバーを割り当て
		for j := 0; j < assignCount && j < len(memberIDs); j++ {
			memberIdx := (idx + j) % len(memberIDs)
			memberID := memberIDs[memberIdx]

			// ShiftAssignment エンティティを作成
			var nilPlanID shift.PlanID
			assignment, err := shift.NewShiftAssignment(
				tenantID,
				nilPlanID,
				slotID,
				memberID,
				shift.AssignmentMethodManual,
				false, // is_outside_preference
			)
			if err != nil {
				log.Printf("Failed to create assignment: %v", err)
				continue
			}

			// 保存
			if err := repo.Save(ctx, assignment); err != nil {
				// 既に存在する場合はスキップ
				log.Printf("Failed to save assignment: %v", err)
				continue
			}

			count++
		}
	}

	return count, nil
}

// createRolesAndAssignToMembers creates roles and assigns them to members
func createRolesAndAssignToMembers(ctx context.Context, pool *pgxpool.Pool, tenantID common.TenantID, memberIDs []common.MemberID) (int, error) {
	now := time.Now()

	// ロール定義
	roles := []struct {
		name         string
		description  string
		color        string
		displayOrder int
	}{
		{"リーダー", "チームをまとめる責任者", "#EF4444", 1},         // 赤
		{"サブリーダー", "リーダーをサポートする役割", "#8B5CF6", 2}, // 紫
		{"ベテラン", "経験豊富なメンバー", "#3B82F6", 3},            // 青
		{"レギュラー", "通常メンバー", "#10B981", 4},                // 緑
		{"新人", "新しく参加したメンバー", "#F59E0B", 5},            // オレンジ
	}

	roleIDs := make([]string, 0, len(roles))

	// ロールを作成
	for _, r := range roles {
		roleID := common.NewRoleID()
		query := `
			INSERT INTO roles (role_id, tenant_id, name, description, color, display_order, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
			ON CONFLICT (role_id) DO NOTHING
		`
		_, err := pool.Exec(ctx, query, string(roleID), string(tenantID), r.name, r.description, r.color, r.displayOrder, now)
		if err != nil {
			return 0, fmt.Errorf("failed to create role %s: %w", r.name, err)
		}
		roleIDs = append(roleIDs, string(roleID))
	}

	// メンバーにロールを割り当て
	// 最初の2人: リーダー
	// 3-4人目: サブリーダー
	// 5-7人目: ベテラン
	// 8-9人目: レギュラー
	// 10人目: 新人
	for i, memberID := range memberIDs {
		var assignedRoleIDs []string

		if i < 2 {
			// リーダー
			assignedRoleIDs = append(assignedRoleIDs, roleIDs[0])
		} else if i < 4 {
			// サブリーダー
			assignedRoleIDs = append(assignedRoleIDs, roleIDs[1])
		} else if i < 7 {
			// ベテラン
			assignedRoleIDs = append(assignedRoleIDs, roleIDs[2])
		} else if i < 9 {
			// レギュラー
			assignedRoleIDs = append(assignedRoleIDs, roleIDs[3])
		} else {
			// 新人
			assignedRoleIDs = append(assignedRoleIDs, roleIDs[4])
		}

		// メンバーロールを保存
		for _, roleID := range assignedRoleIDs {
			query := `
				INSERT INTO member_roles (member_id, role_id, assigned_at)
				VALUES ($1, $2, $3)
				ON CONFLICT (member_id, role_id) DO NOTHING
			`
			_, err := pool.Exec(ctx, query, string(memberID), roleID, now)
			if err != nil {
				log.Printf("Failed to assign role to member: %v", err)
			}
		}
	}

	return len(roles), nil
}
