package service

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/omniful/payment-platform/shared/kafka"
	"github.com/omniful/payment-platform/services/payment-service/internal/repository"
)

type OutboxWorker struct {
	repo     *repository.PaymentRepository
	producer *kafka.Producer
	log      *zap.Logger
}

func NewOutboxWorker(repo *repository.PaymentRepository, producer *kafka.Producer, log *zap.Logger) *OutboxWorker {
	return &OutboxWorker{repo: repo, producer: producer, log: log}
}

func (w *OutboxWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *OutboxWorker) processBatch(ctx context.Context) {
	events, err := w.repo.FetchUnprocessedOutbox(ctx, 50)
	if err != nil {
		return
	}
	for _, e := range events {
		var payload map[string]interface{}
		_ = json.Unmarshal(e.Payload, &payload)
		topic := e.EventType
		if err := w.producer.Publish(ctx, topic, e.AggregateID, payload); err != nil {
			w.log.Warn("outbox publish failed", zap.Error(err), zap.String("event", e.EventType))
			continue
		}
		_ = w.repo.MarkOutboxProcessed(ctx, e.ID)
	}
}
