package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PasswordResetTokenRepository implements auth.PasswordResetTokenRepository for PostgreSQL
type PasswordResetTokenRepository struct {
	db *pgxpool.Pool
}

// NewPasswordResetTokenRepository creates a new PasswordResetTokenRepository
func NewPasswordResetTokenRepository(db *pgxpool.Pool) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{db: db}
}

// Save saves a password reset token (insert or update)
func (r *PasswordResetTokenRepository) Save(ctx context.Context, prt *auth.PasswordResetToken) error {
	query := `
		INSERT INTO password_reset_tokens (
			token_id, admin_id, token, expires_at, used_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (token_id) DO UPDATE SET
			used_at = EXCLUDED.used_at
	`

	_, err := r.db.Exec(ctx, query,
		prt.TokenID().String(),
		prt.AdminID().String(),
		prt.Token(),
		prt.ExpiresAt(),
		prt.UsedAt(),
		prt.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save password reset token: %w", err)
	}

	return nil
}

// FindByToken finds a password reset token by token string
// Returns only unused tokens
func (r *PasswordResetTokenRepository) FindByToken(ctx context.Context, token string) (*auth.PasswordResetToken, error) {
	query := `
		SELECT
			token_id, admin_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token = $1 AND used_at IS NULL
		LIMIT 1
	`

	var (
		tokenIDStr string
		adminIDStr string
		tokenStr   string
		expiresAt  time.Time
		usedAt     sql.NullTime
		createdAt  time.Time
	)

	err := r.db.QueryRow(ctx, query, token).Scan(
		&tokenIDStr,
		&adminIDStr,
		&tokenStr,
		&expiresAt,
		&usedAt,
		&createdAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("PasswordResetToken", token)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find password reset token: %w", err)
	}

	tokenID, err := common.ParsePasswordResetTokenID(tokenIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token_id: %w", err)
	}

	adminID, err := common.ParseAdminID(adminIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse admin_id: %w", err)
	}

	var usedAtPtr *time.Time
	if usedAt.Valid {
		usedAtPtr = &usedAt.Time
	}

	return auth.ReconstructPasswordResetToken(
		tokenID,
		adminID,
		tokenStr,
		expiresAt,
		usedAtPtr,
		createdAt,
	)
}

// FindValidByAdminID finds a valid (unused and not expired) token for an admin
func (r *PasswordResetTokenRepository) FindValidByAdminID(ctx context.Context, adminID common.AdminID) (*auth.PasswordResetToken, error) {
	query := `
		SELECT
			token_id, admin_id, token, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE admin_id = $1 AND used_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`

	var (
		tokenIDStr string
		adminIDStr string
		tokenStr   string
		expiresAt  time.Time
		usedAt     sql.NullTime
		createdAt  time.Time
	)

	err := r.db.QueryRow(ctx, query, adminID.String()).Scan(
		&tokenIDStr,
		&adminIDStr,
		&tokenStr,
		&expiresAt,
		&usedAt,
		&createdAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("PasswordResetToken", adminID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find valid password reset token: %w", err)
	}

	tokenID, err := common.ParsePasswordResetTokenID(tokenIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token_id: %w", err)
	}

	parsedAdminID, err := common.ParseAdminID(adminIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse admin_id: %w", err)
	}

	var usedAtPtr *time.Time
	if usedAt.Valid {
		usedAtPtr = &usedAt.Time
	}

	return auth.ReconstructPasswordResetToken(
		tokenID,
		parsedAdminID,
		tokenStr,
		expiresAt,
		usedAtPtr,
		createdAt,
	)
}

// InvalidateAllByAdminID invalidates all tokens for an admin
func (r *PasswordResetTokenRepository) InvalidateAllByAdminID(ctx context.Context, adminID common.AdminID) error {
	query := `
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE admin_id = $1 AND used_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, adminID.String())
	if err != nil {
		return fmt.Errorf("failed to invalidate password reset tokens: %w", err)
	}

	return nil
}
