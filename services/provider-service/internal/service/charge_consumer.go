package service

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/omniful/payment-platform/shared/kafka"
	"github.com/omniful/payment-platform/services/provider-service/internal/usecase"
)

type ChargeConsumer struct {
	uc       *usecase.ChargeUsecase
	producer *kafka.Producer
	log      *zap.Logger
}

func NewChargeConsumer(uc *usecase.ChargeUsecase, producer *kafka.Producer, log *zap.Logger) *ChargeConsumer {
	return &ChargeConsumer{uc: uc, producer: producer, log: log}
}

func (c *ChargeConsumer) HandleProviderCharge(ctx context.Context, key string, raw json.RawMessage) error {
	var evt struct {
		PaymentID string  `json:"payment_id"`
		Amount    float64 `json:"amount"`
		Currency  string  `json:"currency"`
		Provider  string  `json:"provider"`
	}
	if err := json.Unmarshal(raw, &evt); err != nil {
		return err
	}
	if evt.Currency == "" {
		evt.Currency = "USD"
	}

	result := c.uc.ProcessCharge(ctx, evt.PaymentID, evt.Provider, evt.Amount, evt.Currency)
	payload := map[string]interface{}{
		"payment_id": evt.PaymentID,
		"provider":   result.Provider,
	}
	if result.Success {
		payload["transaction_id"] = result.TransactionID
		return c.producer.Publish(ctx, kafka.TopicProviderSuccess, key, payload)
	}
	payload["reason"] = result.FailureReason
	return c.producer.Publish(ctx, kafka.TopicProviderFailed, key, payload)
}
