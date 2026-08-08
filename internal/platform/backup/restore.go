package backup

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	platformdatabase "github.com/txyyddss/Remna-User-Panel/internal/platform/database"
	"github.com/txyyddss/Remna-User-Panel/internal/platform/ids"
	_ "modernc.org/sqlite"
)

const (
	markerSuffix = ".restore-pending.json"
	resultSuffix = ".restore-result.json"
)

// ErrRestoreConflict reports an invalid confirmation, incompatible snapshot,
// concurrent restore, or restore record that changed while being staged.
var ErrRestoreConflict = errors.New("restore conflicts with current state")

// RestoreJob describes a durable staged restore operation.
type RestoreJob struct {
	ID           string     `json:"id"`
	BackupID     string     `json:"backupId"`
	ActorUserID  *string    `json:"actorUserId,omitempty"`
	Status       string     `json:"status"`
	StagedPath   string     `json:"-"`
	RescuePath   string     `json:"-"`
	SourceSHA256 string     `json:"sourceSha256"`
	Error        string     `json:"error,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
}

type restoreMarker struct {
	Version      int       `json:"version"`
	JobID        string    `json:"jobId"`
	BackupID     string    `json:"backupId"`
	ActorUserID  *string   `json:"actorUserId,omitempty"`
	DatabasePath string    `json:"databasePath"`
	StagedPath   string    `json:"stagedPath"`
	RescuePath   string    `json:"rescuePath"`
	SourcePath   string    `json:"sourcePath"`
	SourceSize   int64     `json:"sourceSize"`
	SourceSHA256 string    `json:"sourceSha256"`
	CreatedAt    time.Time `json:"createdAt"`
}

// StartupRestoreResult is written beside the database during the pre-open
// atomic swap and imported into restore_jobs after SQLite is opened.
type StartupRestoreResult struct {
	restoreMarker
	Status      string    `json:"status"`
	Error       string    `json:"error,omitempty"`
	CompletedAt time.Time `json:"completedAt"`
}

// StageRestore verifies a stored snapshot, copies it beside the live database,
// creates a verified rescue backup, and publishes a durable pre-open marker.
// No live database file is replaced by this method.
func (s *Service) StageRestore(ctx context.Context, backupID, actorUserID, reason, confirmation string, supportedMigrations []string) (RestoreJob, error) {
	s.restoreMu.Lock()
	defer s.restoreMu.Unlock()

	reason = strings.TrimSpace(reason)
	if len(reason) < 4 || len(reason) > 500 {
		return RestoreJob{}, fmt.Errorf("%w: restore reason must contain 4 to 500 characters", ErrRestoreConflict)
	}
	if len(supportedMigrations) == 0 {
		return RestoreJob{}, fmt.Errorf("%w: migration allowlist is empty", ErrRestoreConflict)
	}
	if active, err := s.pendingMarkerExists(ctx); err != nil {
		return RestoreJob{}, err
	} else if active {
		return RestoreJob{}, fmt.Errorf("%w: another restore is already staged", ErrRestoreConflict)
	}

	download, err := s.OpenDownload(ctx, backupID)
	if err != nil {
		return RestoreJob{}, err
	}
	defer func() { _ = download.File.Close() }()
	if confirmation != "RESTORE "+download.Name {
		return RestoreJob{}, fmt.Errorf("%w: typed restore confirmation does not match", ErrRestoreConflict)
	}
	databasePath, err := s.mainDatabasePath(ctx)
	if err != nil {
		return RestoreJob{}, err
	}
	jobID, err := ids.New()
	if err != nil {
		return RestoreJob{}, fmt.Errorf("create restore identifier: %w", err)
	}
	now := s.now().UTC()
	stagePath := filepath.Join(filepath.Dir(databasePath), "."+filepath.Base(databasePath)+".restore-"+jobID+".stage")
	_, err = copyAndSync(download.File, stagePath)
	if err != nil {
		return RestoreJob{}, fmt.Errorf("stage restore snapshot: %w", err)
	}
	removeStage := true
	defer func() {
		if removeStage {
			_ = os.Remove(stagePath)
		}
	}()
	if err := verifySnapshot(ctx, stagePath, supportedMigrations); err != nil {
		return RestoreJob{}, fmt.Errorf("%w: %v", ErrRestoreConflict, err)
	}
	if err := platformdatabase.PrepareRestoreSnapshot(ctx, stagePath); err != nil {
		return RestoreJob{}, fmt.Errorf("%w: preflight staged snapshot: %v", ErrRestoreConflict, err)
	}
	if err := verifyCurrentSnapshot(ctx, stagePath, supportedMigrations); err != nil {
		return RestoreJob{}, fmt.Errorf("%w: verify migrated staged snapshot: %v", ErrRestoreConflict, err)
	}
	sourceHash, err := hashFile(stagePath)
	if err != nil {
		return RestoreJob{}, fmt.Errorf("hash prepared restore snapshot: %w", err)
	}

	rescue, err := s.Run(ctx)
	if err != nil {
		return RestoreJob{}, fmt.Errorf("create rescue backup: %w", err)
	}
	actor := nullableActor(actorUserID)
	job := RestoreJob{
		ID: jobID, BackupID: backupID, ActorUserID: actor, Status: "staging", StagedPath: stagePath,
		RescuePath: rescue.Path, SourceSHA256: sourceHash, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.insertRestoreJob(ctx, job); err != nil {
		return RestoreJob{}, err
	}
	marker := restoreMarker{
		Version: 2, JobID: jobID, BackupID: backupID, ActorUserID: actor, DatabasePath: databasePath,
		StagedPath: stagePath, RescuePath: rescue.Path, SourcePath: filepath.Join(s.directory, download.Name),
		SourceSize: download.Size, SourceSHA256: sourceHash, CreatedAt: now,
	}
	if err := writeJSONAtomic(markerPath(databasePath), marker); err != nil {
		_ = s.failRestoreJob(ctx, jobID, err)
		return RestoreJob{}, fmt.Errorf("publish restore marker: %w", err)
	}
	if err := s.markRestoreReady(ctx, job, reason, rescue.ID); err != nil {
		_ = os.Remove(markerPath(databasePath))
		_ = s.failRestoreJob(ctx, jobID, err)
		return RestoreJob{}, err
	}
	removeStage = false
	job.Status = "ready"
	job.UpdatedAt = s.now().UTC()
	return job, nil
}

func nullableActor(actor string) *string {
	actor = strings.TrimSpace(actor)
	if actor == "" {
		return nil
	}
	return &actor
}

func (s *Service) pendingMarkerExists(ctx context.Context) (bool, error) {
	if s.db == nil {
		return false, errors.New("backup service database is unavailable")
	}
	path, err := s.mainDatabasePath(ctx)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(markerPath(path))
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("inspect restore marker: %w", err)
}

func (s *Service) mainDatabasePath(ctx context.Context) (string, error) {
	rows, err := s.db.QueryContext(ctx, `PRAGMA database_list`)
	if err != nil {
		return "", fmt.Errorf("locate live database: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sequence int
		var name, path string
		if err := rows.Scan(&sequence, &name, &path); err != nil {
			return "", fmt.Errorf("scan database location: %w", err)
		}
		if name == "main" {
			if strings.TrimSpace(path) == "" {
				return "", errors.New("live database does not have a filesystem path")
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				return "", fmt.Errorf("resolve live database path: %w", err)
			}
			return filepath.Clean(absolute), nil
		}
	}
	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("iterate database locations: %w", err)
	}
	return "", errors.New("main SQLite database was not found")
}

func copyAndSync(source *os.File, destination string) (hash string, returnedErr error) {
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("rewind source snapshot: %w", err)
	}
	file, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("create staged snapshot: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); returnedErr == nil && closeErr != nil {
			returnedErr = fmt.Errorf("close staged snapshot: %w", closeErr)
		}
		if returnedErr != nil {
			_ = os.Remove(destination)
		}
	}()
	digest := sha256.New()
	if _, err := io.Copy(io.MultiWriter(file, digest), source); err != nil {
		return "", fmt.Errorf("copy snapshot: %w", err)
	}
	if err := file.Sync(); err != nil {
		return "", fmt.Errorf("sync staged snapshot: %w", err)
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	digest := sha256.New()
	if _, err := io.Copy(digest, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func verifySnapshot(ctx context.Context, path string, supportedMigrations []string) error {
	return verifySnapshotMigrations(ctx, path, supportedMigrations, false)
}

func verifyCurrentSnapshot(ctx context.Context, path string, supportedMigrations []string) error {
	return verifySnapshotMigrations(ctx, path, supportedMigrations, true)
}

func verifySnapshotMigrations(ctx context.Context, path string, supportedMigrations []string, requireCurrent bool) error {
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?mode=ro&_pragma=query_only(1)&_pragma=foreign_keys(1)")
	if err != nil {
		return fmt.Errorf("open snapshot: %w", err)
	}
	defer func() { _ = db.Close() }()
	var integrity string
	if err := db.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&integrity); err != nil {
		return fmt.Errorf("run integrity check: %w", err)
	}
	if integrity != "ok" {
		return fmt.Errorf("integrity check returned %q", integrity)
	}
	foreignRows, err := db.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		return fmt.Errorf("run foreign-key check: %w", err)
	}
	if foreignRows.Next() {
		_ = foreignRows.Close()
		return errors.New("foreign-key check found violations")
	}
	if err := foreignRows.Err(); err != nil {
		_ = foreignRows.Close()
		return fmt.Errorf("read foreign-key check: %w", err)
	}
	if err := foreignRows.Close(); err != nil {
		return fmt.Errorf("close foreign-key check: %w", err)
	}
	for index, version := range supportedMigrations {
		if strings.TrimSpace(version) == "" {
			return errors.New("supported migration list contains an empty version")
		}
		if index > 0 && supportedMigrations[index-1] >= version {
			return errors.New("supported migration list is not strictly ordered")
		}
	}
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return fmt.Errorf("read snapshot migrations: %w", err)
	}
	defer rows.Close()
	seen := 0
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("scan snapshot migration: %w", err)
		}
		if seen >= len(supportedMigrations) || supportedMigrations[seen] != version {
			return fmt.Errorf("snapshot migrations are not an ordered prefix at %q", version)
		}
		seen++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate snapshot migrations: %w", err)
	}
	if seen == 0 {
		return errors.New("snapshot does not contain application migrations")
	}
	if requireCurrent && seen != len(supportedMigrations) {
		return fmt.Errorf("snapshot has %d of %d required migrations", seen, len(supportedMigrations))
	}
	return nil
}

func (s *Service) insertRestoreJob(ctx context.Context, job RestoreJob) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO restore_jobs(id,backup_run_id,actor_user_id,status,staged_path,rescue_path,source_sha256,error,created_at,updated_at)
		VALUES(?,?,?,?,?,?,?,?,?,?)`, job.ID, job.BackupID, job.ActorUserID, job.Status, job.StagedPath, job.RescuePath,
		job.SourceSHA256, "", job.CreatedAt.Format(time.RFC3339Nano), job.UpdatedAt.Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("record staged restore: %w", err)
	}
	return nil
}

func (s *Service) markRestoreReady(ctx context.Context, job RestoreJob, reason, rescueBackupID string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin restore audit: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	now := s.now().UTC()
	result, err := tx.ExecContext(ctx, `UPDATE restore_jobs SET status='ready',updated_at=? WHERE id=? AND status='staging'`, now.Format(time.RFC3339Nano), job.ID)
	if err != nil {
		return fmt.Errorf("mark restore ready: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		return fmt.Errorf("%w: restore record changed while staging", ErrRestoreConflict)
	}
	detail, err := json.Marshal(map[string]any{
		"backupId": job.BackupID, "rescueBackupId": rescueBackupID, "reason": reason,
		"sourceSha256": job.SourceSHA256, "warning": "direct restore bypasses domain synchronization hooks",
	})
	if err != nil {
		return fmt.Errorf("encode restore audit: %w", err)
	}
	auditID, err := ids.New()
	if err != nil {
		return fmt.Errorf("create restore audit identifier: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO audit_events(id,actor_user_id,action,target_type,target_id,detail,created_at) VALUES(?,?,?,?,?,?,?)`,
		auditID, job.ActorUserID, "database_restore_staged", "restore_job", job.ID, string(detail), now.Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("record restore audit: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM audit_events WHERE id IN (SELECT id FROM audit_events ORDER BY created_at DESC,id DESC LIMIT -1 OFFSET 200)`); err != nil {
		return fmt.Errorf("retain restore audits: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit restore staging: %w", err)
	}
	return nil
}

func (s *Service) failRestoreJob(ctx context.Context, id string, restoreErr error) error {
	message := strings.TrimSpace(restoreErr.Error())
	if len(message) > 500 {
		message = message[:500]
	}
	_, err := s.db.ExecContext(ctx, `UPDATE restore_jobs SET status='failed',error=?,updated_at=?,completed_at=? WHERE id=?`,
		message, s.now().UTC().Format(time.RFC3339Nano), s.now().UTC().Format(time.RFC3339Nano), id)
	return err
}

// Restore returns one restore job by identifier.
func (s *Service) Restore(ctx context.Context, id string) (RestoreJob, error) {
	var job RestoreJob
	var actor, completed sql.NullString
	var created, updated string
	err := s.db.QueryRowContext(ctx, `SELECT id,backup_run_id,actor_user_id,status,staged_path,rescue_path,source_sha256,error,created_at,updated_at,completed_at FROM restore_jobs WHERE id=?`, id).
		Scan(&job.ID, &job.BackupID, &actor, &job.Status, &job.StagedPath, &job.RescuePath, &job.SourceSHA256, &job.Error, &created, &updated, &completed)
	if errors.Is(err, sql.ErrNoRows) {
		return RestoreJob{}, ErrBackupNotFound
	}
	if err != nil {
		return RestoreJob{}, fmt.Errorf("lookup restore: %w", err)
	}
	if actor.Valid {
		job.ActorUserID = &actor.String
	}
	job.CreatedAt, err = time.Parse(time.RFC3339Nano, created)
	if err != nil {
		return RestoreJob{}, fmt.Errorf("parse restore creation time: %w", err)
	}
	job.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated)
	if err != nil {
		return RestoreJob{}, fmt.Errorf("parse restore update time: %w", err)
	}
	if completed.Valid {
		value, parseErr := time.Parse(time.RFC3339Nano, completed.String)
		if parseErr != nil {
			return RestoreJob{}, fmt.Errorf("parse restore completion time: %w", parseErr)
		}
		job.CompletedAt = &value
	}
	return job, nil
}

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

func truncateError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 500 {
		return value[:500]
	}
	return value
}

func markerPath(databasePath string) string { return databasePath + markerSuffix }
func resultPath(databasePath string) string { return databasePath + resultSuffix }

func validStagePath(databasePath, stagePath, jobID string) bool {
	expected := filepath.Join(filepath.Dir(databasePath), "."+filepath.Base(databasePath)+".restore-"+jobID+".stage")
	return samePath(expected, stagePath)
}

func samePath(left, right string) bool {
	leftAbsolute, leftErr := filepath.Abs(filepath.Clean(left))
	rightAbsolute, rightErr := filepath.Abs(filepath.Clean(right))
	if leftErr != nil || rightErr != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(leftAbsolute, rightAbsolute)
	}
	return leftAbsolute == rightAbsolute
}

func readRestoreMarker(path string) (restoreMarker, error) {
	data, err := readSmallFile(path)
	if err != nil {
		return restoreMarker{}, err
	}
	var marker restoreMarker
	if err := json.Unmarshal(data, &marker); err != nil {
		return restoreMarker{}, fmt.Errorf("decode restore marker: %w", err)
	}
	return marker, nil
}

func readStartupResult(path string) (StartupRestoreResult, error) {
	data, err := readSmallFile(path)
	if err != nil {
		return StartupRestoreResult{}, err
	}
	var result StartupRestoreResult
	if err := json.Unmarshal(data, &result); err != nil {
		return StartupRestoreResult{}, fmt.Errorf("decode restore result: %w", err)
	}
	return result, nil
}

func readSmallFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	return io.ReadAll(io.LimitReader(file, 64<<10))
}

func writeJSONAtomic(path string, value any) (returnedErr error) {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode durable state: %w", err)
	}
	temporary := path + ".tmp"
	file, err := os.OpenFile(temporary, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	closed := false
	defer func() {
		if !closed {
			if closeErr := file.Close(); returnedErr == nil && closeErr != nil {
				returnedErr = closeErr
			}
		}
		if returnedErr != nil {
			_ = os.Remove(temporary)
		}
	}()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	closed = true
	if err := os.Rename(temporary, path); err != nil {
		return err
	}
	if runtime.GOOS != "windows" {
		directory, err := os.Open(filepath.Dir(path))
		if err != nil {
			return err
		}
		syncErr := directory.Sync()
		closeErr := directory.Close()
		if syncErr != nil {
			return syncErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}
