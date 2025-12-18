package attendance

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// ListCollectionsUsecase handles listing attendance collections for a tenant
type ListCollectionsUsecase struct {
	repo attendance.AttendanceCollectionRepository
}

// NewListCollectionsUsecase creates a new ListCollectionsUsecase
func NewListCollectionsUsecase(repo attendance.AttendanceCollectionRepository) *ListCollectionsUsecase {
	return &ListCollectionsUsecase{
		repo: repo,
	}
}

// ListCollectionsInput represents the input for listing collections
type ListCollectionsInput struct {
	TenantID string
}

// ListCollectionsOutput represents the output for listing collections
type ListCollectionsOutput struct {
	Collections []CollectionSummary `json:"collections"`
}

// CollectionSummary represents a summary of an attendance collection
type CollectionSummary struct {
	CollectionID    string     `json:"collection_id"`
	TenantID        string     `json:"tenant_id"`
	Title           string     `json:"title"`
	Description     string     `json:"description"`
	TargetType      string     `json:"target_type"`
	TargetID        string     `json:"target_id"`
	PublicToken     string     `json:"public_token"`
	Status          string     `json:"status"`
	Deadline        *time.Time `json:"deadline"`
	TargetDateCount int        `json:"target_date_count"`
	ResponseCount   int        `json:"response_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Execute executes the list collections use case
func (u *ListCollectionsUsecase) Execute(ctx context.Context, input ListCollectionsInput) (*ListCollectionsOutput, error) {
	// 1. Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// 2. Find all collections for this tenant
	collections, err := u.repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 3. Get counts for each collection
	summaries := make([]CollectionSummary, 0, len(collections))
	for _, c := range collections {
		// Get target dates count
		targetDates, err := u.repo.FindTargetDatesByCollectionID(ctx, c.CollectionID())
		if err != nil {
			return nil, err
		}

		// Get responses for this collection
		responses, err := u.repo.FindResponsesByCollectionID(ctx, c.CollectionID())
		if err != nil {
			return nil, err
		}

		// Count unique member responses
		memberMap := make(map[string]bool)
		for _, resp := range responses {
			memberMap[resp.MemberID().String()] = true
		}

		summaries = append(summaries, CollectionSummary{
			CollectionID:    c.CollectionID().String(),
			TenantID:        c.TenantID().String(),
			Title:           c.Title(),
			Description:     c.Description(),
			TargetType:      c.TargetType().String(),
			TargetID:        c.TargetID(),
			PublicToken:     c.PublicToken().String(),
			Status:          c.Status().String(),
			Deadline:        c.Deadline(),
			TargetDateCount: len(targetDates),
			ResponseCount:   len(memberMap),
			CreatedAt:       c.CreatedAt(),
			UpdatedAt:       c.UpdatedAt(),
		})
	}

	return &ListCollectionsOutput{
		Collections: summaries,
	}, nil
}
