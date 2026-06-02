package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/omniful/payment-platform/shared/auth"
	apperrors "github.com/omniful/payment-platform/shared/errors"
	"github.com/omniful/payment-platform/shared/logger"
	"github.com/omniful/payment-platform/shared/metrics"
)

const (
	HeaderCorrelationID = "X-Correlation-ID"
	HeaderUserID        = "X-User-ID"
	ContextUserID       = "user_id"
	ContextClaims       = "claims"
)

func CorrelationID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(HeaderCorrelationID)
		if id == "" {
			id = uuid.New().String()
		}
		c.Set(HeaderCorrelationID, id)
		c.Header(HeaderCorrelationID, id)
		ctx := logger.WithFields(c.Request.Context(), logger.FieldTraceID(id))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		metrics.HTTPRequests.WithLabelValues(
			c.GetString("service_name"),
			c.Request.Method,
			c.FullPath(),
			http.StatusText(c.Writer.Status()),
		).Inc()
		logger.FromContext(c.Request.Context()).Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

func JWTAuth(jwtMgr *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			abortError(c, apperrors.Unauthorized("missing or invalid authorization header"))
			return
		}
		token := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtMgr.Validate(token)
		if err != nil {
			abortError(c, apperrors.Unauthorized("invalid token"))
			return
		}
		c.Set(ContextClaims, claims)
		c.Set(ContextUserID, claims.UserID)
		c.Header(HeaderUserID, claims.UserID)
		c.Next()
	}
}

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get(ContextClaims)
		if !exists {
			abortError(c, apperrors.New(apperrors.CodeForbidden, "forbidden", http.StatusForbidden))
			return
		}
		claims, ok := val.(*auth.Claims)
		if !ok {
			abortError(c, apperrors.New(apperrors.CodeForbidden, "forbidden", http.StatusForbidden))
			return
		}
		for _, r := range claims.Roles {
			if r == role {
				c.Next()
				return
			}
		}
		abortError(c, apperrors.New(apperrors.CodeForbidden, "insufficient permissions", http.StatusForbidden))
	}
}

func abortError(c *gin.Context, err *apperrors.AppError) {
	c.AbortWithStatusJSON(err.HTTPStatus, gin.H{
		"code":    err.Code,
		"message": err.Message,
	})
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}
		err := c.Errors.Last().Err
		if appErr, ok := apperrors.IsAppError(err); ok {
			c.JSON(appErr.HTTPStatus, gin.H{"code": appErr.Code, "message": appErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    apperrors.CodeInternal,
			"message": "internal server error",
		})
	}
}

func ServiceName(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("service_name", name)
		c.Next()
	}
}

func AbortRateLimit(c *gin.Context) {
	abortError(c, apperrors.RateLimit("rate limit exceeded"))
}
