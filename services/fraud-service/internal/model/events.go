package model

type FraudValidateEvent struct {
	PaymentID  string  `json:"payment_id"`
	UserID     string  `json:"user_id"`
	Amount     float64 `json:"amount"`
	IP         string  `json:"ip,omitempty"`
	DeviceID   string  `json:"device_id,omitempty"`
	Country    string  `json:"country,omitempty"`
}

type FraudValidatedEvent struct {
	PaymentID  string  `json:"payment_id"`
	Amount     float64 `json:"amount"`
	Approved   bool    `json:"approved"`
	Reason     string  `json:"reason,omitempty"`
	RiskScore  int     `json:"risk_score"`
}
