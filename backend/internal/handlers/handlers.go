package handlers

import (
	"github.com/user/remna-user-panel/internal/services"
)

// Handler holds all HTTP handler dependencies
type Handler struct {
	Credit    *services.CreditService
	Payment   *services.PaymentService
	IPChanges *services.IPChangeService
}

// NewHandler creates a new Handler
func NewHandler() *Handler {
	credit := services.NewCreditService()
	payment := services.NewPaymentService(credit)
	ipChange := services.NewIPChangeService()
	return &Handler{
		Credit:    credit,
		Payment:   payment,
		IPChanges: ipChange,
	}
}
