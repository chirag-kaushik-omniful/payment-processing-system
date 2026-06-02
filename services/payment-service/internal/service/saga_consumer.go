package service

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/omniful/payment-platform/shared/events"
	"github.com/omniful/payment-platform/shared/kafka"
	"github.com/omniful/payment-platform/shared/metrics"
	"github.com/omniful/payment-platform/services/payment-service/internal/domain"
	"github.com/omniful/payment-platform/services/payment-service/internal/repository"
)

type SagaConsumer struct {
	repo     *repository.PaymentRepository
	producer *kafka.Producer
	log      *zap.Logger
}

func NewSagaConsumer(repo *repository.PaymentRepository, producer *kafka.Producer, log *zap.Logger) *SagaConsumer {
	return &SagaConsumer{repo: repo, producer: producer, log: log}
}

func (s *SagaConsumer) publish(ctx context.Context, topic, key string, payload interface{}) error {
	data, err := events.Wrap(topic, payload)
	if err != nil {
		return err
	}
	var m map[string]interface{}
	_ = json.Unmarshal(data, &m)
	return s.producer.Publish(ctx, topic, key, payload)
}

func (s *SagaConsumer) HandlePaymentCreated(ctx context.Context, key string, raw json.RawMessage) error {
	var evt struct {
		PaymentID string  `json:"payment_id"`
		UserID    string  `json:"user_id"`
		Amount    float64 `json:"amount"`
	}
	_ = json.Unmarshal(raw, &evt)
	_ = s.repo.UpdateStatus(ctx, evt.PaymentID, domain.StatusProcessing)
	_ = s.producer.Publish(ctx, kafka.TopicPaymentProcessing, key, evt)
	return s.publish(ctx, kafka.TopicFraudValidate, key, map[string]interface{}{
		"payment_id": evt.PaymentID,
		"user_id":    evt.UserID,
		"amount":     evt.Amount,
	})
}

func (s *SagaConsumer) HandleFraudValidated(ctx context.Context, key string, raw json.RawMessage) error {
	var evt struct {
		PaymentID string  `json:"payment_id"`
		Amount    float64 `json:"amount"`
		Approved  bool    `json:"approved"`
	}
	_ = json.Unmarshal(raw, &evt)
	if !evt.Approved {
		_ = s.repo.UpdateStatus(ctx, evt.PaymentID, domain.StatusFailed)
		metrics.PaymentFailed.WithLabelValues("payment-service").Inc()
		return s.producer.Publish(ctx, kafka.TopicPaymentFailed, key, evt)
	}
	return s.publish(ctx, kafka.TopicProviderCharge, key, evt)
}

func (s *SagaConsumer) HandleProviderSuccess(ctx context.Context, key string, raw json.RawMessage) error {
	var evt struct {
		PaymentID string  `json:"payment_id"`
		UserID    string  `json:"user_id"`
		Amount    float64 `json:"amount"`
	}
	_ = json.Unmarshal(raw, &evt)
	_ = s.repo.UpdateStatus(ctx, evt.PaymentID, domain.StatusSuccess)
	metrics.PaymentSuccess.WithLabelValues("payment-service").Inc()
	return s.publish(ctx, kafka.TopicLedgerEntryCreated, key, evt)
}

func (s *SagaConsumer) HandleProviderFailed(ctx context.Context, key string, raw json.RawMessage) error {
	var evt struct {
		PaymentID string `json:"payment_id"`
	}
	_ = json.Unmarshal(raw, &evt)
	_ = s.repo.UpdateStatus(ctx, evt.PaymentID, domain.StatusFailed)
	metrics.PaymentFailed.WithLabelValues("payment-service").Inc()
	return s.producer.Publish(ctx, kafka.TopicPaymentFailed, key, evt)
}

func (s *SagaConsumer) HandleLedgerWritten(ctx context.Context, key string, raw json.RawMessage) error {
	var evt struct {
		PaymentID string  `json:"payment_id"`
		Amount    float64 `json:"amount"`
	}
	_ = json.Unmarshal(raw, &evt)
	_ = s.producer.Publish(ctx, kafka.TopicWalletDebited, key, map[string]interface{}{
		"user_id": "", "amount": evt.Amount, "payment_id": evt.PaymentID,
	})
	if err := s.publish(ctx, kafka.TopicNotificationSend, key, map[string]interface{}{
		"payment_id": evt.PaymentID,
		"status":     "completed",
	}); err != nil {
		return err
	}
	_ = s.repo.UpdateStatus(ctx, evt.PaymentID, domain.StatusCaptured)
	return s.producer.Publish(ctx, kafka.TopicPaymentCompleted, key, evt)
}

func (s *SagaConsumer) HandleRefundRequested(ctx context.Context, key string, raw json.RawMessage) error {
	var evt struct {
		RefundID  string  `json:"refund_id"`
		PaymentID string  `json:"payment_id"`
		Amount    float64 `json:"amount"`
		Provider  string  `json:"provider"`
	}
	_ = json.Unmarshal(raw, &evt)
	return s.publish(ctx, kafka.TopicProviderRefund, key, evt)
}

func (s *SagaConsumer) HandleRefundCompleted(ctx context.Context, key string, raw json.RawMessage) error {
	var evt struct {
		PaymentID string  `json:"payment_id"`
		Amount    float64 `json:"amount"`
	}
	_ = json.Unmarshal(raw, &evt)
	_ = s.repo.UpdateStatus(ctx, evt.PaymentID, domain.StatusRefunded)
	return s.publish(ctx, kafka.TopicLedgerCompensation, key, evt)
}

func (s *SagaConsumer) HandleWebhookReceived(ctx context.Context, key string, raw json.RawMessage) error {
	var evt struct {
		PaymentID string `json:"payment_id"`
		Status    string `json:"status"`
	}
	_ = json.Unmarshal(raw, &evt)
	if evt.Status == "dispute.opened" {
		_ = s.repo.UpdateStatus(ctx, evt.PaymentID, domain.StatusDisputed)
		return s.producer.Publish(ctx, kafka.TopicDisputeOpened, key, evt)
	}
	return nil
}
