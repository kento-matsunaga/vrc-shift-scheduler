package tutorial

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// TutorialID represents tutorial ID
type TutorialID string

func (id TutorialID) String() string {
	return string(id)
}

// Tutorial represents a tutorial entity
type Tutorial struct {
	id           TutorialID
	category     string
	title        string
	body         string
	displayOrder int
	isPublished  bool
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
}

// NewTutorial creates a new Tutorial
func NewTutorial(
	now time.Time,
	category string,
	title string,
	body string,
	displayOrder int,
) (*Tutorial, error) {
	if category == "" {
		return nil, ErrCategoryRequired
	}
	if len(category) > 50 {
		return nil, ErrCategoryTooLong
	}
	if title == "" {
		return nil, ErrTitleRequired
	}
	if len(title) > 200 {
		return nil, ErrTitleTooLong
	}
	if body == "" {
		return nil, ErrBodyRequired
	}

	return &Tutorial{
		id:           TutorialID(common.NewULID()),
		category:     category,
		title:        title,
		body:         body,
		displayOrder: displayOrder,
		isPublished:  false,
		createdAt:    now,
		updatedAt:    now,
		deletedAt:    nil,
	}, nil
}

// Reconstruct reconstructs a Tutorial from persistence
func Reconstruct(
	id TutorialID,
	category string,
	title string,
	body string,
	displayOrder int,
	isPublished bool,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *Tutorial {
	return &Tutorial{
		id:           id,
		category:     category,
		title:        title,
		body:         body,
		displayOrder: displayOrder,
		isPublished:  isPublished,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
	}
}

// Getters
func (t *Tutorial) ID() TutorialID     { return t.id }
func (t *Tutorial) Category() string   { return t.category }
func (t *Tutorial) Title() string      { return t.title }
func (t *Tutorial) Body() string       { return t.body }
func (t *Tutorial) DisplayOrder() int  { return t.displayOrder }
func (t *Tutorial) IsPublished() bool  { return t.isPublished }
func (t *Tutorial) CreatedAt() time.Time { return t.createdAt }
func (t *Tutorial) UpdatedAt() time.Time { return t.updatedAt }
func (t *Tutorial) DeletedAt() *time.Time { return t.deletedAt }

// Update updates the tutorial
func (t *Tutorial) Update(now time.Time, category, title, body string, displayOrder int) error {
	if category == "" {
		return ErrCategoryRequired
	}
	if len(category) > 50 {
		return ErrCategoryTooLong
	}
	if title == "" {
		return ErrTitleRequired
	}
	if len(title) > 200 {
		return ErrTitleTooLong
	}
	if body == "" {
		return ErrBodyRequired
	}

	t.category = category
	t.title = title
	t.body = body
	t.displayOrder = displayOrder
	t.updatedAt = now
	return nil
}

// Publish publishes the tutorial
func (t *Tutorial) Publish(now time.Time) {
	t.isPublished = true
	t.updatedAt = now
}

// Unpublish unpublishes the tutorial
func (t *Tutorial) Unpublish(now time.Time) {
	t.isPublished = false
	t.updatedAt = now
}

// Delete soft-deletes the tutorial
func (t *Tutorial) Delete(now time.Time) {
	t.deletedAt = &now
	t.updatedAt = now
}
