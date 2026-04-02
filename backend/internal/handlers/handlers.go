package handlers

import (
	"github.com/user/remna-user-panel/internal/services"
)

// Handler holds all HTTP handler dependencies
type Handler struct {
	Credit  *services.CreditService
	Payment *services.PaymentService
}

// NewHandler creates a new Handler
func NewHandler() *Handler {
	credit := services.NewCreditService()
	payment := services.NewPaymentService(credit)
	return &Handler{
		Credit:  credit,
		Payment: payment,
	}
}
