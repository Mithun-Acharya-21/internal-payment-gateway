package response

import (
	"errors"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)
type APIResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: requestID(c),
	})
}

func BadRequest(c *gin.Context, message string) {
	errorResponse(c, http.StatusBadRequest, message, nil)
}

func NotFound(c *gin.Context, message string) {
	errorResponse(c, http.StatusNotFound, message, nil)
}

func Unauthorized(c *gin.Context) {
	errorResponse(c, http.StatusUnauthorized, "authentication required", nil)
}

func UnprocessableEntity(c *gin.Context, message string) {
	errorResponse(c, http.StatusUnprocessableEntity, message, nil)
}

func InternalServerError(c *gin.Context) {
	errorResponse(c, http.StatusInternalServerError, "an unexpected error occurred", nil)
}

func ValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if ok := errors.As(err, &ve); ok {
		fields := make(map[string]string, len(ve))
		for _, fe := range ve {
			fields[fe.Field()] = fieldMessage(fe)
		}
		errorResponse(c, http.StatusBadRequest, "validation failed", fields)
		return
	}
	BadRequest(c, err.Error())
}

func errorResponse(c *gin.Context, code int, message string, detail interface{}) {
	c.AbortWithStatusJSON(code, APIResponse{
		Success:   false,
		Message:   message,
		Error:     detail,
		RequestID: requestID(c),
	})
}

func requestID(c *gin.Context) string {
	if id, ok := c.Get("request_id"); ok {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "uuid":
		return "must be a valid UUID"
	case "gt":
		return "must be greater than " + fe.Param()
	case "max":
		return "must not exceed " + fe.Param() + " characters"
	case "oneof":
		return "must be one of: " + fe.Param()
	default:
		return "invalid value"
	}
}

