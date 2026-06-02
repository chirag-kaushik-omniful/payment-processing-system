package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/omniful/payment-platform/shared/httputil"
	"github.com/omniful/payment-platform/services/reconciliation-service/internal/model"
	"github.com/omniful/payment-platform/services/reconciliation-service/internal/usecase"
)

type ReconciliationHandler struct {
	uc *usecase.ReconciliationUsecase
}

func NewReconciliationHandler(uc *usecase.ReconciliationUsecase) *ReconciliationHandler {
	return &ReconciliationHandler{uc: uc}
}

func (h *ReconciliationHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.POST("/reconcile", h.Reconcile)
	r.POST("/reconcile/auto", h.AutoReconcile)
}

func (h *ReconciliationHandler) AutoReconcile(c *gin.Context) {
	var req struct {
		Settlements []model.SettlementRecord `json:"settlements" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, err)
		return
	}
	resp, err := h.uc.AutoReconcile(req.Settlements)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, resp)
}

func (h *ReconciliationHandler) Reconcile(c *gin.Context) {
	var req model.ReconcileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, err)
		return
	}
	resp := h.uc.Reconcile(req)
	httputil.OK(c, resp)
}
