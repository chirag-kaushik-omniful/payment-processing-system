package service

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/omniful/payment-platform/services/ledger-service/internal/usecase"
)

type CompensationConsumer struct {
	uc  *usecase.LedgerUsecase
	log *zap.Logger
}

func NewCompensationConsumer(uc *usecase.LedgerUsecase, log *zap.Logger) *CompensationConsumer {
	return &CompensationConsumer{uc: uc, log: log}
}

func (c *CompensationConsumer) HandleCompensation(ctx context.Context, _ string, raw json.RawMessage) error {
	var evt struct {
		PaymentID string  `json:"payment_id"`
		Amount    float64 `json:"amount"`
	}
	if err := json.Unmarshal(raw, &evt); err != nil {
		return err
	}
	if err := c.uc.RecordCompensation(ctx, evt.PaymentID, evt.Amount); err != nil {
		c.log.Warn("compensation failed", zap.Error(err))
		return err
	}
	return nil
}
