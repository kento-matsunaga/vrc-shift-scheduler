package shift_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appshift "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/shift"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// =====================================================
// Mock Repositories
// =====================================================

type MockShiftSlotRepository struct {
	saveFunc                             func(ctx context.Context, slot *shift.ShiftSlot) error
	findByIDFunc                         func(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) (*shift.ShiftSlot, error)
	findByBusinessDayFunc                func(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) ([]*shift.ShiftSlot, error)
	findByInstanceIDFunc                 func(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) ([]*shift.ShiftSlot, error)
	findByBusinessDayIDAndInstanceIDFunc func(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID, instanceID shift.InstanceID) ([]*shift.ShiftSlot, error)
}

func (m *MockShiftSlotRepository) Save(ctx context.Context, slot *shift.ShiftSlot) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, slot)
	}
	return nil
}

func (m *MockShiftSlotRepository) FindByID(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) (*shift.ShiftSlot, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, slotID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockShiftSlotRepository) FindByBusinessDayID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) ([]*shift.ShiftSlot, error) {
	if m.findByBusinessDayFunc != nil {
		return m.findByBusinessDayFunc(ctx, tenantID, businessDayID)
	}
	return nil, nil
}

func (m *MockShiftSlotRepository) FindByInstanceID(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) ([]*shift.ShiftSlot, error) {
	if m.findByInstanceIDFunc != nil {
		return m.findByInstanceIDFunc(ctx, tenantID, instanceID)
	}
	return nil, nil
}

func (m *MockShiftSlotRepository) FindByBusinessDayIDAndInstanceID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID, instanceID shift.InstanceID) ([]*shift.ShiftSlot, error) {
	if m.findByBusinessDayIDAndInstanceIDFunc != nil {
		return m.findByBusinessDayIDAndInstanceIDFunc(ctx, tenantID, businessDayID, instanceID)
	}
	return nil, nil
}

func (m *MockShiftSlotRepository) Delete(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) error {
	return nil
}

type MockShiftAssignmentRepository struct {
	saveFunc                 func(ctx context.Context, assignment *shift.ShiftAssignment) error
	findByIDFunc             func(ctx context.Context, tenantID common.TenantID, assignmentID shift.AssignmentID) (*shift.ShiftAssignment, error)
	findBySlotIDFunc         func(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) ([]*shift.ShiftAssignment, error)
	findByMemberIDFunc       func(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) ([]*shift.ShiftAssignment, error)
	countConfirmedBySlotFunc func(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) (int, error)
	deleteFunc               func(ctx context.Context, tenantID common.TenantID, assignmentID shift.AssignmentID) error
}

func (m *MockShiftAssignmentRepository) Save(ctx context.Context, assignment *shift.ShiftAssignment) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, assignment)
	}
	return nil
}

func (m *MockShiftAssignmentRepository) FindByID(ctx context.Context, tenantID common.TenantID, assignmentID shift.AssignmentID) (*shift.ShiftAssignment, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, assignmentID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockShiftAssignmentRepository) FindBySlotID(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) ([]*shift.ShiftAssignment, error) {
	if m.findBySlotIDFunc != nil {
		return m.findBySlotIDFunc(ctx, tenantID, slotID)
	}
	return nil, nil
}

func (m *MockShiftAssignmentRepository) FindConfirmedBySlotID(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) ([]*shift.ShiftAssignment, error) {
	return nil, nil
}

func (m *MockShiftAssignmentRepository) FindByMemberID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) ([]*shift.ShiftAssignment, error) {
	if m.findByMemberIDFunc != nil {
		return m.findByMemberIDFunc(ctx, tenantID, memberID)
	}
	return nil, nil
}

func (m *MockShiftAssignmentRepository) FindConfirmedByMemberID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) ([]*shift.ShiftAssignment, error) {
	return nil, nil
}

func (m *MockShiftAssignmentRepository) FindByPlanID(ctx context.Context, tenantID common.TenantID, planID shift.PlanID) ([]*shift.ShiftAssignment, error) {
	return nil, nil
}

func (m *MockShiftAssignmentRepository) CountConfirmedBySlotID(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID) (int, error) {
	if m.countConfirmedBySlotFunc != nil {
		return m.countConfirmedBySlotFunc(ctx, tenantID, slotID)
	}
	return 0, nil
}

func (m *MockShiftAssignmentRepository) Delete(ctx context.Context, tenantID common.TenantID, assignmentID shift.AssignmentID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, tenantID, assignmentID)
	}
	return nil
}

func (m *MockShiftAssignmentRepository) ExistsBySlotIDAndMemberID(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID, memberID common.MemberID) (bool, error) {
	return false, nil
}

func (m *MockShiftAssignmentRepository) HasConfirmedByMemberAndBusinessDayID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID, businessDayID string) (bool, error) {
	return false, nil
}

type MockBusinessDayRepository struct {
	findByIDFunc func(ctx context.Context, tenantID common.TenantID, id event.BusinessDayID) (*event.EventBusinessDay, error)
}

func (m *MockBusinessDayRepository) Save(ctx context.Context, bd *event.EventBusinessDay) error {
	return nil
}

func (m *MockBusinessDayRepository) FindByID(ctx context.Context, tenantID common.TenantID, id event.BusinessDayID) (*event.EventBusinessDay, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockBusinessDayRepository) FindByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *MockBusinessDayRepository) ExistsByEventIDAndDate(ctx context.Context, tenantID common.TenantID, eventID common.EventID, date time.Time, startTime time.Time) (bool, error) {
	return false, nil
}

func (m *MockBusinessDayRepository) FindByEventIDAndDateRange(ctx context.Context, tenantID common.TenantID, eventID common.EventID, startDate, endDate time.Time) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *MockBusinessDayRepository) FindActiveByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *MockBusinessDayRepository) FindByTenantIDAndDate(ctx context.Context, tenantID common.TenantID, date time.Time) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *MockBusinessDayRepository) Delete(ctx context.Context, tenantID common.TenantID, id event.BusinessDayID) error {
	return nil
}

func (m *MockBusinessDayRepository) FindRecentByTenantID(ctx context.Context, tenantID common.TenantID, limit int) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

func (m *MockBusinessDayRepository) FindRecentByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID, limit int, includeFuture bool) ([]*event.EventBusinessDay, error) {
	return nil, nil
}

type MockMemberRepository struct {
	findByIDFunc func(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) (*member.Member, error)
}

func (m *MockMemberRepository) Save(ctx context.Context, mem *member.Member) error {
	return nil
}

func (m *MockMemberRepository) FindByID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) (*member.Member, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, memberID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockMemberRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error) {
	return nil, nil
}

func (m *MockMemberRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error) {
	return nil, nil
}

func (m *MockMemberRepository) Delete(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) error {
	return nil
}

func (m *MockMemberRepository) ExistsByDiscordUserID(ctx context.Context, tenantID common.TenantID, discordUserID string) (bool, error) {
	return false, nil
}

type MockInstanceRepository struct {
	findByIDFunc             func(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) (*shift.Instance, error)
	findByEventIDFunc        func(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*shift.Instance, error)
	findByEventIDAndNameFunc func(ctx context.Context, tenantID common.TenantID, eventID common.EventID, name string) (*shift.Instance, error)
	deleteFunc               func(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) error
}

func (m *MockInstanceRepository) Save(ctx context.Context, instance *shift.Instance) error {
	return nil
}

func (m *MockInstanceRepository) FindByID(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) (*shift.Instance, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, instanceID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockInstanceRepository) FindByEventID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) ([]*shift.Instance, error) {
	if m.findByEventIDFunc != nil {
		return m.findByEventIDFunc(ctx, tenantID, eventID)
	}
	return nil, nil
}

func (m *MockInstanceRepository) FindByEventIDAndName(ctx context.Context, tenantID common.TenantID, eventID common.EventID, name string) (*shift.Instance, error) {
	if m.findByEventIDAndNameFunc != nil {
		return m.findByEventIDAndNameFunc(ctx, tenantID, eventID, name)
	}
	return nil, nil
}

func (m *MockInstanceRepository) Delete(ctx context.Context, tenantID common.TenantID, instanceID shift.InstanceID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, tenantID, instanceID)
	}
	return nil
}

func (m *MockMemberRepository) FindByDiscordUserID(ctx context.Context, tenantID common.TenantID, discordUserID string) (*member.Member, error) {
	return nil, nil
}

func (m *MockMemberRepository) FindByEmail(ctx context.Context, tenantID common.TenantID, email string) (*member.Member, error) {
	return nil, nil
}

func (m *MockMemberRepository) ExistsByEmail(ctx context.Context, tenantID common.TenantID, email string) (bool, error) {
	return false, nil
}

// MockTxManager is a mock implementation of TxManager for testing
type MockTxManager struct {
	withTxFunc func(ctx context.Context, fn func(context.Context) error) error
}

func (m *MockTxManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	if m.withTxFunc != nil {
		return m.withTxFunc(ctx, fn)
	}
	// Default implementation: just call the function without actual transaction
	return fn(ctx)
}

// =====================================================
// Helper functions
// =====================================================

func createTestShiftSlot(t *testing.T, tenantID common.TenantID) *shift.ShiftSlot {
	t.Helper()
	now := time.Now()
	businessDayID := event.NewBusinessDayID()

	slot, err := shift.NewShiftSlot(
		now,
		tenantID,
		businessDayID,
		nil, // instanceID
		"テストシフト",
		"VRChat Japan",
		time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
		3,
		1,
	)
	if err != nil {
		t.Fatalf("Failed to create test shift slot: %v", err)
	}
	return slot
}

func createTestMember(t *testing.T, tenantID common.TenantID) *member.Member {
	t.Helper()
	now := time.Now()
	mem, err := member.NewMember(
		now,
		tenantID,
		"テストメンバー",
		"discord_user_123",
		"test@example.com",
	)
	if err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}
	return mem
}

func createTestBusinessDay(t *testing.T, tenantID common.TenantID, eventID common.EventID) *event.EventBusinessDay {
	t.Helper()
	now := time.Now()
	targetDate := now.Add(24 * time.Hour)
	startTime := time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC)
	endTime := time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC)

	bd, err := event.NewEventBusinessDay(
		now,
		tenantID,
		eventID,
		targetDate,
		startTime,
		endTime,
		event.OccurrenceTypeRecurring,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test business day: %v", err)
	}
	return bd
}

// =====================================================
// CreateShiftSlotUsecase Tests
// =====================================================

func TestCreateShiftSlotUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	businessDay := createTestBusinessDay(t, tenantID, eventID)
	businessDayID := businessDay.BusinessDayID()

	bdRepo := &MockBusinessDayRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id event.BusinessDayID) (*event.EventBusinessDay, error) {
			return businessDay, nil
		},
	}

	slotRepo := &MockShiftSlotRepository{
		saveFunc: func(ctx context.Context, slot *shift.ShiftSlot) error {
			return nil
		},
	}

	instanceRepo := &MockInstanceRepository{}

	usecase := appshift.NewCreateShiftSlotUsecase(slotRepo, bdRepo, instanceRepo)

	input := appshift.CreateShiftSlotInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		SlotName:      "受付",
		InstanceName:  "VRChat Japan",
		StartTime:     time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
		EndTime:       time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
		RequiredCount: 2,
		Priority:      1,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.SlotName() != "受付" {
		t.Errorf("SlotName mismatch: got %v, want '受付'", result.SlotName())
	}

	if result.RequiredCount() != 2 {
		t.Errorf("RequiredCount mismatch: got %v, want 2", result.RequiredCount())
	}
}

func TestCreateShiftSlotUsecase_Execute_ErrorWhenBusinessDayNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()

	bdRepo := &MockBusinessDayRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id event.BusinessDayID) (*event.EventBusinessDay, error) {
			return nil, common.NewNotFoundError("business_day", id.String())
		},
	}

	slotRepo := &MockShiftSlotRepository{}
	instanceRepo := &MockInstanceRepository{}

	usecase := appshift.NewCreateShiftSlotUsecase(slotRepo, bdRepo, instanceRepo)

	input := appshift.CreateShiftSlotInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		SlotName:      "受付",
		InstanceName:  "VRChat Japan",
		StartTime:     time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
		EndTime:       time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
		RequiredCount: 2,
		Priority:      1,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when business day not found")
	}
}

func TestCreateShiftSlotUsecase_Execute_ErrorWhenSaveFails(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	businessDay := createTestBusinessDay(t, tenantID, eventID)
	businessDayID := businessDay.BusinessDayID()

	bdRepo := &MockBusinessDayRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id event.BusinessDayID) (*event.EventBusinessDay, error) {
			return businessDay, nil
		},
	}

	slotRepo := &MockShiftSlotRepository{
		saveFunc: func(ctx context.Context, slot *shift.ShiftSlot) error {
			return errors.New("database error")
		},
	}

	instanceRepo := &MockInstanceRepository{}

	usecase := appshift.NewCreateShiftSlotUsecase(slotRepo, bdRepo, instanceRepo)

	input := appshift.CreateShiftSlotInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		SlotName:      "受付",
		InstanceName:  "VRChat Japan",
		StartTime:     time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
		EndTime:       time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
		RequiredCount: 2,
		Priority:      1,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when save fails")
	}
}

func TestCreateShiftSlotUsecase_Execute_SuccessWithInstanceID(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	businessDay := createTestBusinessDay(t, tenantID, eventID)
	businessDayID := businessDay.BusinessDayID()

	// 同じイベントに属するインスタンスを作成
	now := time.Now()
	instanceID := shift.NewInstanceIDWithTime(now)
	instance, err := shift.NewInstance(now, tenantID, eventID, "第一インスタンス", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create test instance: %v", err)
	}

	bdRepo := &MockBusinessDayRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id event.BusinessDayID) (*event.EventBusinessDay, error) {
			return businessDay, nil
		},
	}

	slotRepo := &MockShiftSlotRepository{
		saveFunc: func(ctx context.Context, slot *shift.ShiftSlot) error {
			return nil
		},
	}

	instanceRepo := &MockInstanceRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, instID shift.InstanceID) (*shift.Instance, error) {
			return instance, nil
		},
	}

	usecase := appshift.NewCreateShiftSlotUsecase(slotRepo, bdRepo, instanceRepo)

	input := appshift.CreateShiftSlotInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		InstanceID:    &instanceID,
		SlotName:      "受付",
		InstanceName:  "第一インスタンス",
		StartTime:     time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
		EndTime:       time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
		RequiredCount: 2,
		Priority:      1,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.InstanceID() == nil {
		t.Fatal("InstanceID should not be nil")
	}
}

func TestCreateShiftSlotUsecase_Execute_ErrorWhenInstanceNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	businessDay := createTestBusinessDay(t, tenantID, eventID)
	businessDayID := businessDay.BusinessDayID()

	instanceID := shift.NewInstanceIDWithTime(time.Now())

	bdRepo := &MockBusinessDayRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id event.BusinessDayID) (*event.EventBusinessDay, error) {
			return businessDay, nil
		},
	}

	slotRepo := &MockShiftSlotRepository{}

	instanceRepo := &MockInstanceRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, instID shift.InstanceID) (*shift.Instance, error) {
			return nil, common.NewNotFoundError("instance", instID.String())
		},
	}

	usecase := appshift.NewCreateShiftSlotUsecase(slotRepo, bdRepo, instanceRepo)

	input := appshift.CreateShiftSlotInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		InstanceID:    &instanceID,
		SlotName:      "受付",
		InstanceName:  "第一インスタンス",
		StartTime:     time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
		EndTime:       time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
		RequiredCount: 2,
		Priority:      1,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when instance not found")
	}
}

func TestCreateShiftSlotUsecase_Execute_ErrorWhenInstanceBelongsToDifferentEvent(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	differentEventID := common.NewEventID() // 異なるイベントID
	businessDay := createTestBusinessDay(t, tenantID, eventID)
	businessDayID := businessDay.BusinessDayID()

	// 異なるイベントに属するインスタンスを作成
	now := time.Now()
	instanceID := shift.NewInstanceIDWithTime(now)
	instance, err := shift.NewInstance(now, tenantID, differentEventID, "別イベントのインスタンス", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create test instance: %v", err)
	}

	bdRepo := &MockBusinessDayRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id event.BusinessDayID) (*event.EventBusinessDay, error) {
			return businessDay, nil
		},
	}

	slotRepo := &MockShiftSlotRepository{}

	instanceRepo := &MockInstanceRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, instID shift.InstanceID) (*shift.Instance, error) {
			return instance, nil
		},
	}

	usecase := appshift.NewCreateShiftSlotUsecase(slotRepo, bdRepo, instanceRepo)

	input := appshift.CreateShiftSlotInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		InstanceID:    &instanceID,
		SlotName:      "受付",
		InstanceName:  "別イベントのインスタンス",
		StartTime:     time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
		EndTime:       time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
		RequiredCount: 2,
		Priority:      1,
	}

	_, err = usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when instance belongs to a different event")
	}

	// エラーメッセージの確認（日本語化対応）
	expectedMsg := "シフトテンプレートは営業日と同じイベントに属している必要があります"
	if !containsString(err.Error(), expectedMsg) {
		t.Errorf("Error message should contain '%s', got: %v", expectedMsg, err.Error())
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// =====================================================
// ListShiftSlotsUsecase Tests
// =====================================================

func TestListShiftSlotsUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()

	testSlots := []*shift.ShiftSlot{
		createTestShiftSlot(t, tenantID),
		createTestShiftSlot(t, tenantID),
	}

	slotRepo := &MockShiftSlotRepository{
		findByBusinessDayFunc: func(ctx context.Context, tid common.TenantID, bdID event.BusinessDayID) ([]*shift.ShiftSlot, error) {
			return testSlots, nil
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			return 1, nil
		},
	}

	usecase := appshift.NewListShiftSlotsUsecase(slotRepo, assignmentRepo)

	input := appshift.ListShiftSlotsInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 slots, got %d", len(result))
	}

	// Check assigned count
	for _, item := range result {
		if item.AssignedCount != 1 {
			t.Errorf("AssignedCount should be 1, got %d", item.AssignedCount)
		}
	}
}

func TestListShiftSlotsUsecase_Execute_EmptyList(t *testing.T) {
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()

	slotRepo := &MockShiftSlotRepository{
		findByBusinessDayFunc: func(ctx context.Context, tid common.TenantID, bdID event.BusinessDayID) ([]*shift.ShiftSlot, error) {
			return []*shift.ShiftSlot{}, nil
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{}

	usecase := appshift.NewListShiftSlotsUsecase(slotRepo, assignmentRepo)

	input := appshift.ListShiftSlotsInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 slots, got %d", len(result))
	}
}

// =====================================================
// GetShiftSlotUsecase Tests
// =====================================================

func TestGetShiftSlotUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	testSlot := createTestShiftSlot(t, tenantID)

	slotRepo := &MockShiftSlotRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (*shift.ShiftSlot, error) {
			return testSlot, nil
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			return 2, nil
		},
	}

	usecase := appshift.NewGetShiftSlotUsecase(slotRepo, assignmentRepo)

	input := appshift.GetShiftSlotInput{
		TenantID: tenantID,
		SlotID:   testSlot.SlotID(),
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result.Slot.SlotID() != testSlot.SlotID() {
		t.Errorf("SlotID mismatch: got %v, want %v", result.Slot.SlotID(), testSlot.SlotID())
	}

	if result.AssignedCount != 2 {
		t.Errorf("AssignedCount should be 2, got %d", result.AssignedCount)
	}
}

func TestGetShiftSlotUsecase_Execute_NotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	slotID := shift.NewSlotID()

	slotRepo := &MockShiftSlotRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, sid shift.SlotID) (*shift.ShiftSlot, error) {
			return nil, common.NewNotFoundError("shift_slot", sid.String())
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{}

	usecase := appshift.NewGetShiftSlotUsecase(slotRepo, assignmentRepo)

	input := appshift.GetShiftSlotInput{
		TenantID: tenantID,
		SlotID:   slotID,
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when slot not found")
	}
}

// =====================================================
// ConfirmManualAssignmentUsecase Tests
// =====================================================

func TestConfirmManualAssignmentUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	testSlot := createTestShiftSlot(t, tenantID)
	testMember := createTestMember(t, tenantID)

	slotRepo := &MockShiftSlotRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (*shift.ShiftSlot, error) {
			return testSlot, nil
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			return 0, nil // Not full yet
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

	input := appshift.ConfirmManualAssignmentInput{
		TenantID: tenantID,
		SlotID:   testSlot.SlotID(),
		MemberID: testMember.MemberID(),
		ActorID:  common.NewMemberID(),
		Note:     "手動割り当て",
	}

	result, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.MemberID() != testMember.MemberID() {
		t.Errorf("MemberID mismatch: got %v, want %v", result.MemberID(), testMember.MemberID())
	}
}

func TestConfirmManualAssignmentUsecase_Execute_ErrorWhenSlotFull(t *testing.T) {
	tenantID := common.NewTenantID()
	testSlot := createTestShiftSlot(t, tenantID) // required_count = 3
	testMember := createTestMember(t, tenantID)

	slotRepo := &MockShiftSlotRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (*shift.ShiftSlot, error) {
			return testSlot, nil
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			return 3, nil // Slot is full
		},
	}

	memberRepo := &MockMemberRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, memID common.MemberID) (*member.Member, error) {
			return testMember, nil
		},
	}

	usecase := appshift.NewConfirmManualAssignmentUsecase(slotRepo, assignmentRepo, memberRepo)

	input := appshift.ConfirmManualAssignmentInput{
		TenantID: tenantID,
		SlotID:   testSlot.SlotID(),
		MemberID: testMember.MemberID(),
		ActorID:  common.NewMemberID(),
		Note:     "",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when slot is full")
	}
}

func TestConfirmManualAssignmentUsecase_Execute_ErrorWhenSlotNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	slotID := shift.NewSlotID()

	slotRepo := &MockShiftSlotRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, sid shift.SlotID) (*shift.ShiftSlot, error) {
			return nil, common.NewNotFoundError("shift_slot", sid.String())
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{}
	memberRepo := &MockMemberRepository{}

	usecase := appshift.NewConfirmManualAssignmentUsecase(slotRepo, assignmentRepo, memberRepo)

	input := appshift.ConfirmManualAssignmentInput{
		TenantID: tenantID,
		SlotID:   slotID,
		MemberID: common.NewMemberID(),
		ActorID:  common.NewMemberID(),
		Note:     "",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when slot not found")
	}
}

func TestConfirmManualAssignmentUsecase_Execute_ErrorWhenMemberNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	testSlot := createTestShiftSlot(t, tenantID)
	memberID := common.NewMemberID()

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

	input := appshift.ConfirmManualAssignmentInput{
		TenantID: tenantID,
		SlotID:   testSlot.SlotID(),
		MemberID: memberID,
		ActorID:  common.NewMemberID(),
		Note:     "",
	}

	_, err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when member not found")
	}
}

// =====================================================
// CancelAssignmentUsecase Tests
// =====================================================

func TestCancelAssignmentUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	assignmentID := shift.NewAssignmentID()

	assignmentRepo := &MockShiftAssignmentRepository{
		deleteFunc: func(ctx context.Context, tid common.TenantID, aid shift.AssignmentID) error {
			return nil
		},
	}

	usecase := appshift.NewCancelAssignmentUsecase(assignmentRepo)

	input := appshift.CancelAssignmentInput{
		TenantID:     tenantID,
		AssignmentID: assignmentID,
	}

	err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}
}

func TestCancelAssignmentUsecase_Execute_ErrorWhenDeleteFails(t *testing.T) {
	tenantID := common.NewTenantID()
	assignmentID := shift.NewAssignmentID()

	assignmentRepo := &MockShiftAssignmentRepository{
		deleteFunc: func(ctx context.Context, tid common.TenantID, aid shift.AssignmentID) error {
			return errors.New("database error")
		},
	}

	usecase := appshift.NewCancelAssignmentUsecase(assignmentRepo)

	input := appshift.CancelAssignmentInput{
		TenantID:     tenantID,
		AssignmentID: assignmentID,
	}

	err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when delete fails")
	}
}

// =====================================================
// Helper function to create test instance
// =====================================================

func createTestInstance(t *testing.T, tenantID common.TenantID, eventID common.EventID) *shift.Instance {
	t.Helper()
	now := time.Now()
	instance, err := shift.NewInstance(now, tenantID, eventID, "テストインスタンス", 1, nil)
	if err != nil {
		t.Fatalf("Failed to create test instance: %v", err)
	}
	return instance
}

func createTestShiftSlotWithInstance(t *testing.T, tenantID common.TenantID, businessDayID event.BusinessDayID, instanceID *shift.InstanceID) *shift.ShiftSlot {
	t.Helper()
	now := time.Now()

	slot, err := shift.NewShiftSlot(
		now,
		tenantID,
		businessDayID,
		instanceID,
		"テストシフト",
		"VRChat Japan",
		time.Date(2024, 1, 1, 20, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 22, 0, 0, 0, time.UTC),
		3,
		1,
	)
	if err != nil {
		t.Fatalf("Failed to create test shift slot: %v", err)
	}
	return slot
}

// =====================================================
// DeleteInstanceUsecase Tests
// =====================================================

func TestDeleteInstanceUsecase_CheckDeletable_Success_CanDelete(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	instance := createTestInstance(t, tenantID, eventID)
	instanceID := instance.InstanceID()

	instanceRepo := &MockInstanceRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, iid shift.InstanceID) (*shift.Instance, error) {
			return instance, nil
		},
	}

	slotRepo := &MockShiftSlotRepository{
		findByInstanceIDFunc: func(ctx context.Context, tid common.TenantID, iid shift.InstanceID) ([]*shift.ShiftSlot, error) {
			return []*shift.ShiftSlot{}, nil // No slots
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{}

	usecase := appshift.NewDeleteInstanceUsecase(&MockTxManager{}, instanceRepo, slotRepo, assignmentRepo)

	input := appshift.DeleteInstanceInput{
		TenantID:   tenantID,
		InstanceID: instanceID,
	}

	result, err := usecase.CheckDeletable(context.Background(), input)

	if err != nil {
		t.Fatalf("CheckDeletable() should succeed, got error: %v", err)
	}

	if !result.CanDelete {
		t.Error("CanDelete should be true when no slots exist")
	}

	if result.SlotCount != 0 {
		t.Errorf("SlotCount should be 0, got %d", result.SlotCount)
	}
}

func TestDeleteInstanceUsecase_CheckDeletable_CannotDelete_HasAssignedSlots(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	instance := createTestInstance(t, tenantID, eventID)
	instanceID := instance.InstanceID()
	businessDayID := event.NewBusinessDayID()

	// シフト枠を作成
	slot := createTestShiftSlotWithInstance(t, tenantID, businessDayID, &instanceID)

	instanceRepo := &MockInstanceRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, iid shift.InstanceID) (*shift.Instance, error) {
			return instance, nil
		},
	}

	slotRepo := &MockShiftSlotRepository{
		findByInstanceIDFunc: func(ctx context.Context, tid common.TenantID, iid shift.InstanceID) ([]*shift.ShiftSlot, error) {
			return []*shift.ShiftSlot{slot}, nil
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			return 2, nil // Has assignments
		},
	}

	usecase := appshift.NewDeleteInstanceUsecase(&MockTxManager{}, instanceRepo, slotRepo, assignmentRepo)

	input := appshift.DeleteInstanceInput{
		TenantID:   tenantID,
		InstanceID: instanceID,
	}

	result, err := usecase.CheckDeletable(context.Background(), input)

	if err != nil {
		t.Fatalf("CheckDeletable() should succeed, got error: %v", err)
	}

	if result.CanDelete {
		t.Error("CanDelete should be false when assigned slots exist")
	}

	if result.SlotCount != 1 {
		t.Errorf("SlotCount should be 1, got %d", result.SlotCount)
	}

	if result.AssignedSlots != 1 {
		t.Errorf("AssignedSlots should be 1, got %d", result.AssignedSlots)
	}

	if result.BlockingReason == "" {
		t.Error("BlockingReason should not be empty")
	}
}

func TestDeleteInstanceUsecase_CheckDeletable_CanDelete_NoAssignedSlots(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	instance := createTestInstance(t, tenantID, eventID)
	instanceID := instance.InstanceID()
	businessDayID := event.NewBusinessDayID()

	// シフト枠を作成（担当なし）
	slot := createTestShiftSlotWithInstance(t, tenantID, businessDayID, &instanceID)

	instanceRepo := &MockInstanceRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, iid shift.InstanceID) (*shift.Instance, error) {
			return instance, nil
		},
	}

	slotRepo := &MockShiftSlotRepository{
		findByInstanceIDFunc: func(ctx context.Context, tid common.TenantID, iid shift.InstanceID) ([]*shift.ShiftSlot, error) {
			return []*shift.ShiftSlot{slot}, nil
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			return 0, nil // No assignments
		},
	}

	usecase := appshift.NewDeleteInstanceUsecase(&MockTxManager{}, instanceRepo, slotRepo, assignmentRepo)

	input := appshift.DeleteInstanceInput{
		TenantID:   tenantID,
		InstanceID: instanceID,
	}

	result, err := usecase.CheckDeletable(context.Background(), input)

	if err != nil {
		t.Fatalf("CheckDeletable() should succeed, got error: %v", err)
	}

	if !result.CanDelete {
		t.Error("CanDelete should be true when no assignments exist")
	}

	if result.SlotCount != 1 {
		t.Errorf("SlotCount should be 1, got %d", result.SlotCount)
	}
}

func TestDeleteInstanceUsecase_CheckDeletable_ErrorWhenInstanceNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	instanceID := shift.NewInstanceIDWithTime(time.Now())

	instanceRepo := &MockInstanceRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, iid shift.InstanceID) (*shift.Instance, error) {
			return nil, common.NewNotFoundError("instance", iid.String())
		},
	}

	slotRepo := &MockShiftSlotRepository{}
	assignmentRepo := &MockShiftAssignmentRepository{}

	usecase := appshift.NewDeleteInstanceUsecase(&MockTxManager{}, instanceRepo, slotRepo, assignmentRepo)

	input := appshift.DeleteInstanceInput{
		TenantID:   tenantID,
		InstanceID: instanceID,
	}

	_, err := usecase.CheckDeletable(context.Background(), input)

	if err == nil {
		t.Fatal("CheckDeletable() should fail when instance not found")
	}
}

func TestDeleteInstanceUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	instance := createTestInstance(t, tenantID, eventID)
	instanceID := instance.InstanceID()
	businessDayID := event.NewBusinessDayID()

	// シフト枠を作成（担当なし）
	slot := createTestShiftSlotWithInstance(t, tenantID, businessDayID, &instanceID)
	savedSlots := []*shift.ShiftSlot{}

	instanceRepo := &MockInstanceRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, iid shift.InstanceID) (*shift.Instance, error) {
			return instance, nil
		},
		deleteFunc: func(ctx context.Context, tid common.TenantID, iid shift.InstanceID) error {
			return nil
		},
	}

	slotRepo := &MockShiftSlotRepository{
		findByInstanceIDFunc: func(ctx context.Context, tid common.TenantID, iid shift.InstanceID) ([]*shift.ShiftSlot, error) {
			return []*shift.ShiftSlot{slot}, nil
		},
		saveFunc: func(ctx context.Context, s *shift.ShiftSlot) error {
			savedSlots = append(savedSlots, s)
			return nil
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			return 0, nil
		},
	}

	usecase := appshift.NewDeleteInstanceUsecase(&MockTxManager{}, instanceRepo, slotRepo, assignmentRepo)

	input := appshift.DeleteInstanceInput{
		TenantID:   tenantID,
		InstanceID: instanceID,
	}

	err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	// シフト枠がソフトデリートされたことを確認
	if len(savedSlots) != 1 {
		t.Errorf("Expected 1 slot to be saved, got %d", len(savedSlots))
	}
}

func TestDeleteInstanceUsecase_Execute_ErrorWhenHasAssignedSlots(t *testing.T) {
	tenantID := common.NewTenantID()
	eventID := common.NewEventID()
	instance := createTestInstance(t, tenantID, eventID)
	instanceID := instance.InstanceID()
	businessDayID := event.NewBusinessDayID()

	slot := createTestShiftSlotWithInstance(t, tenantID, businessDayID, &instanceID)

	instanceRepo := &MockInstanceRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, iid shift.InstanceID) (*shift.Instance, error) {
			return instance, nil
		},
	}

	slotRepo := &MockShiftSlotRepository{
		findByInstanceIDFunc: func(ctx context.Context, tid common.TenantID, iid shift.InstanceID) ([]*shift.ShiftSlot, error) {
			return []*shift.ShiftSlot{slot}, nil
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			return 1, nil // Has assignments
		},
	}

	usecase := appshift.NewDeleteInstanceUsecase(&MockTxManager{}, instanceRepo, slotRepo, assignmentRepo)

	input := appshift.DeleteInstanceInput{
		TenantID:   tenantID,
		InstanceID: instanceID,
	}

	err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when assigned slots exist")
	}
}

// =====================================================
// DeleteSlotsByInstanceUsecase Tests
// =====================================================

func TestDeleteSlotsByInstanceUsecase_CheckDeletable_Success_NoSlots(t *testing.T) {
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()
	instanceID := shift.NewInstanceIDWithTime(time.Now())

	slotRepo := &MockShiftSlotRepository{
		findByBusinessDayIDAndInstanceIDFunc: func(ctx context.Context, tid common.TenantID, bdID event.BusinessDayID, iid shift.InstanceID) ([]*shift.ShiftSlot, error) {
			return []*shift.ShiftSlot{}, nil // No slots
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{}

	usecase := appshift.NewDeleteSlotsByInstanceUsecase(&MockTxManager{}, slotRepo, assignmentRepo)

	input := appshift.DeleteSlotsByInstanceInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		InstanceID:    instanceID,
	}

	result, err := usecase.CheckDeletable(context.Background(), input)

	if err != nil {
		t.Fatalf("CheckDeletable() should succeed, got error: %v", err)
	}

	if !result.CanDelete {
		t.Error("CanDelete should be true when no slots exist")
	}

	if result.SlotCount != 0 {
		t.Errorf("SlotCount should be 0, got %d", result.SlotCount)
	}
}

func TestDeleteSlotsByInstanceUsecase_CheckDeletable_CanDelete_NoAssignments(t *testing.T) {
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()
	instanceID := shift.NewInstanceIDWithTime(time.Now())

	slot1 := createTestShiftSlotWithInstance(t, tenantID, businessDayID, &instanceID)
	slot2 := createTestShiftSlotWithInstance(t, tenantID, businessDayID, &instanceID)

	slotRepo := &MockShiftSlotRepository{
		findByBusinessDayIDAndInstanceIDFunc: func(ctx context.Context, tid common.TenantID, bdID event.BusinessDayID, iid shift.InstanceID) ([]*shift.ShiftSlot, error) {
			return []*shift.ShiftSlot{slot1, slot2}, nil
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			return 0, nil // No assignments
		},
	}

	usecase := appshift.NewDeleteSlotsByInstanceUsecase(&MockTxManager{}, slotRepo, assignmentRepo)

	input := appshift.DeleteSlotsByInstanceInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		InstanceID:    instanceID,
	}

	result, err := usecase.CheckDeletable(context.Background(), input)

	if err != nil {
		t.Fatalf("CheckDeletable() should succeed, got error: %v", err)
	}

	if !result.CanDelete {
		t.Error("CanDelete should be true when no assignments exist")
	}

	if result.SlotCount != 2 {
		t.Errorf("SlotCount should be 2, got %d", result.SlotCount)
	}
}

func TestDeleteSlotsByInstanceUsecase_CheckDeletable_CannotDelete_HasAssignments(t *testing.T) {
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()
	instanceID := shift.NewInstanceIDWithTime(time.Now())

	slot1 := createTestShiftSlotWithInstance(t, tenantID, businessDayID, &instanceID)
	slot2 := createTestShiftSlotWithInstance(t, tenantID, businessDayID, &instanceID)

	slotRepo := &MockShiftSlotRepository{
		findByBusinessDayIDAndInstanceIDFunc: func(ctx context.Context, tid common.TenantID, bdID event.BusinessDayID, iid shift.InstanceID) ([]*shift.ShiftSlot, error) {
			return []*shift.ShiftSlot{slot1, slot2}, nil
		},
	}

	callCount := 0
	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			callCount++
			if callCount == 1 {
				return 1, nil // First slot has assignments
			}
			return 0, nil
		},
	}

	usecase := appshift.NewDeleteSlotsByInstanceUsecase(&MockTxManager{}, slotRepo, assignmentRepo)

	input := appshift.DeleteSlotsByInstanceInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		InstanceID:    instanceID,
	}

	result, err := usecase.CheckDeletable(context.Background(), input)

	if err != nil {
		t.Fatalf("CheckDeletable() should succeed, got error: %v", err)
	}

	if result.CanDelete {
		t.Error("CanDelete should be false when assignments exist")
	}

	if result.SlotCount != 2 {
		t.Errorf("SlotCount should be 2, got %d", result.SlotCount)
	}

	if result.AssignedSlots != 1 {
		t.Errorf("AssignedSlots should be 1, got %d", result.AssignedSlots)
	}

	if result.BlockingReason == "" {
		t.Error("BlockingReason should not be empty")
	}
}

func TestDeleteSlotsByInstanceUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()
	instanceID := shift.NewInstanceIDWithTime(time.Now())

	slot1 := createTestShiftSlotWithInstance(t, tenantID, businessDayID, &instanceID)
	slot2 := createTestShiftSlotWithInstance(t, tenantID, businessDayID, &instanceID)
	savedSlots := []*shift.ShiftSlot{}

	slotRepo := &MockShiftSlotRepository{
		findByBusinessDayIDAndInstanceIDFunc: func(ctx context.Context, tid common.TenantID, bdID event.BusinessDayID, iid shift.InstanceID) ([]*shift.ShiftSlot, error) {
			return []*shift.ShiftSlot{slot1, slot2}, nil
		},
		saveFunc: func(ctx context.Context, s *shift.ShiftSlot) error {
			savedSlots = append(savedSlots, s)
			return nil
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			return 0, nil
		},
	}

	usecase := appshift.NewDeleteSlotsByInstanceUsecase(&MockTxManager{}, slotRepo, assignmentRepo)

	input := appshift.DeleteSlotsByInstanceInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		InstanceID:    instanceID,
	}

	err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, got error: %v", err)
	}

	// 2つのシフト枠がソフトデリートされたことを確認
	if len(savedSlots) != 2 {
		t.Errorf("Expected 2 slots to be saved, got %d", len(savedSlots))
	}
}

func TestDeleteSlotsByInstanceUsecase_Execute_ErrorWhenHasAssignments(t *testing.T) {
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()
	instanceID := shift.NewInstanceIDWithTime(time.Now())

	slot := createTestShiftSlotWithInstance(t, tenantID, businessDayID, &instanceID)

	slotRepo := &MockShiftSlotRepository{
		findByBusinessDayIDAndInstanceIDFunc: func(ctx context.Context, tid common.TenantID, bdID event.BusinessDayID, iid shift.InstanceID) ([]*shift.ShiftSlot, error) {
			return []*shift.ShiftSlot{slot}, nil
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{
		countConfirmedBySlotFunc: func(ctx context.Context, tid common.TenantID, slotID shift.SlotID) (int, error) {
			return 1, nil // Has assignments
		},
	}

	usecase := appshift.NewDeleteSlotsByInstanceUsecase(&MockTxManager{}, slotRepo, assignmentRepo)

	input := appshift.DeleteSlotsByInstanceInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		InstanceID:    instanceID,
	}

	err := usecase.Execute(context.Background(), input)

	if err == nil {
		t.Fatal("Execute() should fail when assignments exist")
	}
}

func TestDeleteSlotsByInstanceUsecase_Execute_ErrorWhenSlotRepoFails(t *testing.T) {
	tenantID := common.NewTenantID()
	businessDayID := event.NewBusinessDayID()
	instanceID := shift.NewInstanceIDWithTime(time.Now())

	slotRepo := &MockShiftSlotRepository{
		findByBusinessDayIDAndInstanceIDFunc: func(ctx context.Context, tid common.TenantID, bdID event.BusinessDayID, iid shift.InstanceID) ([]*shift.ShiftSlot, error) {
			return nil, errors.New("database error")
		},
	}

	assignmentRepo := &MockShiftAssignmentRepository{}

	usecase := appshift.NewDeleteSlotsByInstanceUsecase(&MockTxManager{}, slotRepo, assignmentRepo)

	input := appshift.DeleteSlotsByInstanceInput{
		TenantID:      tenantID,
		BusinessDayID: businessDayID,
		InstanceID:    instanceID,
	}

	_, err := usecase.CheckDeletable(context.Background(), input)

	if err == nil {
		t.Fatal("CheckDeletable() should fail when repository fails")
	}
}
