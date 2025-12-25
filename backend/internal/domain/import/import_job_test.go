package importjob

import (
	"testing"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

func TestNewImportJob(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantIDWithTime(now)
	adminID := common.NewAdminIDWithTime(now)

	tests := []struct {
		name       string
		importType ImportType
		fileName   string
		options    ImportOptions
		wantErr    bool
	}{
		{
			name:       "正常系: メンバーインポート",
			importType: ImportTypeMembers,
			fileName:   "members.csv",
			options:    ImportOptions{SkipExisting: true},
			wantErr:    false,
		},
		{
			name:       "正常系: 出席データインポート",
			importType: ImportTypeActualAttendance,
			fileName:   "attendance.csv",
			options:    ImportOptions{CreateMissingEvents: true},
			wantErr:    false,
		},
		{
			name:       "異常系: 無効なインポートタイプ",
			importType: ImportType("invalid"),
			fileName:   "test.csv",
			options:    ImportOptions{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job, err := NewImportJob(now, tenantID, tt.importType, tt.fileName, tt.options, adminID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if job.Status() != ImportStatusPending {
				t.Errorf("initial status = %v, want %v", job.Status(), ImportStatusPending)
			}

			if job.ImportType() != tt.importType {
				t.Errorf("import_type = %v, want %v", job.ImportType(), tt.importType)
			}

			if job.FileName() != tt.fileName {
				t.Errorf("file_name = %v, want %v", job.FileName(), tt.fileName)
			}
		})
	}
}

func TestImportJob_Start(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantIDWithTime(now)
	adminID := common.NewAdminIDWithTime(now)

	job, _ := NewImportJob(now, tenantID, ImportTypeMembers, "test.csv", ImportOptions{}, adminID)

	// Start the job
	err := job.Start(now.Add(time.Second), 100)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	if job.Status() != ImportStatusProcessing {
		t.Errorf("status after Start() = %v, want %v", job.Status(), ImportStatusProcessing)
	}

	if job.TotalRows() != 100 {
		t.Errorf("total_rows = %d, want 100", job.TotalRows())
	}

	if job.StartedAt() == nil {
		t.Error("started_at should be set after Start()")
	}

	// Try to start again (should fail)
	err = job.Start(now.Add(2*time.Second), 50)
	if err == nil {
		t.Error("expected error when starting already started job")
	}
}

func TestImportJob_RecordSuccess(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantIDWithTime(now)
	adminID := common.NewAdminIDWithTime(now)

	job, _ := NewImportJob(now, tenantID, ImportTypeMembers, "test.csv", ImportOptions{}, adminID)
	_ = job.Start(now, 10)

	job.RecordSuccess()
	job.RecordSuccess()
	job.RecordSuccess()

	if job.ProcessedRows() != 3 {
		t.Errorf("processed_rows = %d, want 3", job.ProcessedRows())
	}

	if job.SuccessCount() != 3 {
		t.Errorf("success_count = %d, want 3", job.SuccessCount())
	}

	if job.ErrorCount() != 0 {
		t.Errorf("error_count = %d, want 0", job.ErrorCount())
	}
}

func TestImportJob_RecordError(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantIDWithTime(now)
	adminID := common.NewAdminIDWithTime(now)

	job, _ := NewImportJob(now, tenantID, ImportTypeMembers, "test.csv", ImportOptions{}, adminID)
	_ = job.Start(now, 10)

	job.RecordError(5, "テストエラー1")
	job.RecordError(8, "テストエラー2")

	if job.ProcessedRows() != 2 {
		t.Errorf("processed_rows = %d, want 2", job.ProcessedRows())
	}

	if job.ErrorCount() != 2 {
		t.Errorf("error_count = %d, want 2", job.ErrorCount())
	}

	errors := job.ErrorDetails()
	if len(errors) != 2 {
		t.Fatalf("error_details length = %d, want 2", len(errors))
	}

	if errors[0].Row != 5 || errors[0].Message != "テストエラー1" {
		t.Errorf("first error = %+v, want row=5, message=テストエラー1", errors[0])
	}
}

func TestImportJob_Progress(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantIDWithTime(now)
	adminID := common.NewAdminIDWithTime(now)

	job, _ := NewImportJob(now, tenantID, ImportTypeMembers, "test.csv", ImportOptions{}, adminID)

	// Before start
	if job.Progress() != 0 {
		t.Errorf("progress before start = %f, want 0", job.Progress())
	}

	_ = job.Start(now, 100)

	// 0%
	if job.Progress() != 0 {
		t.Errorf("progress at 0/100 = %f, want 0", job.Progress())
	}

	// 50%
	for i := 0; i < 50; i++ {
		job.RecordSuccess()
	}
	if job.Progress() != 50 {
		t.Errorf("progress at 50/100 = %f, want 50", job.Progress())
	}

	// 100%
	for i := 0; i < 50; i++ {
		job.RecordSuccess()
	}
	if job.Progress() != 100 {
		t.Errorf("progress at 100/100 = %f, want 100", job.Progress())
	}
}

func TestImportJob_Complete(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantIDWithTime(now)
	adminID := common.NewAdminIDWithTime(now)

	job, _ := NewImportJob(now, tenantID, ImportTypeMembers, "test.csv", ImportOptions{}, adminID)
	_ = job.Start(now, 10)

	for i := 0; i < 10; i++ {
		job.RecordSuccess()
	}

	err := job.Complete(now.Add(time.Minute))
	if err != nil {
		t.Errorf("Complete() error = %v", err)
	}

	if job.Status() != ImportStatusCompleted {
		t.Errorf("status after Complete() = %v, want %v", job.Status(), ImportStatusCompleted)
	}

	if job.CompletedAt() == nil {
		t.Error("completed_at should be set after Complete()")
	}
}

func TestImportJob_Fail(t *testing.T) {
	now := time.Now()
	tenantID := common.NewTenantIDWithTime(now)
	adminID := common.NewAdminIDWithTime(now)

	job, _ := NewImportJob(now, tenantID, ImportTypeMembers, "test.csv", ImportOptions{}, adminID)
	_ = job.Start(now, 10)

	err := job.Fail(now.Add(time.Minute), "致命的なエラー")
	if err != nil {
		t.Errorf("Fail() error = %v", err)
	}

	if job.Status() != ImportStatusFailed {
		t.Errorf("status after Fail() = %v, want %v", job.Status(), ImportStatusFailed)
	}

	errors := job.ErrorDetails()
	if len(errors) == 0 {
		t.Error("error_details should contain the failure reason")
	}
}

func TestImportType_IsValid(t *testing.T) {
	tests := []struct {
		importType ImportType
		want       bool
	}{
		{ImportTypeMembers, true},
		{ImportTypeActualAttendance, true},
		{ImportType("invalid"), false},
		{ImportType(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.importType), func(t *testing.T) {
			got := tt.importType.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImportStatus_IsValid(t *testing.T) {
	tests := []struct {
		status ImportStatus
		want   bool
	}{
		{ImportStatusPending, true},
		{ImportStatusProcessing, true},
		{ImportStatusCompleted, true},
		{ImportStatusFailed, true},
		{ImportStatus("invalid"), false},
		{ImportStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			got := tt.status.IsValid()
			if got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImportOptions_ToJSON(t *testing.T) {
	opts := ImportOptions{
		SkipExisting:        true,
		UpdateExisting:      false,
		DefaultRoleIDs:      []string{"role1", "role2"},
		CreateMissingEvents: true,
	}

	json, err := opts.ToJSON()
	if err != nil {
		t.Errorf("ToJSON() error = %v", err)
	}

	if len(json) == 0 {
		t.Error("ToJSON() returned empty bytes")
	}

	// Parse back
	parsed, err := ParseImportOptions(json)
	if err != nil {
		t.Errorf("ParseImportOptions() error = %v", err)
	}

	if parsed.SkipExisting != opts.SkipExisting {
		t.Errorf("SkipExisting = %v, want %v", parsed.SkipExisting, opts.SkipExisting)
	}

	if len(parsed.DefaultRoleIDs) != len(opts.DefaultRoleIDs) {
		t.Errorf("DefaultRoleIDs length = %d, want %d", len(parsed.DefaultRoleIDs), len(opts.DefaultRoleIDs))
	}
}
