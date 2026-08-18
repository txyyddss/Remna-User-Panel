package app

import (
	"fmt"

	"github.com/txyyddss/Remna-User-Panel/internal/admin"
	"github.com/txyyddss/Remna-User-Panel/internal/catalog"
	"github.com/txyyddss/Remna-User-Panel/internal/emby"
	"github.com/txyyddss/Remna-User-Panel/internal/providerops"
	"github.com/txyyddss/Remna-User-Panel/internal/questionnaires"
)

func registerMutationOperationHandlers(dispatcher *providerops.Dispatcher, catalogService *catalog.Service,
	embyOperations *emby.OperationService, questionnaireService *questionnaires.Service, adminService *admin.Service) error {
	adminWorker, err := admin.NewMutationWorker(adminService)
	if err != nil {
		return err
	}
	registrations := []struct {
		kind    string
		handler providerops.Handler
	}{
		{providerops.KindSubscriptionRevoke, catalogService},
		{providerops.KindEmbySetup, embyOperations},
		{providerops.KindEmbyPreferences, embyOperations},
		{providerops.KindEmbyPassword, embyOperations},
		{providerops.KindEmbyProvisionRetry, embyOperations},
		{providerops.KindQuestionnaireSettlement, questionnaireService},
		{providerops.KindOutboxRetry, adminWorker},
		{providerops.KindPaymentRefund, adminWorker},
	}
	for _, registration := range registrations {
		if err := dispatcher.Register(registration.kind, registration.handler); err != nil {
			return fmt.Errorf("register %s operation: %w", registration.kind, err)
		}
	}
	return nil
}
