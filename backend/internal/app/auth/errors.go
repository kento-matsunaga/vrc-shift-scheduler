package auth

import "errors"

var (
	// ErrInvalidCredentials is returned when the credentials are invalid
	// メールアドレスが存在しない / パスワードが間違っている を区別しない
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrAccountDisabled is returned when the account is disabled
	ErrAccountDisabled = errors.New("account is disabled")
)
