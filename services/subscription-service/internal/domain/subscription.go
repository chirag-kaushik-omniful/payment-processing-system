package domain

import "time"

const (
	StatusActive    = "active"
	StatusCancelled = "cancelled"
	StatusPaused    = "paused"

	IntervalMonthly = "monthly"
	IntervalYearly  = "yearly"
)

type Subscription struct {
	ID            string
	UserID        string
	PlanID        string
	Amount        float64
	Currency      string
	Interval      string
	Status        string
	NextBillingAt time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
