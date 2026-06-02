package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/omniful/payment-platform/shared/httputil"
	"github.com/omniful/payment-platform/services/analytics-service/internal/usecase"
)

type AnalyticsHandler struct {
	uc *usecase.AnalyticsUsecase
}

func NewAnalyticsHandler(uc *usecase.AnalyticsUsecase) *AnalyticsHandler {
	return &AnalyticsHandler{uc: uc}
}

func (h *AnalyticsHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/analytics/summary", h.GetSummary)
}

func (h *AnalyticsHandler) GetSummary(c *gin.Context) {
	httputil.OK(c, h.uc.Summary())
}
