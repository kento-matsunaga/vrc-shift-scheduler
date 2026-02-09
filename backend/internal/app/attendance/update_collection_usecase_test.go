package attendance_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appattendance "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

func TestUpdateCollectionUsecase_Execute_Success(t *testing.T) {
	tenantID := common.NewTenantID()
	collection := createTestCollection(t, tenantID)
	collectionID := collection.CollectionID()

	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return collection, nil
		},
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
	}
	txManager := &MockTxManager{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appattendance.NewUpdateCollectionUsecase(repo, txManager, clock)

	input := appattendance.UpdateCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: collectionID.String(),
		Title:        "Updated Title",
		Description:  "Updated Description",
	}

	output, err := usecase.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() should succeed: %v", err)
	}
	if output.Title != "Updated Title" {
		t.Errorf("Title mismatch: got %v, want %v", output.Title, "Updated Title")
	}
}

func TestUpdateCollectionUsecase_Execute_WithTargetDates(t *testing.T) {
	tenantID := common.NewTenantID()
	collection := createTestCollection(t, tenantID)
	collectionID := collection.CollectionID()

	existingTargetDateID := common.NewTargetDateID()
	existingTD, err := attendance.ReconstructTargetDate(
		existingTargetDateID, collectionID,
		time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		nil, nil, 0, time.Now(),
	)
	if err != nil {
		t.Fatalf("ReconstructTargetDate() should succeed: %v", err)
	}

	var replacedDates []*attendance.TargetDate
	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return collection, nil
		},
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
		findTargetDatesByCollectionIDFunc: func(ctx context.Context, cid common.CollectionID) ([]*attendance.TargetDate, error) {
			return []*attendance.TargetDate{existingTD}, nil
		},
		replaceTargetDatesFunc: func(ctx context.Context, cid common.CollectionID, dates []*attendance.TargetDate) error {
			replacedDates = dates
			return nil
		},
	}
	txManager := &MockTxManager{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appattendance.NewUpdateCollectionUsecase(repo, txManager, clock)

	startTime := "20:00"
	endTime := "23:00"
	input := appattendance.UpdateCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: collectionID.String(),
		Title:        "Updated",
		Description:  "",
		TargetDates: []appattendance.UpdateTargetDateInput{
			{
				TargetDateID: existingTargetDateID.String(),
				TargetDate:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
				StartTime:    &startTime,
				EndTime:      &endTime,
			},
			{
				TargetDate: time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	_, err = usecase.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() should succeed: %v", err)
	}

	if len(replacedDates) != 2 {
		t.Fatalf("ReplaceTargetDates should receive 2 dates, got %d", len(replacedDates))
	}
	if replacedDates[0].TargetDateID() != existingTargetDateID {
		t.Errorf("First date should be the existing one: got %v, want %v", replacedDates[0].TargetDateID(), existingTargetDateID)
	}
	if *replacedDates[0].StartTime() != "20:00" {
		t.Errorf("First date start_time should be updated: got %v", *replacedDates[0].StartTime())
	}
}

func TestUpdateCollectionUsecase_Execute_ErrorWhenInvalidTargetDateID(t *testing.T) {
	tenantID := common.NewTenantID()
	collection := createTestCollection(t, tenantID)
	collectionID := collection.CollectionID()

	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return collection, nil
		},
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
		findTargetDatesByCollectionIDFunc: func(ctx context.Context, cid common.CollectionID) ([]*attendance.TargetDate, error) {
			return []*attendance.TargetDate{}, nil
		},
	}
	txManager := &MockTxManager{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appattendance.NewUpdateCollectionUsecase(repo, txManager, clock)

	input := appattendance.UpdateCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: collectionID.String(),
		Title:        "Updated",
		Description:  "",
		TargetDates: []appattendance.UpdateTargetDateInput{
			{
				TargetDateID: common.NewTargetDateID().String(),
				TargetDate:   time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	_, err := usecase.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("Execute() should fail when target_date_id is not found in collection")
	}

	// エラーメッセージに具体的なIDが含まれていないことを確認（情報漏洩防止）
	if domainErr, ok := err.(*common.DomainError); ok {
		if domainErr.Code() != common.ErrInvalidInput {
			t.Errorf("Error code should be INVALID_INPUT, got %v", domainErr.Code())
		}
	}
}

func TestUpdateCollectionUsecase_Execute_ErrorWhenFindTargetDatesFails(t *testing.T) {
	tenantID := common.NewTenantID()
	collection := createTestCollection(t, tenantID)
	collectionID := collection.CollectionID()

	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return collection, nil
		},
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
		findTargetDatesByCollectionIDFunc: func(ctx context.Context, cid common.CollectionID) ([]*attendance.TargetDate, error) {
			return nil, errors.New("database error")
		},
	}
	txManager := &MockTxManager{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appattendance.NewUpdateCollectionUsecase(repo, txManager, clock)

	input := appattendance.UpdateCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: collectionID.String(),
		Title:        "Updated",
		Description:  "",
		TargetDates:  []appattendance.UpdateTargetDateInput{},
	}

	_, err := usecase.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("Execute() should fail when FindTargetDatesByCollectionID fails")
	}
}

func TestUpdateCollectionUsecase_Execute_ErrorWhenCollectionNotFound(t *testing.T) {
	tenantID := common.NewTenantID()

	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return nil, common.NewNotFoundError("AttendanceCollection", cid.String())
		},
	}
	txManager := &MockTxManager{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appattendance.NewUpdateCollectionUsecase(repo, txManager, clock)

	input := appattendance.UpdateCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: common.NewCollectionID().String(),
		Title:        "Updated",
		Description:  "",
	}

	_, err := usecase.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("Execute() should fail when collection not found")
	}
}

func TestUpdateCollectionUsecase_Execute_ErrorWhenReplaceTargetDatesFails(t *testing.T) {
	tenantID := common.NewTenantID()
	collection := createTestCollection(t, tenantID)
	collectionID := collection.CollectionID()

	repo := &MockAttendanceCollectionRepository{
		findByIDFunc: func(ctx context.Context, tid common.TenantID, cid common.CollectionID) (*attendance.AttendanceCollection, error) {
			return collection, nil
		},
		saveFunc: func(ctx context.Context, c *attendance.AttendanceCollection) error {
			return nil
		},
		findTargetDatesByCollectionIDFunc: func(ctx context.Context, cid common.CollectionID) ([]*attendance.TargetDate, error) {
			return []*attendance.TargetDate{}, nil
		},
		replaceTargetDatesFunc: func(ctx context.Context, cid common.CollectionID, dates []*attendance.TargetDate) error {
			return errors.New("database error")
		},
	}
	txManager := &MockTxManager{}
	clock := &MockClock{nowFunc: func() time.Time { return time.Now() }}

	usecase := appattendance.NewUpdateCollectionUsecase(repo, txManager, clock)

	input := appattendance.UpdateCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: collectionID.String(),
		Title:        "Updated",
		Description:  "",
		TargetDates: []appattendance.UpdateTargetDateInput{
			{TargetDate: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)},
		},
	}

	_, err := usecase.Execute(context.Background(), input)
	if err == nil {
		t.Fatal("Execute() should fail when ReplaceTargetDates fails")
	}
}
