package usecase

import (
	"context"
	"time"

	"github.com/omniful/payment-platform/shared/metrics"
	"github.com/omniful/payment-platform/services/provider-service/internal/domain"
	"github.com/omniful/payment-platform/services/provider-service/internal/provider"
	"github.com/omniful/payment-platform/services/provider-service/internal/provider/factory"
)

type ChargeUsecase struct {
	factory *factory.Factory
}

func NewChargeUsecase(f *factory.Factory) *ChargeUsecase {
	return &ChargeUsecase{factory: f}
}

func (u *ChargeUsecase) ProcessCharge(ctx context.Context, paymentID, providerName string, amount float64, currency string) domain.ChargeResult {
	p, err := u.factory.Get(providerName)
	if err != nil {
		return domain.ChargeResult{
			PaymentID:     paymentID,
			Provider:      providerName,
			Success:       false,
			FailureReason: err.Error(),
		}
	}

	start := time.Now()
	txnID, err := p.Charge(ctx, provider.ChargeRequest{
		PaymentID: paymentID,
		Amount:    amount,
		Currency:  currency,
	})
	metrics.ProviderLatency.WithLabelValues(p.Name()).Observe(time.Since(start).Seconds())

	if err != nil {
		return domain.ChargeResult{
			PaymentID:     paymentID,
			Provider:      p.Name(),
			Success:       false,
			FailureReason: err.Error(),
		}
	}
	return domain.ChargeResult{
		PaymentID:     paymentID,
		Provider:      p.Name(),
		TransactionID: txnID,
		Success:       true,
	}
}
