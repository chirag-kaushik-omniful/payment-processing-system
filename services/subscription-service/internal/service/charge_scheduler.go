package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/omniful/payment-platform/shared/kafka"
	"github.com/omniful/payment-platform/services/subscription-service/internal/domain"
	"github.com/omniful/payment-platform/services/subscription-service/internal/usecase"
)

type ChargeScheduler struct {
	uc       *usecase.SubscriptionUsecase
	producer *kafka.Producer
	log      *zap.Logger
	interval time.Duration
}

func NewChargeScheduler(uc *usecase.SubscriptionUsecase, producer *kafka.Producer, log *zap.Logger, tickInterval time.Duration) *ChargeScheduler {
	if tickInterval <= 0 {
		tickInterval = time.Minute
	}
	return &ChargeScheduler{uc: uc, producer: producer, log: log, interval: tickInterval}
}

func (s *ChargeScheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *ChargeScheduler) tick(ctx context.Context) {
	now := time.Now().UTC()
	due, err := s.uc.ListDue(ctx, now, 100)
	if err != nil {
		s.log.Error("list due subscriptions failed", zap.Error(err))
		return
	}
	for i := range due {
		sub := due[i]
		if err := s.publishCharge(ctx, &sub); err != nil {
			s.log.Error("publish subscription.charge failed",
				zap.String("subscription_id", sub.ID),
				zap.Error(err),
			)
			continue
		}
		if err := s.uc.AdvanceBilling(ctx, &sub); err != nil {
			s.log.Error("advance billing failed",
				zap.String("subscription_id", sub.ID),
				zap.Error(err),
			)
		}
	}
}

func (s *ChargeScheduler) publishCharge(ctx context.Context, sub *domain.Subscription) error {
	payload := map[string]interface{}{
		"subscription_id": sub.ID,
		"user_id":         sub.UserID,
		"plan_id":         sub.PlanID,
		"amount":          sub.Amount,
		"currency":        sub.Currency,
		"interval":        sub.Interval,
	}
	return s.producer.Publish(ctx, kafka.TopicSubscriptionCharge, sub.ID, payload)
}
