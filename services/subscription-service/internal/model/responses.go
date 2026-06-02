package model

import "time"

type SubscriptionResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	PlanID        string    `json:"plan_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Interval      string    `json:"interval"`
	Status        string    `json:"status"`
	NextBillingAt time.Time `json:"next_billing_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
