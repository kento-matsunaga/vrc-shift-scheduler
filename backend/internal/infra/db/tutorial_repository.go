package db

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tutorial"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TutorialRepository struct {
	pool *pgxpool.Pool
}

func NewTutorialRepository(pool *pgxpool.Pool) *TutorialRepository {
	return &TutorialRepository{pool: pool}
}

func (r *TutorialRepository) Save(ctx context.Context, t *tutorial.Tutorial) error {
	query := `
		INSERT INTO tutorials (id, category, title, body, display_order, is_published, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET
			category = EXCLUDED.category,
			title = EXCLUDED.title,
			body = EXCLUDED.body,
			display_order = EXCLUDED.display_order,
			is_published = EXCLUDED.is_published,
			updated_at = EXCLUDED.updated_at,
			deleted_at = EXCLUDED.deleted_at
	`

	_, err := r.pool.Exec(ctx, query,
		t.ID().String(),
		t.Category(),
		t.Title(),
		t.Body(),
		t.DisplayOrder(),
		t.IsPublished(),
		t.CreatedAt(),
		t.UpdatedAt(),
		t.DeletedAt(),
	)
	return err
}

func (r *TutorialRepository) FindByID(ctx context.Context, id tutorial.TutorialID) (*tutorial.Tutorial, error) {
	query := `
		SELECT id, category, title, body, display_order, is_published, created_at, updated_at, deleted_at
		FROM tutorials
		WHERE id = $1 AND deleted_at IS NULL
	`

	var (
		tid          string
		category     string
		title        string
		body         string
		displayOrder int
		isPublished  bool
		createdAt    time.Time
		updatedAt    time.Time
		deletedAt    *time.Time
	)

	err := r.pool.QueryRow(ctx, query, id.String()).Scan(
		&tid, &category, &title, &body, &displayOrder, &isPublished, &createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		return nil, err
	}

	return tutorial.Reconstruct(
		tutorial.TutorialID(tid),
		category,
		title,
		body,
		displayOrder,
		isPublished,
		createdAt,
		updatedAt,
		deletedAt,
	), nil
}

func (r *TutorialRepository) FindAll(ctx context.Context) ([]*tutorial.Tutorial, error) {
	query := `
		SELECT id, category, title, body, display_order, is_published, created_at, updated_at, deleted_at
		FROM tutorials
		WHERE deleted_at IS NULL
		ORDER BY category, display_order
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tutorials []*tutorial.Tutorial
	for rows.Next() {
		var (
			id           string
			category     string
			title        string
			body         string
			displayOrder int
			isPublished  bool
			createdAt    time.Time
			updatedAt    time.Time
			deletedAt    *time.Time
		)

		if err := rows.Scan(&id, &category, &title, &body, &displayOrder, &isPublished, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}

		tutorials = append(tutorials, tutorial.Reconstruct(
			tutorial.TutorialID(id),
			category,
			title,
			body,
			displayOrder,
			isPublished,
			createdAt,
			updatedAt,
			deletedAt,
		))
	}

	return tutorials, nil
}

func (r *TutorialRepository) FindPublished(ctx context.Context) ([]*tutorial.Tutorial, error) {
	query := `
		SELECT id, category, title, body, display_order, is_published, created_at, updated_at, deleted_at
		FROM tutorials
		WHERE deleted_at IS NULL AND is_published = true
		ORDER BY category, display_order
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tutorials []*tutorial.Tutorial
	for rows.Next() {
		var (
			id           string
			category     string
			title        string
			body         string
			displayOrder int
			isPublished  bool
			createdAt    time.Time
			updatedAt    time.Time
			deletedAt    *time.Time
		)

		if err := rows.Scan(&id, &category, &title, &body, &displayOrder, &isPublished, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}

		tutorials = append(tutorials, tutorial.Reconstruct(
			tutorial.TutorialID(id),
			category,
			title,
			body,
			displayOrder,
			isPublished,
			createdAt,
			updatedAt,
			deletedAt,
		))
	}

	return tutorials, nil
}

func (r *TutorialRepository) FindByCategory(ctx context.Context, category string) ([]*tutorial.Tutorial, error) {
	query := `
		SELECT id, category, title, body, display_order, is_published, created_at, updated_at, deleted_at
		FROM tutorials
		WHERE deleted_at IS NULL AND category = $1
		ORDER BY display_order
	`

	rows, err := r.pool.Query(ctx, query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tutorials []*tutorial.Tutorial
	for rows.Next() {
		var (
			id           string
			cat          string
			title        string
			body         string
			displayOrder int
			isPublished  bool
			createdAt    time.Time
			updatedAt    time.Time
			deletedAt    *time.Time
		)

		if err := rows.Scan(&id, &cat, &title, &body, &displayOrder, &isPublished, &createdAt, &updatedAt, &deletedAt); err != nil {
			return nil, err
		}

		tutorials = append(tutorials, tutorial.Reconstruct(
			tutorial.TutorialID(id),
			cat,
			title,
			body,
			displayOrder,
			isPublished,
			createdAt,
			updatedAt,
			deletedAt,
		))
	}

	return tutorials, nil
}
