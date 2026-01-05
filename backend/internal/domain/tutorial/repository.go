package tutorial

import "context"

// Repository defines the interface for tutorial persistence
type Repository interface {
	Save(ctx context.Context, tutorial *Tutorial) error
	FindByID(ctx context.Context, id TutorialID) (*Tutorial, error)
	FindAll(ctx context.Context) ([]*Tutorial, error)
	FindPublished(ctx context.Context) ([]*Tutorial, error)
	FindByCategory(ctx context.Context, category string) ([]*Tutorial, error)
}
