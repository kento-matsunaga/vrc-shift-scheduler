package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// LicenseKeyRepository implements billing.LicenseKeyRepository for PostgreSQL
type LicenseKeyRepository struct {
	db *pgxpool.Pool
}

// NewLicenseKeyRepository creates a new LicenseKeyRepository
func NewLicenseKeyRepository(db *pgxpool.Pool) *LicenseKeyRepository {
	return &LicenseKeyRepository{db: db}
}

// Save saves a license key
func (r *LicenseKeyRepository) Save(ctx context.Context, k *billing.LicenseKey) error {
	query := `
		INSERT INTO license_keys (
			key_id, key_hash, status, issued_batch_id, expires_at, memo,
			used_at, used_tenant_id, revoked_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (key_id) DO UPDATE SET
			status = EXCLUDED.status,
			used_at = EXCLUDED.used_at,
			used_tenant_id = EXCLUDED.used_tenant_id,
			revoked_at = EXCLUDED.revoked_at
	`

	var usedTenantIDStr *string
	if k.UsedTenantID() != nil {
		s := k.UsedTenantID().String()
		usedTenantIDStr = &s
	}

	// Use GetTx to support transaction context
	_, err := GetTx(ctx, r.db).Exec(ctx, query,
		k.KeyID().String(),
		k.KeyHash(),
		k.Status().String(),
		k.BatchID(),
		k.ExpiresAt(),
		k.Memo(),
		k.UsedAt(),
		usedTenantIDStr,
		k.RevokedAt(),
		k.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save license key: %w", err)
	}

	return nil
}

// SaveBatch saves multiple license keys
func (r *LicenseKeyRepository) SaveBatch(ctx context.Context, keys []*billing.LicenseKey) error {
	if len(keys) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	query := `
		INSERT INTO license_keys (
			key_id, key_hash, status, issued_batch_id, expires_at, memo,
			used_at, used_tenant_id, revoked_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	for _, k := range keys {
		var usedTenantIDStr *string
		if k.UsedTenantID() != nil {
			s := k.UsedTenantID().String()
			usedTenantIDStr = &s
		}

		batch.Queue(query,
			k.KeyID().String(),
			k.KeyHash(),
			k.Status().String(),
			k.BatchID(),
			k.ExpiresAt(),
			k.Memo(),
			k.UsedAt(),
			usedTenantIDStr,
			k.RevokedAt(),
			k.CreatedAt(),
		)
	}

	results := r.db.SendBatch(ctx, batch)
	defer results.Close()

	for range keys {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("failed to save license key batch: %w", err)
		}
	}

	return nil
}

// FindByHashForUpdate finds a license key by its hash with row lock
func (r *LicenseKeyRepository) FindByHashForUpdate(ctx context.Context, keyHash string) (*billing.LicenseKey, error) {
	query := `
		SELECT
			key_id, key_hash, status, issued_batch_id, expires_at, memo,
			used_at, used_tenant_id, revoked_at, created_at
		FROM license_keys
		WHERE key_hash = $1
		FOR UPDATE
	`

	return r.scanLicenseKey(ctx, query, keyHash)
}

// FindByID finds a license key by ID
func (r *LicenseKeyRepository) FindByID(ctx context.Context, keyID billing.LicenseKeyID) (*billing.LicenseKey, error) {
	query := `
		SELECT
			key_id, key_hash, status, issued_batch_id, expires_at, memo,
			used_at, used_tenant_id, revoked_at, created_at
		FROM license_keys
		WHERE key_id = $1
	`

	return r.scanLicenseKey(ctx, query, keyID.String())
}

// FindByBatchID finds all license keys in a batch
func (r *LicenseKeyRepository) FindByBatchID(ctx context.Context, batchID string) ([]*billing.LicenseKey, error) {
	query := `
		SELECT
			key_id, key_hash, status, issued_batch_id, expires_at, memo,
			used_at, used_tenant_id, revoked_at, created_at
		FROM license_keys
		WHERE issued_batch_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Query(ctx, query, batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to query license keys: %w", err)
	}
	defer rows.Close()

	return r.scanLicenseKeys(rows)
}

// CountByStatus counts license keys by status
func (r *LicenseKeyRepository) CountByStatus(ctx context.Context, status billing.LicenseKeyStatus) (int, error) {
	query := `SELECT COUNT(*) FROM license_keys WHERE status = $1`

	var count int
	err := r.db.QueryRow(ctx, query, status.String()).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count license keys: %w", err)
	}

	return count, nil
}

// RevokeBatch revokes all keys in a batch
func (r *LicenseKeyRepository) RevokeBatch(ctx context.Context, batchID string) error {
	query := `
		UPDATE license_keys
		SET status = 'revoked', revoked_at = NOW()
		WHERE issued_batch_id = $1 AND status = 'unused'
	`

	_, err := r.db.Exec(ctx, query, batchID)
	if err != nil {
		return fmt.Errorf("failed to revoke batch: %w", err)
	}

	return nil
}

// List returns license keys with optional status filter
func (r *LicenseKeyRepository) List(ctx context.Context, status *billing.LicenseKeyStatus, limit, offset int) ([]*billing.LicenseKey, int, error) {
	// Count query
	countQuery := `SELECT COUNT(*) FROM license_keys`
	countArgs := []interface{}{}

	if status != nil {
		countQuery += ` WHERE status = $1`
		countArgs = append(countArgs, status.String())
	}

	var totalCount int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count license keys: %w", err)
	}

	// Data query
	query := `
		SELECT
			key_id, key_hash, status, issued_batch_id, expires_at, memo,
			used_at, used_tenant_id, revoked_at, created_at
		FROM license_keys
	`
	args := []interface{}{}
	argIdx := 1

	if status != nil {
		query += fmt.Sprintf(` WHERE status = $%d`, argIdx)
		args = append(args, status.String())
		argIdx++
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query license keys: %w", err)
	}
	defer rows.Close()

	keys, err := r.scanLicenseKeys(rows)
	if err != nil {
		return nil, 0, err
	}

	return keys, totalCount, nil
}

// FindByHashAndTenant はハッシュとテナントIDで使用済みライセンスキーを検索
// PWリセット時の本人確認に使用
func (r *LicenseKeyRepository) FindByHashAndTenant(ctx context.Context, keyHash string, tenantID common.TenantID) (*billing.LicenseKey, error) {
	query := `
		SELECT
			key_id, key_hash, status, issued_batch_id, expires_at, memo,
			used_at, used_tenant_id, revoked_at, created_at
		FROM license_keys
		WHERE key_hash = $1 AND used_tenant_id = $2 AND status = 'used'
	`

	return r.scanLicenseKey(ctx, query, keyHash, tenantID.String())
}

func (r *LicenseKeyRepository) scanLicenseKey(ctx context.Context, query string, args ...interface{}) (*billing.LicenseKey, error) {
	var (
		keyIDStr     string
		keyHash      string
		status       string
		batchID      sql.NullString
		expiresAt    sql.NullTime
		memo         sql.NullString
		usedAt       sql.NullTime
		usedTenantID sql.NullString
		revokedAt    sql.NullTime
		createdAt    time.Time
	)

	err := r.db.QueryRow(ctx, query, args...).Scan(
		&keyIDStr,
		&keyHash,
		&status,
		&batchID,
		&expiresAt,
		&memo,
		&usedAt,
		&usedTenantID,
		&revokedAt,
		&createdAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find license key: %w", err)
	}

	keyID, err := billing.ParseLicenseKeyID(keyIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse key_id: %w", err)
	}

	var batchIDPtr *string
	if batchID.Valid {
		batchIDPtr = &batchID.String
	}

	var expiresAtPtr *time.Time
	if expiresAt.Valid {
		expiresAtPtr = &expiresAt.Time
	}

	var memoStr string
	if memo.Valid {
		memoStr = memo.String
	}

	var usedAtPtr *time.Time
	if usedAt.Valid {
		usedAtPtr = &usedAt.Time
	}

	var usedTenantIDPtr *common.TenantID
	if usedTenantID.Valid {
		tid, err := common.ParseTenantID(usedTenantID.String)
		if err != nil {
			return nil, fmt.Errorf("failed to parse used_tenant_id: %w", err)
		}
		usedTenantIDPtr = &tid
	}

	var revokedAtPtr *time.Time
	if revokedAt.Valid {
		revokedAtPtr = &revokedAt.Time
	}

	return billing.ReconstructLicenseKey(
		keyID,
		keyHash,
		billing.LicenseKeyStatus(status),
		batchIDPtr,
		expiresAtPtr,
		memoStr,
		usedAtPtr,
		usedTenantIDPtr,
		revokedAtPtr,
		createdAt,
	)
}

func (r *LicenseKeyRepository) scanLicenseKeys(rows pgx.Rows) ([]*billing.LicenseKey, error) {
	var keys []*billing.LicenseKey

	for rows.Next() {
		var (
			keyIDStr     string
			keyHash      string
			status       string
			batchID      sql.NullString
			expiresAt    sql.NullTime
			memo         sql.NullString
			usedAt       sql.NullTime
			usedTenantID sql.NullString
			revokedAt    sql.NullTime
			createdAt    time.Time
		)

		if err := rows.Scan(
			&keyIDStr,
			&keyHash,
			&status,
			&batchID,
			&expiresAt,
			&memo,
			&usedAt,
			&usedTenantID,
			&revokedAt,
			&createdAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan license key: %w", err)
		}

		keyID, err := billing.ParseLicenseKeyID(keyIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse key_id: %w", err)
		}

		var batchIDPtr *string
		if batchID.Valid {
			batchIDPtr = &batchID.String
		}

		var expiresAtPtr *time.Time
		if expiresAt.Valid {
			expiresAtPtr = &expiresAt.Time
		}

		var memoStr string
		if memo.Valid {
			memoStr = memo.String
		}

		var usedAtPtr *time.Time
		if usedAt.Valid {
			usedAtPtr = &usedAt.Time
		}

		var usedTenantIDPtr *common.TenantID
		if usedTenantID.Valid {
			tid, err := common.ParseTenantID(usedTenantID.String)
			if err != nil {
				return nil, fmt.Errorf("failed to parse used_tenant_id: %w", err)
			}
			usedTenantIDPtr = &tid
		}

		var revokedAtPtr *time.Time
		if revokedAt.Valid {
			revokedAtPtr = &revokedAt.Time
		}

		k, err := billing.ReconstructLicenseKey(
			keyID,
			keyHash,
			billing.LicenseKeyStatus(status),
			batchIDPtr,
			expiresAtPtr,
			memoStr,
			usedAtPtr,
			usedTenantIDPtr,
			revokedAtPtr,
			createdAt,
		)
		if err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating license keys: %w", err)
	}

	return keys, nil
}
