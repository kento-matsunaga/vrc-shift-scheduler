package schedule_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appschedule "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/schedule"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =====================================================
// Mock Implementations for ConvertToAttendanceUsecase
// =====================================================

// MockAttendanceCollectionRepository is a mock implementation of attendance.AttendanceCollectionRepository
type MockAttendanceCollectionRepository struct {
	saveFunc                              func(ctx context.Context, collection *attendance.AttendanceCollection) error
	findByIDFunc                          func(ctx context.Context, tenantID common.TenantID, id common.CollectionID) (*attendance.AttendanceCollection, error)
	findByTokenFunc                       func(ctx context.Context, token common.PublicToken) (*attendance.AttendanceCollection, error)
	findByTenantIDFunc                    func(ctx context.Context, tenantID common.TenantID) ([]*attendance.AttendanceCollection, error)
	upsertResponseFunc                    func(ctx context.Context, response *attendance.AttendanceResponse) error
	findResponsesByCollectionIDFunc       func(ctx context.Context, collectionID common.CollectionID) ([]*attendance.AttendanceResponse, error)
	findResponsesByMemberIDFunc           func(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) ([]*attendance.AttendanceResponse, error)
	findResponsesByCollectionIDAndMemberIDFunc func(ctx context.Context, tenantID common.TenantID, collectionID common.CollectionID, memberID common.MemberID) ([]*attendance.AttendanceResponse, error)
	saveTargetDatesFunc                   func(ctx context.Context, collectionID common.CollectionID, targetDates []*attendance.TargetDate) error
	findTargetDatesByCollectionIDFunc     func(ctx context.Context, collectionID common.CollectionID) ([]*attendance.TargetDate, error)
	saveGroupAssignmentsFunc              func(ctx context.Context, collectionID common.CollectionID, assignments []*attendance.CollectionGroupAssignment) error
	findGroupAssignmentsByCollectionIDFunc func(ctx context.Context, collectionID common.CollectionID) ([]*attendance.CollectionGroupAssignment, error)
	saveRoleAssignmentsFunc               func(ctx context.Context, collectionID common.CollectionID, assignments []*attendance.CollectionRoleAssignment) error
	findRoleAssignmentsByCollectionIDFunc func(ctx context.Context, collectionID common.CollectionID) ([]*attendance.CollectionRoleAssignment, error)
}

func (m *MockAttendanceCollectionRepository) Save(ctx context.Context, collection *attendance.AttendanceCollection) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, collection)
	}
	return nil
}

func (m *MockAttendanceCollectionRepository) FindByID(ctx context.Context, tenantID common.TenantID, id common.CollectionID) (*attendance.AttendanceCollection, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, tenantID, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAttendanceCollectionRepository) FindByToken(ctx context.Context, token common.PublicToken) (*attendance.AttendanceCollection, error) {
	if m.findByTokenFunc != nil {
		return m.findByTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAttendanceCollectionRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*attendance.AttendanceCollection, error) {
	if m.findByTenantIDFunc != nil {
		return m.findByTenantIDFunc(ctx, tenantID)
	}
	return nil, errors.New("not implemented")
}

func (m *MockAttendanceCollectionRepository) UpsertResponse(ctx context.Context, response *attendance.AttendanceResponse) error {
	if m.upsertResponseFunc != nil {
		return m.upsertResponseFunc(ctx, response)
	}
	return nil
}

func (m *MockAttendanceCollectionRepository) FindResponsesByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*attendance.AttendanceResponse, error) {
	if m.findResponsesByCollectionIDFunc != nil {
		return m.findResponsesByCollectionIDFunc(ctx, collectionID)
	}
	return []*attendance.AttendanceResponse{}, nil
}

func (m *MockAttendanceCollectionRepository) FindResponsesByMemberID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) ([]*attendance.AttendanceResponse, error) {
	if m.findResponsesByMemberIDFunc != nil {
		return m.findResponsesByMemberIDFunc(ctx, tenantID, memberID)
	}
	return []*attendance.AttendanceResponse{}, nil
}

func (m *MockAttendanceCollectionRepository) FindResponsesByCollectionIDAndMemberID(ctx context.Context, tenantID common.TenantID, collectionID common.CollectionID, memberID common.MemberID) ([]*attendance.AttendanceResponse, error) {
	if m.findResponsesByCollectionIDAndMemberIDFunc != nil {
		return m.findResponsesByCollectionIDAndMemberIDFunc(ctx, tenantID, collectionID, memberID)
	}
	return []*attendance.AttendanceResponse{}, nil
}

func (m *MockAttendanceCollectionRepository) SaveTargetDates(ctx context.Context, collectionID common.CollectionID, targetDates []*attendance.TargetDate) error {
	if m.saveTargetDatesFunc != nil {
		return m.saveTargetDatesFunc(ctx, collectionID, targetDates)
	}
	return nil
}

func (m *MockAttendanceCollectionRepository) FindTargetDatesByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*attendance.TargetDate, error) {
	if m.findTargetDatesByCollectionIDFunc != nil {
		return m.findTargetDatesByCollectionIDFunc(ctx, collectionID)
	}
	return []*attendance.TargetDate{}, nil
}

func (m *MockAttendanceCollectionRepository) SaveGroupAssignments(ctx context.Context, collectionID common.CollectionID, assignments []*attendance.CollectionGroupAssignment) error {
	if m.saveGroupAssignmentsFunc != nil {
		return m.saveGroupAssignmentsFunc(ctx, collectionID, assignments)
	}
	return nil
}

func (m *MockAttendanceCollectionRepository) FindGroupAssignmentsByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*attendance.CollectionGroupAssignment, error) {
	if m.findGroupAssignmentsByCollectionIDFunc != nil {
		return m.findGroupAssignmentsByCollectionIDFunc(ctx, collectionID)
	}
	return []*attendance.CollectionGroupAssignment{}, nil
}

func (m *MockAttendanceCollectionRepository) SaveRoleAssignments(ctx context.Context, collectionID common.CollectionID, assignments []*attendance.CollectionRoleAssignment) error {
	if m.saveRoleAssignmentsFunc != nil {
		return m.saveRoleAssignmentsFunc(ctx, collectionID, assignments)
	}
	return nil
}

func (m *MockAttendanceCollectionRepository) FindRoleAssignmentsByCollectionID(ctx context.Context, collectionID common.CollectionID) ([]*attendance.CollectionRoleAssignment, error) {
	if m.findRoleAssignmentsByCollectionIDFunc != nil {
		return m.findRoleAssignmentsByCollectionIDFunc(ctx, collectionID)
	}
	return []*attendance.CollectionRoleAssignment{}, nil
}

// MockMemberGroupRepository is a mock implementation of member.MemberGroupRepository
type MockMemberGroupRepository struct {
	findMemberIDsByGroupIDFunc func(ctx context.Context, groupID common.MemberGroupID) ([]common.MemberID, error)
}

func (m *MockMemberGroupRepository) Save(ctx context.Context, group *member.MemberGroup) error {
	return nil
}

func (m *MockMemberGroupRepository) FindByID(ctx context.Context, tenantID common.TenantID, groupID common.MemberGroupID) (*member.MemberGroup, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMemberGroupRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.MemberGroup, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMemberGroupRepository) Delete(ctx context.Context, tenantID common.TenantID, groupID common.MemberGroupID) error {
	return nil
}

func (m *MockMemberGroupRepository) AssignMember(ctx context.Context, groupID common.MemberGroupID, memberID common.MemberID) error {
	return nil
}

func (m *MockMemberGroupRepository) RemoveMember(ctx context.Context, groupID common.MemberGroupID, memberID common.MemberID) error {
	return nil
}

func (m *MockMemberGroupRepository) FindMemberIDsByGroupID(ctx context.Context, groupID common.MemberGroupID) ([]common.MemberID, error) {
	if m.findMemberIDsByGroupIDFunc != nil {
		return m.findMemberIDsByGroupIDFunc(ctx, groupID)
	}
	return []common.MemberID{}, nil
}

func (m *MockMemberGroupRepository) FindGroupIDsByMemberID(ctx context.Context, memberID common.MemberID) ([]common.MemberGroupID, error) {
	return []common.MemberGroupID{}, nil
}

func (m *MockMemberGroupRepository) SetMemberGroups(ctx context.Context, memberID common.MemberID, groupIDs []common.MemberGroupID) error {
	return nil
}

// MockTxManager is a mock implementation of services.TxManager
type MockTxManager struct {
	withTxFunc func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *MockTxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.withTxFunc != nil {
		return m.withTxFunc(ctx, fn)
	}
	// Default: execute function without transaction
	return fn(ctx)
}

// =====================================================
// Test Helper Functions
// =====================================================

func createTestScheduleWithCandidates(t *testing.T, tenantID common.TenantID, numCandidates int) *schedule.DateSchedule {
	t.Helper()
	now := time.Now()
	scheduleID := common.NewScheduleID()

	candidates := make([]*schedule.CandidateDate, numCandidates)
	for i := 0; i < numCandidates; i++ {
		candidate, err := schedule.NewCandidateDate(now, scheduleID, now.AddDate(0, 0, i+1), nil, nil, i)
		require.NoError(t, err, "Failed to create test candidate")
		candidates[i] = candidate
	}

	sch, err := schedule.NewDateSchedule(now, scheduleID, tenantID, "Test Schedule", "Test Description", nil, candidates, nil)
	require.NoError(t, err, "Failed to create test schedule")
	return sch
}

func createTestScheduleResponse(t *testing.T, tenantID common.TenantID, scheduleID common.ScheduleID, memberID common.MemberID, candidateID common.CandidateID, avail schedule.Availability) *schedule.DateScheduleResponse {
	t.Helper()
	now := time.Now()
	resp, err := schedule.NewDateScheduleResponse(now, scheduleID, tenantID, memberID, candidateID, avail, "")
	require.NoError(t, err, "Failed to create test response")
	return resp
}

// =====================================================
// ConvertToAttendanceUsecase Tests
// =====================================================

func TestConvertToAttendanceUsecase_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	testSchedule := createTestScheduleWithCandidates(t, tenantID, 2)
	candidateID := testSchedule.Candidates()[0].CandidateID()

	scheduleRepo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return testSchedule, nil
		},
		findGroupAssignmentsByScheduleIDFunc: func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.ScheduleGroupAssignment, error) {
			return []*schedule.ScheduleGroupAssignment{}, nil
		},
		findResponsesByScheduleIDFunc: func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.DateScheduleResponse, error) {
			return []*schedule.DateScheduleResponse{}, nil
		},
	}

	attendanceRepo := &MockAttendanceCollectionRepository{}
	memberGroupRepo := &MockMemberGroupRepository{}
	txManager := &MockTxManager{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewConvertToAttendanceUsecase(
		scheduleRepo,
		attendanceRepo,
		memberGroupRepo,
		txManager,
		clock,
	)

	input := appschedule.ConvertToAttendanceInput{
		TenantID:     tenantID.String(),
		ScheduleID:   testSchedule.ScheduleID().String(),
		CandidateIDs: []string{candidateID.String()},
		Title:        "New Attendance Title",
	}

	result, err := usecase.Execute(context.Background(), input)

	require.NoError(t, err, "Execute() should succeed")
	assert.NotNil(t, result, "Result should not be nil")
	assert.NotEmpty(t, result.CollectionID, "CollectionID should not be empty")
	assert.NotEmpty(t, result.PublicToken, "PublicToken should not be empty")
	assert.Equal(t, "New Attendance Title", result.Title, "Title should match input")
}

func TestConvertToAttendanceUsecase_ErrorWhenScheduleNotFound(t *testing.T) {
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()
	candidateID := common.NewCandidateID()

	scheduleRepo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return nil, common.NewNotFoundError("schedule", id.String())
		},
	}

	attendanceRepo := &MockAttendanceCollectionRepository{}
	memberGroupRepo := &MockMemberGroupRepository{}
	txManager := &MockTxManager{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewConvertToAttendanceUsecase(
		scheduleRepo,
		attendanceRepo,
		memberGroupRepo,
		txManager,
		clock,
	)

	input := appschedule.ConvertToAttendanceInput{
		TenantID:     tenantID.String(),
		ScheduleID:   scheduleID.String(),
		CandidateIDs: []string{candidateID.String()},
	}

	_, err := usecase.Execute(context.Background(), input)

	assert.Error(t, err, "Execute() should fail when schedule not found")
}

func TestConvertToAttendanceUsecase_ErrorWhenInvalidCandidateID(t *testing.T) {
	tenantID := common.NewTenantID()
	testSchedule := createTestScheduleWithCandidates(t, tenantID, 2)

	scheduleRepo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return testSchedule, nil
		},
		findGroupAssignmentsByScheduleIDFunc: func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.ScheduleGroupAssignment, error) {
			return []*schedule.ScheduleGroupAssignment{}, nil
		},
		findResponsesByScheduleIDFunc: func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.DateScheduleResponse, error) {
			return []*schedule.DateScheduleResponse{}, nil
		},
	}

	attendanceRepo := &MockAttendanceCollectionRepository{}
	memberGroupRepo := &MockMemberGroupRepository{}
	txManager := &MockTxManager{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewConvertToAttendanceUsecase(
		scheduleRepo,
		attendanceRepo,
		memberGroupRepo,
		txManager,
		clock,
	)

	// Use a candidate ID that doesn't exist in the schedule
	nonExistentCandidateID := common.NewCandidateID()
	input := appschedule.ConvertToAttendanceInput{
		TenantID:     tenantID.String(),
		ScheduleID:   testSchedule.ScheduleID().String(),
		CandidateIDs: []string{nonExistentCandidateID.String()},
	}

	_, err := usecase.Execute(context.Background(), input)

	assert.Error(t, err, "Execute() should fail when candidate ID is not in the schedule")
}

func TestConvertToAttendanceUsecase_ErrorWhenNoCandidateIDsProvided(t *testing.T) {
	tenantID := common.NewTenantID()
	scheduleID := common.NewScheduleID()

	scheduleRepo := &MockDateScheduleRepository{}
	attendanceRepo := &MockAttendanceCollectionRepository{}
	memberGroupRepo := &MockMemberGroupRepository{}
	txManager := &MockTxManager{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewConvertToAttendanceUsecase(
		scheduleRepo,
		attendanceRepo,
		memberGroupRepo,
		txManager,
		clock,
	)

	input := appschedule.ConvertToAttendanceInput{
		TenantID:     tenantID.String(),
		ScheduleID:   scheduleID.String(),
		CandidateIDs: []string{}, // Empty slice
	}

	_, err := usecase.Execute(context.Background(), input)

	assert.Error(t, err, "Execute() should fail when no candidate IDs are provided")
}

func TestConvertToAttendanceUsecase_ResponseMappingCorrect(t *testing.T) {
	tenantID := common.NewTenantID()
	testSchedule := createTestScheduleWithCandidates(t, tenantID, 1)
	candidateID := testSchedule.Candidates()[0].CandidateID()

	memberID1 := common.NewMemberID()
	memberID2 := common.NewMemberID()
	memberID3 := common.NewMemberID()

	// Create responses with different availability types
	responses := []*schedule.DateScheduleResponse{
		createTestScheduleResponse(t, tenantID, testSchedule.ScheduleID(), memberID1, candidateID, schedule.AvailabilityAvailable),
		createTestScheduleResponse(t, tenantID, testSchedule.ScheduleID(), memberID2, candidateID, schedule.AvailabilityUnavailable),
		createTestScheduleResponse(t, tenantID, testSchedule.ScheduleID(), memberID3, candidateID, schedule.AvailabilityMaybe),
	}

	// Track the response types that were saved
	savedResponses := make([]*attendance.AttendanceResponse, 0)

	scheduleRepo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return testSchedule, nil
		},
		findGroupAssignmentsByScheduleIDFunc: func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.ScheduleGroupAssignment, error) {
			return []*schedule.ScheduleGroupAssignment{}, nil
		},
		findResponsesByScheduleIDFunc: func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.DateScheduleResponse, error) {
			return responses, nil
		},
	}

	attendanceRepo := &MockAttendanceCollectionRepository{
		upsertResponseFunc: func(ctx context.Context, response *attendance.AttendanceResponse) error {
			savedResponses = append(savedResponses, response)
			return nil
		},
	}
	memberGroupRepo := &MockMemberGroupRepository{}
	txManager := &MockTxManager{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewConvertToAttendanceUsecase(
		scheduleRepo,
		attendanceRepo,
		memberGroupRepo,
		txManager,
		clock,
	)

	input := appschedule.ConvertToAttendanceInput{
		TenantID:     tenantID.String(),
		ScheduleID:   testSchedule.ScheduleID().String(),
		CandidateIDs: []string{candidateID.String()},
	}

	_, err := usecase.Execute(context.Background(), input)

	require.NoError(t, err, "Execute() should succeed")
	require.Len(t, savedResponses, 3, "Should save 3 responses")

	// Verify mappings: available->attending, unavailable->absent, maybe->undecided
	responseTypeByMember := make(map[string]attendance.ResponseType)
	for _, resp := range savedResponses {
		responseTypeByMember[resp.MemberID().String()] = resp.Response()
	}

	assert.Equal(t, attendance.ResponseTypeAttending, responseTypeByMember[memberID1.String()], "available should map to attending")
	assert.Equal(t, attendance.ResponseTypeAbsent, responseTypeByMember[memberID2.String()], "unavailable should map to absent")
	assert.Equal(t, attendance.ResponseTypeUndecided, responseTypeByMember[memberID3.String()], "maybe should map to undecided")
}

func TestConvertToAttendanceUsecase_GroupAssignmentsCopied(t *testing.T) {
	tenantID := common.NewTenantID()
	testSchedule := createTestScheduleWithCandidates(t, tenantID, 1)
	candidateID := testSchedule.Candidates()[0].CandidateID()

	groupID1 := common.NewMemberGroupID()
	groupID2 := common.NewMemberGroupID()

	// Create schedule group assignments
	now := time.Now()
	scheduleGroupAssignment1, _ := schedule.NewScheduleGroupAssignment(now, testSchedule.ScheduleID(), groupID1)
	scheduleGroupAssignment2, _ := schedule.NewScheduleGroupAssignment(now, testSchedule.ScheduleID(), groupID2)
	scheduleGroupAssignments := []*schedule.ScheduleGroupAssignment{scheduleGroupAssignment1, scheduleGroupAssignment2}

	// Track what group assignments were saved
	var savedGroupAssignments []*attendance.CollectionGroupAssignment

	scheduleRepo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return testSchedule, nil
		},
		findGroupAssignmentsByScheduleIDFunc: func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.ScheduleGroupAssignment, error) {
			return scheduleGroupAssignments, nil
		},
		findResponsesByScheduleIDFunc: func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.DateScheduleResponse, error) {
			return []*schedule.DateScheduleResponse{}, nil
		},
	}

	attendanceRepo := &MockAttendanceCollectionRepository{
		saveGroupAssignmentsFunc: func(ctx context.Context, collectionID common.CollectionID, assignments []*attendance.CollectionGroupAssignment) error {
			savedGroupAssignments = assignments
			return nil
		},
	}
	memberGroupRepo := &MockMemberGroupRepository{}
	txManager := &MockTxManager{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewConvertToAttendanceUsecase(
		scheduleRepo,
		attendanceRepo,
		memberGroupRepo,
		txManager,
		clock,
	)

	input := appschedule.ConvertToAttendanceInput{
		TenantID:     tenantID.String(),
		ScheduleID:   testSchedule.ScheduleID().String(),
		CandidateIDs: []string{candidateID.String()},
	}

	_, err := usecase.Execute(context.Background(), input)

	require.NoError(t, err, "Execute() should succeed")
	require.Len(t, savedGroupAssignments, 2, "Should copy 2 group assignments")

	// Verify the group IDs were copied
	savedGroupIDs := make(map[string]bool)
	for _, ga := range savedGroupAssignments {
		savedGroupIDs[ga.GroupID().String()] = true
	}

	assert.True(t, savedGroupIDs[groupID1.String()], "Group 1 should be copied")
	assert.True(t, savedGroupIDs[groupID2.String()], "Group 2 should be copied")
}

func TestConvertToAttendanceUsecase_UnrespondedMembersIncluded(t *testing.T) {
	tenantID := common.NewTenantID()
	testSchedule := createTestScheduleWithCandidates(t, tenantID, 1)
	candidateID := testSchedule.Candidates()[0].CandidateID()

	// Create group with 3 members
	groupID := common.NewMemberGroupID()
	respondedMemberID := common.NewMemberID()
	unrespondedMemberID1 := common.NewMemberID()
	unrespondedMemberID2 := common.NewMemberID()
	allMemberIDs := []common.MemberID{respondedMemberID, unrespondedMemberID1, unrespondedMemberID2}

	// Only one member has responded
	responses := []*schedule.DateScheduleResponse{
		createTestScheduleResponse(t, tenantID, testSchedule.ScheduleID(), respondedMemberID, candidateID, schedule.AvailabilityAvailable),
	}

	// Create schedule group assignment
	now := time.Now()
	scheduleGroupAssignment, _ := schedule.NewScheduleGroupAssignment(now, testSchedule.ScheduleID(), groupID)

	// Track saved responses
	savedResponses := make([]*attendance.AttendanceResponse, 0)

	scheduleRepo := &MockDateScheduleRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, id common.ScheduleID) (*schedule.DateSchedule, error) {
			return testSchedule, nil
		},
		findGroupAssignmentsByScheduleIDFunc: func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.ScheduleGroupAssignment, error) {
			return []*schedule.ScheduleGroupAssignment{scheduleGroupAssignment}, nil
		},
		findResponsesByScheduleIDFunc: func(ctx context.Context, scheduleID common.ScheduleID) ([]*schedule.DateScheduleResponse, error) {
			return responses, nil
		},
	}

	attendanceRepo := &MockAttendanceCollectionRepository{
		upsertResponseFunc: func(ctx context.Context, response *attendance.AttendanceResponse) error {
			savedResponses = append(savedResponses, response)
			return nil
		},
	}

	memberGroupRepo := &MockMemberGroupRepository{
		findMemberIDsByGroupIDFunc: func(ctx context.Context, gid common.MemberGroupID) ([]common.MemberID, error) {
			return allMemberIDs, nil
		},
	}
	txManager := &MockTxManager{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appschedule.NewConvertToAttendanceUsecase(
		scheduleRepo,
		attendanceRepo,
		memberGroupRepo,
		txManager,
		clock,
	)

	input := appschedule.ConvertToAttendanceInput{
		TenantID:     tenantID.String(),
		ScheduleID:   testSchedule.ScheduleID().String(),
		CandidateIDs: []string{candidateID.String()},
	}

	_, err := usecase.Execute(context.Background(), input)

	require.NoError(t, err, "Execute() should succeed")

	// Verify all 3 members have responses
	require.Len(t, savedResponses, 3, "Should have responses for all 3 members (including unresponded)")

	// Verify response types
	responseTypeByMember := make(map[string]attendance.ResponseType)
	for _, resp := range savedResponses {
		responseTypeByMember[resp.MemberID().String()] = resp.Response()
	}

	assert.Equal(t, attendance.ResponseTypeAttending, responseTypeByMember[respondedMemberID.String()], "Responded member should have attending status")
	assert.Equal(t, attendance.ResponseTypeUndecided, responseTypeByMember[unrespondedMemberID1.String()], "Unresponded member 1 should have undecided status")
	assert.Equal(t, attendance.ResponseTypeUndecided, responseTypeByMember[unrespondedMemberID2.String()], "Unresponded member 2 should have undecided status")
}
