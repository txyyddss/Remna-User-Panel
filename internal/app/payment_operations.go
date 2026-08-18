package app

import (
	"fmt"

	"github.com/txyyddss/Remna-User-Panel/internal/billing"
	"github.com/txyyddss/Remna-User-Panel/internal/providerops"
)

func registerPaymentOperationHandlers(dispatcher *providerops.Dispatcher, service *billing.Service) error {
	worker, err := billing.NewOperationWorker(service)
	if err != nil {
		return err
	}
	for _, kind := range []string{billing.OperationCreateKind, billing.OperationCancelKind} {
		if err := dispatcher.Register(kind, worker); err != nil {
			return fmt.Errorf("register %s: %w", kind, err)
		}
	}
	return nil
}
