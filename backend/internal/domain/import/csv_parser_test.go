package importjob

import (
	"strings"
	"testing"
)

func TestCSVParser_ParseMembersCSV(t *testing.T) {
	parser := NewCSVParser()

	tests := []struct {
		name        string
		input       string
		wantRows    int
		wantErr     bool
		errContains string
	}{
		{
			name:     "正常系: 3件のメンバー",
			input:    "name,display_name,note\nラット,らっと,一期生\nもやし,,二期生\nおおちゃん,おおちゃん,",
			wantRows: 3,
			wantErr:  false,
		},
		{
			name:     "正常系: display_nameが空の場合はnameを使用",
			input:    "name,display_name,note\nテスト,,メモ",
			wantRows: 1,
			wantErr:  false,
		},
		{
			name:        "異常系: 必須カラム欠落",
			input:       "display_name,note\nらっと,一期生",
			wantErr:     true,
			errContains: "required column 'name' not found",
		},
		{
			name:     "正常系: 空のCSV（ヘッダーのみ）",
			input:    "name,display_name,note\n",
			wantRows: 0,
			wantErr:  false,
		},
		{
			name:     "正常系: カラム順序が異なる",
			input:    "note,name,display_name\n一期生,ラット,らっと",
			wantRows: 1,
			wantErr:  false,
		},
		{
			name:     "正常系: 大文字小文字混在のヘッダー",
			input:    "Name,Display_Name,NOTE\nラット,らっと,一期生",
			wantRows: 1,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			rows, err := parser.ParseMembersCSV(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(rows) != tt.wantRows {
				t.Errorf("got %d rows, want %d", len(rows), tt.wantRows)
			}
		})
	}
}

func TestCSVParser_ParseMembersCSV_RowContent(t *testing.T) {
	parser := NewCSVParser()
	input := "name,display_name,note\nラット,らっと,一期生・IL"

	reader := strings.NewReader(input)
	rows, err := parser.ParseMembersCSV(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	row := rows[0]
	if row.Name != "ラット" {
		t.Errorf("Name = %q, want %q", row.Name, "ラット")
	}
	if row.DisplayName != "らっと" {
		t.Errorf("DisplayName = %q, want %q", row.DisplayName, "らっと")
	}
	if row.Note != "一期生・IL" {
		t.Errorf("Note = %q, want %q", row.Note, "一期生・IL")
	}
	if row.RowNumber != 2 {
		t.Errorf("RowNumber = %d, want %d", row.RowNumber, 2)
	}
}

func TestCSVParser_ParseActualAttendanceCSV(t *testing.T) {
	parser := NewCSVParser()

	tests := []struct {
		name        string
		input       string
		wantRows    int
		wantErr     bool
		errContains string
	}{
		{
			name:     "正常系: 2件の出席データ",
			input:    "date,member_name,event_name,slot_name,start_time,end_time,note\n2024-12-01,ラット,通常営業,A卓,20:00,22:00,\n2024-12-01,もやし,通常営業,IL,20:00,22:00,遅刻",
			wantRows: 2,
			wantErr:  false,
		},
		{
			name:        "異常系: date欠落",
			input:       "member_name,event_name\nラット,通常営業",
			wantErr:     true,
			errContains: "required column 'date' not found",
		},
		{
			name:        "異常系: member_name欠落",
			input:       "date,event_name\n2024-12-01,通常営業",
			wantErr:     true,
			errContains: "required column 'member_name' not found",
		},
		{
			name:     "正常系: 最小限のカラム",
			input:    "date,member_name\n2024-12-01,ラット",
			wantRows: 1,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			rows, err := parser.ParseActualAttendanceCSV(reader)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(rows) != tt.wantRows {
				t.Errorf("got %d rows, want %d", len(rows), tt.wantRows)
			}
		})
	}
}

func TestMemberRow_Validate(t *testing.T) {
	tests := []struct {
		name    string
		row     MemberRow
		wantErr bool
	}{
		{
			name: "正常系: 有効な行",
			row: MemberRow{
				RowNumber:   1,
				Name:        "ラット",
				DisplayName: "らっと",
			},
			wantErr: false,
		},
		{
			name: "異常系: 名前が空",
			row: MemberRow{
				RowNumber:   1,
				Name:        "",
				DisplayName: "らっと",
			},
			wantErr: true,
		},
		{
			name: "異常系: 名前が空白のみ",
			row: MemberRow{
				RowNumber:   1,
				Name:        "   ",
				DisplayName: "らっと",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.row.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestActualAttendanceRow_Validate(t *testing.T) {
	tests := []struct {
		name    string
		row     ActualAttendanceRow
		wantErr bool
	}{
		{
			name: "正常系: 有効な行",
			row: ActualAttendanceRow{
				RowNumber:  1,
				Date:       "2024-12-01",
				MemberName: "ラット",
			},
			wantErr: false,
		},
		{
			name: "異常系: 日付が空",
			row: ActualAttendanceRow{
				RowNumber:  1,
				Date:       "",
				MemberName: "ラット",
			},
			wantErr: true,
		},
		{
			name: "異常系: メンバー名が空",
			row: ActualAttendanceRow{
				RowNumber:  1,
				Date:       "2024-12-01",
				MemberName: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.row.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCSVParser_CountRows(t *testing.T) {
	parser := NewCSVParser()

	tests := []struct {
		name      string
		input     string
		wantCount int
		wantErr   bool
	}{
		{
			name:      "正常系: 3行",
			input:     "name,note\nラット,一期生\nもやし,二期生\nおおちゃん,一期生",
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "正常系: 0行（ヘッダーのみ）",
			input:     "name,note\n",
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			count, err := parser.CountRows(reader)

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

			if count != tt.wantCount {
				t.Errorf("got count %d, want %d", count, tt.wantCount)
			}
		})
	}
}
