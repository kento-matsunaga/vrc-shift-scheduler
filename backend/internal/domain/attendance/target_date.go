package attendance

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// TargetDate は出欠確認の対象日エンティティ
type TargetDate struct {
	targetDateID  common.TargetDateID
	collectionID  common.CollectionID
	targetDate    time.Time // 日付部分のみ使用
	startTime     *string   // 開始時間（HH:MM形式、任意）
	endTime       *string   // 終了時間（HH:MM形式、任意）
	displayOrder  int
	createdAt     time.Time
}

// NewTargetDate creates a new TargetDate entity
func NewTargetDate(
	now time.Time,
	collectionID common.CollectionID,
	targetDate time.Time,
	startTime *string,
	endTime *string,
	displayOrder int,
) (*TargetDate, error) {
	td := &TargetDate{
		targetDateID: common.NewTargetDateID(),
		collectionID: collectionID,
		targetDate:   targetDate,
		startTime:    startTime,
		endTime:      endTime,
		displayOrder: displayOrder,
		createdAt:    now,
	}

	if err := td.validate(); err != nil {
		return nil, err
	}

	return td, nil
}

// ReconstructTargetDate reconstructs a TargetDate entity from persistence
func ReconstructTargetDate(
	targetDateID common.TargetDateID,
	collectionID common.CollectionID,
	targetDate time.Time,
	startTime *string,
	endTime *string,
	displayOrder int,
	createdAt time.Time,
) (*TargetDate, error) {
	td := &TargetDate{
		targetDateID: targetDateID,
		collectionID: collectionID,
		targetDate:   targetDate,
		startTime:    startTime,
		endTime:      endTime,
		displayOrder: displayOrder,
		createdAt:    createdAt,
	}

	if err := td.validate(); err != nil {
		return nil, err
	}

	return td, nil
}

func (td *TargetDate) validate() error {
	// CollectionID の必須性チェック
	if err := td.collectionID.Validate(); err != nil {
		return common.NewValidationError("collection_id is required", err)
	}

	// 開始時間のフォーマットチェック (HH:MM形式)
	if td.startTime != nil {
		if !isValidTimeFormat(*td.startTime) {
			return common.NewValidationError("start_time must be in HH:MM format (00:00-23:59)", nil)
		}
	}

	// 終了時間のフォーマットチェック (HH:MM形式)
	if td.endTime != nil {
		if !isValidTimeFormat(*td.endTime) {
			return common.NewValidationError("end_time must be in HH:MM format (00:00-23:59)", nil)
		}
	}

	// 開始時間と終了時間の論理チェック
	if td.startTime != nil && td.endTime != nil {
		if *td.startTime >= *td.endTime {
			return common.NewValidationError("start_time must be before end_time", nil)
		}
	}

	return nil
}

// isValidTimeFormat checks if time string is in HH:MM format (00:00-23:59)
func isValidTimeFormat(timeStr string) bool {
	return timeFormatRegex.MatchString(timeStr)
}

// Getters

func (td *TargetDate) TargetDateID() common.TargetDateID {
	return td.targetDateID
}

func (td *TargetDate) CollectionID() common.CollectionID {
	return td.collectionID
}

func (td *TargetDate) TargetDateValue() time.Time {
	return td.targetDate
}

func (td *TargetDate) StartTime() *string {
	return td.startTime
}

func (td *TargetDate) EndTime() *string {
	return td.endTime
}

func (td *TargetDate) DisplayOrder() int {
	return td.displayOrder
}

func (td *TargetDate) CreatedAt() time.Time {
	return td.createdAt
}
