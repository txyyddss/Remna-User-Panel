# Abuse member UI

- `AbuseRecordsPage.vue` renders privacy-safe member detector records with wrapping mobile rows and locale-aware timestamps. It owns distinct loading, retryable error, populated, and empty views; `useAbuseRecords` remains responsible for loading the records. A failed request never appears as an empty account history.
