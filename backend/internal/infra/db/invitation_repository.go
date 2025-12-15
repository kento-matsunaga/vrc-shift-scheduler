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

// InvitationRepository implements auth.InvitationRepository for PostgreSQL
type InvitationRepository struct {
	db *pgxpool.Pool
}

// NewInvitationRepository creates a new InvitationRepository
func NewInvitationRepository(db *pgxpool.Pool) *InvitationRepository {
	return &InvitationRepository{db: db}
}

// Save saves an invitation (insert or update)
func (r *InvitationRepository) Save(ctx context.Context, inv *auth.Invitation) error {
	query := `
		INSERT INTO invitations (
			invitation_id, tenant_id, email, role, token,
			created_by_admin_id, expires_at, accepted_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (invitation_id) DO UPDATE SET
			email = EXCLUDED.email,
			role = EXCLUDED.role,
			accepted_at = EXCLUDED.accepted_at
	`

	_, err := r.db.Exec(ctx, query,
		inv.InvitationID().String(),
		inv.TenantID().String(),
		inv.Email(),
		inv.Role().String(),
		inv.Token(),
		inv.CreatedByAdminID().String(),
		inv.ExpiresAt(),
		inv.AcceptedAt(),
		inv.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save invitation: %w", err)
	}

	return nil
}

// FindByToken finds an invitation by token
func (r *InvitationRepository) FindByToken(ctx context.Context, token string) (*auth.Invitation, error) {
	query := `
		SELECT
			invitation_id, tenant_id, email, role, token,
			created_by_admin_id, expires_at, accepted_at, created_at
		FROM invitations
		WHERE token = $1
		LIMIT 1
	`

	var (
		invitationIDStr    string
		tenantIDStr        string
		email              string
		roleStr            string
		tokenStr           string
		createdByAdminIDStr string
		expiresAt          time.Time
		acceptedAt         sql.NullTime
		createdAt          time.Time
	)

	err := r.db.QueryRow(ctx, query, token).Scan(
		&invitationIDStr,
		&tenantIDStr,
		&email,
		&roleStr,
		&tokenStr,
		&createdByAdminIDStr,
		&expiresAt,
		&acceptedAt,
		&createdAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Invitation", token)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find invitation: %w", err)
	}

	tenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
	}

	createdByAdminID, err := common.ParseAdminID(createdByAdminIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_by_admin_id: %w", err)
	}

	role, err := auth.NewRole(roleStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse role: %w", err)
	}

	var acceptedAtPtr *time.Time
	if acceptedAt.Valid {
		acceptedAtPtr = &acceptedAt.Time
	}

	return auth.ReconstructInvitation(
		auth.InvitationID(invitationIDStr),
		tenantID,
		email,
		role,
		tokenStr,
		createdByAdminID,
		expiresAt,
		acceptedAtPtr,
		createdAt,
	)
}

// FindByTenantID finds all invitations within a tenant
func (r *InvitationRepository) FindByTenantID(ctx context.Context, tenantID common.TenantID) ([]*auth.Invitation, error) {
	query := `
		SELECT
			invitation_id, tenant_id, email, role, token,
			created_by_admin_id, expires_at, accepted_at, created_at
		FROM invitations
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, tenantID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to query invitations: %w", err)
	}
	defer rows.Close()

	invitations := make([]*auth.Invitation, 0)

	for rows.Next() {
		var (
			invitationIDStr    string
			tenantIDStr        string
			email              string
			roleStr            string
			tokenStr           string
			createdByAdminIDStr string
			expiresAt          time.Time
			acceptedAt         sql.NullTime
			createdAt          time.Time
		)

		err := rows.Scan(
			&invitationIDStr,
			&tenantIDStr,
			&email,
			&roleStr,
			&tokenStr,
			&createdByAdminIDStr,
			&expiresAt,
			&acceptedAt,
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}

		parsedTenantID, err := common.ParseTenantID(tenantIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tenant_id: %w", err)
		}

		createdByAdminID, err := common.ParseAdminID(createdByAdminIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_by_admin_id: %w", err)
		}

		role, err := auth.NewRole(roleStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse role: %w", err)
		}

		var acceptedAtPtr *time.Time
		if acceptedAt.Valid {
			acceptedAtPtr = &acceptedAt.Time
		}

		inv, err := auth.ReconstructInvitation(
			auth.InvitationID(invitationIDStr),
			parsedTenantID,
			email,
			role,
			tokenStr,
			createdByAdminID,
			expiresAt,
			acceptedAtPtr,
			createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to reconstruct invitation: %w", err)
		}

		invitations = append(invitations, inv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate invitations: %w", err)
	}

	return invitations, nil
}

// ExistsPendingByEmail checks if a pending invitation exists for the email
func (r *InvitationRepository) ExistsPendingByEmail(ctx context.Context, tenantID common.TenantID, email string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM invitations
			WHERE tenant_id = $1 AND email = $2 AND accepted_at IS NULL
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, tenantID.String(), email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check pending invitation: %w", err)
	}

	return exists, nil
}

// Delete deletes an invitation (physical delete)
func (r *InvitationRepository) Delete(ctx context.Context, invitationID auth.InvitationID) error {
	query := `DELETE FROM invitations WHERE invitation_id = $1`

	result, err := r.db.Exec(ctx, query, invitationID.String())
	if err != nil {
		return fmt.Errorf("failed to delete invitation: %w", err)
	}

	if result.RowsAffected() == 0 {
		return common.NewNotFoundError("Invitation", invitationID.String())
	}

	return nil
}
