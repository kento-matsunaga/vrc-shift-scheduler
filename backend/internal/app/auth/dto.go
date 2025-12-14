package auth

import "time"

// LoginInput represents the input for the login use case
type LoginInput struct {
	TenantID string // ログイン時のみ Body で受け取る（認証前なのでJWTがない）
	Email    string
	Password string
}

// LoginOutput represents the output for the login use case
type LoginOutput struct {
	Token     string    `json:"token"`
	AdminID   string    `json:"admin_id"`
	TenantID  string    `json:"tenant_id"`
	Role      string    `json:"role"`
	ExpiresAt time.Time `json:"expires_at"`
}
