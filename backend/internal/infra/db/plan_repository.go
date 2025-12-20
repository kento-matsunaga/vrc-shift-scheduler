package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/billing"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PlanRepository implements billing.PlanRepository for PostgreSQL
type PlanRepository struct {
	db *pgxpool.Pool
}

// NewPlanRepository creates a new PlanRepository
func NewPlanRepository(db *pgxpool.Pool) *PlanRepository {
	return &PlanRepository{db: db}
}

// FindByCode finds a plan by its code
func (r *PlanRepository) FindByCode(ctx context.Context, planCode string) (*billing.Plan, error) {
	query := `
		SELECT
			plan_code, plan_type, display_name, price_jpy, features_json,
			created_at, updated_at
		FROM plans
		WHERE plan_code = $1
	`

	var (
		code         string
		planType     string
		displayName  string
		priceJPY     sql.NullInt32
		featuresJSON string
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := r.db.QueryRow(ctx, query, planCode).Scan(
		&code,
		&planType,
		&displayName,
		&priceJPY,
		&featuresJSON,
		&createdAt,
		&updatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, common.NewNotFoundError("Plan", planCode)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find plan: %w", err)
	}

	var pricePtr *int
	if priceJPY.Valid {
		price := int(priceJPY.Int32)
		pricePtr = &price
	}

	return billing.ReconstructPlan(
		code,
		billing.PlanType(planType),
		displayName,
		pricePtr,
		featuresJSON,
		createdAt,
		updatedAt,
	), nil
}

// FindAll retrieves all plans
func (r *PlanRepository) FindAll(ctx context.Context) ([]*billing.Plan, error) {
	query := `
		SELECT
			plan_code, plan_type, display_name, price_jpy, features_json,
			created_at, updated_at
		FROM plans
		ORDER BY plan_code
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query plans: %w", err)
	}
	defer rows.Close()

	var plans []*billing.Plan
	for rows.Next() {
		var (
			code         string
			planType     string
			displayName  string
			priceJPY     sql.NullInt32
			featuresJSON string
			createdAt    time.Time
			updatedAt    time.Time
		)

		if err := rows.Scan(
			&code,
			&planType,
			&displayName,
			&priceJPY,
			&featuresJSON,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}

		var pricePtr *int
		if priceJPY.Valid {
			price := int(priceJPY.Int32)
			pricePtr = &price
		}

		plan := billing.ReconstructPlan(
			code,
			billing.PlanType(planType),
			displayName,
			pricePtr,
			featuresJSON,
			createdAt,
			updatedAt,
		)
		plans = append(plans, plan)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating plans: %w", err)
	}

	return plans, nil
}
