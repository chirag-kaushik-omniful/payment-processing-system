package model

type CreateSubscriptionRequest struct {
	UserID   string  `json:"user_id" binding:"required"`
	PlanID   string  `json:"plan_id" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency"`
	Interval string  `json:"interval"`
}
