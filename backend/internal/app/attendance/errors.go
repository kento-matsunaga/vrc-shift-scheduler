package attendance

import "errors"

var (
	// ErrCollectionNotFound is returned when a collection is not found (token error)
	// トークンエラー → 404 を返す（詳細は返さない）
	ErrCollectionNotFound = errors.New("collection not found")

	// ErrMemberNotAllowed is returned when a member is not allowed to respond
	// メンバーエラー → 400 を返す（詳細は返さない）
	ErrMemberNotAllowed = errors.New("member not allowed")

	// ErrMemberNotFound is returned when a member is not found in the tenant
	// メンバーが見つからない → 404 を返す
	ErrMemberNotFound = errors.New("member not found")
)
