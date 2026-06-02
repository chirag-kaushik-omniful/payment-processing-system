package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omniful/payment-platform/services/payment-service/internal/domain"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) CreateWithOutbox(ctx context.Context, p *domain.Payment, eventType string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO payments (id, user_id, amount, currency, status, provider, idempotency_key)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		p.ID, p.UserID, p.Amount, p.Currency, p.Status, p.Provider, p.IdempotencyKey,
	)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO outbox (id, aggregate_id, event_type, payload) VALUES ($1, $2, $3, $4)`,
		uuid.New().String(), p.ID, eventType, payloadBytes,
	)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PaymentRepository) FindByID(ctx context.Context, id string) (*domain.Payment, error) {
	var p domain.Payment
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, amount, currency, status, COALESCE(provider,''), idempotency_key, created_at, updated_at
		 FROM payments WHERE id = $1`, id,
	).Scan(&p.ID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.Provider, &p.IdempotencyKey, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) FindByIdempotencyKey(ctx context.Context, key string) (*domain.Payment, error) {
	var p domain.Payment
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, amount, currency, status, COALESCE(provider,''), idempotency_key, created_at, updated_at
		 FROM payments WHERE idempotency_key = $1`, key,
	).Scan(&p.ID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.Provider, &p.IdempotencyKey, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payments SET status = $1, updated_at = NOW() WHERE id = $2`, status, id,
	)
	return err
}

func (r *PaymentRepository) FetchUnprocessedOutbox(ctx context.Context, limit int) ([]domain.OutboxEvent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, aggregate_id, event_type, payload::text, processed, created_at
		 FROM outbox WHERE processed = FALSE ORDER BY created_at LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []domain.OutboxEvent
	for rows.Next() {
		var e domain.OutboxEvent
		var payloadStr string
		if err := rows.Scan(&e.ID, &e.AggregateID, &e.EventType, &payloadStr, &e.Processed, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Payload = []byte(payloadStr)
		events = append(events, e)
	}
	return events, nil
}

func (r *PaymentRepository) MarkOutboxProcessed(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE outbox SET processed = TRUE WHERE id = $1`, id)
	return err
}

func (r *PaymentRepository) CreateRefund(ctx context.Context, refund *domain.Refund) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refunds (id, payment_id, amount, status) VALUES ($1, $2, $3, $4)`,
		refund.ID, refund.PaymentID, refund.Amount, refund.Status,
	)
	return err
}

func (r *PaymentRepository) AppendOutbox(ctx context.Context, aggregateID, eventType string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO outbox (id, aggregate_id, event_type, payload) VALUES ($1, $2, $3, $4)`,
		uuid.New().String(), aggregateID, eventType, payloadBytes,
	)
	return err
}

func (r *PaymentRepository) CreatePaymentMethod(ctx context.Context, pm *domain.PaymentMethod) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO payment_methods (id, user_id, provider, token, last_four, brand) VALUES ($1,$2,$3,$4,$5,$6)`,
		pm.ID, pm.UserID, pm.Provider, pm.Token, pm.LastFour, pm.Brand,
	)
	return err
}

func (r *PaymentRepository) CreateHold(ctx context.Context, h *domain.Hold) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO payment_holds (id, payment_id, user_id, amount, currency, status) VALUES ($1,$2,$3,$4,$5,$6)`,
		h.ID, h.PaymentID, h.UserID, h.Amount, h.Currency, h.Status,
	)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx, `UPDATE payments SET hold_id = $1 WHERE id = $2`, h.ID, h.PaymentID)
	return err
}

func (r *PaymentRepository) CaptureHold(ctx context.Context, holdID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE payment_holds SET status = 'CAPTURED', captured_at = NOW() WHERE id = $1`, holdID,
	)
	return err
}

func (r *PaymentRepository) ListByUser(ctx context.Context, userID string, limit int) ([]domain.Payment, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, amount, currency, status, COALESCE(provider,''), idempotency_key, created_at, updated_at
		 FROM payments WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.Provider, &p.IdempotencyKey, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (r *PaymentRepository) CreateRefundWithOutbox(ctx context.Context, refund *domain.Refund, p *domain.Payment, eventType string, payload interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx,
		`INSERT INTO refunds (id, payment_id, amount, status) VALUES ($1, $2, $3, $4)`,
		refund.ID, refund.PaymentID, refund.Amount, refund.Status,
	)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO outbox (id, aggregate_id, event_type, payload) VALUES ($1, $2, $3, $4)`,
		uuid.New().String(), p.ID, eventType, payloadBytes,
	)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *PaymentRepository) RecordAudit(ctx context.Context, actorID, action, resourceType, resourceID string, metadata []byte) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO audit_events (id, actor_id, action, resource_type, resource_id, metadata) VALUES ($1,$2,$3,$4,$5,$6)`,
		uuid.New().String(), actorID, action, resourceType, resourceID, metadata,
	)
	return err
}

func (r *PaymentRepository) CreateDispute(ctx context.Context, d *domain.Dispute) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO disputes (id, payment_id, provider, status, reason) VALUES ($1,$2,$3,$4,$5)`,
		d.ID, d.PaymentID, d.Provider, d.Status, d.Reason,
	)
	return err
}

func (r *PaymentRepository) WithTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
