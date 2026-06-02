package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/omniful/payment-platform/shared/httputil"
	"github.com/omniful/payment-platform/services/subscription-service/internal/model"
	"github.com/omniful/payment-platform/services/subscription-service/internal/usecase"
)

type SubscriptionHandler struct {
	uc *usecase.SubscriptionUsecase
}

func NewSubscriptionHandler(uc *usecase.SubscriptionUsecase) *SubscriptionHandler {
	return &SubscriptionHandler{uc: uc}
}

func (h *SubscriptionHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.POST("/subscriptions", h.CreateSubscription)
	r.GET("/subscriptions/:id", h.GetSubscription)
}

func (h *SubscriptionHandler) CreateSubscription(c *gin.Context) {
	var req model.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, err)
		return
	}
	resp, err := h.uc.Create(c.Request.Context(), req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.Created(c, resp)
}

func (h *SubscriptionHandler) GetSubscription(c *gin.Context) {
	resp, err := h.uc.GetByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, resp)
}
