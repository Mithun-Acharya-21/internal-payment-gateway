package middleware

import (
	"github.com/Mithun-Acharya-21/internal-payment-gateway/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered",
					zap.Any("error", r),
					zap.String("path", c.Request.URL.Path),
					zap.String("request_id", c.GetString("request_id")),
				)
				response.InternalServerError(c)
			}
		}()
		c.Next()
	}
}

