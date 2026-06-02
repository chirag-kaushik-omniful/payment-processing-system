package service

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/omniful/payment-platform/services/analytics-service/internal/usecase"
)

type AnalyticsConsumer struct {
	uc  *usecase.AnalyticsUsecase
	log *zap.Logger
}

func NewAnalyticsConsumer(uc *usecase.AnalyticsUsecase, log *zap.Logger) *AnalyticsConsumer {
	return &AnalyticsConsumer{uc: uc, log: log}
}

func (c *AnalyticsConsumer) HandlePaymentCreated(ctx context.Context, key string, raw json.RawMessage) error {
	var evt paymentEvent
	_ = json.Unmarshal(raw, &evt)
	c.uc.RecordCreated(evt.Amount)
	return nil
}

func (c *AnalyticsConsumer) HandlePaymentCompleted(ctx context.Context, key string, raw json.RawMessage) error {
	var evt paymentEvent
	_ = json.Unmarshal(raw, &evt)
	c.uc.RecordCompleted(evt.Amount)
	c.log.Debug("payment completed metric", zap.String("payment_id", evt.PaymentID))
	return nil
}

func (c *AnalyticsConsumer) HandlePaymentFailed(ctx context.Context, key string, raw json.RawMessage) error {
	c.uc.RecordFailed()
	return nil
}

type paymentEvent struct {
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
}
