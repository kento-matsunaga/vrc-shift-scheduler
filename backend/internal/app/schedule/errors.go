package schedule

import "errors"

var (
	// ErrScheduleNotFound is returned when a schedule is not found (token error)
	// トークンエラー → 404 を返す（詳細は返さない）
	ErrScheduleNotFound = errors.New("schedule not found")

	// ErrMemberNotAllowed is returned when a member is not allowed to respond
	// メンバーエラー → 400 を返す（詳細は返さない）
	ErrMemberNotAllowed = errors.New("member not allowed")
)
