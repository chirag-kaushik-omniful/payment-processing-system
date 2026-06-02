package model

type CreatePaymentRequest struct {
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	Currency        string  `json:"currency" binding:"required"`
	Provider        string  `json:"provider"`
	PaymentMethodID string  `json:"payment_method_id"`
	CaptureMode     string  `json:"capture_mode"` // auto | manual (hold)
}

type CreatePaymentMethodRequest struct {
	Provider string `json:"provider" binding:"required"`
	LastFour string `json:"last_four"`
	Brand    string `json:"brand"`
}

type PaymentMethodResponse struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	LastFour string `json:"last_four"`
	Brand    string `json:"brand"`
}

type RefundRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

type PaymentResponse struct {
	ID       string  `json:"id"`
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Status   string  `json:"status"`
	Provider string  `json:"provider,omitempty"`
}
