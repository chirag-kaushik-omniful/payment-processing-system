package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	apperrors "github.com/omniful/payment-platform/shared/errors"
	"github.com/omniful/payment-platform/services/wallet-service/internal/domain"
)

const maxOptimisticRetries = 5

type WalletRepository struct {
	pool *pgxpool.Pool
}

func NewWalletRepository(pool *pgxpool.Pool) *WalletRepository {
	return &WalletRepository{pool: pool}
}

func (r *WalletRepository) Get(ctx context.Context, userID string) (*domain.Wallet, error) {
	w, err := r.getRow(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.NotFound("wallet not found")
		}
		return nil, err
	}
	return w, nil
}

func (r *WalletRepository) EnsureWallet(ctx context.Context, userID string) (*domain.Wallet, error) {
	w, err := r.getRow(ctx, userID)
	if err == nil {
		return w, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	_, err = r.pool.Exec(ctx,
		`INSERT INTO wallets (user_id, balance, version) VALUES ($1, 0, 0) ON CONFLICT (user_id) DO NOTHING`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	return r.getRow(ctx, userID)
}

func (r *WalletRepository) getRow(ctx context.Context, userID string) (*domain.Wallet, error) {
	var w domain.Wallet
	err := r.pool.QueryRow(ctx,
		`SELECT user_id, balance, version, updated_at FROM wallets WHERE user_id = $1`,
		userID,
	).Scan(&w.UserID, &w.Balance, &w.Version, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *WalletRepository) Credit(ctx context.Context, userID string, amount float64) (*domain.Wallet, error) {
	return r.adjustBalance(ctx, userID, amount)
}

func (r *WalletRepository) Debit(ctx context.Context, userID string, amount float64) (*domain.Wallet, error) {
	return r.adjustBalance(ctx, userID, -amount)
}

func (r *WalletRepository) adjustBalance(ctx context.Context, userID string, delta float64) (*domain.Wallet, error) {
	if _, err := r.EnsureWallet(ctx, userID); err != nil {
		return nil, err
	}
	for attempt := 0; attempt < maxOptimisticRetries; attempt++ {
		w, err := r.Get(ctx, userID)
		if err != nil {
			return nil, err
		}
		newBalance := w.Balance + delta
		if newBalance < 0 {
			return nil, apperrors.Validation("insufficient wallet balance")
		}
		tag, err := r.pool.Exec(ctx,
			`UPDATE wallets SET balance = $1, version = version + 1, updated_at = NOW()
			 WHERE user_id = $2 AND version = $3`,
			newBalance, userID, w.Version,
		)
		if err != nil {
			return nil, err
		}
		if tag.RowsAffected() == 1 {
			w.Balance = newBalance
			w.Version++
			return w, nil
		}
	}
	return nil, apperrors.Conflict(fmt.Sprintf("wallet update conflict for user %s", userID))
}

