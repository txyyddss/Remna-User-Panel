package billing

import (
	"context"
	"time"
)

type paymentCreateOperationResolver interface {
	ResolvePaymentCreateOperation(context.Context, string, string, time.Time) error
}

func (s *Service) resolvePaymentCreateOperation(ctx context.Context, event ProviderEvent) error {
	repository, ok := s.repository.(paymentCreateOperationResolver)
	if !ok {
		return nil
	}
	reference := event.TradeID
	if reference == "" {
		reference = event.ChargeID
	}
	return repository.ResolvePaymentCreateOperation(ctx, event.OrderID, reference, s.now().UTC())
}
