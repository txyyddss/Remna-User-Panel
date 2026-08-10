# Database administration

This package provides the allowlisted, typed SQLite administration surface. It
introspects real schema metadata, masks sensitive data, uses opaque cursors, and
requires reviewed mutations with rescue backups and immutable audit evidence.

## Files

- `types.go` — public service contracts, table/column/value representations,
  query and mutation requests, reviews, results, and reason validation.
- `schema.go` — identifier allowlisting, SQLite schema introspection, affinity,
  boolean detection, and sensitive-column classification.
- `query.go` — filtered record queries, stable fingerprints, and typed query
  cursor encoding/decoding.
- `records.go` — basic table paging, select/key construction, scanning, and public
  record assembly.
- `record_values.go` — sensitive-value masking, canonical hashing, and canonical
  key representation.
- `record_keys.go` — public-key parsing plus legacy record cursor encoding and
  decoding.
- `mutations.go` — mutation review/application orchestration, confirmations,
  digests, and one-time review consumption.
- `mutation_prepare.go` — current-record lookup, requested-value preparation,
  default/null handling, and encrypted-setting replacement.
- `mutation_values.go` — strict editor-value conversion and mutation key recovery.
- `mutation_execute.go` — transactional insert/update/delete execution, audited
  rescue-backup linkage, and defensive request cloning.
- `service_test.go` — schema, typed value, rowid, default insert, and cursor tests.
- `mutation_service_test.go` — reviewed mutation, rollback/conflict, encrypted
  setting, shared fixture, and fake backup tests.
