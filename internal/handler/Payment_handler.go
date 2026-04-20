package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"STRIPE/internal/service"
	"STRIPE/pkg/response"
)

type PaymentHandler struct {
	service *service.PaymentService
}

func NewPaymentHandler(service *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

type CreatePaymentRequest struct {
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")

	if idempotencyKey == "" {
		response.Error(w, http.StatusBadRequest, "Idempotency-Key header required")
		return
	}

	var req CreatePaymentRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid json body")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	payment, err := h.service.ProcessPayment(
		ctx,
		req.UserID,
		req.Amount,
		req.Currency,
		idempotencyKey,
	)

	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, payment)
}
