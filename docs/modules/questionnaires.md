# Questionnaires module

## Ownership and interfaces

Questionnaires owns the single-active-questionnaire lifecycle, durable participant validation codes, bounded CSV ingestion, pre-settlement analysis, and idempotent background awards. Members retrieve the active questionnaire and history, then create or retrieve participation at `/api/v1/questionnaires/{id}/participation`. Administrators manage definitions and activate one questionnaire at a time under `/api/v1/admin/questionnaires`.

External form URLs must use HTTPS and an allowlisted Google Forms or Microsoft Forms host. Each member/questionnaire pair receives one cryptographically generated code. The code is stored durably and remains retrievable; retries never rotate it.

## CSV import state machine

Upload accepts one multipart `file`, at most 5 MiB. The parser accepts UTF-8 with an optional BOM, automatically selects comma, semicolon, or tab, and rejects documents over 50,000 data rows or 100 columns. It stores headers, row counts, a bounded sample, and the server-generated or supplied idempotency key.

An administrator then selects one exact header as the validation-code column. Analysis is read-only and reports matched, duplicate, unknown, malformed, and already-awarded counts. The generated ASCII codes are trimmed and compared case-insensitively; after the first known match, later duplicate rows do not produce another candidate award.

Confirmation changes the import to queued state and inserts a `questionnaire_settlement` outbox job in the same transaction. The kind-specific worker settles known, not-yet-rewarded participants. Each award record, TXB ledger credit, participant status update, and import progress update commits atomically with a semantic unique reference. Worker retries return the stored report and cannot duplicate credit.

Only one questionnaire can be active. Activation retires the prior active definition transactionally. Upload, analysis, confirmation, and polling verify that the import belongs to the questionnaire in the route.

## Security and failure behavior

- Raw CSV content is never executed as SQL or interpreted as HTML.
- Browser-supplied reward amounts, participant IDs, or match counts are ignored.
- Malformed encodings, oversized multipart bodies, duplicate headers, invalid delimiter inference, and unknown code columns fail before settlement.
- A crash after queuing or during a settlement batch is recoverable through durable state and idempotent award references.

## Verification

- Parser tests cover BOM, all delimiters, quoted cells, malformed rows, size/row/column bounds, and case-insensitive code matching.
- SQLite tests cover duplicate codes and rows, already-awarded participants, concurrent confirmation, rollback, replay, and one-ledger-credit-per-participant.
- Vue tests cover upload, sample, column selection, analysis review, queued progress, final report, error recovery, keyboard focus, and the 320 px mobile drawer.
