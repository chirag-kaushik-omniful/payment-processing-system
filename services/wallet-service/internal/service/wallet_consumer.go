package service

import (
	"context"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/omniful/payment-platform/services/wallet-service/internal/usecase"
)

type WalletConsumer struct {
	uc  *usecase.WalletUsecase
	log *zap.Logger
}

func NewWalletConsumer(uc *usecase.WalletUsecase, log *zap.Logger) *WalletConsumer {
	return &WalletConsumer{uc: uc, log: log}
}

func (c *WalletConsumer) HandleWalletCredited(ctx context.Context, _ string, raw json.RawMessage) error {
	var evt struct {
		UserID string  `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := json.Unmarshal(raw, &evt); err != nil {
		return err
	}
	if evt.UserID == "" || evt.Amount <= 0 {
		return nil
	}
	return c.uc.Credit(ctx, evt.UserID, evt.Amount)
}

func (c *WalletConsumer) HandleWalletDebited(ctx context.Context, _ string, raw json.RawMessage) error {
	var evt struct {
		UserID string  `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := json.Unmarshal(raw, &evt); err != nil {
		return err
	}
	if evt.UserID == "" || evt.Amount <= 0 {
		return nil
	}
	return c.uc.Debit(ctx, evt.UserID, evt.Amount)
}
