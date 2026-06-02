package usecase

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/omniful/payment-platform/services/wallet-service/internal/model"
	"github.com/omniful/payment-platform/services/wallet-service/internal/repository"
)

const balanceCacheTTL = 5 * time.Minute

type WalletUsecase struct {
	repo  *repository.WalletRepository
	redis *redis.Client
}

func NewWalletUsecase(repo *repository.WalletRepository, redisClient *redis.Client) *WalletUsecase {
	return &WalletUsecase{repo: repo, redis: redisClient}
}

func balanceCacheKey(userID string) string {
	return fmt.Sprintf("wallet:balance:%s", userID)
}

func (u *WalletUsecase) GetWallet(ctx context.Context, userID string) (*model.WalletResponse, error) {
	if cached, err := u.redis.Get(ctx, balanceCacheKey(userID)).Result(); err == nil {
		if balance, err := strconv.ParseFloat(cached, 64); err == nil {
			w, err := u.repo.Get(ctx, userID)
			if err == nil {
				return &model.WalletResponse{UserID: userID, Balance: balance, Version: w.Version}, nil
			}
		}
	}

	w, err := u.repo.EnsureWallet(ctx, userID)
	if err != nil {
		return nil, err
	}
	_ = u.setCache(ctx, userID, w.Balance)
	return &model.WalletResponse{UserID: w.UserID, Balance: w.Balance, Version: w.Version}, nil
}

func (u *WalletUsecase) Credit(ctx context.Context, userID string, amount float64) error {
	w, err := u.repo.Credit(ctx, userID, amount)
	if err != nil {
		return err
	}
	return u.setCache(ctx, userID, w.Balance)
}

func (u *WalletUsecase) Debit(ctx context.Context, userID string, amount float64) error {
	w, err := u.repo.Debit(ctx, userID, amount)
	if err != nil {
		return err
	}
	return u.setCache(ctx, userID, w.Balance)
}

func (u *WalletUsecase) setCache(ctx context.Context, userID string, balance float64) error {
	return u.redis.Set(ctx, balanceCacheKey(userID), fmt.Sprintf("%.2f", balance), balanceCacheTTL).Err()
}

func (u *WalletUsecase) InvalidateCache(ctx context.Context, userID string) {
	_ = u.redis.Del(ctx, balanceCacheKey(userID)).Err()
}
