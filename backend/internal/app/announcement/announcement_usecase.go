package announcement

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/announcement"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// AnnouncementOutput represents an announcement in output
type AnnouncementOutput struct {
	ID          string    `json:"id"`
	TenantID    *string   `json:"tenant_id"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	IsRead      bool      `json:"is_read"`
}

// ListAnnouncementsUsecase lists announcements for a tenant
type ListAnnouncementsUsecase struct {
	repo     announcement.Repository
	readRepo announcement.ReadRepository
}

func NewListAnnouncementsUsecase(repo announcement.Repository, readRepo announcement.ReadRepository) *ListAnnouncementsUsecase {
	return &ListAnnouncementsUsecase{repo: repo, readRepo: readRepo}
}

func (uc *ListAnnouncementsUsecase) Execute(ctx context.Context, tenantID common.TenantID, adminID common.AdminID) ([]AnnouncementOutput, error) {
	announcements, err := uc.repo.FindPublishedForTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	readIDs, err := uc.readRepo.GetReadAnnouncementIDs(ctx, adminID)
	if err != nil {
		return nil, err
	}

	readMap := make(map[string]bool)
	for _, id := range readIDs {
		readMap[id.String()] = true
	}

	outputs := make([]AnnouncementOutput, 0, len(announcements))
	for _, a := range announcements {
		var tenantIDStr *string
		if a.TenantID() != nil {
			tid := a.TenantID().String()
			tenantIDStr = &tid
		}

		outputs = append(outputs, AnnouncementOutput{
			ID:          a.ID().String(),
			TenantID:    tenantIDStr,
			Title:       a.Title(),
			Body:        a.Body(),
			PublishedAt: a.PublishedAt(),
			CreatedAt:   a.CreatedAt(),
			IsRead:      readMap[a.ID().String()],
		})
	}

	return outputs, nil
}

// GetUnreadCountUsecase gets unread announcement count
type GetUnreadCountUsecase struct {
	readRepo announcement.ReadRepository
}

func NewGetUnreadCountUsecase(readRepo announcement.ReadRepository) *GetUnreadCountUsecase {
	return &GetUnreadCountUsecase{readRepo: readRepo}
}

func (uc *GetUnreadCountUsecase) Execute(ctx context.Context, adminID common.AdminID, tenantID common.TenantID) (int, error) {
	return uc.readRepo.GetUnreadCount(ctx, adminID, tenantID)
}

// MarkAsReadUsecase marks an announcement as read
type MarkAsReadUsecase struct {
	readRepo announcement.ReadRepository
}

func NewMarkAsReadUsecase(readRepo announcement.ReadRepository) *MarkAsReadUsecase {
	return &MarkAsReadUsecase{readRepo: readRepo}
}

func (uc *MarkAsReadUsecase) Execute(ctx context.Context, announcementID string, adminID common.AdminID) error {
	return uc.readRepo.MarkAsRead(ctx, announcement.AnnouncementID(announcementID), adminID)
}

// MarkAllAsReadUsecase marks all announcements as read
type MarkAllAsReadUsecase struct {
	readRepo announcement.ReadRepository
}

func NewMarkAllAsReadUsecase(readRepo announcement.ReadRepository) *MarkAllAsReadUsecase {
	return &MarkAllAsReadUsecase{readRepo: readRepo}
}

func (uc *MarkAllAsReadUsecase) Execute(ctx context.Context, adminID common.AdminID, tenantID common.TenantID) error {
	return uc.readRepo.MarkAllAsRead(ctx, adminID, tenantID)
}

// CreateAnnouncementInput represents input for creating announcement
type CreateAnnouncementInput struct {
	TenantID    *string   `json:"tenant_id"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
}

// CreateAnnouncementUsecase creates an announcement (admin)
type CreateAnnouncementUsecase struct {
	repo announcement.Repository
}

func NewCreateAnnouncementUsecase(repo announcement.Repository) *CreateAnnouncementUsecase {
	return &CreateAnnouncementUsecase{repo: repo}
}

func (uc *CreateAnnouncementUsecase) Execute(ctx context.Context, input CreateAnnouncementInput) (*AnnouncementOutput, error) {
	var tenantID *common.TenantID
	if input.TenantID != nil {
		tid := common.TenantID(*input.TenantID)
		tenantID = &tid
	}

	a, err := announcement.NewAnnouncement(time.Now(), tenantID, input.Title, input.Body, input.PublishedAt)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, a); err != nil {
		return nil, err
	}

	var tenantIDStr *string
	if a.TenantID() != nil {
		tid := a.TenantID().String()
		tenantIDStr = &tid
	}

	return &AnnouncementOutput{
		ID:          a.ID().String(),
		TenantID:    tenantIDStr,
		Title:       a.Title(),
		Body:        a.Body(),
		PublishedAt: a.PublishedAt(),
		CreatedAt:   a.CreatedAt(),
		IsRead:      false,
	}, nil
}

// UpdateAnnouncementInput represents input for updating announcement
type UpdateAnnouncementInput struct {
	ID          string    `json:"-"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	PublishedAt time.Time `json:"published_at"`
}

// UpdateAnnouncementUsecase updates an announcement (admin)
type UpdateAnnouncementUsecase struct {
	repo announcement.Repository
}

func NewUpdateAnnouncementUsecase(repo announcement.Repository) *UpdateAnnouncementUsecase {
	return &UpdateAnnouncementUsecase{repo: repo}
}

func (uc *UpdateAnnouncementUsecase) Execute(ctx context.Context, input UpdateAnnouncementInput) (*AnnouncementOutput, error) {
	a, err := uc.repo.FindByID(ctx, announcement.AnnouncementID(input.ID))
	if err != nil {
		return nil, err
	}

	if err := a.Update(time.Now(), input.Title, input.Body, input.PublishedAt); err != nil {
		return nil, err
	}

	if err := uc.repo.Save(ctx, a); err != nil {
		return nil, err
	}

	var tenantIDStr *string
	if a.TenantID() != nil {
		tid := a.TenantID().String()
		tenantIDStr = &tid
	}

	return &AnnouncementOutput{
		ID:          a.ID().String(),
		TenantID:    tenantIDStr,
		Title:       a.Title(),
		Body:        a.Body(),
		PublishedAt: a.PublishedAt(),
		CreatedAt:   a.CreatedAt(),
		IsRead:      false,
	}, nil
}

// DeleteAnnouncementUsecase deletes an announcement (admin)
type DeleteAnnouncementUsecase struct {
	repo announcement.Repository
}

func NewDeleteAnnouncementUsecase(repo announcement.Repository) *DeleteAnnouncementUsecase {
	return &DeleteAnnouncementUsecase{repo: repo}
}

func (uc *DeleteAnnouncementUsecase) Execute(ctx context.Context, id string) error {
	a, err := uc.repo.FindByID(ctx, announcement.AnnouncementID(id))
	if err != nil {
		return err
	}

	a.Delete(time.Now())
	return uc.repo.Save(ctx, a)
}

// ListAllAnnouncementsUsecase lists all announcements (admin)
type ListAllAnnouncementsUsecase struct {
	repo announcement.Repository
}

func NewListAllAnnouncementsUsecase(repo announcement.Repository) *ListAllAnnouncementsUsecase {
	return &ListAllAnnouncementsUsecase{repo: repo}
}

func (uc *ListAllAnnouncementsUsecase) Execute(ctx context.Context) ([]AnnouncementOutput, error) {
	announcements, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]AnnouncementOutput, 0, len(announcements))
	for _, a := range announcements {
		var tenantIDStr *string
		if a.TenantID() != nil {
			tid := a.TenantID().String()
			tenantIDStr = &tid
		}

		outputs = append(outputs, AnnouncementOutput{
			ID:          a.ID().String(),
			TenantID:    tenantIDStr,
			Title:       a.Title(),
			Body:        a.Body(),
			PublishedAt: a.PublishedAt(),
			CreatedAt:   a.CreatedAt(),
			IsRead:      false,
		})
	}

	return outputs, nil
}
