# Questionnaires
- `csv_part2.go` continues the focused implementation from its original package module.
- `questionnaires_part2.go` continues the focused implementation from its original package module.

- `questionnaires.go` defines form, participant, import, settlement, and validation domain types.
- `service.go` coordinates questionnaire participation and bounded CSV import workflows.
- `operations.go` atomically confirms an analyzed import and creates its idempotent settlement receipt.
- `operation_worker.go` applies the local settlement transaction and completes its receipt exactly once.
- `csv.go` parses and normalizes uploaded questionnaire result files.
- `questionnaires_test.go` verifies validation, participation, and import behavior.
