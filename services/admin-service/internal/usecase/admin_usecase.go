package usecase

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	apperrors "github.com/omniful/payment-platform/shared/errors"
	"github.com/omniful/payment-platform/shared/kafka"
	"github.com/omniful/payment-platform/services/admin-service/internal/domain"
	"github.com/omniful/payment-platform/services/admin-service/internal/repository"
)

type AdminUsecase struct {
	repo     *repository.AdminRepository
	producer *kafka.Producer
}

func NewAdminUsecase(repo *repository.AdminRepository, producer *kafka.Producer) *AdminUsecase {
	return &AdminUsecase{repo: repo, producer: producer}
}

func (u *AdminUsecase) ListPayments(ctx context.Context, limit int) ([]domain.Payment, error) {
	payments, err := u.repo.ListPayments(ctx, limit)
	if err != nil {
		return nil, apperrors.Internal("list payments failed", err)
	}
	return payments, nil
}

func (u *AdminUsecase) ListAuditEvents(ctx context.Context, limit int) ([]domain.AuditEvent, error) {
	events, err := u.repo.ListAuditEvents(ctx, limit)
	if err != nil {
		return nil, apperrors.Internal("list audit events failed", err)
	}
	return events, nil
}

func (u *AdminUsecase) ReplaySaga(ctx context.Context, paymentID string) error {
	payment, err := u.repo.FindPaymentByID(ctx, paymentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.NotFound("payment not found")
		}
		return apperrors.Internal("find payment failed", err)
	}

	payload := map[string]interface{}{
		"payment_id": payment.ID,
		"user_id":    payment.UserID,
		"amount":     payment.Amount,
		"currency":   payment.Currency,
		"provider":   payment.Provider,
		"replay":     true,
	}
	if err := u.producer.Publish(ctx, kafka.TopicPaymentCreated, payment.ID, payload); err != nil {
		return apperrors.Internal("publish payment.created failed", err)
	}
	return nil
}
