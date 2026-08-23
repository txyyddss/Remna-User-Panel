package app

import (
	"fmt"

	"github.com/txyyddss/Remna-User-Panel/internal/admin"
	"github.com/txyyddss/Remna-User-Panel/internal/connections"
	"github.com/txyyddss/Remna-User-Panel/internal/emby"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	platformoutbox "github.com/txyyddss/Remna-User-Panel/internal/platform/outbox"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/secret"
	"github.com/txyyddss/Remna-User-Panel/internal/providerops"
	"github.com/txyyddss/Remna-User-Panel/internal/purchaseops"
)

type memberWorkflowServices struct {
	connections *connections.Service
	drops       *connections.DropService
	purchases   *purchaseops.Service
}

func newMemberWorkflows(store *database.Store, remna remnaAdapter, vault *secret.Vault, signingKey []byte) (memberWorkflowServices, *connections.Worker, *connections.ExpiryWorker, *providerops.Dispatcher, error) {
	signer, err := connections.NewSigner(signingKey)
	if err != nil {
		return memberWorkflowServices{}, nil, nil, nil, err
	}
	secrets := emby.NewSecretBox(vault)
	services := memberWorkflowServices{
		connections: connections.NewService(store, remna, signer),
		drops:       connections.NewDropService(store, signer, secrets),
		purchases:   purchaseops.NewService(store, remna),
	}
	dispatcher := providerops.NewDispatcher(store)
	purchaseWorker := purchaseops.NewWorker(store, remna)
	for _, kind := range []string{purchaseops.OperationResetKind, purchaseops.OperationRefundKind} {
		if err := dispatcher.Register(kind, purchaseWorker); err != nil {
			return memberWorkflowServices{}, nil, nil, nil, fmt.Errorf("register %s: %w", kind, err)
		}
	}
	if err := dispatcher.Register(connections.DropOperationKind, connections.NewDropWorker(store, remna, signer, secrets)); err != nil {
		return memberWorkflowServices{}, nil, nil, nil, fmt.Errorf("register connection drop: %w", err)
	}
	if err := dispatcher.Register(connections.BlockOperationKind, connections.NewBlockWorker(store, remna, secrets)); err != nil {
		return memberWorkflowServices{}, nil, nil, nil, fmt.Errorf("register connection block: %w", err)
	}
	if err := dispatcher.Register(connections.UnblockOperationKind, connections.NewUnblockWorker(store, remna, secrets)); err != nil {
		return memberWorkflowServices{}, nil, nil, nil, fmt.Errorf("register connection unblock: %w", err)
	}
	return services, connections.NewWorker(store, remna), connections.NewExpiryWorker(store, remna, secrets), dispatcher, nil
}

func registerAdminOperationHandlers(dispatcher *providerops.Dispatcher, store *database.Store, remna remnaAdapter) error {
	worker := admin.NewUserOperationWorker(store, remna)
	for _, kind := range []string{providerops.KindAdminEntitlementEdit, providerops.KindAdminEntitlementRefund,
		providerops.KindAdminComboReplacement, providerops.KindAdminBulkExtension, providerops.KindNodeCompensation} {
		if err := dispatcher.Register(kind, worker); err != nil {
			return err
		}
	}
	return nil
}

func registerProviderDispatcher(worker *platformoutbox.Worker, dispatcher *providerops.Dispatcher) error {
	return worker.Register(providerops.OutboxKind, dispatcher)
}
