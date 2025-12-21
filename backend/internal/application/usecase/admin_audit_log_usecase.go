package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
)

// AdminAuditLogUsecase handles admin operations for audit logs
type AdminAuditLogUsecase struct {
	auditLogRepo billing.BillingAuditLogRepository
}

// NewAdminAuditLogUsecase creates a new AdminAuditLogUsecase
func NewAdminAuditLogUsecase(auditLogRepo billing.BillingAuditLogRepository) *AdminAuditLogUsecase {
	return &AdminAuditLogUsecase{
		auditLogRepo: auditLogRepo,
	}
}

// AuditLogListInput represents input for listing audit logs
type AuditLogListInput struct {
	Action *string
	Limit  int
	Offset int
}

// AuditLogListItem represents an audit log in the list
type AuditLogListItem struct {
	LogID       billing.BillingAuditLogID
	ActorType   billing.ActorType
	ActorID     *string
	Action      string
	TargetType  *string
	TargetID    *string
	BeforeJSON  *string
	AfterJSON   *string
	IPAddress   *string
	UserAgent   *string
	CreatedAt   time.Time
}

// AuditLogListOutput represents output from listing audit logs
type AuditLogListOutput struct {
	Logs       []AuditLogListItem
	TotalCount int
}

// List returns a list of audit logs
func (uc *AdminAuditLogUsecase) List(ctx context.Context, input AuditLogListInput) (*AuditLogListOutput, error) {
	if input.Limit <= 0 {
		input.Limit = 50
	}
	if input.Limit > 100 {
		input.Limit = 100
	}

	logs, totalCount, err := uc.auditLogRepo.List(ctx, input.Action, input.Limit, input.Offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list audit logs: %w", err)
	}

	items := make([]AuditLogListItem, len(logs))
	for i, l := range logs {
		items[i] = AuditLogListItem{
			LogID:      l.LogID(),
			ActorType:  l.ActorType(),
			ActorID:    l.ActorID(),
			Action:     l.Action(),
			TargetType: l.TargetType(),
			TargetID:   l.TargetID(),
			BeforeJSON: l.BeforeJSON(),
			AfterJSON:  l.AfterJSON(),
			IPAddress:  l.IPAddress(),
			UserAgent:  l.UserAgent(),
			CreatedAt:  l.CreatedAt(),
		}
	}

	return &AuditLogListOutput{
		Logs:       items,
		TotalCount: totalCount,
	}, nil
}
