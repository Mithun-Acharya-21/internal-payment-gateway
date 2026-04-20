package model

import "time"

type PaymentStatus string

//allowed status constants for payment processing

const (
	StatusPending PaymentStatus = "pending"
	StatusProcessed PaymentStatus = "processed"
	StatusFailed PaymentStatus = "failed"
)
//it bridges three diffrent layers of the application - database, business logic, and API responses. It defines the structure of a payment record and provides a constructor function to create new payment instances with default values. This model is used throughout the application to ensure consistency in how payment data is represented and manipulated.
type Payment struct {
	ID int `json:"id"`
	UserID string `json:"user_id"`
	Amount float64 `json:"amount"`
	Currency string `json:"currency"`
	Status PaymentStatus `json:"status"`
	IdempotencyKey string `json:"idempotency_key"`
	CreatedAt time.Time `json:"created_at"`
}
//safe entry point for creating new payment instances with default values, ensuring that all required fields are set and that the status is initialized to pending. This promotes consistency and reduces the risk of errors when creating payment records throughout the application.
func NewPayment(userID string,amount float64,currency string,idempotencykey string) *Payment {
	return &Payment{
		UserID: userID,
		Amount: amount,
		Currency: currency,
		Status: StatusPending,
		IdempotencyKey: idempotencykey,
		CreatedAt: time.Now(),
}
}
//perfect place to add logic later 