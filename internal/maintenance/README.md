# Daily maintenance

`Service` owns the once-per-local-date maintenance order: claim the durable
lease, create and verify a SQLite backup, then invoke the database's atomic
fact compaction and retention transaction. Failed backups are recorded and
never permit cleanup. The database layer preserves active and queued purchases,
immutable ledger/audit evidence, pending work, provider replay tombstones, and
the compact rollups required after detail rows are removed.

The daily transaction also bounds provider-operation receipts, notification
deduplication, sessions, abandoned onboarding identities, and maintenance-run
history. Stale locally debited operations are compensated before removal.

- `service.go` coordinates the locked backup, verification, compaction, and purge sequence once per configured local date.
- `service_test.go` covers backup-gated cleanup and same-date locking.
- `README.md` documents the package ownership boundary.
