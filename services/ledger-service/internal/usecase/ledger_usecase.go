package usecase

import (
	"context"
	"fmt"
	"time"

	apperrors "github.com/omniful/payment-platform/shared/errors"
	"github.com/omniful/payment-platform/services/ledger-service/internal/domain"
	"github.com/omniful/payment-platform/services/ledger-service/internal/repository"
)

type LedgerUsecase struct {
	repo *repository.LedgerRepository
}

func NewLedgerUsecase(repo *repository.LedgerRepository) *LedgerUsecase {
	return &LedgerUsecase{repo: repo}
}

type CreateEntryInput struct {
	PaymentID     string
	UserID        string
	Amount        float64
	DebitAccount  string
	CreditAccount string
}

func (u *LedgerUsecase) RecordDoubleEntry(ctx context.Context, in CreateEntryInput) (*domain.LedgerEntry, error) {
	if in.PaymentID == "" {
		return nil, apperrors.Validation("payment_id required")
	}
	if in.Amount <= 0 {
		return nil, apperrors.Validation("amount must be positive")
	}

	exists, err := u.repo.ExistsForTransaction(ctx, in.PaymentID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.New(apperrors.CodeIdempotent, "ledger entry already recorded", 200)
	}

	debit := in.DebitAccount
	credit := in.CreditAccount
	if debit == "" {
		if in.UserID != "" {
			debit = fmt.Sprintf("wallet:%s", in.UserID)
		} else {
			debit = "payments:receivable"
		}
	}
	if credit == "" {
		credit = "revenue:payments"
	}

	entry := domain.LedgerEntry{
		TransactionID: in.PaymentID,
		DebitAccount:  debit,
		CreditAccount: credit,
		Amount:        in.Amount,
	}
	return u.repo.AppendEntry(ctx, entry)
}

func (u *LedgerUsecase) Export(ctx context.Context, since, until time.Time, limit int) ([]repository.ExportRecord, error) {
	return u.repo.Export(ctx, since, until, limit)
}

func (u *LedgerUsecase) RecordCompensation(ctx context.Context, paymentID string, amount float64) error {
	return u.repo.AppendCompensation(ctx, paymentID, amount)
}
