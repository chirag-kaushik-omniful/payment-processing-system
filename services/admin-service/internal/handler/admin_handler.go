package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/omniful/payment-platform/shared/auth"
	"github.com/omniful/payment-platform/shared/httputil"
	"github.com/omniful/payment-platform/shared/middleware"
	"github.com/omniful/payment-platform/services/admin-service/internal/usecase"
)

type AdminHandler struct {
	uc     *usecase.AdminUsecase
	jwtMgr *auth.JWTManager
}

func NewAdminHandler(uc *usecase.AdminUsecase, jwtMgr *auth.JWTManager) *AdminHandler {
	return &AdminHandler{uc: uc, jwtMgr: jwtMgr}
}

func (h *AdminHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	admin := r.Group("/admin")
	admin.Use(middleware.JWTAuth(h.jwtMgr))
	admin.Use(middleware.RequireRole("admin"))
	{
		admin.GET("/payments", h.ListPayments)
		admin.POST("/payments/:id/replay-saga", h.ReplaySaga)
		admin.GET("/audit", h.ListAudit)
	}
}

func (h *AdminHandler) ListPayments(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	payments, err := h.uc.ListPayments(c.Request.Context(), limit)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, gin.H{"payments": payments})
}

func (h *AdminHandler) ReplaySaga(c *gin.Context) {
	if err := h.uc.ReplaySaga(c.Request.Context(), c.Param("id")); err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, gin.H{"status": "published", "topic": "payment.created"})
}

func (h *AdminHandler) ListAudit(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	events, err := h.uc.ListAuditEvents(c.Request.Context(), limit)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, gin.H{"events": events})
}
