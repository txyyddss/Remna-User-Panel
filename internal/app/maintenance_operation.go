package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/maintenance"
	"github.com/txyyddss/Remna-User-Panel/internal/model"
	"github.com/txyyddss/Remna-User-Panel/internal/providerops"
)

type maintenanceOperationRepository interface {
	ProviderOperationItems(context.Context, string) ([]providerops.Item, error)
	BeginProviderOperationAttempt(context.Context, string, time.Time) (providerops.Operation, error)
	BeginProviderOperationItemAttempt(context.Context, string, string, time.Time) (providerops.Item, error)
	CompleteProviderOperationItem(context.Context, string, string, providerops.Completion, time.Time) (providerops.Item, error)
	CompleteProviderOperation(context.Context, string, providerops.Completion, time.Time) (providerops.Operation, error)
}

type maintenanceOperationHandler struct {
	repository   maintenanceOperationRepository
	maintenance  *maintenance.Service
	databasePath string
	now          func() time.Time
}

func registerMaintenanceOperationHandler(dispatcher *providerops.Dispatcher, repository maintenanceOperationRepository,
	service *maintenance.Service, databasePath string) error {
	if dispatcher == nil || repository == nil || service == nil || databasePath == "" {
		return errors.New("maintenance operation dependencies are incomplete")
	}
	handler := &maintenanceOperationHandler{repository: repository, maintenance: service, databasePath: databasePath, now: time.Now}
	if err := dispatcher.Register(providerops.KindAdminMaintenance, handler); err != nil {
		return fmt.Errorf("register %s operation: %w", providerops.KindAdminMaintenance, err)
	}
	return nil
}

func (h *maintenanceOperationHandler) HandleProviderOperation(ctx context.Context, operation providerops.Operation, _ model.OutboxJob) error {
	operation, item, err := h.start(ctx, operation)
	if err != nil {
		return err
	}
	runCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()
	result, runErr := runMaintenanceTask(runCtx, h.maintenance, h.databasePath, h.now().UTC(), true)
	if errors.Is(runErr, maintenance.ErrBusy) {
		return runErr
	}
	if runErr != nil {
		if errors.Is(runErr, context.Canceled) || errors.Is(runErr, context.DeadlineExceeded) {
			return runErr
		}
		return h.complete(ctx, operation, item, providerops.Completion{Status: providerops.StatusFailed, ErrorCode: "MAINTENANCE_FAILED"})
	}
	resultJSON := "{}"
	if result.BackupRetentionWarning != nil {
		resultJSON = `{"backupRetentionWarning":true}`
	}
	return h.complete(ctx, operation, item, providerops.Completion{Status: providerops.StatusSucceeded, ResultJSON: resultJSON})
}

func (h *maintenanceOperationHandler) start(ctx context.Context, operation providerops.Operation) (providerops.Operation, providerops.Item, error) {
	items, err := h.repository.ProviderOperationItems(ctx, operation.Receipt.ID)
	if err != nil || len(items) != 1 || items[0].TargetType != "maintenance" || items[0].TargetID != "retention" {
		if err == nil {
			err = errors.New("maintenance operation has an invalid target")
		}
		return providerops.Operation{}, providerops.Item{}, err
	}
	item := items[0]
	if operation.Receipt.Status == string(providerops.StatusQueued) {
		operation, err = h.repository.BeginProviderOperationAttempt(ctx, operation.Receipt.ID, h.now().UTC())
		if err != nil {
			return providerops.Operation{}, providerops.Item{}, err
		}
	}
	if item.Status == providerops.StatusQueued {
		item, err = h.repository.BeginProviderOperationItemAttempt(ctx, operation.Receipt.ID, item.Key, h.now().UTC())
	}
	return operation, item, err
}

func (h *maintenanceOperationHandler) complete(ctx context.Context, operation providerops.Operation, item providerops.Item,
	completion providerops.Completion) error {
	if item.Status == providerops.StatusProcessing {
		if _, err := h.repository.CompleteProviderOperationItem(ctx, operation.Receipt.ID, item.Key, completion, h.now().UTC()); err != nil {
			return err
		}
	}
	_, err := h.repository.CompleteProviderOperation(ctx, operation.Receipt.ID, completion, h.now().UTC())
	return err
}

var _ providerops.Handler = (*maintenanceOperationHandler)(nil)
