package service

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/omniful/payment-platform/shared/kafka"
	"github.com/omniful/payment-platform/services/provider-service/internal/provider"
	"github.com/omniful/payment-platform/services/provider-service/internal/provider/factory"
)

type RefundConsumer struct {
	factory  *factory.Factory
	producer *kafka.Producer
	log      *zap.Logger
}

func NewRefundConsumer(f *factory.Factory, producer *kafka.Producer, log *zap.Logger) *RefundConsumer {
	return &RefundConsumer{factory: f, producer: producer, log: log}
}

func (c *RefundConsumer) HandleProviderRefund(ctx context.Context, key string, raw json.RawMessage) error {
	var evt struct {
		RefundID  string  `json:"refund_id"`
		PaymentID string  `json:"payment_id"`
		Amount    float64 `json:"amount"`
		Provider  string  `json:"provider"`
	}
	if err := json.Unmarshal(raw, &evt); err != nil {
		return err
	}
	p, err := c.factory.Get(evt.Provider)
	if err != nil {
		return err
	}
	if err := p.Refund(ctx, provider.RefundRequest{PaymentID: evt.PaymentID, Amount: evt.Amount}); err != nil {
		c.log.Warn("refund failed", zap.Error(err))
		return err
	}
	return c.producer.Publish(ctx, kafka.TopicRefundCompleted, key, map[string]interface{}{
		"refund_id": evt.RefundID, "payment_id": evt.PaymentID, "amount": evt.Amount,
	})
}
