package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/omniful/payment-platform/shared/httputil"
	"github.com/omniful/payment-platform/services/fraud-service/internal/usecase"
)

type FraudHandler struct {
	uc *usecase.FraudUsecase
}

func NewFraudHandler(uc *usecase.FraudUsecase) *FraudHandler {
	return &FraudHandler{uc: uc}
}

func (h *FraudHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/fraud/blocklist", h.AddBlocklist)
}

func (h *FraudHandler) AddBlocklist(c *gin.Context) {
	var req struct {
		Kind string `json:"kind" binding:"required"`
		ID   string `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, err)
		return
	}
	_ = h.uc.AddToBlocklist(c.Request.Context(), req.Kind, req.ID)
	httputil.OK(c, gin.H{"blocked": true})
}
