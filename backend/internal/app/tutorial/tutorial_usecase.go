package tutorial

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tutorial"
)

// TutorialOutput represents a tutorial in output
type TutorialOutput struct {
	ID           string    `json:"id"`
	Category     string    `json:"category"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	DisplayOrder int       `json:"display_order"`
	IsPublished  bool      `json:"is_published"`
	CreatedAt    time.Time `json:"created_at"`
}

// ListTutorialsUsecase lists published tutorials
type ListTutorialsUsecase struct {
	repo tutorial.Repository
}

func NewListTutorialsUsecase(repo tutorial.Repository) *ListTutorialsUsecase {
	return &ListTutorialsUsecase{repo: repo}
}

func (uc *ListTutorialsUsecase) Execute(ctx context.Context) ([]TutorialOutput, error) {
	tutorials, err := uc.repo.FindPublished(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]TutorialOutput, 0, len(tutorials))
	for _, t := range tutorials {
		outputs = append(outputs, TutorialOutput{
			ID:           t.ID().String(),
			Category:     t.Category(),
			Title:        t.Title(),
			Body:         t.Body(),
			DisplayOrder: t.DisplayOrder(),
			IsPublished:  t.IsPublished(),
			CreatedAt:    t.CreatedAt(),
		})
	}

	return outputs, nil
}

// GetTutorialUsecase gets a tutorial by ID
type GetTutorialUsecase struct {
	repo tutorial.Repository
}

func NewGetTutorialUsecase(repo tutorial.Repository) *GetTutorialUsecase {
	return &GetTutorialUsecase{repo: repo}
}

func (uc *GetTutorialUsecase) Execute(ctx context.Context, id string) (*TutorialOutput, error) {
	t, err := uc.repo.FindByID(ctx, tutorial.TutorialID(id))
	if err != nil {
		return nil, err
	}

	return &TutorialOutput{
		ID:           t.ID().String(),
		Category:     t.Category(),
		Title:        t.Title(),
		Body:         t.Body(),
		DisplayOrder: t.DisplayOrder(),
		IsPublished:  t.IsPublished(),
		CreatedAt:    t.CreatedAt(),
	}, nil
}

// CreateTutorialInput represents input for creating tutorial
type CreateTutorialInput struct {
	Category     string
	Title        string
	Body         string
	DisplayOrder int
	IsPublished  bool
}

// CreateTutorialUsecase creates a tutorial (admin)
type CreateTutorialUsecase struct {
	repo tutorial.Repository
}

func NewCreateTutorialUsecase(repo tutorial.Repository) *CreateTutorialUsecase {
	return &CreateTutorialUsecase{repo: repo}
}

func (uc *CreateTutorialUsecase) Execute(ctx context.Context, input CreateTutorialInput) (*TutorialOutput, error) {
	now := time.Now()
	t, err := tutorial.NewTutorial(now, input.Category, input.Title, input.Body, input.DisplayOrder)
	if err != nil {
		return nil, err
	}

	if input.IsPublished {
		t.Publish(now)
	}

	if err := uc.repo.Save(ctx, t); err != nil {
		return nil, err
	}

	return &TutorialOutput{
		ID:           t.ID().String(),
		Category:     t.Category(),
		Title:        t.Title(),
		Body:         t.Body(),
		DisplayOrder: t.DisplayOrder(),
		IsPublished:  t.IsPublished(),
		CreatedAt:    t.CreatedAt(),
	}, nil
}

// UpdateTutorialInput represents input for updating tutorial
type UpdateTutorialInput struct {
	ID           string
	Category     string
	Title        string
	Body         string
	DisplayOrder int
	IsPublished  bool
}

// UpdateTutorialUsecase updates a tutorial (admin)
type UpdateTutorialUsecase struct {
	repo tutorial.Repository
}

func NewUpdateTutorialUsecase(repo tutorial.Repository) *UpdateTutorialUsecase {
	return &UpdateTutorialUsecase{repo: repo}
}

func (uc *UpdateTutorialUsecase) Execute(ctx context.Context, input UpdateTutorialInput) (*TutorialOutput, error) {
	now := time.Now()
	t, err := uc.repo.FindByID(ctx, tutorial.TutorialID(input.ID))
	if err != nil {
		return nil, err
	}

	if err := t.Update(now, input.Category, input.Title, input.Body, input.DisplayOrder); err != nil {
		return nil, err
	}

	if input.IsPublished {
		t.Publish(now)
	} else {
		t.Unpublish(now)
	}

	if err := uc.repo.Save(ctx, t); err != nil {
		return nil, err
	}

	return &TutorialOutput{
		ID:           t.ID().String(),
		Category:     t.Category(),
		Title:        t.Title(),
		Body:         t.Body(),
		DisplayOrder: t.DisplayOrder(),
		IsPublished:  t.IsPublished(),
		CreatedAt:    t.CreatedAt(),
	}, nil
}

// DeleteTutorialUsecase deletes a tutorial (admin)
type DeleteTutorialUsecase struct {
	repo tutorial.Repository
}

func NewDeleteTutorialUsecase(repo tutorial.Repository) *DeleteTutorialUsecase {
	return &DeleteTutorialUsecase{repo: repo}
}

func (uc *DeleteTutorialUsecase) Execute(ctx context.Context, id string) error {
	t, err := uc.repo.FindByID(ctx, tutorial.TutorialID(id))
	if err != nil {
		return err
	}

	t.Delete(time.Now())
	return uc.repo.Save(ctx, t)
}

// ListAllTutorialsUsecase lists all tutorials (admin)
type ListAllTutorialsUsecase struct {
	repo tutorial.Repository
}

func NewListAllTutorialsUsecase(repo tutorial.Repository) *ListAllTutorialsUsecase {
	return &ListAllTutorialsUsecase{repo: repo}
}

func (uc *ListAllTutorialsUsecase) Execute(ctx context.Context) ([]TutorialOutput, error) {
	tutorials, err := uc.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]TutorialOutput, 0, len(tutorials))
	for _, t := range tutorials {
		outputs = append(outputs, TutorialOutput{
			ID:           t.ID().String(),
			Category:     t.Category(),
			Title:        t.Title(),
			Body:         t.Body(),
			DisplayOrder: t.DisplayOrder(),
			IsPublished:  t.IsPublished(),
			CreatedAt:    t.CreatedAt(),
		})
	}

	return outputs, nil
}
