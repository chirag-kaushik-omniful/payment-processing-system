package handler

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	goredis "github.com/redis/go-redis/v9"

	"github.com/omniful/payment-platform/shared/auth"
	"github.com/omniful/payment-platform/shared/circuitbreaker"
	"github.com/omniful/payment-platform/shared/middleware"
	redisutil "github.com/omniful/payment-platform/shared/redis"
	"github.com/omniful/payment-platform/services/api-gateway/internal/proxy"
)

type GatewayHandler struct {
	proxy          *proxy.ServiceProxy
	jwtMgr         *auth.JWTManager
	rateLimiter    *redisutil.RateLimiter
	paymentBreaker *circuitbreaker.Breaker
	paymentURL     string
	walletURL      string
	authURL        string
	adminURL       string
}

func NewGatewayHandler(jwtMgr *auth.JWTManager, redisClient *goredis.Client) *GatewayHandler {
	return &GatewayHandler{
		proxy:          proxy.NewServiceProxy(),
		jwtMgr:         jwtMgr,
		rateLimiter:    redisutil.NewRateLimiter(redisClient, 100, time.Minute),
		paymentBreaker: circuitbreaker.New(5, 30*time.Second),
		paymentURL:     getEnv("PAYMENT_SERVICE_URL", "http://localhost:8002"),
		walletURL:      getEnv("WALLET_SERVICE_URL", "http://localhost:8004"),
		authURL:        getEnv("AUTH_SERVICE_URL", "http://localhost:8001"),
		adminURL:       getEnv("ADMIN_SERVICE_URL", "http://localhost:8011"),
	}
}

func (h *GatewayHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/signup", h.forwardAuth("/signup"))
		authGroup.POST("/login", h.forwardAuth("/login"))
		authGroup.POST("/refresh", h.forwardAuth("/refresh"))
	}

	api := r.Group("/")
	api.Use(middleware.JWTAuth(h.jwtMgr))
	api.Use(h.rateLimit())
	{
		api.POST("/payment-methods", h.forwardPayment("/payment-methods"))
		api.POST("/payments", h.forwardPaymentWithBreaker("/payments"))
		api.GET("/payments", h.forwardPaymentWithBreaker("/payments"))
		api.GET("/payments/:id", h.forwardPaymentWithBreaker("/payments/:id"))
		api.POST("/payments/:id/capture", h.forwardPaymentWithBreaker("/payments/:id/capture"))
		api.POST("/refunds", middleware.RequireRole("admin"), h.forwardPayment("/refunds"))
		api.GET("/wallet", h.forwardWallet("/wallet"))
	}

	merchant := r.Group("/merchant")
	merchant.Use(middleware.JWTAuth(h.jwtMgr))
	{
		merchant.GET("/payments", h.forwardPayment("/payments"))
	}

	admin := r.Group("/admin")
	admin.Use(middleware.JWTAuth(h.jwtMgr))
	admin.Use(middleware.RequireRole("admin"))
	{
		admin.GET("/payments", h.forwardAdmin("/admin/payments"))
		admin.GET("/audit", h.forwardAdmin("/admin/audit"))
		admin.POST("/payments/:id/replay-saga", h.forwardAdmin("/admin/payments/:id/replay-saga"))
	}
}

func (h *GatewayHandler) forwardPaymentWithBreaker(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !h.paymentBreaker.Allow() {
			c.JSON(503, gin.H{"code": "SERVICE_UNAVAILABLE", "message": "payment service circuit open"})
			return
		}
		h.forwardPayment(path)(c)
		if c.Writer.Status() < 500 {
			h.paymentBreaker.RecordSuccess()
		} else {
			h.paymentBreaker.RecordFailure()
		}
	}
}

func (h *GatewayHandler) rateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString(middleware.ContextUserID)
		if userID == "" {
			userID = c.ClientIP()
		}
		allowed, err := h.rateLimiter.Allow(c.Request.Context(), userID)
		if err != nil || !allowed {
			middleware.AbortRateLimit(c)
			return
		}
		c.Next()
	}
}

func (h *GatewayHandler) forwardAuth(path string) gin.HandlerFunc {
	return func(c *gin.Context) { h.proxy.Forward(c, h.authURL, path) }
}

func (h *GatewayHandler) forwardPayment(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := path
		if id := c.Param("id"); id != "" {
			if path == "/payments/:id/capture" {
				p = "/payments/" + id + "/capture"
			} else {
				p = "/payments/" + id
			}
		}
		h.proxy.Forward(c, h.paymentURL, p)
	}
}

func (h *GatewayHandler) forwardWallet(path string) gin.HandlerFunc {
	return func(c *gin.Context) { h.proxy.Forward(c, h.walletURL, path) }
}

func (h *GatewayHandler) forwardAdmin(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		p := path
		if id := c.Param("id"); id != "" {
			p = "/admin/payments/" + id + "/replay-saga"
		}
		h.proxy.Forward(c, h.adminURL, p)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
