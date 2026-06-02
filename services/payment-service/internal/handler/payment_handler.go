package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/omniful/payment-platform/shared/httputil"
	"github.com/omniful/payment-platform/shared/middleware"
	"github.com/omniful/payment-platform/services/payment-service/internal/model"
	"github.com/omniful/payment-platform/services/payment-service/internal/usecase"
)

type PaymentHandler struct {
	uc *usecase.PaymentUsecase
}

func NewPaymentHandler(uc *usecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{uc: uc}
}

func (h *PaymentHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.POST("/payment-methods", h.CreatePaymentMethod)
	r.POST("/payments", h.CreatePayment)
	r.GET("/payments", h.ListPayments)
	r.GET("/payments/:id", h.GetPayment)
	r.POST("/payments/:id/capture", h.CaptureHold)
	r.POST("/refunds", h.RefundPayment)
	r.POST("/disputes", h.OpenDispute)
}

func (h *PaymentHandler) CreatePaymentMethod(c *gin.Context) {
	userID := c.GetHeader(middleware.HeaderUserID)
	var req model.CreatePaymentMethodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, err)
		return
	}
	resp, err := h.uc.CreatePaymentMethod(c.Request.Context(), userID, req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.Created(c, resp)
}

func (h *PaymentHandler) ListPayments(c *gin.Context) {
	userID := c.GetHeader(middleware.HeaderUserID)
	resp, err := h.uc.ListPayments(c.Request.Context(), userID, 50)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, resp)
}

func (h *PaymentHandler) CaptureHold(c *gin.Context) {
	userID := c.GetHeader(middleware.HeaderUserID)
	resp, err := h.uc.CaptureHold(c.Request.Context(), c.Param("id"), userID)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, resp)
}

func (h *PaymentHandler) OpenDispute(c *gin.Context) {
	var req struct {
		PaymentID string `json:"payment_id" binding:"required"`
		Provider  string `json:"provider" binding:"required"`
		Reason    string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, err)
		return
	}
	if err := h.uc.OpenDispute(c.Request.Context(), req.PaymentID, req.Provider, req.Reason); err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, gin.H{"status": "dispute_opened"})
}

func (h *PaymentHandler) CreatePayment(c *gin.Context) {
	userID := c.GetHeader(middleware.HeaderUserID)
	if userID == "" {
		userID = c.Query("user_id")
	}
	var req model.CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, err)
		return
	}
	idemKey := c.GetHeader("Idempotency-Key")
	resp, err := h.uc.CreatePayment(c.Request.Context(), userID, req, idemKey)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.Created(c, resp)
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	userID := c.GetHeader(middleware.HeaderUserID)
	resp, err := h.uc.GetPayment(c.Request.Context(), c.Param("id"), userID)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, resp)
}

func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	userID := c.GetHeader(middleware.HeaderUserID)
	var req struct {
		PaymentID string  `json:"payment_id" binding:"required"`
		model.RefundRequest
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, err)
		return
	}
	resp, err := h.uc.RefundPayment(c.Request.Context(), req.PaymentID, userID, req.RefundRequest)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, resp)
}
