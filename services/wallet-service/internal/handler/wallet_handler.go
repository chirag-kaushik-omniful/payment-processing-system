package handler

import (
	"github.com/gin-gonic/gin"

	apperrors "github.com/omniful/payment-platform/shared/errors"
	"github.com/omniful/payment-platform/shared/httputil"
	"github.com/omniful/payment-platform/shared/middleware"
	"github.com/omniful/payment-platform/services/wallet-service/internal/usecase"
)

type WalletHandler struct {
	uc *usecase.WalletUsecase
}

func NewWalletHandler(uc *usecase.WalletUsecase) *WalletHandler {
	return &WalletHandler{uc: uc}
}

func (h *WalletHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	r.GET("/wallet", h.GetWallet)
}

func (h *WalletHandler) GetWallet(c *gin.Context) {
	userID := c.GetHeader(middleware.HeaderUserID)
	if userID == "" {
		userID = c.Query("user_id")
	}
	if userID == "" {
		httputil.HandleError(c, apperrors.Validation("user_id required"))
		return
	}
	resp, err := h.uc.GetWallet(c.Request.Context(), userID)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, resp)
}
