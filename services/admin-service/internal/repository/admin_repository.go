package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omniful/payment-platform/services/admin-service/internal/domain"
)

type AdminRepository struct {
	pool *pgxpool.Pool
}

func NewAdminRepository(pool *pgxpool.Pool) *AdminRepository {
	return &AdminRepository{pool: pool}
}

func (r *AdminRepository) ListPayments(ctx context.Context, limit int) ([]domain.Payment, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, amount, currency, status, COALESCE(provider,''), created_at, updated_at
		 FROM payments ORDER BY created_at DESC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.Provider, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, rows.Err()
}

func (r *AdminRepository) FindPaymentByID(ctx context.Context, id string) (*domain.Payment, error) {
	var p domain.Payment
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, amount, currency, status, COALESCE(provider,''), created_at, updated_at
		 FROM payments WHERE id = $1`, id,
	).Scan(&p.ID, &p.UserID, &p.Amount, &p.Currency, &p.Status, &p.Provider, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *AdminRepository) ListAuditEvents(ctx context.Context, limit int) ([]domain.AuditEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx,
		`SELECT id, COALESCE(actor_id,''), action, resource_type, resource_id, COALESCE(metadata::text,'{}'), created_at
		 FROM audit_events ORDER BY created_at DESC LIMIT $1`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.AuditEvent
	for rows.Next() {
		var e domain.AuditEvent
		var metadata string
		if err := rows.Scan(&e.ID, &e.ActorID, &e.Action, &e.ResourceType, &e.ResourceID, &metadata, &e.CreatedAt); err != nil {
			return nil, err
		}
		e.Metadata = []byte(metadata)
		events = append(events, e)
	}
	return events, rows.Err()
}
