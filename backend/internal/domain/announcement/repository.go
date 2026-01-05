package announcement

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// Repository defines the interface for announcement persistence
type Repository interface {
	Save(ctx context.Context, announcement *Announcement) error
	FindByID(ctx context.Context, id AnnouncementID) (*Announcement, error)
	FindAll(ctx context.Context) ([]*Announcement, error)
	FindPublishedForTenant(ctx context.Context, tenantID common.TenantID) ([]*Announcement, error)
}

// ReadRepository defines the interface for announcement read status
type ReadRepository interface {
	MarkAsRead(ctx context.Context, announcementID AnnouncementID, adminID common.AdminID) error
	MarkAllAsRead(ctx context.Context, adminID common.AdminID, tenantID common.TenantID) error
	IsRead(ctx context.Context, announcementID AnnouncementID, adminID common.AdminID) (bool, error)
	GetUnreadCount(ctx context.Context, adminID common.AdminID, tenantID common.TenantID) (int, error)
	GetReadAnnouncementIDs(ctx context.Context, adminID common.AdminID) ([]AnnouncementID, error)
}
