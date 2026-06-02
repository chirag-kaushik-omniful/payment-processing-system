package handler

import (
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"

	"github.com/omniful/payment-platform/shared/httputil"
	"github.com/omniful/payment-platform/services/webhook-service/internal/model"
	"github.com/omniful/payment-platform/services/webhook-service/internal/usecase"
)

type WebhookHandler struct {
	uc *usecase.WebhookUsecase
}

func NewWebhookHandler(uc *usecase.WebhookUsecase) *WebhookHandler {
	return &WebhookHandler{uc: uc}
}

func (h *WebhookHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.POST("/webhooks/:provider", h.ReceiveWebhook)
}

func (h *WebhookHandler) ReceiveWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}

	provider := c.Param("provider")
	signature := c.GetHeader("X-Webhook-Signature")
	if signature == "" {
		signature = c.GetHeader("X-Signature")
	}
	if err := h.uc.VerifySignature(provider, signature, body); err != nil {
		httputil.HandleError(c, err)
		return
	}

	var payload model.WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		httputil.HandleError(c, err)
		return
	}
	if payload.EventID == "" {
		payload.EventID = c.GetHeader("X-Event-Id")
	}

	resp, err := h.uc.Process(c.Request.Context(), provider, payload)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, resp)
}
