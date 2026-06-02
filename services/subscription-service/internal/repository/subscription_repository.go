package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/omniful/payment-platform/shared/errors"
	"github.com/omniful/payment-platform/services/subscription-service/internal/domain"
)

type SubscriptionRepository struct {
	pool *pgxpool.Pool
}

func NewSubscriptionRepository(pool *pgxpool.Pool) *SubscriptionRepository {
	return &SubscriptionRepository{pool: pool}
}

func (r *SubscriptionRepository) Create(ctx context.Context, sub *domain.Subscription) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO subscriptions (id, user_id, plan_id, amount, currency, interval, status, next_billing_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		sub.ID, sub.UserID, sub.PlanID, sub.Amount, sub.Currency, sub.Interval, sub.Status, sub.NextBillingAt,
	)
	return err
}

func (r *SubscriptionRepository) FindByID(ctx context.Context, id string) (*domain.Subscription, error) {
	return r.scanOne(ctx,
		`SELECT id, user_id, plan_id, amount, currency, interval, status, next_billing_at, created_at, updated_at
		 FROM subscriptions WHERE id = $1`, id,
	)
}

func (r *SubscriptionRepository) Update(ctx context.Context, sub *domain.Subscription) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE subscriptions SET plan_id = $2, amount = $3, currency = $4, interval = $5,
		 status = $6, next_billing_at = $7, updated_at = NOW() WHERE id = $1`,
		sub.ID, sub.PlanID, sub.Amount, sub.Currency, sub.Interval, sub.Status, sub.NextBillingAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *SubscriptionRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *SubscriptionRepository) ListDue(ctx context.Context, asOf time.Time, limit int) ([]domain.Subscription, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, plan_id, amount, currency, interval, status, next_billing_at, created_at, updated_at
		 FROM subscriptions
		 WHERE status = $1 AND next_billing_at <= $2
		 ORDER BY next_billing_at ASC
		 LIMIT $3`,
		domain.StatusActive, asOf, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []domain.Subscription
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, *sub)
	}
	return subs, rows.Err()
}

func (r *SubscriptionRepository) AdvanceBilling(ctx context.Context, id string, next time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE subscriptions SET next_billing_at = $2, updated_at = NOW() WHERE id = $1`, id, next,
	)
	return err
}

func (r *SubscriptionRepository) scanOne(ctx context.Context, query string, id string) (*domain.Subscription, error) {
	row := r.pool.QueryRow(ctx, query, id)
	sub, err := scanSubscription(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("subscription not found")
		}
		return nil, err
	}
	return sub, nil
}

type scannable interface {
	Scan(dest ...interface{}) error
}

func scanSubscription(row scannable) (*domain.Subscription, error) {
	var sub domain.Subscription
	err := row.Scan(
		&sub.ID, &sub.UserID, &sub.PlanID, &sub.Amount, &sub.Currency,
		&sub.Interval, &sub.Status, &sub.NextBillingAt, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}
