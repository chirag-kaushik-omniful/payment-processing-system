package razorpay

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/omniful/payment-platform/services/provider-service/internal/provider"
)

// Adapter mocks Razorpay payment API.
type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Name() string { return "razorpay" }

func (a *Adapter) Charge(ctx context.Context, req provider.ChargeRequest) (string, error) {
	_ = ctx
	if req.Amount <= 0 {
		return "", fmt.Errorf("razorpay: invalid amount")
	}
	// Mock failure for test amounts ending in .99
	if int(req.Amount*100)%100 == 99 {
		return "", fmt.Errorf("razorpay: payment failed (mock)")
	}
	time.Sleep(10 * time.Millisecond)
	return "rzp_" + uuid.New().String(), nil
}

func (a *Adapter) Refund(ctx context.Context, req provider.RefundRequest) error {
	_ = ctx
	_ = req
	return nil
}
