package db

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/announcement"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnnouncementRepository struct {
	pool *pgxpool.Pool
}

func NewAnnouncementRepository(pool *pgxpool.Pool) *AnnouncementRepository {
	return &AnnouncementRepository{pool: pool}
}

func (r *AnnouncementRepository) Save(ctx context.Context, a *announcement.Announcement) error {
	query := `
		INSERT INTO announcements (id, tenant_id, title, body, published_at, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			tenant_id = EXCLUDED.tenant_id,
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			published_at = EXCLUDED.published_at,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	var tenantID *string
	if a.TenantID() != nil {
		tid := a.TenantID().String()
		tenantID = &tid
	}

	_, err := r.pool.Exec(ctx, query,
		a.ID().String(),
		tenantID,
		a.Title(),
		a.Body(),
		a.PublishedAt(),
		a.CreatedAt(),
		a.UpdatedAt(),
		a.DeletedAt(),
	)
	return err
}

func (r *AnnouncementRepository) FindByID(ctx context.Context, id announcement.AnnouncementID) (*announcement.Announcement, error) {
	query := `
		SELECT id, tenant_id, title, body, published_at, created_at, updated_at, deleted_at
		FROM announcements
		WHERE id = $1 AND deleted_at IS NULL
	`

	var (
		aid         string
		tenantID    *string
		title       string
		body        string
		publishedAt time.Time
		createdAt   time.Time
		updatedAt   time.Time
		deletedAt   *time.Time
	)

	err := r.pool.QueryRow(ctx, query, id.String()).Scan(
		&aid, &tenantID, &title, &body, &publishedAt, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		return nil, err
	}

	var tid *common.TenantID
	if tenantID != nil {
		t := common.TenantID(*tenantID)
		tid = &t
	}

	return announcement.Reconstruct(
		announcement.AnnouncementID(aid),
		tid,
		title,
		body,
		publishedAt,
		createdAt,
		updatedAt,
		deletedAt,
	), nil
}

func (r *AnnouncementRepository) FindAll(ctx context.Context) ([]*announcement.Announcement, error) {
	query := `
		SELECT id, tenant_id, title, body, published_at, created_at, updated_at, deleted_at
		FROM announcements
		WHERE deleted_at IS NULL
		ORDER BY published_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var announcements []*announcement.Announcement
	for rows.Next() {
		var (
			id          string
			tenantID    *string
			title       string
			body        string
			publishedAt time.Time
			createdAt   time.Time
			updatedAt   time.Time
			deletedAt   *time.Time
		)

		if err := rows.Scan(&id, &tenantID, &title, &body, &publishedAt, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}

		var tid *common.TenantID
		if tenantID != nil {
			t := common.TenantID(*tenantID)
			tid = &t
		}

		announcements = append(announcements, announcement.Reconstruct(
			announcement.AnnouncementID(id),
			tid,
			title,
			body,
			publishedAt,
			createdAt,
			updatedAt,
			deletedAt,
		))
	}

	return announcements, nil
}

func (r *AnnouncementRepository) FindPublishedForTenant(ctx context.Context, tenantID common.TenantID) ([]*announcement.Announcement, error) {
	query := `
		SELECT id, tenant_id, title, body, published_at, created_at, updated_at, deleted_at
		FROM announcements
		WHERE deleted_at IS NULL
		  AND published_at <= NOW()
		  AND (tenant_id IS NULL OR tenant_id = $1)
		ORDER BY published_at DESC
	`

	rows, err := r.pool.Query(ctx, query, tenantID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var announcements []*announcement.Announcement
	for rows.Next() {
		var (
			id          string
			tid         *string
			title       string
			body        string
			publishedAt time.Time
			createdAt   time.Time
			updatedAt   time.Time
			deletedAt   *time.Time
		)

		if err := rows.Scan(&id, &tid, &title, &body, &publishedAt, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}

		var tenantIDPtr *common.TenantID
		if tid != nil {
			t := common.TenantID(*tid)
			tenantIDPtr = &t
		}

		announcements = append(announcements, announcement.Reconstruct(
			announcement.AnnouncementID(id),
			tenantIDPtr,
			title,
			body,
			publishedAt,
			createdAt,
			updatedAt,
			deletedAt,
		))
	}

	return announcements, nil
}

// AnnouncementReadRepository handles read status
type AnnouncementReadRepository struct {
	pool *pgxpool.Pool
}

func NewAnnouncementReadRepository(pool *pgxpool.Pool) *AnnouncementReadRepository {
	return &AnnouncementReadRepository{pool: pool}
}

func (r *AnnouncementReadRepository) MarkAsRead(ctx context.Context, announcementID announcement.AnnouncementID, adminID common.AdminID) error {
	query := `
		INSERT INTO announcement_reads (id, announcement_id, admin_id, read_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (announcement_id, admin_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query,
		common.NewULID(),
		announcementID.String(),
		adminID.String(),
		time.Now(),
	)
	return err
}

func (r *AnnouncementReadRepository) MarkAllAsRead(ctx context.Context, adminID common.AdminID, tenantID common.TenantID) error {
	// INSERT ... SELECT で一括挿入（N+1問題を解消）
	// 疑似ULID: gen_random_uuid()を26文字に切り詰め（039_migrate_instance_dataで実績あり）
	query := `
		INSERT INTO announcement_reads (id, announcement_id, admin_id, read_at)
		SELECT
			SUBSTRING(UPPER(REPLACE(gen_random_uuid()::text, '-', '')), 1, 26),
			a.id,
			$1,
			NOW()
		FROM announcements a
		WHERE a.deleted_at IS NULL
		  AND a.published_at <= NOW()
		  AND (a.tenant_id IS NULL OR a.tenant_id = $2)
		  AND NOT EXISTS (
			SELECT 1 FROM announcement_reads ar
			WHERE ar.announcement_id = a.id AND ar.admin_id = $1
		  )
		ON CONFLICT (announcement_id, admin_id) DO NOTHING
	`

	_, err := r.pool.Exec(ctx, query, adminID.String(), tenantID.String())
	return err
}

func (r *AnnouncementReadRepository) IsRead(ctx context.Context, announcementID announcement.AnnouncementID, adminID common.AdminID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM announcement_reads
			WHERE announcement_id = $1 AND admin_id = $2
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, announcementID.String(), adminID.String()).Scan(&exists)
	return exists, err
}

func (r *AnnouncementReadRepository) GetUnreadCount(ctx context.Context, adminID common.AdminID, tenantID common.TenantID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM announcements a
		WHERE a.deleted_at IS NULL
		  AND a.published_at <= NOW()
		  AND (a.tenant_id IS NULL OR a.tenant_id = $1)
		  AND NOT EXISTS (
			SELECT 1 FROM announcement_reads ar
			WHERE ar.announcement_id = a.id AND ar.admin_id = $2
		  )
	`

	var count int
	err := r.pool.QueryRow(ctx, query, tenantID.String(), adminID.String()).Scan(&count)
	return count, err
}

func (r *AnnouncementReadRepository) GetReadAnnouncementIDs(ctx context.Context, adminID common.AdminID) ([]announcement.AnnouncementID, error) {
	query := `
		SELECT announcement_id FROM announcement_reads WHERE admin_id = $1
	`

	rows, err := r.pool.Query(ctx, query, adminID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []announcement.AnnouncementID
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, announcement.AnnouncementID(id))
	}

	return ids, nil
}
