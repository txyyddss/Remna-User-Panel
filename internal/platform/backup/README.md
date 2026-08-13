# Backup and restore
- `restore_part2.go` continues the focused implementation from its original package module.
- `restore_startup_part2.go` continues the focused implementation from its original package module.
- `service_part2.go` continues the focused implementation from its original package module.

This package creates verified SQLite backups, serves contained downloads, stages
administrator restore requests, and performs the pre-open database swap. Restore
markers and results are written atomically so process crashes remain recoverable.

## Files

The implementation is split so backup publication and restore startup remain independently reviewable.

- `restore_startup_part2.go` contains restore marker startup transitions.
- `service_part2.go` contains backup publication and retention helpers.

- `service.go` — service construction, backup deletion/containment, verified
  online backup creation, and retention cleanup.
- `download.go` — safe opening of completed backup files for HTTP download.
- `restore.go` — restore contracts, confirmation, staging, database-path lookup,
  and pending-marker checks.
- `restore_snapshot.go` — snapshot copying, hashing, integrity checks, and ordered
  migration-prefix validation.
- `restore_jobs.go` — restore-job insertion, ready/failure transitions, rescue
  backup linkage, and staged-job reads.
- `restore_startup.go` — pre-open swap, SQLite sidecar handling, rollback, startup
  failure recording, and post-open completion recording.
- `restore_files.go` — bounded marker/result reads, path validation, error
  truncation, and atomic JSON file replacement.
- `service_test.go` — backup creation, retention, download containment, staged
  swap/rollback, and migration-marker verification tests.
