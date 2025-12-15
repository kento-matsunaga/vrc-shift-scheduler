package auth

import "time"

// LoginInput represents the input for the login use case
type LoginInput struct {
	// TenantID削除: email + password のみでログイン
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
