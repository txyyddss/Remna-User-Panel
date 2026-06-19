package payments

type providerPaymentRequest struct {
	OrderID     string
	Amount      float64
	Currency    string
	Description string
	ClientIP    string
}

type providerPaymentResponse struct {
	ProviderPaymentID string
	PaymentURL        string
	QRContent         string
	DisplayAmount     string
	DisplayCurrency   string
	PaymentAddress    string
	Network           string
	URLScheme         string
	ExpiresInSeconds  int
}

type webhookResult struct {
	OrderID           string
	ProviderPaymentID string
	Paid              bool
	Raw               map[string]any
}
