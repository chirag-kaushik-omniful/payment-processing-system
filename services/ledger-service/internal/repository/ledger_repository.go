package repository

import (
	"context"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omniful/payment-platform/services/ledger-service/internal/domain"
)

type LedgerRepository struct {
	pool *pgxpool.Pool
}

func NewLedgerRepository(pool *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{pool: pool}
}

func (r *LedgerRepository) ExistsForTransaction(ctx context.Context, transactionID string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM ledger_entries WHERE transaction_id = $1)`,
		transactionID,
	).Scan(&exists)
	return exists, err
}

func (r *LedgerRepository) AppendEntry(ctx context.Context, entry domain.LedgerEntry) (*domain.LedgerEntry, error) {
	id := entry.ID
	if id == "" {
		id = uuid.New().String()
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ledger_entries (id, transaction_id, debit_account, credit_account, amount)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, entry.TransactionID, entry.DebitAccount, entry.CreditAccount, entry.Amount,
	)
	if err != nil {
		return nil, err
	}
	entry.ID = id
	return &entry, nil
}

type ExportRecord struct {
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount"`
	Debit     string  `json:"debit_account"`
	Credit    string  `json:"credit_account"`
}

func (r *LedgerRepository) Export(ctx context.Context, since, until time.Time, limit int) ([]ExportRecord, error) {
	query := `SELECT transaction_id, amount, debit_account, credit_account FROM ledger_entries WHERE 1=1`
	args := []interface{}{}
	i := 1
	if !since.IsZero() {
		query += ` AND created_at >= $` + strconv.Itoa(i)
		args = append(args, since)
		i++
	}
	if !until.IsZero() {
		query += ` AND created_at <= $` + strconv.Itoa(i)
		args = append(args, until)
		i++
	}
	query += ` ORDER BY created_at DESC LIMIT $` + strconv.Itoa(i)
	args = append(args, limit)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ExportRecord
	for rows.Next() {
		var rec ExportRecord
		if err := rows.Scan(&rec.PaymentID, &rec.Amount, &rec.Debit, &rec.Credit); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, nil
}

func (r *LedgerRepository) AppendCompensation(ctx context.Context, paymentID string, amount float64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO ledger_entries (id, transaction_id, debit_account, credit_account, amount, entry_type, is_compensation)
		 VALUES ($1, $2, $3, $4, $5, 'REFUND', TRUE)`,
		uuid.New().String(), paymentID, "revenue:payments", "wallet:refund", amount,
	)
	return err
}
