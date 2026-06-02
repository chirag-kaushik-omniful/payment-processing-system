package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/omniful/payment-platform/shared/httputil"
	"github.com/omniful/payment-platform/services/auth-service/internal/model"
	"github.com/omniful/payment-platform/services/auth-service/internal/usecase"
)

type AuthHandler struct {
	uc *usecase.AuthUsecase
}

func NewAuthHandler(uc *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{uc: uc}
}

func (h *AuthHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/signup", h.Signup)
	r.POST("/login", h.Login)
	r.POST("/refresh", h.Refresh)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req model.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, err)
		return
	}
	resp, err := h.uc.Signup(c.Request.Context(), req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.Created(c, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, err)
		return
	}
	resp, err := h.uc.Login(c.Request.Context(), req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req model.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.HandleError(c, err)
		return
	}
	resp, err := h.uc.Refresh(c.Request.Context(), req)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, resp)
}
