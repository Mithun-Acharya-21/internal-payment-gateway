package handler

import (
	"errors"
	"net/http"

	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/domain"
	"github.com/Mithun-Acharya-21/internal-payment-gateway/internal/service"
	"github.com/Mithun-Acharya-21/internal-payment-gateway/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WalletHandler struct {
	svc    *service.WalletService
	logger *zap.Logger
}

func NewWalletHandler(svc *service.WalletService, logger *zap.Logger) *WalletHandler {
	return &WalletHandler{svc: svc, logger: logger}
}

type createWalletRequest struct {
	UserID   string          `json:"user_id" binding:"required,max=255"`
	Currency domain.Currency `json:"currency" binding:"required,oneof=INR USD EUR"`
}

type topupRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

func (h *WalletHandler) CreateWallet(c *gin.Context) {
	var req createWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	w, err := h.svc.CreateWallet(c.Request.Context(), req.UserID, req.Currency)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidAmount) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusCreated, "wallet created", w)
}

func (h *WalletHandler) GetWallet(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid wallet id")
		return
	}
	w, err := h.svc.GetWallet(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrWalletNotFound) {
			response.NotFound(c, "wallet not found")
			return
		}
		response.InternalServerError(c)
		return
	}
	response.Success(c, http.StatusOK, "wallet retrieved", w)
}

func (h *WalletHandler) GetBalance(c *gin.Context) {
	h.GetWallet(c)
}

func (h *WalletHandler) TopUp(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid wallet id")
		return
	}

	var req topupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	w, err := h.svc.TopUp(c.Request.Context(), id, req.Amount)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrWalletNotFound):
			response.NotFound(c, "wallet not found")
		case errors.Is(err, domain.ErrInvalidAmount):
			response.BadRequest(c, "amount must be greater than zero")
		default:
			response.InternalServerError(c)
		}
		return
	}
	response.Success(c, http.StatusOK, "wallet topped up", w)
}

