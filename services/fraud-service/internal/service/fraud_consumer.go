package service

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/omniful/payment-platform/shared/kafka"
	"github.com/omniful/payment-platform/services/fraud-service/internal/model"
	"github.com/omniful/payment-platform/services/fraud-service/internal/usecase"
)


type FraudConsumer struct {
	uc       *usecase.FraudUsecase
	producer *kafka.Producer
	log      *zap.Logger
}

func NewFraudConsumer(uc *usecase.FraudUsecase, producer *kafka.Producer, log *zap.Logger) *FraudConsumer {
	return &FraudConsumer{uc: uc, producer: producer, log: log}
}

func (c *FraudConsumer) HandleFraudValidate(ctx context.Context, key string, raw json.RawMessage) error {
	var evt model.FraudValidateEvent
	if err := json.Unmarshal(raw, &evt); err != nil {
		return err
	}
	result := c.uc.Evaluate(ctx, evt)
	if err := c.producer.Publish(ctx, kafka.TopicFraudValidated, key, result); err != nil {
		c.log.Error("publish fraud.validated failed", zap.Error(err))
		return err
	}
	c.log.Info("fraud validated",
		zap.String("payment_id", result.PaymentID),
		zap.Bool("approved", result.Approved),
	)
	return nil
}
