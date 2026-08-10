package backup

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	platformdatabase "github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ApplyPendingRestore atomically swaps a staged snapshot into place before the
// application opens SQLite. It restores the original file on any failed
// post-swap verification and writes a sidecar result for startup recording.
func ApplyPendingRestore(ctx context.Context, databasePath string, supportedMigrations []string) (*StartupRestoreResult, error) {
	databasePath, err := filepath.Abs(filepath.Clean(databasePath))
	if err != nil {
		return nil, fmt.Errorf("resolve restore database path: %w", err)
	}
	markerFile := markerPath(databasePath)
	marker, err := readRestoreMarker(markerFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !samePath(marker.DatabasePath, databasePath) || (marker.Version != 1 && marker.Version != 2) {
		return nil, fmt.Errorf("%w: restore marker target or version is invalid", ErrRestoreConflict)
	}
	if !validStagePath(databasePath, marker.StagedPath, marker.JobID) {
		return nil, fmt.Errorf("%w: restore staging path is invalid", ErrRestoreConflict)
	}
	if existing, readErr := readStartupResult(resultPath(databasePath)); readErr == nil && existing.JobID == marker.JobID && existing.Status == "complete" {
		_ = os.Remove(markerFile)
		return &existing, nil
	}
	actualHash, err := hashFile(marker.StagedPath)
	if err != nil {
		return nil, recordStartupFailure(databasePath, marker, "failed", fmt.Errorf("hash staged restore: %w", err))
	}
	if !strings.EqualFold(actualHash, marker.SourceSHA256) {
		return nil, recordStartupFailure(databasePath, marker, "failed", errors.New("staged restore checksum changed"))
	}
	if err := verifySnapshot(ctx, marker.StagedPath, supportedMigrations); err != nil {
		return nil, recordStartupFailure(databasePath, marker, "failed", err)
	}
	if err := platformdatabase.PrepareRestoreSnapshot(ctx, marker.StagedPath); err != nil {
		return nil, recordStartupFailure(databasePath, marker, "failed", fmt.Errorf("preflight staged restore: %w", err))
	}
	if err := verifyCurrentSnapshot(ctx, marker.StagedPath, supportedMigrations); err != nil {
		return nil, recordStartupFailure(databasePath, marker, "failed", fmt.Errorf("verify prepared staged restore: %w", err))
	}
	preparedHash, err := hashFile(marker.StagedPath)
	if err != nil {
		return nil, recordStartupFailure(databasePath, marker, "failed", fmt.Errorf("hash prepared staged restore: %w", err))
	}
	if marker.Version == 2 && !strings.EqualFold(preparedHash, marker.SourceSHA256) {
		return nil, recordStartupFailure(databasePath, marker, "failed", errors.New("prepared restore changed during startup preflight"))
	}
	marker.SourceSHA256 = preparedHash

	rollbackPath := databasePath + ".restore-rollback-" + marker.JobID
	if _, err := os.Stat(rollbackPath); err == nil {
		return nil, recordStartupFailure(databasePath, marker, "failed", errors.New("restore rollback path already exists"))
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, recordStartupFailure(databasePath, marker, "failed", err)
	}
	if err := os.Rename(databasePath, rollbackPath); err != nil {
		return nil, recordStartupFailure(databasePath, marker, "failed", fmt.Errorf("preserve live database: %w", err))
	}
	movedSidecars, sidecarErr := moveSQLiteSidecars(databasePath, rollbackPath)
	if sidecarErr != nil {
		rollbackErr := restoreRollback(databasePath, rollbackPath, movedSidecars)
		if rollbackErr != nil {
			sidecarErr = errors.Join(sidecarErr, rollbackErr)
		}
		return nil, recordStartupFailure(databasePath, marker, "rolled_back", sidecarErr)
	}
	if err := os.Rename(marker.StagedPath, databasePath); err != nil {
		restoreErr := restoreRollback(databasePath, rollbackPath, movedSidecars)
		if restoreErr != nil {
			err = errors.Join(err, restoreErr)
		}
		return nil, recordStartupFailure(databasePath, marker, "rolled_back", fmt.Errorf("activate staged restore: %w", err))
	}
	if err := verifyCurrentSnapshot(ctx, databasePath, supportedMigrations); err != nil {
		failedPath := marker.StagedPath + ".failed"
		_ = os.Rename(databasePath, failedPath)
		rollbackErr := restoreRollback(databasePath, rollbackPath, movedSidecars)
		if rollbackErr != nil {
			err = errors.Join(err, rollbackErr)
		}
		return nil, recordStartupFailure(databasePath, marker, "rolled_back", fmt.Errorf("verify activated restore: %w", err))
	}
	_ = os.Remove(rollbackPath)
	for _, path := range movedSidecars {
		_ = os.Remove(path)
	}
	result := StartupRestoreResult{restoreMarker: marker, Status: "complete", CompletedAt: time.Now().UTC()}
	if err := writeJSONAtomic(resultPath(databasePath), result); err != nil {
		return &result, fmt.Errorf("record completed startup restore: %w", err)
	}
	if err := os.Remove(markerFile); err != nil && !errors.Is(err, os.ErrNotExist) {
		return &result, fmt.Errorf("remove applied restore marker: %w", err)
	}
	return &result, nil
}

func moveSQLiteSidecars(databasePath, rollbackPath string) ([]string, error) {
	moved := make([]string, 0, 2)
	for _, suffix := range []string{"-wal", "-shm"} {
		source, destination := databasePath+suffix, rollbackPath+suffix
		info, err := os.Lstat(source)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return moved, fmt.Errorf("inspect SQLite sidecar %s: %w", suffix, err)
		}
		if !info.Mode().IsRegular() {
			return moved, fmt.Errorf("SQLite sidecar %s is not a regular file", suffix)
		}
		if err := os.Rename(source, destination); err != nil {
			return moved, fmt.Errorf("preserve SQLite sidecar %s: %w", suffix, err)
		}
		moved = append(moved, destination)
	}
	return moved, nil
}

func restoreRollback(databasePath, rollbackPath string, movedSidecars []string) error {
	if err := os.Rename(rollbackPath, databasePath); err != nil {
		return fmt.Errorf("restore original database: %w", err)
	}
	for _, source := range movedSidecars {
		suffix := strings.TrimPrefix(source, rollbackPath)
		if err := os.Rename(source, databasePath+suffix); err != nil {
			return fmt.Errorf("restore SQLite sidecar %s: %w", suffix, err)
		}
	}
	return nil
}

func recordStartupFailure(databasePath string, marker restoreMarker, status string, restoreErr error) error {
	result := StartupRestoreResult{restoreMarker: marker, Status: status, Error: restoreErr.Error(), CompletedAt: time.Now().UTC()}
	if err := writeJSONAtomic(resultPath(databasePath), result); err != nil {
		return errors.Join(restoreErr, fmt.Errorf("record restore failure: %w", err))
	}
	_ = os.Remove(markerPath(databasePath))
	_ = os.Remove(marker.StagedPath)
	return restoreErr
}

// RecordStartupRestore imports the pre-open result into the active database.
// Call it immediately after migrations complete; it is idempotent.
func RecordStartupRestore(ctx context.Context, db *sql.DB, databasePath string) (*StartupRestoreResult, error) {
	result, err := readStartupResult(resultPath(databasePath))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return &result, fmt.Errorf("begin startup restore record: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	stamp := result.CompletedAt.Format(time.RFC3339Nano)
	if _, err := tx.ExecContext(ctx, `INSERT INTO backup_runs(id,path,size_bytes,status,error,created_at,completed_at)
		VALUES(?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET path=excluded.path,size_bytes=excluded.size_bytes,status='complete',error='',completed_at=excluded.completed_at`,
		result.BackupID, result.SourcePath, result.SourceSize, "complete", "", result.CreatedAt.Format(time.RFC3339Nano), stamp); err != nil {
		return &result, fmt.Errorf("restore source backup record: %w", err)
	}
	var actor any
	if result.ActorUserID != nil {
		var exists int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE id=?`, *result.ActorUserID).Scan(&exists); err != nil {
			return &result, fmt.Errorf("verify restore actor: %w", err)
		}
		if exists > 0 {
			actor = *result.ActorUserID
		}
	}
	status := result.Status
	if status != "complete" && status != "failed" && status != "rolled_back" {
		status = "failed"
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO restore_jobs(id,backup_run_id,actor_user_id,status,staged_path,rescue_path,source_sha256,error,created_at,updated_at,completed_at)
		VALUES(?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(id) DO UPDATE SET status=excluded.status,error=excluded.error,updated_at=excluded.updated_at,completed_at=excluded.completed_at`,
		result.JobID, result.BackupID, actor, status, result.StagedPath, result.RescuePath, result.SourceSHA256,
		truncateError(result.Error), result.CreatedAt.Format(time.RFC3339Nano), stamp, stamp); err != nil {
		return &result, fmt.Errorf("record startup restore job: %w", err)
	}
	auditID, err := ids.New()
	if err != nil {
		return &result, fmt.Errorf("create startup restore audit identifier: %w", err)
	}
	detail, _ := json.Marshal(map[string]any{"backupId": result.BackupID, "status": status, "error": truncateError(result.Error)})
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_events(id,actor_user_id,action,target_type,target_id,detail,created_at) VALUES(?,?,?,?,?,?,?)`,
		auditID, actor, "database_restore_"+status, "restore_job", result.JobID, string(detail), stamp); err != nil {
		return &result, fmt.Errorf("audit startup restore: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM audit_events WHERE id IN (SELECT id FROM audit_events ORDER BY created_at DESC,id DESC LIMIT -1 OFFSET 200)`); err != nil {
		return &result, fmt.Errorf("retain startup restore audits: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return &result, fmt.Errorf("commit startup restore record: %w", err)
	}
	if err := os.Remove(resultPath(databasePath)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return &result, fmt.Errorf("remove recorded restore result: %w", err)
	}
	return &result, nil
}
