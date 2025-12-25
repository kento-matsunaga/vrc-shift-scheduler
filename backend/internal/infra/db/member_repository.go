package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MemberRepository implements member.MemberRepository for PostgreSQL
type MemberRepository struct {
	db *pgxpool.Pool
}

// NewMemberRepository creates a new MemberRepository
func NewMemberRepository(db *pgxpool.Pool) *MemberRepository {
	return &MemberRepository{db: db}
}

// Save saves a member (insert or update)
func (r *MemberRepository) Save(ctx context.Context, m *member.Member) error {
	query := `
		INSERT INTO members (
			member_id, tenant_id, display_name, discord_user_id, email,
			is_active, created_at, updated_at, deleted_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (member_id) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			discord_user_id = EXCLUDED.discord_user_id,
			email = EXCLUDED.email,
			is_active = EXCLUDED.is_active,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.db.Exec(ctx, query,
		m.MemberID().String(),
		m.TenantID().String(),
		m.DisplayName(),
		nullString(m.DiscordUserID()),
		nullString(m.Email()),
		m.IsActive(),
		m.CreatedAt(),
		m.UpdatedAt(),
		m.DeletedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save member: %w", err)
	}

	return nil
}

// FindByID finds a member by ID within a tenant
func (r *MemberRepository) FindByID(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) (*member.Member, error) {
	query := `
		SELECT
			member_id, tenant_id, display_name, discord_user_id, email,
			is_active, created_at, updated_at, deleted_at
		FROM members
		WHERE tenant_id = $1 AND member_id = $2 AND deleted_at IS NULL
	`

	var (
		memberIDStr     string
		tenantIDStr     string
		displayName     string
		discordUserID   sql.NullString
		email           sql.NullString
		isActive        bool
		createdAt       time.Time
		updatedAt       time.Time
		deletedAt       sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), memberID.String()).Scan(
		&memberIDStr,
		&tenantIDStr,
		&displayName,
		&discordUserID,
		&email,
		&isActive,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Member", memberID.String())
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find member: %w", err)
	}

	return r.scanToMember(
		memberIDStr, tenantIDStr, displayName, discordUserID, email,
		isActive, createdAt, updatedAt, deletedAt,
	)
}

// FindByTenantID finds all members within a tenant
func (r *MemberRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error) {
	query := `
		SELECT
			member_id, tenant_id, display_name, discord_user_id, email,
			is_active, created_at, updated_at, deleted_at
		FROM members
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	return r.queryMembers(ctx, query, tenantID.String())
}

// FindActiveByTenantID finds all active members within a tenant
func (r *MemberRepository) FindActiveByTenantID(ctx context.Context, tenantID common.TenantID) ([]*member.Member, error) {
	query := `
		SELECT
			member_id, tenant_id, display_name, discord_user_id, email,
			is_active, created_at, updated_at, deleted_at
		FROM members
		WHERE tenant_id = $1 AND is_active = true AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	return r.queryMembers(ctx, query, tenantID.String())
}

// FindByDiscordUserID finds a member by Discord user ID within a tenant
func (r *MemberRepository) FindByDiscordUserID(ctx context.Context, tenantID common.TenantID, discordUserID string) (*member.Member, error) {
	query := `
		SELECT
			member_id, tenant_id, display_name, discord_user_id, email,
			is_active, created_at, updated_at, deleted_at
		FROM members
		WHERE tenant_id = $1 AND discord_user_id = $2 AND deleted_at IS NULL
	`

	var (
		memberIDStr       string
		tenantIDStr       string
		displayName       string
		discordUserIDVal  sql.NullString
		email             sql.NullString
		isActive          bool
		createdAt         time.Time
		updatedAt         time.Time
		deletedAt         sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), discordUserID).Scan(
		&memberIDStr,
		&tenantIDStr,
		&displayName,
		&discordUserIDVal,
		&email,
		&isActive,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Member", discordUserID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find member by discord_user_id: %w", err)
	}

	return r.scanToMember(
		memberIDStr, tenantIDStr, displayName, discordUserIDVal, email,
		isActive, createdAt, updatedAt, deletedAt,
	)
}

// FindByEmail finds a member by email within a tenant
func (r *MemberRepository) FindByEmail(ctx context.Context, tenantID common.TenantID, emailAddr string) (*member.Member, error) {
	query := `
		SELECT
			member_id, tenant_id, display_name, discord_user_id, email,
			is_active, created_at, updated_at, deleted_at
		FROM members
		WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL
	`

	var (
		memberIDStr     string
		tenantIDStr     string
		displayName     string
		discordUserID   sql.NullString
		email           sql.NullString
		isActive        bool
		createdAt       time.Time
		updatedAt       time.Time
		deletedAt       sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), emailAddr).Scan(
		&memberIDStr,
		&tenantIDStr,
		&displayName,
		&discordUserID,
		&email,
		&isActive,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Member", emailAddr)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find member by email: %w", err)
	}

	return r.scanToMember(
		memberIDStr, tenantIDStr, displayName, discordUserID, email,
		isActive, createdAt, updatedAt, deletedAt,
	)
}

// FindByDisplayName finds a member by display name within a tenant
func (r *MemberRepository) FindByDisplayName(ctx context.Context, tenantID common.TenantID, displayName string) (*member.Member, error) {
	query := `
		SELECT
			member_id, tenant_id, display_name, discord_user_id, email,
			is_active, created_at, updated_at, deleted_at
		FROM members
		WHERE tenant_id = $1 AND display_name = $2 AND deleted_at IS NULL
	`

	var (
		memberIDStr       string
		tenantIDStr       string
		displayNameVal    string
		discordUserID     sql.NullString
		email             sql.NullString
		isActive          bool
		createdAt         time.Time
		updatedAt         time.Time
		deletedAt         sql.NullTime
	)

	err := r.db.QueryRow(ctx, query, tenantID.String(), displayName).Scan(
		&memberIDStr,
		&tenantIDStr,
		&displayNameVal,
		&discordUserID,
		&email,
		&isActive,
		&createdAt,
		&updatedAt,
		&deletedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil // Not found - return nil without error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find member by display_name: %w", err)
	}

	return r.scanToMember(
		memberIDStr, tenantIDStr, displayNameVal, discordUserID, email,
		isActive, createdAt, updatedAt, deletedAt,
	)
}

// Delete deletes a member (physical delete)
func (r *MemberRepository) Delete(ctx context.Context, tenantID common.TenantID, memberID common.MemberID) error {
	query := `
		DELETE FROM members
		WHERE tenant_id = $1 AND member_id = $2
	`

	result, err := r.db.Exec(ctx, query, tenantID.String(), memberID.String())
	if err != nil {
		return fmt.Errorf("failed to delete member: %w", err)
	}

	if result.RowsAffected() == 0 {
		return common.NewNotFoundError("Member", memberID.String())
	}

	return nil
}

// ExistsByDiscordUserID checks if a member with the given Discord user ID exists within a tenant
func (r *MemberRepository) ExistsByDiscordUserID(ctx context.Context, tenantID common.TenantID, discordUserID string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM members
			WHERE tenant_id = $1 AND discord_user_id = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, tenantID.String(), discordUserID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check member existence by discord_user_id: %w", err)
	}

	return exists, nil
}

// ExistsByEmail checks if a member with the given email exists within a tenant
func (r *MemberRepository) ExistsByEmail(ctx context.Context, tenantID common.TenantID, emailAddr string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM members
			WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, tenantID.String(), emailAddr).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check member existence by email: %w", err)
	}

	return exists, nil
}

// queryMembers executes a query and returns a list of members
func (r *MemberRepository) queryMembers(ctx context.Context, query string, args ...interface{}) ([]*member.Member, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query members: %w", err)
	}
	defer rows.Close()

	var members []*member.Member
	for rows.Next() {
		var (
			memberIDStr     string
			tenantIDStr     string
			displayName     string
			discordUserID   sql.NullString
			email           sql.NullString
			isActive        bool
			createdAt       time.Time
			updatedAt       time.Time
			deletedAt       sql.NullTime
		)

		err := rows.Scan(
			&memberIDStr,
			&tenantIDStr,
			&displayName,
			&discordUserID,
			&email,
			&isActive,
			&createdAt,
			&updatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member row: %w", err)
		}

		m, err := r.scanToMember(
			memberIDStr, tenantIDStr, displayName, discordUserID, email,
			isActive, createdAt, updatedAt, deletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct member: %w", err)
		}

		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating member rows: %w", err)
	}

	return members, nil
}

// scanToMember converts scanned row data to Member entity
func (r *MemberRepository) scanToMember(
	memberIDStr, tenantIDStr, displayName string,
	discordUserID, email sql.NullString,
	isActive bool,
	createdAt, updatedAt time.Time,
	deletedAt sql.NullTime,
) (*member.Member, error) {
	var deletedAtPtr *time.Time
	if deletedAt.Valid {
		deletedAtPtr = &deletedAt.Time
	}

	return member.ReconstructMember(
		common.MemberID(memberIDStr),
		common.TenantID(tenantIDStr),
		displayName,
		stringValue(discordUserID),
		stringValue(email),
		isActive,
		createdAt,
		updatedAt,
		deletedAtPtr,
	)
}

// nullString converts a string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// stringValue converts sql.NullString to string
func stringValue(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

