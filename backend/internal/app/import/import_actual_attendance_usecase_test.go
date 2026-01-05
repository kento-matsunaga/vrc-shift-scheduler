package importapp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	importjob "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/import"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// =====================================================
// Mock Implementations
// =====================================================

type MockImportJobRepository struct {
	saveFunc   func(ctx context.Context, job *importjob.ImportJob) error
	updateFunc func(ctx context.Context, job *importjob.ImportJob) error
}

func (m *MockImportJobRepository) Save(ctx context.Context, job *importjob.ImportJob) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, job)
	}
	return nil
}

func (m *MockImportJobRepository) Update(ctx context.Context, job *importjob.ImportJob) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, job)
	}
	return nil
}

func (m *MockImportJobRepository) FindByID(ctx context.Context, jobID common.ImportJobID) (*importjob.ImportJob, error) {
	return nil, errors.New("not implemented")
}

func (m *MockImportJobRepository) FindByIDAndTenantID(ctx context.Context, jobID common.ImportJobID, tenantID common.TenantID) (*importjob.ImportJob, error) {
	return nil, errors.New("not implemented")
}

func (m *MockImportJobRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID, limit, offset int) ([]*importjob.ImportJob, error) {
	return nil, errors.New("not implemented")
}

func (m *MockImportJobRepository) CountByTenantID(ctx context.Context, tenantID common.TenantID) (int, error) {
	return 0, nil
}

type MockMemberRepository struct {
	members []*member.Member
}

func (m *MockMemberRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error) {
	return m.members, nil
}

func (m *MockMemberRepository) Save(ctx context.Context, mem *member.Member) error {
	return nil
}

func (m *MockMemberRepository) FindByDisplayName(ctx context.Context, tenantID common.TenantID, displayName string) (*member.Member, error) {
	return nil, errors.New("not implemented")
}

type MockEventRepository struct {
	events []*event.Event
}

func (m *MockEventRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*event.Event, error) {
	return m.events, nil
}

func (m *MockEventRepository) FindByID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) (*event.Event, error) {
	for _, e := range m.events {
		if e.EventID() == eventID {
			return e, nil
		}
	}
	return nil, errors.New("not found")
}

type MockEventBusinessDayRepository struct {
	businessDays []*event.EventBusinessDay
}

func (m *MockEventBusinessDayRepository) FindByTenantIDAndDate(ctx context.Context, tenantID common.TenantID, date time.Time) ([]*event.EventBusinessDay, error) {
	var result []*event.EventBusinessDay
	for _, bd := range m.businessDays {
		if bd.TargetDate().Format("2006-01-02") == date.Format("2006-01-02") {
			result = append(result, bd)
		}
	}
	return result, nil
}

func (m *MockEventBusinessDayRepository) FindByTenantIDAndDateRange(ctx context.Context, tenantID common.TenantID, startDate, endDate time.Time) ([]*event.EventBusinessDay, error) {
	var result []*event.EventBusinessDay
	for _, bd := range m.businessDays {
		if !bd.TargetDate().Before(startDate) && !bd.TargetDate().After(endDate) {
			result = append(result, bd)
		}
	}
	return result, nil
}

func (m *MockEventBusinessDayRepository) Save(ctx context.Context, businessDay *event.EventBusinessDay) error {
	return nil
}

type MockShiftSlotRepository struct {
	slots []*shift.ShiftSlot
}

func (m *MockShiftSlotRepository) FindByBusinessDayIDAndSlotName(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID, slotName string) (*shift.ShiftSlot, error) {
	for _, s := range m.slots {
		if s.BusinessDayID() == businessDayID && s.SlotName() == slotName {
			return s, nil
		}
	}
	return nil, nil
}

func (m *MockShiftSlotRepository) FindByBusinessDayID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) ([]*shift.ShiftSlot, error) {
	var result []*shift.ShiftSlot
	for _, s := range m.slots {
		if s.BusinessDayID() == businessDayID {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *MockShiftSlotRepository) Save(ctx context.Context, slot *shift.ShiftSlot) error {
	return nil
}

type MockShiftAssignmentRepository struct {
	assignments map[string]bool // key: slotID:memberID
}

func (m *MockShiftAssignmentRepository) ExistsBySlotIDAndMemberID(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID, memberID common.MemberID) (bool, error) {
	key := string(slotID) + ":" + string(memberID)
	return m.assignments[key], nil
}

func (m *MockShiftAssignmentRepository) Save(ctx context.Context, assignment *shift.ShiftAssignment) error {
	return nil
}

type MockPositionRepository struct {
	positions []*shift.Position
}

func (m *MockPositionRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*shift.Position, error) {
	return m.positions, nil
}

// MockTxManager is a mock transaction manager for testing
type MockTxManager struct{}

func (m *MockTxManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	// Simply execute the function without actual transaction
	return fn(ctx)
}

// =====================================================
// Test Helper Functions
// =====================================================

func createTestMember(t *testing.T, tenantID common.TenantID, displayName string) *member.Member {
	t.Helper()
	m, err := member.NewMember(time.Now(), tenantID, displayName, "", "")
	if err != nil {
		t.Fatalf("Failed to create test member: %v", err)
	}
	return m
}

func createTestEvent(t *testing.T, tenantID common.TenantID, eventName string) *event.Event {
	t.Helper()
	e, err := event.NewEvent(
		time.Now(),
		tenantID,
		eventName,
		event.EventTypeNormal,
		"",
		event.RecurrenceTypeNone,
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to create test event: %v", err)
	}
	return e
}

func createTestBusinessDay(t *testing.T, tenantID common.TenantID, eventID common.EventID, targetDate time.Time) *event.EventBusinessDay {
	t.Helper()
	startTime, _ := time.Parse("15:04", "20:00")
	endTime, _ := time.Parse("15:04", "22:00")
	bd, err := event.NewEventBusinessDay(
		time.Now(),
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

func createTestPosition(t *testing.T, tenantID common.TenantID, positionName string) *shift.Position {
	t.Helper()
	p, err := shift.NewPosition(tenantID, positionName, "", 0)
	if err != nil {
		t.Fatalf("Failed to create test position: %v", err)
	}
	return p
}

func createTestShiftSlot(t *testing.T, tenantID common.TenantID, businessDayID event.BusinessDayID, positionID shift.PositionID, slotName string) *shift.ShiftSlot {
	t.Helper()
	startTime, _ := time.Parse("15:04", "20:00")
	endTime, _ := time.Parse("15:04", "22:00")
	s, err := shift.NewShiftSlot(time.Now(), tenantID, businessDayID, positionID, slotName, "", startTime, endTime, 2, 1)
	if err != nil {
		t.Fatalf("Failed to create test shift slot: %v", err)
	}
	return s
}

// =====================================================
// Success Cases
// =====================================================

func TestImportActualAttendanceUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	adminID := common.NewAdminID()

	// Create test data
	testMember := createTestMember(t, tenantID, "たろう")
	testEvent := createTestEvent(t, tenantID, "週末イベント")
	testBusinessDay := createTestBusinessDay(t, tenantID, testEvent.EventID(), time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC))
	testPosition := createTestPosition(t, tenantID, "スタッフ")
	testSlot := createTestShiftSlot(t, tenantID, testBusinessDay.BusinessDayID(), testPosition.PositionID(), "受付")

	// CSV data
	csvData := []byte("date,member_name,event_name,slot_name\n2025-01-15,たろう,週末イベント,受付\n")

	usecase := NewImportActualAttendanceUsecase(
		&MockImportJobRepository{},
		&MockMemberRepository{members: []*member.Member{testMember}},
		&MockEventBusinessDayRepository{businessDays: []*event.EventBusinessDay{testBusinessDay}},
		&MockEventRepository{events: []*event.Event{testEvent}},
		&MockShiftSlotRepository{slots: []*shift.ShiftSlot{testSlot}},
		&MockShiftAssignmentRepository{assignments: map[string]bool{}},
		&MockPositionRepository{positions: []*shift.Position{testPosition}},
		&MockTxManager{},
	)

	input := ImportActualAttendanceInput{
		TenantID: tenantID,
		AdminID:  adminID,
		FileName: "test.csv",
		FileData: csvData,
		Options:  importjob.ImportOptions{},
	}

	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, but got error: %v", err)
	}

	if output == nil {
		t.Fatal("Execute() returned nil output")
	}

	if output.Status != importjob.ImportStatusCompleted {
		t.Errorf("Status: expected %s, got %s", importjob.ImportStatusCompleted, output.Status)
	}

	if output.TotalRows != 1 {
		t.Errorf("TotalRows: expected 1, got %d", output.TotalRows)
	}

	if output.SuccessCount != 1 {
		t.Errorf("SuccessCount: expected 1, got %d", output.SuccessCount)
	}

	if output.ErrorCount != 0 {
		t.Errorf("ErrorCount: expected 0, got %d", output.ErrorCount)
	}
}

func TestImportActualAttendanceUsecase_Execute_MultipleRows(t *testing.T) {
	tenantID := common.NewTenantID()
	adminID := common.NewAdminID()

	// Create test data
	member1 := createTestMember(t, tenantID, "たろう")
	member2 := createTestMember(t, tenantID, "はなこ")
	testEvent := createTestEvent(t, tenantID, "週末イベント")
	testBusinessDay := createTestBusinessDay(t, tenantID, testEvent.EventID(), time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC))
	testPosition := createTestPosition(t, tenantID, "スタッフ")
	slot1 := createTestShiftSlot(t, tenantID, testBusinessDay.BusinessDayID(), testPosition.PositionID(), "受付")
	slot2 := createTestShiftSlot(t, tenantID, testBusinessDay.BusinessDayID(), testPosition.PositionID(), "案内")

	// CSV data with 2 rows
	csvData := []byte("date,member_name,event_name,slot_name\n2025-01-15,たろう,週末イベント,受付\n2025-01-15,はなこ,週末イベント,案内\n")

	usecase := NewImportActualAttendanceUsecase(
		&MockImportJobRepository{},
		&MockMemberRepository{members: []*member.Member{member1, member2}},
		&MockEventBusinessDayRepository{businessDays: []*event.EventBusinessDay{testBusinessDay}},
		&MockEventRepository{events: []*event.Event{testEvent}},
		&MockShiftSlotRepository{slots: []*shift.ShiftSlot{slot1, slot2}},
		&MockShiftAssignmentRepository{assignments: map[string]bool{}},
		&MockPositionRepository{positions: []*shift.Position{testPosition}},
		&MockTxManager{},
	)

	input := ImportActualAttendanceInput{
		TenantID: tenantID,
		AdminID:  adminID,
		FileName: "test.csv",
		FileData: csvData,
		Options:  importjob.ImportOptions{},
	}

	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should succeed, but got error: %v", err)
	}

	if output.SuccessCount != 2 {
		t.Errorf("SuccessCount: expected 2, got %d", output.SuccessCount)
	}
}

// =====================================================
// Error Cases
// =====================================================

func TestImportActualAttendanceUsecase_Execute_InvalidCSV(t *testing.T) {
	tenantID := common.NewTenantID()
	adminID := common.NewAdminID()

	// Invalid CSV (missing required column)
	csvData := []byte("invalid_column\ntest\n")

	usecase := NewImportActualAttendanceUsecase(
		&MockImportJobRepository{},
		&MockMemberRepository{},
		&MockEventBusinessDayRepository{},
		&MockEventRepository{},
		&MockShiftSlotRepository{},
		&MockShiftAssignmentRepository{assignments: map[string]bool{}},
		&MockPositionRepository{},
		&MockTxManager{},
	)

	input := ImportActualAttendanceInput{
		TenantID: tenantID,
		AdminID:  adminID,
		FileName: "test.csv",
		FileData: csvData,
		Options:  importjob.ImportOptions{},
	}

	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should not return error, but got: %v", err)
	}

	if output.Status != importjob.ImportStatusFailed {
		t.Errorf("Status: expected %s, got %s", importjob.ImportStatusFailed, output.Status)
	}
}

func TestImportActualAttendanceUsecase_Execute_MemberNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	adminID := common.NewAdminID()

	testEvent := createTestEvent(t, tenantID, "週末イベント")
	testBusinessDay := createTestBusinessDay(t, tenantID, testEvent.EventID(), time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC))
	testPosition := createTestPosition(t, tenantID, "スタッフ")
	testSlot := createTestShiftSlot(t, tenantID, testBusinessDay.BusinessDayID(), testPosition.PositionID(), "受付")

	// CSV with member that doesn't exist
	csvData := []byte("date,member_name,event_name,slot_name\n2025-01-15,存在しないメンバー,週末イベント,受付\n")

	usecase := NewImportActualAttendanceUsecase(
		&MockImportJobRepository{},
		&MockMemberRepository{members: []*member.Member{}}, // No members
		&MockEventBusinessDayRepository{businessDays: []*event.EventBusinessDay{testBusinessDay}},
		&MockEventRepository{events: []*event.Event{testEvent}},
		&MockShiftSlotRepository{slots: []*shift.ShiftSlot{testSlot}},
		&MockShiftAssignmentRepository{assignments: map[string]bool{}},
		&MockPositionRepository{positions: []*shift.Position{testPosition}},
		&MockTxManager{},
	)

	input := ImportActualAttendanceInput{
		TenantID: tenantID,
		AdminID:  adminID,
		FileName: "test.csv",
		FileData: csvData,
		Options:  importjob.ImportOptions{},
	}

	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should not return error, but got: %v", err)
	}

	if output.ErrorCount != 1 {
		t.Errorf("ErrorCount: expected 1, got %d", output.ErrorCount)
	}
}

func TestImportActualAttendanceUsecase_Execute_SlotNotFoundWithoutCreate(t *testing.T) {
	tenantID := common.NewTenantID()
	adminID := common.NewAdminID()

	testMember := createTestMember(t, tenantID, "たろう")
	testEvent := createTestEvent(t, tenantID, "週末イベント")
	testBusinessDay := createTestBusinessDay(t, tenantID, testEvent.EventID(), time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC))

	// CSV with slot that doesn't exist and CreateMissingSlots = false
	csvData := []byte("date,member_name,event_name,slot_name\n2025-01-15,たろう,週末イベント,存在しないスロット\n")

	usecase := NewImportActualAttendanceUsecase(
		&MockImportJobRepository{},
		&MockMemberRepository{members: []*member.Member{testMember}},
		&MockEventBusinessDayRepository{businessDays: []*event.EventBusinessDay{testBusinessDay}},
		&MockEventRepository{events: []*event.Event{testEvent}},
		&MockShiftSlotRepository{slots: []*shift.ShiftSlot{}}, // No slots
		&MockShiftAssignmentRepository{assignments: map[string]bool{}},
		&MockPositionRepository{},
		&MockTxManager{},
	)

	input := ImportActualAttendanceInput{
		TenantID: tenantID,
		AdminID:  adminID,
		FileName: "test.csv",
		FileData: csvData,
		Options: importjob.ImportOptions{
			CreateMissingSlots: false,
		},
	}

	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should not return error, but got: %v", err)
	}

	if output.ErrorCount != 1 {
		t.Errorf("ErrorCount: expected 1, got %d", output.ErrorCount)
	}
}

func TestImportActualAttendanceUsecase_Execute_DuplicateWithSkip(t *testing.T) {
	tenantID := common.NewTenantID()
	adminID := common.NewAdminID()

	testMember := createTestMember(t, tenantID, "たろう")
	testEvent := createTestEvent(t, tenantID, "週末イベント")
	testBusinessDay := createTestBusinessDay(t, tenantID, testEvent.EventID(), time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC))
	testPosition := createTestPosition(t, tenantID, "スタッフ")
	testSlot := createTestShiftSlot(t, tenantID, testBusinessDay.BusinessDayID(), testPosition.PositionID(), "受付")

	// CSV with already assigned member
	csvData := []byte("date,member_name,event_name,slot_name\n2025-01-15,たろう,週末イベント,受付\n")

	// Pre-existing assignment
	assignments := map[string]bool{
		string(testSlot.SlotID()) + ":" + string(testMember.MemberID()): true,
	}

	usecase := NewImportActualAttendanceUsecase(
		&MockImportJobRepository{},
		&MockMemberRepository{members: []*member.Member{testMember}},
		&MockEventBusinessDayRepository{businessDays: []*event.EventBusinessDay{testBusinessDay}},
		&MockEventRepository{events: []*event.Event{testEvent}},
		&MockShiftSlotRepository{slots: []*shift.ShiftSlot{testSlot}},
		&MockShiftAssignmentRepository{assignments: assignments},
		&MockPositionRepository{positions: []*shift.Position{testPosition}},
		&MockTxManager{},
	)

	input := ImportActualAttendanceInput{
		TenantID: tenantID,
		AdminID:  adminID,
		FileName: "test.csv",
		FileData: csvData,
		Options: importjob.ImportOptions{
			SkipExisting: true, // Skip duplicates
		},
	}

	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should not return error, but got: %v", err)
	}

	if output.SkippedCount != 1 {
		t.Errorf("SkippedCount: expected 1, got %d", output.SkippedCount)
	}

	if output.ErrorCount != 0 {
		t.Errorf("ErrorCount: expected 0, got %d", output.ErrorCount)
	}
}

func TestImportActualAttendanceUsecase_Execute_DuplicateWithoutSkip(t *testing.T) {
	tenantID := common.NewTenantID()
	adminID := common.NewAdminID()

	testMember := createTestMember(t, tenantID, "たろう")
	testEvent := createTestEvent(t, tenantID, "週末イベント")
	testBusinessDay := createTestBusinessDay(t, tenantID, testEvent.EventID(), time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC))
	testPosition := createTestPosition(t, tenantID, "スタッフ")
	testSlot := createTestShiftSlot(t, tenantID, testBusinessDay.BusinessDayID(), testPosition.PositionID(), "受付")

	csvData := []byte("date,member_name,event_name,slot_name\n2025-01-15,たろう,週末イベント,受付\n")

	// Pre-existing assignment
	assignments := map[string]bool{
		string(testSlot.SlotID()) + ":" + string(testMember.MemberID()): true,
	}

	usecase := NewImportActualAttendanceUsecase(
		&MockImportJobRepository{},
		&MockMemberRepository{members: []*member.Member{testMember}},
		&MockEventBusinessDayRepository{businessDays: []*event.EventBusinessDay{testBusinessDay}},
		&MockEventRepository{events: []*event.Event{testEvent}},
		&MockShiftSlotRepository{slots: []*shift.ShiftSlot{testSlot}},
		&MockShiftAssignmentRepository{assignments: assignments},
		&MockPositionRepository{positions: []*shift.Position{testPosition}},
		&MockTxManager{},
	)

	input := ImportActualAttendanceInput{
		TenantID: tenantID,
		AdminID:  adminID,
		FileName: "test.csv",
		FileData: csvData,
		Options: importjob.ImportOptions{
			SkipExisting: false, // Don't skip, report as error
		},
	}

	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should not return error, but got: %v", err)
	}

	if output.ErrorCount != 1 {
		t.Errorf("ErrorCount: expected 1, got %d", output.ErrorCount)
	}
}

// =====================================================
// Row Limit Test
// =====================================================

func TestImportActualAttendanceUsecase_Execute_RowLimitExceeded(t *testing.T) {
	tenantID := common.NewTenantID()
	adminID := common.NewAdminID()

	// Create CSV with header and > 10000 rows
	csvData := []byte("date,member_name,event_name,slot_name\n")
	for i := 0; i < 10001; i++ {
		csvData = append(csvData, []byte("2025-01-15,test,test,test\n")...)
	}

	usecase := NewImportActualAttendanceUsecase(
		&MockImportJobRepository{},
		&MockMemberRepository{},
		&MockEventBusinessDayRepository{},
		&MockEventRepository{},
		&MockShiftSlotRepository{},
		&MockShiftAssignmentRepository{assignments: map[string]bool{}},
		&MockPositionRepository{},
		&MockTxManager{},
	)

	input := ImportActualAttendanceInput{
		TenantID: tenantID,
		AdminID:  adminID,
		FileName: "test.csv",
		FileData: csvData,
		Options:  importjob.ImportOptions{},
	}

	output, err := usecase.Execute(context.Background(), input)

	if err != nil {
		t.Fatalf("Execute() should not return error, but got: %v", err)
	}

	if output.Status != importjob.ImportStatusFailed {
		t.Errorf("Status: expected %s, got %s", importjob.ImportStatusFailed, output.Status)
	}
}
