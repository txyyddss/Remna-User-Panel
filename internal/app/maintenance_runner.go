package app

import (
	"context"
	"fmt"
	"time"

	"github.com/txyyddss/Remna-User-Panel/internal/maintenance"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/database"
)

// runMaintenanceTask keeps the scheduled and manual commands on the same
// backup-gated cleanup path, including adjacent migration snapshot pruning.
func runMaintenanceTask(ctx context.Context, service *maintenance.Service, databasePath string,
	at time.Time, force bool) (maintenance.RunResult, error) {
	var (
		result maintenance.RunResult
		err    error
	)
	if force {
		result, err = service.RunManual(ctx, at)
	} else {
		result, err = service.Run(ctx, at)
	}
	if err != nil {
		return result, err
	}
	if err := database.PruneExpansionBackups(databasePath, at); err != nil {
		return result, fmt.Errorf("prune migration snapshots: %w", err)
	}
	return result, nil
}
