package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BillingAuditLogRepository implements billing.BillingAuditLogRepository for PostgreSQL
type BillingAuditLogRepository struct {
	db *pgxpool.Pool
}

// NewBillingAuditLogRepository creates a new BillingAuditLogRepository
func NewBillingAuditLogRepository(db *pgxpool.Pool) *BillingAuditLogRepository {
	return &BillingAuditLogRepository{db: db}
}

// Save saves a billing audit log
func (r *BillingAuditLogRepository) Save(ctx context.Context, log *billing.BillingAuditLog) error {
	query := `
		INSERT INTO billing_audit_logs (
			log_id, actor_type, actor_id, action, target_type, target_id,
			before_json, after_json, ip_address, user_agent, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.Exec(ctx, query,
		log.LogID().String(),
		log.ActorType().String(),
		log.ActorID(),
		log.Action(),
		log.TargetType(),
		log.TargetID(),
		log.BeforeJSON(),
		log.AfterJSON(),
		log.IPAddress(),
		log.UserAgent(),
		log.CreatedAt(),
	)

	if err != nil {
		return fmt.Errorf("failed to save billing audit log: %w", err)
	}

	return nil
}

// FindByDateRange finds audit logs within a date range with pagination
func (r *BillingAuditLogRepository) FindByDateRange(ctx context.Context, startDate, endDate string, limit, offset int) ([]*billing.BillingAuditLog, error) {
	query := `
		SELECT
			log_id, actor_type, actor_id, action, target_type, target_id,
			before_json, after_json, ip_address, user_agent, created_at
		FROM billing_audit_logs
		WHERE created_at >= $1::date AND created_at < ($2::date + INTERVAL '1 day')
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.db.Query(ctx, query, startDate, endDate, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query billing audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*billing.BillingAuditLog
	for rows.Next() {
		log, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating billing audit logs: %w", err)
	}

	return logs, nil
}

// FindByAction finds audit logs by action
func (r *BillingAuditLogRepository) FindByAction(ctx context.Context, action string, limit, offset int) ([]*billing.BillingAuditLog, error) {
	query := `
		SELECT
			log_id, actor_type, actor_id, action, target_type, target_id,
			before_json, after_json, ip_address, user_agent, created_at
		FROM billing_audit_logs
		WHERE action = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, action, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query billing audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*billing.BillingAuditLog
	for rows.Next() {
		log, err := r.scanRow(rows)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating billing audit logs: %w", err)
	}

	return logs, nil
}

// CountByDateRange counts audit logs within a date range
func (r *BillingAuditLogRepository) CountByDateRange(ctx context.Context, startDate, endDate string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM billing_audit_logs
		WHERE created_at >= $1::date AND created_at < ($2::date + INTERVAL '1 day')
	`

	var count int
	err := r.db.QueryRow(ctx, query, startDate, endDate).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count billing audit logs: %w", err)
	}

	return count, nil
}

type scannable interface {
	Scan(dest ...interface{}) error
}

func (r *BillingAuditLogRepository) scanRow(row scannable) (*billing.BillingAuditLog, error) {
	var (
		logIDStr   string
		actorType  string
		actorID    sql.NullString
		action     string
		targetType sql.NullString
		targetID   sql.NullString
		beforeJSON sql.NullString
		afterJSON  sql.NullString
		ipAddress  sql.NullString
		userAgent  sql.NullString
		createdAt  time.Time
	)

	if err := row.Scan(
		&logIDStr,
		&actorType,
		&actorID,
		&action,
		&targetType,
		&targetID,
		&beforeJSON,
		&afterJSON,
		&ipAddress,
		&userAgent,
		&createdAt,
	); err != nil {
		return nil, fmt.Errorf("failed to scan billing audit log: %w", err)
	}

	var actorIDPtr *string
	if actorID.Valid {
		actorIDPtr = &actorID.String
	}

	var targetTypePtr *string
	if targetType.Valid {
		targetTypePtr = &targetType.String
	}

	var targetIDPtr *string
	if targetID.Valid {
		targetIDPtr = &targetID.String
	}

	var beforeJSONPtr *string
	if beforeJSON.Valid {
		beforeJSONPtr = &beforeJSON.String
	}

	var afterJSONPtr *string
	if afterJSON.Valid {
		afterJSONPtr = &afterJSON.String
	}

	var ipAddressPtr *string
	if ipAddress.Valid {
		ipAddressPtr = &ipAddress.String
	}

	var userAgentPtr *string
	if userAgent.Valid {
		userAgentPtr = &userAgent.String
	}

	return billing.ReconstructBillingAuditLog(
		billing.BillingAuditLogID(logIDStr),
		billing.ActorType(actorType),
		actorIDPtr,
		action,
		targetTypePtr,
		targetIDPtr,
		beforeJSONPtr,
		afterJSONPtr,
		ipAddressPtr,
		userAgentPtr,
		createdAt,
	)
}
