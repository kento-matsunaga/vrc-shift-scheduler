package importapp

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/event"
	importjob "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/import"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/shift"
)

// EventBusinessDayRepository defines the interface for business day persistence
type EventBusinessDayRepository interface {
	FindByTenantIDAndDate(ctx context.Context, tenantID common.TenantID, date time.Time) ([]*event.EventBusinessDay, error)
	FindByTenantIDAndDateRange(ctx context.Context, tenantID common.TenantID, startDate, endDate time.Time) ([]*event.EventBusinessDay, error)
	Save(ctx context.Context, businessDay *event.EventBusinessDay) error
}

// EventRepository defines the interface for event persistence
type EventRepository interface {
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*event.Event, error)
	FindByID(ctx context.Context, tenantID common.TenantID, eventID common.EventID) (*event.Event, error)
}

// ShiftSlotRepository defines the interface for shift slot persistence
type ShiftSlotRepository interface {
	FindByBusinessDayIDAndSlotName(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID, slotName string) (*shift.ShiftSlot, error)
	FindByBusinessDayID(ctx context.Context, tenantID common.TenantID, businessDayID event.BusinessDayID) ([]*shift.ShiftSlot, error)
	Save(ctx context.Context, slot *shift.ShiftSlot) error
}

// ShiftAssignmentRepository defines the interface for shift assignment persistence
type ShiftAssignmentRepository interface {
	ExistsBySlotIDAndMemberID(ctx context.Context, tenantID common.TenantID, slotID shift.SlotID, memberID common.MemberID) (bool, error)
	Save(ctx context.Context, assignment *shift.ShiftAssignment) error
}

// PositionRepository defines the interface for position persistence
type PositionRepository interface {
	FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*shift.Position, error)
}

// ImportActualAttendanceInput represents the input for importing actual attendance
type ImportActualAttendanceInput struct {
	TenantID common.TenantID
	AdminID  common.AdminID
	FileName string
	FileData []byte
	Options  importjob.ImportOptions
}

// ImportActualAttendanceOutput represents the output of actual attendance import
type ImportActualAttendanceOutput struct {
	ImportJobID  common.ImportJobID
	Status       importjob.ImportStatus
	TotalRows    int
	SuccessCount int
	ErrorCount   int
	SkippedCount int
	Errors       []importjob.ErrorDetail
}

// ImportActualAttendanceUsecase handles the actual attendance import use case
type ImportActualAttendanceUsecase struct {
	importJobRepo      importjob.ImportJobRepository
	memberRepo         MemberRepository
	businessDayRepo    EventBusinessDayRepository
	eventRepo          EventRepository
	shiftSlotRepo      ShiftSlotRepository
	shiftAssignmentRepo ShiftAssignmentRepository
	positionRepo       PositionRepository
	csvParser          *importjob.CSVParser
}

// NewImportActualAttendanceUsecase creates a new ImportActualAttendanceUsecase
func NewImportActualAttendanceUsecase(
	importJobRepo importjob.ImportJobRepository,
	memberRepo MemberRepository,
	businessDayRepo EventBusinessDayRepository,
	eventRepo EventRepository,
	shiftSlotRepo ShiftSlotRepository,
	shiftAssignmentRepo ShiftAssignmentRepository,
	positionRepo PositionRepository,
) *ImportActualAttendanceUsecase {
	return &ImportActualAttendanceUsecase{
		importJobRepo:      importJobRepo,
		memberRepo:         memberRepo,
		businessDayRepo:    businessDayRepo,
		eventRepo:          eventRepo,
		shiftSlotRepo:      shiftSlotRepo,
		shiftAssignmentRepo: shiftAssignmentRepo,
		positionRepo:       positionRepo,
		csvParser:          importjob.NewCSVParser(),
	}
}

// Execute imports actual attendance from CSV
func (uc *ImportActualAttendanceUsecase) Execute(ctx context.Context, input ImportActualAttendanceInput) (*ImportActualAttendanceOutput, error) {
	now := time.Now()

	// Create import job
	job, err := importjob.NewImportJob(
		now,
		input.TenantID,
		importjob.ImportTypeActualAttendance,
		input.FileName,
		input.Options,
		input.AdminID,
	)
	if err != nil {
		return nil, err
	}

	// Save initial job
	if err := uc.importJobRepo.Save(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to save import job: %w", err)
	}

	// Parse CSV
	reader := bytes.NewReader(input.FileData)
	rows, err := uc.csvParser.ParseActualAttendanceCSV(reader)
	if err != nil {
		_ = job.Fail(time.Now(), fmt.Sprintf("CSVパースエラー: %v", err))
		_ = uc.importJobRepo.Update(ctx, job)
		return &ImportActualAttendanceOutput{
			ImportJobID:  job.ImportJobID(),
			Status:       job.Status(),
			TotalRows:    0,
			SuccessCount: 0,
			ErrorCount:   1,
			SkippedCount: 0,
			Errors:       job.ErrorDetails(),
		}, nil
	}

	// Check row limit
	if len(rows) > importjob.MaxImportRows {
		_ = job.Fail(time.Now(), fmt.Sprintf("行数が上限を超えています: %d行 (上限: %d行)", len(rows), importjob.MaxImportRows))
		_ = uc.importJobRepo.Update(ctx, job)
		return &ImportActualAttendanceOutput{
			ImportJobID:  job.ImportJobID(),
			Status:       job.Status(),
			TotalRows:    len(rows),
			SuccessCount: 0,
			ErrorCount:   1,
			SkippedCount: 0,
			Errors:       job.ErrorDetails(),
		}, nil
	}

	// Start processing
	if err := job.Start(time.Now(), len(rows)); err != nil {
		return nil, err
	}
	if err := uc.importJobRepo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to update import job: %w", err)
	}

	// Collect data for bulk fetch
	dateRange := uc.collectDateRange(rows)
	if dateRange.minDate.IsZero() {
		_ = job.Fail(time.Now(), "有効な日付が見つかりません")
		_ = uc.importJobRepo.Update(ctx, job)
		return &ImportActualAttendanceOutput{
			ImportJobID:  job.ImportJobID(),
			Status:       job.Status(),
			TotalRows:    len(rows),
			SuccessCount: 0,
			ErrorCount:   1,
			SkippedCount: 0,
			Errors:       job.ErrorDetails(),
		}, nil
	}

	// Bulk fetch data (N+1 prevention)
	members, err := uc.memberRepo.FindByTenantID(ctx, input.TenantID)
	if err != nil {
		_ = job.Fail(time.Now(), fmt.Sprintf("メンバー取得エラー: %v", err))
		_ = uc.importJobRepo.Update(ctx, job)
		return &ImportActualAttendanceOutput{
			ImportJobID:  job.ImportJobID(),
			Status:       job.Status(),
			TotalRows:    len(rows),
			SuccessCount: 0,
			ErrorCount:   1,
			SkippedCount: 0,
			Errors:       job.ErrorDetails(),
		}, nil
	}

	events, err := uc.eventRepo.FindByTenantID(ctx, input.TenantID)
	if err != nil {
		_ = job.Fail(time.Now(), fmt.Sprintf("イベント取得エラー: %v", err))
		_ = uc.importJobRepo.Update(ctx, job)
		return &ImportActualAttendanceOutput{
			ImportJobID:  job.ImportJobID(),
			Status:       job.Status(),
			TotalRows:    len(rows),
			SuccessCount: 0,
			ErrorCount:   1,
			SkippedCount: 0,
			Errors:       job.ErrorDetails(),
		}, nil
	}

	businessDays, err := uc.businessDayRepo.FindByTenantIDAndDateRange(ctx, input.TenantID, dateRange.minDate, dateRange.maxDate)
	if err != nil {
		_ = job.Fail(time.Now(), fmt.Sprintf("営業日取得エラー: %v", err))
		_ = uc.importJobRepo.Update(ctx, job)
		return &ImportActualAttendanceOutput{
			ImportJobID:  job.ImportJobID(),
			Status:       job.Status(),
			TotalRows:    len(rows),
			SuccessCount: 0,
			ErrorCount:   1,
			SkippedCount: 0,
			Errors:       job.ErrorDetails(),
		}, nil
	}

	positions, err := uc.positionRepo.FindByTenantID(ctx, input.TenantID)
	if err != nil {
		_ = job.Fail(time.Now(), fmt.Sprintf("ポジション取得エラー: %v", err))
		_ = uc.importJobRepo.Update(ctx, job)
		return &ImportActualAttendanceOutput{
			ImportJobID:  job.ImportJobID(),
			Status:       job.Status(),
			TotalRows:    len(rows),
			SuccessCount: 0,
			ErrorCount:   1,
			SkippedCount: 0,
			Errors:       job.ErrorDetails(),
		}, nil
	}

	// Build lookup maps
	memberMatcher := importjob.NewMemberMatcher(members, input.Options.FuzzyMemberMatch)
	eventByName := uc.buildEventByNameMap(events)
	businessDayMap := uc.buildBusinessDayMap(businessDays, events)
	positionByName := uc.buildPositionByNameMap(positions)

	// Process each row
	for _, row := range rows {
		uc.processRow(ctx, job, row, input, memberMatcher, eventByName, businessDayMap, positionByName)
	}

	// Complete job
	if err := job.Complete(time.Now()); err != nil {
		return nil, fmt.Errorf("failed to complete import job: %w", err)
	}
	if err := uc.importJobRepo.Update(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to update import job: %w", err)
	}

	skippedCount := job.ProcessedRows() - job.SuccessCount() - job.ErrorCount()
	if skippedCount < 0 {
		skippedCount = 0
	}

	return &ImportActualAttendanceOutput{
		ImportJobID:  job.ImportJobID(),
		Status:       job.Status(),
		TotalRows:    job.TotalRows(),
		SuccessCount: job.SuccessCount(),
		ErrorCount:   job.ErrorCount(),
		SkippedCount: skippedCount,
		Errors:       job.ErrorDetails(),
	}, nil
}

type dateRange struct {
	minDate time.Time
	maxDate time.Time
}

func (uc *ImportActualAttendanceUsecase) collectDateRange(rows []importjob.ActualAttendanceRow) dateRange {
	var minDate, maxDate time.Time

	for _, row := range rows {
		date, err := time.Parse("2006-01-02", row.Date)
		if err != nil {
			continue
		}

		if minDate.IsZero() || date.Before(minDate) {
			minDate = date
		}
		if maxDate.IsZero() || date.After(maxDate) {
			maxDate = date
		}
	}

	return dateRange{minDate: minDate, maxDate: maxDate}
}

func (uc *ImportActualAttendanceUsecase) buildEventByNameMap(events []*event.Event) map[string]*event.Event {
	m := make(map[string]*event.Event)
	for _, e := range events {
		m[e.EventName()] = e
	}
	return m
}

// businessDayKey is "date:event_id"
func (uc *ImportActualAttendanceUsecase) buildBusinessDayMap(businessDays []*event.EventBusinessDay, events []*event.Event) map[string]*event.EventBusinessDay {
	eventByID := make(map[string]*event.Event)
	for _, e := range events {
		eventByID[e.EventID().String()] = e
	}

	m := make(map[string]*event.EventBusinessDay)
	for _, bd := range businessDays {
		dateStr := bd.TargetDate().Format("2006-01-02")
		key := fmt.Sprintf("%s:%s", dateStr, bd.EventID().String())
		m[key] = bd
	}
	return m
}

func (uc *ImportActualAttendanceUsecase) buildPositionByNameMap(positions []*shift.Position) map[string]*shift.Position {
	m := make(map[string]*shift.Position)
	for _, p := range positions {
		m[p.PositionName()] = p
	}
	return m
}

func (uc *ImportActualAttendanceUsecase) processRow(
	ctx context.Context,
	job *importjob.ImportJob,
	row importjob.ActualAttendanceRow,
	input ImportActualAttendanceInput,
	memberMatcher *importjob.MemberMatcher,
	eventByName map[string]*event.Event,
	businessDayMap map[string]*event.EventBusinessDay,
	positionByName map[string]*shift.Position,
) {
	// Validate row
	if err := row.Validate(); err != nil {
		job.RecordError(row.RowNumber, err.Error())
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", row.Date)
	if err != nil {
		job.RecordError(row.RowNumber, fmt.Sprintf("日付形式が不正です: %s (YYYY-MM-DD形式で入力)", row.Date))
		return
	}

	// Find member
	matchedMember, _ := memberMatcher.Match(row.MemberName)
	if matchedMember == nil {
		job.RecordError(row.RowNumber, fmt.Sprintf("メンバー '%s' が見つかりません", row.MemberName))
		return
	}

	// Find event and business day
	businessDay, err := uc.findBusinessDay(ctx, row, date, input, eventByName, businessDayMap)
	if err != nil {
		job.RecordError(row.RowNumber, err.Error())
		return
	}
	if businessDay == nil {
		job.RecordError(row.RowNumber, fmt.Sprintf("日付 '%s' の営業日が見つかりません", row.Date))
		return
	}

	// Find or create shift slot
	slotName := row.SlotName
	if slotName == "" {
		slotName = "通常シフト"
	}

	slot, err := uc.shiftSlotRepo.FindByBusinessDayIDAndSlotName(ctx, input.TenantID, businessDay.BusinessDayID(), slotName)
	if err != nil {
		job.RecordError(row.RowNumber, fmt.Sprintf("シフト枠検索エラー: %v", err))
		return
	}

	if slot == nil {
		if !input.Options.CreateMissingSlots {
			job.RecordError(row.RowNumber, fmt.Sprintf("シフト枠 '%s' が見つかりません", slotName))
			return
		}

		// Need position to create slot
		if row.PositionName == "" {
			job.RecordError(row.RowNumber, "シフト枠作成には position_name が必要です")
			return
		}

		position := positionByName[row.PositionName]
		if position == nil {
			job.RecordError(row.RowNumber, fmt.Sprintf("ポジション '%s' が見つかりません", row.PositionName))
			return
		}

		// Parse times
		startTime, endTime, err := uc.parseSlotTimes(row)
		if err != nil {
			job.RecordError(row.RowNumber, err.Error())
			return
		}

		// Create new slot
		newSlot, err := shift.NewShiftSlot(
			time.Now(),
			input.TenantID,
			businessDay.BusinessDayID(),
			position.PositionID(),
			slotName,
			"",
			startTime,
			endTime,
			1,
			0,
		)
		if err != nil {
			job.RecordError(row.RowNumber, fmt.Sprintf("シフト枠作成エラー: %v", err))
			return
		}

		if err := uc.shiftSlotRepo.Save(ctx, newSlot); err != nil {
			job.RecordError(row.RowNumber, fmt.Sprintf("シフト枠保存エラー: %v", err))
			return
		}

		slot = newSlot
	}

	// Check for duplicate assignment
	exists, err := uc.shiftAssignmentRepo.ExistsBySlotIDAndMemberID(ctx, input.TenantID, slot.SlotID(), matchedMember.MemberID())
	if err != nil {
		job.RecordError(row.RowNumber, fmt.Sprintf("重複チェックエラー: %v", err))
		return
	}

	if exists {
		if input.Options.SkipExisting {
			job.RecordSkip()
			return
		}
		job.RecordError(row.RowNumber, fmt.Sprintf("'%s' は既に %s に割り当て済みです", row.MemberName, row.Date))
		return
	}

	// Create shift assignment
	assignment, err := shift.NewShiftAssignment(
		time.Now(),
		input.TenantID,
		"", // plan_id is NULL for manual assignments
		slot.SlotID(),
		matchedMember.MemberID(),
		shift.AssignmentMethodManual,
		false,
	)
	if err != nil {
		job.RecordError(row.RowNumber, fmt.Sprintf("シフト割り当て作成エラー: %v", err))
		return
	}

	if err := uc.shiftAssignmentRepo.Save(ctx, assignment); err != nil {
		job.RecordError(row.RowNumber, fmt.Sprintf("シフト割り当て保存エラー: %v", err))
		return
	}

	job.RecordSuccess()
}

func (uc *ImportActualAttendanceUsecase) findBusinessDay(
	ctx context.Context,
	row importjob.ActualAttendanceRow,
	date time.Time,
	input ImportActualAttendanceInput,
	eventByName map[string]*event.Event,
	businessDayMap map[string]*event.EventBusinessDay,
) (*event.EventBusinessDay, error) {
	dateStr := date.Format("2006-01-02")

	// If event_name is specified, look up specific event
	if row.EventName != "" {
		evt := eventByName[row.EventName]
		if evt == nil {
			return nil, fmt.Errorf("イベント '%s' が見つかりません", row.EventName)
		}
		key := fmt.Sprintf("%s:%s", dateStr, evt.EventID().String())
		bd := businessDayMap[key]
		if bd == nil {
			return nil, fmt.Errorf("日付 '%s' にイベント '%s' の営業日がありません", dateStr, row.EventName)
		}
		return bd, nil
	}

	// If default_event_id is specified
	if input.Options.DefaultEventID != "" {
		key := fmt.Sprintf("%s:%s", dateStr, input.Options.DefaultEventID)
		bd := businessDayMap[key]
		if bd != nil {
			return bd, nil
		}
	}

	// Find all business days on this date
	var matchingDays []*event.EventBusinessDay
	for k, bd := range businessDayMap {
		if len(k) >= 10 && k[:10] == dateStr {
			matchingDays = append(matchingDays, bd)
		}
	}

	if len(matchingDays) == 0 {
		return nil, nil
	}

	if len(matchingDays) == 1 {
		return matchingDays[0], nil
	}

	// Multiple events on same date - need event_name
	return nil, fmt.Errorf("日付 '%s' に複数のイベントがあります。event_name を指定してください", dateStr)
}

func (uc *ImportActualAttendanceUsecase) parseSlotTimes(row importjob.ActualAttendanceRow) (time.Time, time.Time, error) {
	if row.StartTime == "" || row.EndTime == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("シフト枠作成には start_time, end_time が必要です")
	}

	// JSTタイムゾーンで時刻をパース
	startTime, err := time.ParseInLocation("15:04", row.StartTime, importjob.DefaultTimezone)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("開始時刻の形式が不正です: %s (HH:MM形式で入力)", row.StartTime)
	}

	endTime, err := time.ParseInLocation("15:04", row.EndTime, importjob.DefaultTimezone)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("終了時刻の形式が不正です: %s (HH:MM形式で入力)", row.EndTime)
	}

	return startTime, endTime, nil
}
