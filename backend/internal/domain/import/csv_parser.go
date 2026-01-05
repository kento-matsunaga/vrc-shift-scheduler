package importjob

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// MemberRow represents a parsed row from members CSV
type MemberRow struct {
	RowNumber   int
	Name        string
	DisplayName string
	Note        string
}

// Validate validates the member row
func (r MemberRow) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return common.NewValidationError(fmt.Sprintf("row %d: name is required", r.RowNumber), nil)
	}
	return nil
}

// ActualAttendanceRow represents a parsed row from actual attendance CSV
type ActualAttendanceRow struct {
	RowNumber  int
	Date       string
	MemberName string
	EventName  string
	SlotName   string
	StartTime  string
	EndTime    string
	Note       string
}

// Validate validates the actual attendance row
func (r ActualAttendanceRow) Validate() error {
	if strings.TrimSpace(r.Date) == "" {
		return common.NewValidationError(fmt.Sprintf("row %d: date is required", r.RowNumber), nil)
	}
	if strings.TrimSpace(r.MemberName) == "" {
		return common.NewValidationError(fmt.Sprintf("row %d: member_name is required", r.RowNumber), nil)
	}
	return nil
}

// CSVParser handles CSV parsing for import operations
type CSVParser struct{}

// NewCSVParser creates a new CSV parser
func NewCSVParser() *CSVParser {
	return &CSVParser{}
}

// sanitizeCSVValue removes potentially dangerous prefixes to prevent CSV injection
// These characters (=, +, -, @) can be interpreted as formulas in spreadsheet applications
func sanitizeCSVValue(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	// Check for dangerous prefixes
	firstChar := s[0]
	if firstChar == '=' || firstChar == '+' || firstChar == '-' || firstChar == '@' {
		// Prefix with a single quote to prevent formula execution
		return "'" + s
	}
	return s
}

// ParseMembersCSV parses a members CSV file
func (p *CSVParser) ParseMembersCSV(reader io.Reader) ([]MemberRow, error) {
	csvReader := csv.NewReader(reader)

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return nil, common.NewValidationError("failed to read CSV header", err)
	}

	// Validate required columns
	columnIndex := p.buildColumnIndex(header)
	if _, ok := columnIndex["name"]; !ok {
		return nil, common.NewValidationError("required column 'name' not found in CSV header", nil)
	}

	var rows []MemberRow
	rowNumber := 1 // Start from 1 (header is row 0)

	for {
		rowNumber++
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, common.NewValidationError(fmt.Sprintf("failed to read row %d", rowNumber), err)
		}

		row := MemberRow{
			RowNumber: rowNumber,
		}

		if idx, ok := columnIndex["name"]; ok && idx < len(record) {
			row.Name = sanitizeCSVValue(record[idx])
		}
		if idx, ok := columnIndex["display_name"]; ok && idx < len(record) {
			row.DisplayName = sanitizeCSVValue(record[idx])
		}
		if idx, ok := columnIndex["note"]; ok && idx < len(record) {
			row.Note = sanitizeCSVValue(record[idx])
		}

		// Use name as display_name if not provided
		if row.DisplayName == "" {
			row.DisplayName = row.Name
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// ParseActualAttendanceCSV parses an actual attendance CSV file
func (p *CSVParser) ParseActualAttendanceCSV(reader io.Reader) ([]ActualAttendanceRow, error) {
	csvReader := csv.NewReader(reader)

	// Read header
	header, err := csvReader.Read()
	if err != nil {
		return nil, common.NewValidationError("failed to read CSV header", err)
	}

	// Validate required columns
	columnIndex := p.buildColumnIndex(header)
	if _, ok := columnIndex["date"]; !ok {
		return nil, common.NewValidationError("required column 'date' not found in CSV header", nil)
	}
	if _, ok := columnIndex["member_name"]; !ok {
		return nil, common.NewValidationError("required column 'member_name' not found in CSV header", nil)
	}

	var rows []ActualAttendanceRow
	rowNumber := 1 // Start from 1 (header is row 0)

	for {
		rowNumber++
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, common.NewValidationError(fmt.Sprintf("failed to read row %d", rowNumber), err)
		}

		row := ActualAttendanceRow{
			RowNumber: rowNumber,
		}

		if idx, ok := columnIndex["date"]; ok && idx < len(record) {
			row.Date = sanitizeCSVValue(record[idx])
		}
		if idx, ok := columnIndex["member_name"]; ok && idx < len(record) {
			row.MemberName = sanitizeCSVValue(record[idx])
		}
		if idx, ok := columnIndex["event_name"]; ok && idx < len(record) {
			row.EventName = sanitizeCSVValue(record[idx])
		}
		if idx, ok := columnIndex["slot_name"]; ok && idx < len(record) {
			row.SlotName = sanitizeCSVValue(record[idx])
		}
		if idx, ok := columnIndex["start_time"]; ok && idx < len(record) {
			row.StartTime = sanitizeCSVValue(record[idx])
		}
		if idx, ok := columnIndex["end_time"]; ok && idx < len(record) {
			row.EndTime = sanitizeCSVValue(record[idx])
		}
		if idx, ok := columnIndex["note"]; ok && idx < len(record) {
			row.Note = sanitizeCSVValue(record[idx])
		}

		rows = append(rows, row)
	}

	return rows, nil
}

// buildColumnIndex creates a map of column names to their indices
func (p *CSVParser) buildColumnIndex(header []string) map[string]int {
	index := make(map[string]int)
	for i, col := range header {
		// Normalize column name (lowercase, trim whitespace)
		normalized := strings.ToLower(strings.TrimSpace(col))
		index[normalized] = i
	}
	return index
}

// CountRows counts the number of data rows in a CSV (excluding header)
func (p *CSVParser) CountRows(reader io.Reader) (int, error) {
	csvReader := csv.NewReader(reader)

	// Skip header
	_, err := csvReader.Read()
	if err != nil {
		return 0, common.NewValidationError("failed to read CSV header", err)
	}

	count := 0
	for {
		_, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, common.NewValidationError("failed to count CSV rows", err)
		}
		count++
	}

	return count, nil
}
