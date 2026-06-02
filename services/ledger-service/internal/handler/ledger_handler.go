package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/omniful/payment-platform/shared/httputil"
	"github.com/omniful/payment-platform/services/ledger-service/internal/usecase"
)

type LedgerHandler struct {
	uc *usecase.LedgerUsecase
}

func NewLedgerHandler(uc *usecase.LedgerUsecase) *LedgerHandler {
	return &LedgerHandler{uc: uc}
}

func (h *LedgerHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/ledger/export", h.Export)
}

func (h *LedgerHandler) Export(c *gin.Context) {
	since := c.Query("since")
	until := c.Query("until")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1000"))
	var sinceT, untilT time.Time
	if since != "" {
		sinceT, _ = time.Parse(time.RFC3339, since)
	}
	if until != "" {
		untilT, _ = time.Parse(time.RFC3339, until)
	}
	records, err := h.uc.Export(c.Request.Context(), sinceT, untilT, limit)
	if err != nil {
		httputil.HandleError(c, err)
		return
	}
	httputil.OK(c, gin.H{"ledger": records})
}
