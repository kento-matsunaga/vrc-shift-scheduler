package billing

import (
	"time"
)

// PlanType represents the type of a plan
type PlanType string

const (
	PlanTypeLifetime     PlanType = "lifetime"
	PlanTypeSubscription PlanType = "subscription"
)

// String returns the string representation
func (pt PlanType) String() string {
	return string(pt)
}

// IsValid checks if the plan type is valid
func (pt PlanType) IsValid() bool {
	return pt == PlanTypeLifetime || pt == PlanTypeSubscription
}

// Plan represents a subscription/purchase plan
type Plan struct {
	planCode     string
	planType     PlanType
	displayName  string
	priceJPY     *int
	featuresJSON string
	createdAt    time.Time
	updatedAt    time.Time
}

// ReconstructPlan reconstructs a Plan entity from persistence
func ReconstructPlan(
	planCode string,
	planType PlanType,
	displayName string,
	priceJPY *int,
	featuresJSON string,
	createdAt time.Time,
	updatedAt time.Time,
) *Plan {
	return &Plan{
		planCode:     planCode,
		planType:     planType,
		displayName:  displayName,
		priceJPY:     priceJPY,
		featuresJSON: featuresJSON,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

// Getters

func (p *Plan) PlanCode() string {
	return p.planCode
}

func (p *Plan) PlanType() PlanType {
	return p.planType
}

func (p *Plan) DisplayName() string {
	return p.displayName
}

func (p *Plan) PriceJPY() *int {
	return p.priceJPY
}

func (p *Plan) FeaturesJSON() string {
	return p.featuresJSON
}

func (p *Plan) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Plan) UpdatedAt() time.Time {
	return p.updatedAt
}

func (p *Plan) IsLifetime() bool {
	return p.planType == PlanTypeLifetime
}

func (p *Plan) IsSubscription() bool {
	return p.planType == PlanTypeSubscription
}
