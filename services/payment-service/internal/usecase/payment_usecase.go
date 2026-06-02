package usecase

import (
	"context"
	"math"

	"github.com/google/uuid"

	"github.com/omniful/payment-platform/shared/featureflags"
	"github.com/omniful/payment-platform/shared/fx"
	apperrors "github.com/omniful/payment-platform/shared/errors"
	"github.com/omniful/payment-platform/shared/kafka"
	redisutil "github.com/omniful/payment-platform/shared/redis"
	"github.com/omniful/payment-platform/services/payment-service/internal/domain"
	"github.com/omniful/payment-platform/services/payment-service/internal/model"
	"github.com/omniful/payment-platform/services/payment-service/internal/repository"
)

type PaymentUsecase struct {
	repo        *repository.PaymentRepository
	idempotency *redisutil.IdempotencyStore
}

func NewPaymentUsecase(repo *repository.PaymentRepository, idempotency *redisutil.IdempotencyStore) *PaymentUsecase {
	return &PaymentUsecase{repo: repo, idempotency: idempotency}
}

func (u *PaymentUsecase) CreatePaymentMethod(ctx context.Context, userID string, req model.CreatePaymentMethodRequest) (*model.PaymentMethodResponse, error) {
	pm := &domain.PaymentMethod{
		ID:       uuid.New().String(),
		UserID:   userID,
		Provider: req.Provider,
		Token:    "tok_" + uuid.New().String()[:12],
		LastFour: req.LastFour,
		Brand:    req.Brand,
	}
	if err := u.repo.CreatePaymentMethod(ctx, pm); err != nil {
		return nil, apperrors.Internal("create payment method failed", err)
	}
	_ = u.repo.RecordAudit(ctx, userID, "payment_method.created", "payment_method", pm.ID, nil)
	return &model.PaymentMethodResponse{ID: pm.ID, Provider: pm.Provider, LastFour: pm.LastFour, Brand: pm.Brand}, nil
}

func (u *PaymentUsecase) CreatePayment(ctx context.Context, userID string, req model.CreatePaymentRequest, idempotencyKey string) (*model.PaymentResponse, error) {
	if idempotencyKey != "" {
		var cached model.PaymentResponse
		if found, _ := u.idempotency.Get(ctx, idempotencyKey, &cached); found {
			return &cached, nil
		}
		if existing, err := u.repo.FindByIdempotencyKey(ctx, idempotencyKey); err == nil {
			return toResponse(existing), nil
		}
	}

	provider := req.Provider
	if provider == "" {
		provider = "stripe"
	}
	if !featureflags.ProviderEnabled(provider) {
		return nil, apperrors.Validation("provider disabled by feature flag")
	}

	amount := req.Amount
	fxRate := 1.0
	baseCurrency := "USD"
	if req.Currency != "" && req.Currency != baseCurrency {
		converted, rate, _ := fx.Convert(amount, req.Currency, baseCurrency)
		amount = converted
		fxRate = rate
	}

	status := domain.StatusCreated
	if featureflags.HoldsEnabled() && req.CaptureMode == "manual" {
		status = domain.StatusAuthorized
	}

	payment := &domain.Payment{
		ID:              uuid.New().String(),
		UserID:          userID,
		Amount:          amount,
		AmountCents:     int64(math.Round(amount * 100)),
		Currency:        req.Currency,
		BaseCurrency:    baseCurrency,
		FXRate:          fxRate,
		Status:          status,
		Provider:        provider,
		IdempotencyKey:  idempotencyKey,
		PaymentMethodID: req.PaymentMethodID,
	}

	eventPayload := map[string]interface{}{
		"payment_id": payment.ID,
		"user_id":    payment.UserID,
		"amount":     payment.Amount,
		"currency":   payment.Currency,
		"provider":   payment.Provider,
	}

	if err := u.repo.CreateWithOutbox(ctx, payment, kafka.TopicPaymentCreated, eventPayload); err != nil {
		return nil, apperrors.Internal("create payment failed", err)
	}

	if featureflags.HoldsEnabled() && status == domain.StatusAuthorized {
		hold := &domain.Hold{
			ID:        uuid.New().String(),
			PaymentID: payment.ID,
			UserID:    userID,
			Amount:    amount,
			Currency:  req.Currency,
			Status:    "HELD",
		}
		_ = u.repo.CreateHold(ctx, hold)
		payment.HoldID = hold.ID
	}

	_ = u.repo.RecordAudit(ctx, userID, "payment.created", "payment", payment.ID, nil)
	resp := toResponse(payment)
	if idempotencyKey != "" {
		_ = u.idempotency.Set(ctx, idempotencyKey, resp)
	}
	return resp, nil
}

func (u *PaymentUsecase) CaptureHold(ctx context.Context, paymentID, userID string) (*model.PaymentResponse, error) {
	p, err := u.repo.FindByID(ctx, paymentID)
	if err != nil || p.UserID != userID {
		return nil, apperrors.NotFound("payment not found")
	}
	if p.Status != domain.StatusAuthorized {
		return nil, apperrors.Validation("payment is not in authorized state")
	}
	_ = u.repo.CaptureHold(ctx, p.HoldID)
	eventPayload := map[string]interface{}{"payment_id": p.ID, "user_id": p.UserID, "amount": p.Amount}
	if err := u.repo.AppendOutbox(ctx, p.ID, kafka.TopicPaymentHoldCaptured, eventPayload); err != nil {
		return nil, apperrors.Internal("capture failed", err)
	}
	return toResponse(p), nil
}

func (u *PaymentUsecase) ListPayments(ctx context.Context, userID string, limit int) ([]model.PaymentResponse, error) {
	payments, err := u.repo.ListByUser(ctx, userID, limit)
	if err != nil {
		return nil, apperrors.Internal("list payments failed", err)
	}
	var out []model.PaymentResponse
	for _, p := range payments {
		out = append(out, *toResponse(&p))
	}
	return out, nil
}

func (u *PaymentUsecase) GetPayment(ctx context.Context, id, userID string) (*model.PaymentResponse, error) {
	p, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.NotFound("payment not found")
	}
	if userID != "" && p.UserID != userID {
		return nil, apperrors.NotFound("payment not found")
	}
	return toResponse(p), nil
}

func (u *PaymentUsecase) RefundPayment(ctx context.Context, paymentID, userID string, req model.RefundRequest) (*model.PaymentResponse, error) {
	p, err := u.repo.FindByID(ctx, paymentID)
	if err != nil || p.UserID != userID {
		return nil, apperrors.NotFound("payment not found")
	}
	if p.Status != domain.StatusSuccess && p.Status != domain.StatusCaptured {
		return nil, apperrors.Validation("only successful payments can be refunded")
	}
	refund := &domain.Refund{
		ID:        uuid.New().String(),
		PaymentID: paymentID,
		Amount:    req.Amount,
		Status:    "CREATED",
	}

	if featureflags.RefundSagaEnabled() {
		if err := u.repo.CreateRefundWithOutbox(ctx, refund, p, kafka.TopicRefundRequested, map[string]interface{}{
			"refund_id":  refund.ID,
			"payment_id": paymentID,
			"amount":     req.Amount,
			"provider":   p.Provider,
		}); err != nil {
			return nil, apperrors.Internal("refund failed", err)
		}
	} else {
		if err := u.repo.CreateRefund(ctx, refund); err != nil {
			return nil, apperrors.Internal("refund failed", err)
		}
		_ = u.repo.UpdateStatus(ctx, paymentID, domain.StatusRefunded)
		p.Status = domain.StatusRefunded
	}
	_ = u.repo.RecordAudit(ctx, userID, "payment.refund_requested", "payment", paymentID, nil)
	return toResponse(p), nil
}

func (u *PaymentUsecase) OpenDispute(ctx context.Context, paymentID, provider, reason string) error {
	if _, err := u.repo.FindByID(ctx, paymentID); err != nil {
		return apperrors.NotFound("payment not found")
	}
	d := &domain.Dispute{ID: uuid.New().String(), PaymentID: paymentID, Provider: provider, Status: "OPEN", Reason: reason}
	if err := u.repo.CreateDispute(ctx, d); err != nil {
		return apperrors.Internal("dispute failed", err)
	}
	_ = u.repo.UpdateStatus(ctx, paymentID, domain.StatusDisputed)
	return nil
}

func toResponse(p *domain.Payment) *model.PaymentResponse {
	return &model.PaymentResponse{
		ID:       p.ID,
		UserID:   p.UserID,
		Amount:   p.Amount,
		Currency: p.Currency,
		Status:   string(p.Status),
		Provider: p.Provider,
	}
}
