package shift_test

import (
	"context"
	"testing"
	"time"

	appshift "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/shift"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// =====================================================
// BDD-style tests for ShiftAssignment
// These tests correspond to: features/shift_assignment.feature
// =====================================================

func TestBDD_ShiftAssignment_ManualAssign_EmptySlot(t *testing.T) {
	// Scenario: 空きのあるシフト枠にメンバーを手動で割り当てる

	// Given: テナント「VRChat Japan」が存在する
	tenantID := common.NewTenantID()

	// And: シフト枠「受付」が存在する（必要人数: 3名）
	testSlot := createTestShiftSlot(t, tenantID)

	// And: シフト枠「受付」の現在の割当人数は 0人 である
	// And: メンバー「佐藤」が存在する
	testMember := createTestMember(t, tenantID)

	slotRepo := &MockShiftSlotRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (*shift.ShiftSlot, error) {
			return testSlot, nil
		},
	}
	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			return 0, nil // 0人
		},
		saveFunc: func(ctx context.Context, assignment *shift.ShiftAssignment) error {
			return nil
		},
	}
	memberRepo := &MockMemberRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, memID common.MemberID) (*member.Member, error) {
			return testMember, nil
		},
	}

	usecase := appshift.NewConfirmManualAssignmentUsecase(slotRepo, assignmentRepo, memberRepo)

	// When: 管理者がメンバー「佐藤」をシフト枠「受付」に手動で割り当てる
	input := appshift.ConfirmManualAssignmentInput{
		TenantID: tenantID,
		SlotID:   testSlot.SlotID(),
		MemberID: testMember.MemberID(),
		ActorID:  common.NewMemberID(),
		Note:     "手動割り当て",
	}
	result, err := usecase.Execute(context.Background(), input)

	// Then: シフト割当が作成される
	if err != nil {
		t.Fatalf("Then failed: assignment should be created: %v", err)
	}
	if result == nil {
		t.Fatal("Then failed: result should not be nil")
	}

	// And: 割当ステータスは「confirmed」である
	if result.AssignmentStatus() != shift.AssignmentStatusConfirmed {
		t.Errorf("Then failed: status should be confirmed, got %s", result.AssignmentStatus())
	}

	// And: 割当方法は「manual」である
	if result.AssignmentMethod() != shift.AssignmentMethodManual {
		t.Errorf("Then failed: method should be manual, got %s", result.AssignmentMethod())
	}
}

func TestBDD_ShiftAssignment_ManualAssign_SlotFull(t *testing.T) {
	// Scenario: 満員のシフト枠への割当は拒否される

	// Given: テナント「VRChat Japan」が存在する
	tenantID := common.NewTenantID()

	// And: シフト枠「受付」が存在する（必要人数: 3名）
	testSlot := createTestShiftSlot(t, tenantID) // required_count = 3

	// And: シフト枠「受付」の現在の割当人数は 3人 である（必要人数と同じ）
	// And: メンバー「鈴木」が存在する
	testMember := createTestMember(t, tenantID)

	slotRepo := &MockShiftSlotRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (*shift.ShiftSlot, error) {
			return testSlot, nil
		},
	}
	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			return 3, nil // 満員
		},
	}
	memberRepo := &MockMemberRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, memID common.MemberID) (*member.Member, error) {
			return testMember, nil
		},
	}

	usecase := appshift.NewConfirmManualAssignmentUsecase(slotRepo, assignmentRepo, memberRepo)

	// When: 管理者がメンバー「鈴木」をシフト枠「受付」に手動で割り当てる
	input := appshift.ConfirmManualAssignmentInput{
		TenantID: tenantID,
		SlotID:   testSlot.SlotID(),
		MemberID: testMember.MemberID(),
		ActorID:  common.NewMemberID(),
		Note:     "",
	}
	result, err := usecase.Execute(context.Background(), input)

	// Then: エラー「slot is full: 3/3」が返される
	if err == nil {
		t.Fatal("Then failed: should return slot full error")
	}

	// And: シフト割当は作成されない
	if result != nil {
		t.Error("Then failed: result should be nil when slot is full")
	}
}

func TestBDD_ShiftAssignment_ManualAssign_SlotNotFound(t *testing.T) {
	// Scenario: 存在しないシフト枠への割当は失敗する

	// Given: テナント「VRChat Japan」が存在する
	tenantID := common.NewTenantID()
	slotID := shift.NewSlotID()

	// And: メンバー「佐藤」が存在する
	slotRepo := &MockShiftSlotRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, sid shift.SlotID) (*shift.ShiftSlot, error) {
			return nil, common.NewNotFoundError("shift_slot", sid.String())
		},
	}
	assignmentRepo := &MockShiftAssignmentRepository{}
	memberRepo := &MockMemberRepository{}

	usecase := appshift.NewConfirmManualAssignmentUsecase(slotRepo, assignmentRepo, memberRepo)

	// When: 管理者がメンバー「佐藤」を存在しないシフト枠に割り当てる
	input := appshift.ConfirmManualAssignmentInput{
		TenantID: tenantID,
		SlotID:   slotID,
		MemberID: common.NewMemberID(),
		ActorID:  common.NewMemberID(),
	}
	_, err := usecase.Execute(context.Background(), input)

	// Then: エラー「shift_slot not found」が返される
	if err == nil {
		t.Fatal("Then failed: should return not found error")
	}
}

func TestBDD_ShiftAssignment_ManualAssign_MemberNotFound(t *testing.T) {
	// Scenario: 存在しないメンバーの割当は失敗する

	// Given: テナント「VRChat Japan」が存在する
	tenantID := common.NewTenantID()

	// And: シフト枠「受付」が存在する（割当人数: 0人）
	testSlot := createTestShiftSlot(t, tenantID)

	slotRepo := &MockShiftSlotRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (*shift.ShiftSlot, error) {
			return testSlot, nil
		},
	}
	assignmentRepo := &MockShiftAssignmentRepository{}
	memberRepo := &MockMemberRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, memID common.MemberID) (*member.Member, error) {
			return nil, common.NewNotFoundError("member", memID.String())
		},
	}

	usecase := appshift.NewConfirmManualAssignmentUsecase(slotRepo, assignmentRepo, memberRepo)

	// When: 管理者が存在しないメンバーをシフト枠「受付」に割り当てる
	input := appshift.ConfirmManualAssignmentInput{
		TenantID: tenantID,
		SlotID:   testSlot.SlotID(),
		MemberID: common.NewMemberID(),
		ActorID:  common.NewMemberID(),
	}
	_, err := usecase.Execute(context.Background(), input)

	// Then: エラー「member not found」が返される
	if err == nil {
		t.Fatal("Then failed: should return not found error")
	}
}

// =====================================================
// シフト割当のキャンセル（ドメインエンティティレベル）
// =====================================================

func TestBDD_ShiftAssignment_Cancel_ConfirmedAssignment(t *testing.T) {
	// Scenario: 確定済みのシフト割当をキャンセルする

	// Given: メンバー「佐藤」のシフト割当が確定済みである
	now := time.Now()
	assignment, err := shift.NewShiftAssignment(
		now,
		common.NewTenantID(),
		shift.PlanID(""),
		shift.NewSlotID(),
		common.NewMemberID(),
		shift.AssignmentMethodManual,
		false,
	)
	if err != nil {
		t.Fatalf("Given failed: %v", err)
	}
	if !assignment.IsConfirmed() {
		t.Fatal("Given failed: assignment should be confirmed")
	}

	// When: メンバー「佐藤」のシフト割当をキャンセルする
	cancelTime := now.Add(1 * time.Hour)
	err = assignment.Cancel(cancelTime)

	// Then: 割当ステータスが「cancelled」に変更される
	if err != nil {
		t.Fatalf("Then failed: cancel should succeed: %v", err)
	}
	if assignment.AssignmentStatus() != shift.AssignmentStatusCancelled {
		t.Errorf("Then failed: expected cancelled, got %s", assignment.AssignmentStatus())
	}

	// And: キャンセル日時が記録される
	if assignment.CancelledAt() == nil {
		t.Error("Then failed: cancelled_at should be set")
	}
	if !assignment.CancelledAt().Equal(cancelTime) {
		t.Errorf("Then failed: cancelled_at should be %v, got %v", cancelTime, assignment.CancelledAt())
	}
}

func TestBDD_ShiftAssignment_Cancel_AlreadyCancelled(t *testing.T) {
	// Scenario: すでにキャンセル済みの割当を再キャンセルするとエラー

	// Given: メンバー「佐藤」のシフト割当がキャンセル済みである
	now := time.Now()
	assignment, _ := shift.NewShiftAssignment(
		now,
		common.NewTenantID(),
		shift.PlanID(""),
		shift.NewSlotID(),
		common.NewMemberID(),
		shift.AssignmentMethodManual,
		false,
	)
	_ = assignment.Cancel(now)
	if !assignment.IsCancelled() {
		t.Fatal("Given failed: assignment should be cancelled")
	}

	// When: メンバー「佐藤」のシフト割当をキャンセルする
	err := assignment.Cancel(now.Add(1 * time.Hour))

	// Then: エラー「assignment is already cancelled」が返される
	if err == nil {
		t.Fatal("Then failed: should return already cancelled error")
	}
}
