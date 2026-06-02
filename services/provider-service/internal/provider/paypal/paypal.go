package paypal

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/omniful/payment-platform/services/provider-service/internal/provider"
)

type PayPal struct{}

func New() *PayPal { return &PayPal{} }

func (p *PayPal) Name() string { return "paypal" }

func (p *PayPal) Charge(ctx context.Context, req provider.ChargeRequest) (string, error) {
	_ = ctx
	if req.Amount <= 0 {
		return "", fmt.Errorf("invalid amount")
	}
	return "paypal_txn_" + uuid.New().String()[:8], nil
}

func (p *PayPal) Refund(ctx context.Context, req provider.RefundRequest) error {
	_ = ctx
	_ = req
	return nil
}
