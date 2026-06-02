package domain

type ChargeResult struct {
	PaymentID     string
	Provider      string
	TransactionID string
	Success       bool
	FailureReason string
}
