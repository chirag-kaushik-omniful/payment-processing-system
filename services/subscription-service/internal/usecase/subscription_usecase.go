package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	apperrors "github.com/omniful/payment-platform/shared/errors"
	"github.com/omniful/payment-platform/services/subscription-service/internal/domain"
	"github.com/omniful/payment-platform/services/subscription-service/internal/model"
	"github.com/omniful/payment-platform/services/subscription-service/internal/repository"
)

type SubscriptionUsecase struct {
	repo *repository.SubscriptionRepository
}

func NewSubscriptionUsecase(repo *repository.SubscriptionRepository) *SubscriptionUsecase {
	return &SubscriptionUsecase{repo: repo}
}

func (u *SubscriptionUsecase) Create(ctx context.Context, req model.CreateSubscriptionRequest) (*model.SubscriptionResponse, error) {
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	interval := req.Interval
	if interval == "" {
		interval = domain.IntervalMonthly
	}
	if interval != domain.IntervalMonthly && interval != domain.IntervalYearly {
		return nil, apperrors.Validation("interval must be monthly or yearly")
	}

	now := time.Now().UTC()
	sub := &domain.Subscription{
		ID:            uuid.New().String(),
		UserID:        req.UserID,
		PlanID:        req.PlanID,
		Amount:        req.Amount,
		Currency:      currency,
		Interval:      interval,
		Status:        domain.StatusActive,
		NextBillingAt: nextBillingAt(now, interval),
	}
	if err := u.repo.Create(ctx, sub); err != nil {
		return nil, apperrors.Internal("create subscription failed", err)
	}
	created, err := u.repo.FindByID(ctx, sub.ID)
	if err != nil {
		return nil, err
	}
	return toResponse(created), nil
}

func (u *SubscriptionUsecase) GetByID(ctx context.Context, id string) (*model.SubscriptionResponse, error) {
	sub, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResponse(sub), nil
}

func (u *SubscriptionUsecase) Update(ctx context.Context, sub *domain.Subscription) error {
	if err := u.repo.Update(ctx, sub); err != nil {
		return apperrors.Internal("update subscription failed", err)
	}
	return nil
}

func (u *SubscriptionUsecase) Delete(ctx context.Context, id string) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return apperrors.Internal("delete subscription failed", err)
	}
	return nil
}

func (u *SubscriptionUsecase) ListDue(ctx context.Context, asOf time.Time, limit int) ([]domain.Subscription, error) {
	return u.repo.ListDue(ctx, asOf, limit)
}

func (u *SubscriptionUsecase) AdvanceBilling(ctx context.Context, sub *domain.Subscription) error {
	next := nextBillingAt(sub.NextBillingAt, sub.Interval)
	return u.repo.AdvanceBilling(ctx, sub.ID, next)
}

func nextBillingAt(from time.Time, interval string) time.Time {
	switch interval {
	case domain.IntervalYearly:
		return from.AddDate(1, 0, 0)
	default:
		return from.AddDate(0, 1, 0)
	}
}

func toResponse(sub *domain.Subscription) *model.SubscriptionResponse {
	return &model.SubscriptionResponse{
		ID:            sub.ID,
		UserID:        sub.UserID,
		PlanID:        sub.PlanID,
		Amount:        sub.Amount,
		Currency:      sub.Currency,
		Interval:      sub.Interval,
		Status:        sub.Status,
		NextBillingAt: sub.NextBillingAt,
		CreatedAt:     sub.CreatedAt,
		UpdatedAt:     sub.UpdatedAt,
	}
}
