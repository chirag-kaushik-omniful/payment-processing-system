package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperrors "github.com/omniful/payment-platform/shared/errors"
)

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

func HandleError(c *gin.Context, err error) {
	if appErr, ok := apperrors.IsAppError(err); ok {
		c.JSON(appErr.HTTPStatus, gin.H{"code": appErr.Code, "message": appErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    apperrors.CodeInternal,
		"message": "internal server error",
	})
}
