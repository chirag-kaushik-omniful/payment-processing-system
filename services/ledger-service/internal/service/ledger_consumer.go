package service

import (
	"context"
	"encoding/json"
	"errors"

	"go.uber.org/zap"

	apperrors "github.com/omniful/payment-platform/shared/errors"
	"github.com/omniful/payment-platform/services/ledger-service/internal/usecase"
)

type LedgerConsumer struct {
	uc  *usecase.LedgerUsecase
	log *zap.Logger
}

func NewLedgerConsumer(uc *usecase.LedgerUsecase, log *zap.Logger) *LedgerConsumer {
	return &LedgerConsumer{uc: uc, log: log}
}

func (c *LedgerConsumer) HandleLedgerEntryCreated(ctx context.Context, _ string, raw json.RawMessage) error {
	var evt struct {
		PaymentID     string  `json:"payment_id"`
		UserID        string  `json:"user_id"`
		Amount        float64 `json:"amount"`
		DebitAccount  string  `json:"debit_account"`
		CreditAccount string  `json:"credit_account"`
	}
	if err := json.Unmarshal(raw, &evt); err != nil {
		return err
	}
	if evt.Amount <= 0 {
		c.log.Warn("ledger entry skipped: missing amount", zap.String("payment_id", evt.PaymentID))
		return nil
	}

	_, err := c.uc.RecordDoubleEntry(ctx, usecase.CreateEntryInput{
		PaymentID:     evt.PaymentID,
		UserID:        evt.UserID,
		Amount:        evt.Amount,
		DebitAccount:  evt.DebitAccount,
		CreditAccount: evt.CreditAccount,
	})
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) && appErr.Code == apperrors.CodeIdempotent {
			return nil
		}
		return err
	}
	return nil
}
