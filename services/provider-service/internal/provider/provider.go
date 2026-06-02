package provider

import "context"

type ChargeRequest struct {
	PaymentID string
	Amount    float64
	Currency  string
}

type RefundRequest struct {
	PaymentID     string
	TransactionID string
	Amount        float64
}

// PaymentProvider is the strategy interface for payment gateways.
type PaymentProvider interface {
	Name() string
	Charge(ctx context.Context, req ChargeRequest) (transactionID string, err error)
	Refund(ctx context.Context, req RefundRequest) error
}
