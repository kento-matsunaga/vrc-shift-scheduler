package announcement

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// AnnouncementID represents announcement ID
type AnnouncementID string

func (id AnnouncementID) String() string {
	return string(id)
}

// Announcement represents an announcement entity
type Announcement struct {
	id          AnnouncementID
	tenantID    *common.TenantID // nil means for all tenants
	title       string
	body        string
	publishedAt time.Time
	createdAt   time.Time
	updatedAt   time.Time
	deletedAt   *time.Time
}

// NewAnnouncement creates a new Announcement
func NewAnnouncement(
	now time.Time,
	tenantID *common.TenantID,
	title string,
	body string,
	publishedAt time.Time,
) (*Announcement, error) {
	if title == "" {
		return nil, ErrTitleRequired
	}
	if len(title) > 200 {
		return nil, ErrTitleTooLong
	}
	if body == "" {
		return nil, ErrBodyRequired
	}

	return &Announcement{
		id:          AnnouncementID(common.NewULID()),
		tenantID:    tenantID,
		title:       title,
		body:        body,
		publishedAt: publishedAt,
		createdAt:   now,
		updatedAt:   now,
		deletedAt:   nil,
	}, nil
}

// Reconstruct reconstructs an Announcement from persistence
func Reconstruct(
	id AnnouncementID,
	tenantID *common.TenantID,
	title string,
	body string,
	publishedAt time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *Announcement {
	return &Announcement{
		id:          id,
		tenantID:    tenantID,
		title:       title,
		body:        body,
		publishedAt: publishedAt,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		deletedAt:   deletedAt,
	}
}

// Getters
func (a *Announcement) ID() AnnouncementID      { return a.id }
func (a *Announcement) TenantID() *common.TenantID { return a.tenantID }
func (a *Announcement) Title() string           { return a.title }
func (a *Announcement) Body() string            { return a.body }
func (a *Announcement) PublishedAt() time.Time  { return a.publishedAt }
func (a *Announcement) CreatedAt() time.Time    { return a.createdAt }
func (a *Announcement) UpdatedAt() time.Time    { return a.updatedAt }
func (a *Announcement) DeletedAt() *time.Time   { return a.deletedAt }

// IsPublished returns true if the announcement is currently published
func (a *Announcement) IsPublished(now time.Time) bool {
	return a.deletedAt == nil && !a.publishedAt.After(now)
}

// IsForAllTenants returns true if the announcement is for all tenants
func (a *Announcement) IsForAllTenants() bool {
	return a.tenantID == nil
}

// Update updates the announcement
func (a *Announcement) Update(now time.Time, title, body string, publishedAt time.Time) error {
	if title == "" {
		return ErrTitleRequired
	}
	if len(title) > 200 {
		return ErrTitleTooLong
	}
	if body == "" {
		return ErrBodyRequired
	}

	a.title = title
	a.body = body
	a.publishedAt = publishedAt
	a.updatedAt = now
	return nil
}

// Delete soft-deletes the announcement
func (a *Announcement) Delete(now time.Time) {
	a.deletedAt = &now
	a.updatedAt = now
}
