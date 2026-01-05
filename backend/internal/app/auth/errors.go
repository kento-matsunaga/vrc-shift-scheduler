package auth

import "errors"

var (
	// ErrInvalidCredentials is returned when the credentials are invalid
	// メールアドレスが存在しない / パスワードが間違っている を区別しない
	ErrInvalidCredentials = errors.New("invalid email or password")

	// ErrAccountDisabled is returned when the account is disabled
	ErrAccountDisabled = errors.New("account is disabled")

	// ErrInvalidInvitation is returned when the invitation is invalid or expired
	ErrInvalidInvitation = errors.New("invitation is invalid or expired")

	// ErrEmailAlreadyExists is returned when the email already exists
	ErrEmailAlreadyExists = errors.New("email already exists")

	// ErrPasswordResetNotAllowed is returned when password reset is not allowed
	// PWリセットが許可されていない または 24時間経過している
	ErrPasswordResetNotAllowed = errors.New("password reset is not allowed or expired")

	// ErrInvalidLicenseKey is returned when the license key verification fails
	// ライセンスキーが不正 または テナントと一致しない
	ErrInvalidLicenseKey = errors.New("invalid license key for this tenant")

	// ErrAdminNotFound is returned when admin is not found
	ErrAdminNotFound = errors.New("admin not found")

	// ErrUnauthorized is returned when the caller lacks permission
	ErrUnauthorized = errors.New("unauthorized operation")
)
