package app

import (
	"context"
	"fmt"
	"log"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ShiftAssignmentService はシフト確定のビジネスロジックを担当する
type ShiftAssignmentService struct {
	dbPool         *pgxpool.Pool
	slotRepo       *db.ShiftSlotRepository
	assignmentRepo *db.ShiftAssignmentRepository
	memberRepo     *db.MemberRepository
}

// NewShiftAssignmentService creates a new ShiftAssignmentService
func NewShiftAssignmentService(dbPool *pgxpool.Pool) *ShiftAssignmentService {
	return &ShiftAssignmentService{
		dbPool:         dbPool,
		slotRepo:       db.NewShiftSlotRepository(dbPool),
		assignmentRepo: db.NewShiftAssignmentRepository(dbPool),
		memberRepo:     db.NewMemberRepository(dbPool),
	}
}

// ConfirmManualAssignment は管理者による手動シフト割り当てを実行する
//
// ロジック:
//  1. ShiftSlot を取得（tenant_id チェック）
//  2. Member を取得（tenant_id チェック）
//  3. トランザクション開始
//  4. SELECT ... FOR UPDATE で該当 slot の assignments をロック
//  5. 既存の confirmed 件数をカウント
//  6. count >= required_count なら ErrSlotFull を返す
//  7. ShiftAssignment を作成・保存
//  8. コミット
//  9. Notification stub 呼び出し（ログ出力）
// 10. AuditLog stub（ログ出力）
func (s *ShiftAssignmentService) ConfirmManualAssignment(
	ctx context.Context,
	tenantID common.TenantID,
	slotID shift.SlotID,
	memberID common.MemberID,
	actorID common.MemberID,
	note string,
) (*shift.ShiftAssignment, error) {
	// 1. ShiftSlot を取得（tenant_id チェック）
	slot, err := s.slotRepo.FindByID(ctx, tenantID, slotID)
	if err != nil {
		return nil, fmt.Errorf("failed to find shift slot: %w", err)
	}

	// 2. Member を取得（tenant_id チェック）
	member, err := s.memberRepo.FindByID(ctx, tenantID, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to find member: %w", err)
	}

	// 3. トランザクション開始
	tx, err := s.dbPool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 4 & 5. SELECT ... FOR UPDATE + 件数カウント
	//
	// NOTE: ShiftPlan を使わない簡易実装のため、plan_id は NULL にする
	// 将来的に ShiftPlan 実装時に plan_id を設定する
	query := `
		SELECT COUNT(*)
		FROM shift_assignments
		WHERE tenant_id = $1
		  AND slot_id = $2
		  AND assignment_status = 'confirmed'
		  AND deleted_at IS NULL
		FOR UPDATE
	`
	var currentCount int
	err = tx.QueryRow(ctx, query, tenantID.String(), slotID.String()).Scan(&currentCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count assignments: %w", err)
	}

	// 6. count >= required_count なら ErrSlotFull を返す
	if currentCount >= slot.RequiredCount() {
		return nil, common.NewDomainError(
			common.ErrConflict,
			fmt.Sprintf("slot is full: %d/%d", currentCount, slot.RequiredCount()),
		)
	}

	// 7. ShiftAssignment を作成
	//
	// NOTE: plan_id は NULL（簡易実装）
	// assignment_method は "manual"
	// is_outside_preference は false（希望収集未実装）
	var nilPlanID shift.PlanID // ゼロ値（NULL として扱う）
	assignment, err := shift.NewShiftAssignment(
		tenantID,
		nilPlanID,
		slotID,
		memberID,
		shift.AssignmentMethodManual,
		false, // is_outside_preference
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create shift assignment: %w", err)
	}

	// 保存（トランザクション内）
	err = s.saveAssignmentInTx(ctx, tx, assignment)
	if err != nil {
		return nil, fmt.Errorf("failed to save shift assignment: %w", err)
	}

	// 8. コミット
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// 9. Notification stub 呼び出し（ログ出力）
	log.Printf("[Notification Stub] シフト確定通知: member=%s, slot=%s, assigned_at=%s",
		member.DisplayName(),
		slot.SlotName(),
		assignment.AssignedAt().Format("2006-01-02 15:04:05"),
	)

	// 10. AuditLog stub（ログ出力）
	log.Printf("[AuditLog Stub] CREATE ShiftAssignment: actor_id=%s, assignment_id=%s, member_id=%s, slot_id=%s",
		actorID.String(),
		assignment.AssignmentID().String(),
		memberID.String(),
		slotID.String(),
	)

	return assignment, nil
}

// saveAssignmentInTx はトランザクション内で ShiftAssignment を保存する
// NOTE: plan_id は NULL で保存（簡易実装）
func (s *ShiftAssignmentService) saveAssignmentInTx(
	ctx context.Context,
	tx pgx.Tx,
	assignment *shift.ShiftAssignment,
) error {
	query := `
		INSERT INTO shift_assignments (
			assignment_id,
			tenant_id,
			plan_id,
			slot_id,
			member_id,
			assignment_status,
			assignment_method,
			is_outside_preference,
			assigned_at,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := tx.Exec(ctx, query,
		assignment.AssignmentID().String(),
		assignment.TenantID().String(),
		nil, // plan_id は NULL（簡易実装）
		assignment.SlotID().String(),
		assignment.MemberID().String(),
		"confirmed", // assignment_status
		"manual",    // assignment_method
		assignment.IsOutsidePreference(),
		assignment.AssignedAt(),
		assignment.CreatedAt(),
		assignment.UpdatedAt(),
	)

	return err
}

