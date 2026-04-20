package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/domain"
	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/service"
	"github.com/Mithun-Acharya-21/internal-payment-gateway/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	svc    *service.PaymentService
	logger *zap.Logger
}

func NewPaymentHandler(svc *service.PaymentService, logger *zap.Logger) *PaymentHandler {
	return &PaymentHandler{svc: svc, logger: logger}
}

type initiatePaymentRequest struct {
	WalletID       string          `json:"wallet_id" binding:"required,uuid"`
	Amount         int64           `json:"amount" binding:"required,gt=0"`
	Currency       domain.Currency `json:"currency" binding:"required,oneof=INR USD EUR"`
	Description    string          `json:"description" binding:"required,max=255"`
	IdempotencyKey string          `json:"idempotency_key" binding:"max=64"`
}

func (h *PaymentHandler) InitiatePayment(c *gin.Context) {
	var req initiatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	walletID, _ := uuid.Parse(req.WalletID)
	idempotencyKey := c.GetHeader("X-Idempotency-Key")
	if idempotencyKey == "" {
		idempotencyKey = req.IdempotencyKey
	}

	tx, err := h.svc.InitiatePayment(c.Request.Context(), service.InitiatePaymentInput{
		WalletID:       walletID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Description:    req.Description,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	statusCode := http.StatusCreated
	if tx.Status == domain.StatusFailed {
		statusCode = http.StatusUnprocessableEntity
	}
	response.Success(c, statusCode, "payment processed", tx)
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid transaction id")
		return
	}

	tx, err := h.svc.GetPayment(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "transaction retrieved", tx)
}

func (h *PaymentHandler) ListPayments(c *gin.Context) {
	walletID, err := uuid.Parse(c.Query("wallet_id"))
	if err != nil {
		response.BadRequest(c, "wallet_id query param is required and must be a valid UUID")
		return
	}

	limit := queryInt(c, "limit", 20)
	offset := queryInt(c, "offset", 0)

	txs, total, err := h.svc.ListPayments(c.Request.Context(), walletID, limit, offset)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	response.Success(c, http.StatusOK, "transactions retrieved", gin.H{
		"transactions": txs,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}

func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid transaction id")
		return
	}

	refund, err := h.svc.RefundPayment(c.Request.Context(), id)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "refund processed", refund)
}

func (h *PaymentHandler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrTransactionNotFound):
		response.NotFound(c, "transaction not found")
	case errors.Is(err, domain.ErrWalletNotFound):
		response.NotFound(c, "wallet not found")
	case errors.Is(err, domain.ErrInsufficientBalance):
		response.UnprocessableEntity(c, "insufficient wallet balance")
	case errors.Is(err, domain.ErrTransactionNotRefundable):
		response.BadRequest(c, "transaction cannot be refunded")
	case errors.Is(err, domain.ErrInvalidAmount):
		response.BadRequest(c, "amount must be greater than zero")
	default:
		h.logger.Error("unexpected service error", zap.Error(err))
		response.InternalServerError(c)
	}
}

func queryInt(c *gin.Context, key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	var i int
	if _, err := fmt.Sscanf(val, "%d", &i); err != nil {
		return defaultVal
	}
	return i
}

