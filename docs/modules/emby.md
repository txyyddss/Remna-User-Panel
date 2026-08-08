# Emby account module

## Ownership and interfaces

The Emby module owns setup pricing, account linkage, temporary-password protection, durable provisioning, restricted policy overlays, preference changes, password changes, terminal refunds, and administrative retry visibility. `GET /api/v1/emby/account` returns current setup price, live Emby rating/library choices, and safe local account state. Members queue setup and update preferences/password under `/api/v1/emby`; administrators list accounts and retry retained transient failures.

`emby.base_url` and `emby.api_token` are encrypted settings. `emby.setup_price_txb` is a human-major decimal parsed exactly into integer hundredths by the server. The browser never supplies a price or raw Emby policy.

## Provisioning saga

Setup validates the selected rating and libraries against live Emby options, derives the base name from the local Remnawave username, and seals the temporary password with the existing AES-GCM vault using `emby.provisioning.password:<localUserId>` as authenticated context. One transaction debits the snapshotted setup price, appends its ledger entry, stores the sealed secret and preferences, and queues `emby_provision_account`.

The kind-specific worker advances idempotently:

1. Reserve the local provisioning attempt and exact candidate name.
2. Use the base Remnawave name if free. If an unrelated Emby account owns it, append `-` plus the stable first eight hexadecimal characters of SHA-256(local user ID).
3. If both candidates exist, fail instead of adopting an unknown account.
4. Mark creation attempted before calling Emby. After an ambiguous error, reconcile only by the exact persisted candidate.
5. Fetch the complete created user, set the password, fetch/overlay policy, and mark active.
6. Erase durable ciphertext on success. A terminal failure erases it and appends one compensating TXB refund; a transient failure retains it for retry.

## Policy and secret boundary

Every policy write starts from the current complete upstream policy. The server overlays only the selected parental rating and folder IDs, sets `EnableAllFolders=false`, and forces both hidden-login flags, remote-control fields, all audio/video transcoding, remux, sync conversion, media conversion, content download, and subtitle download fields off. `EnableRemoteAccess` and unrelated provider fields are preserved.

Plaintext passwords exist only in request memory or the short worker decryption window, are zeroed on return, and never enter responses, provider payload snapshots, audits, or logs. Linked password changes are synchronous and never persist either password. Options and account responses expose only safe identifiers, preferences, status, retryability, and redacted failure text.

## Verification

- Provider fixtures assert `X-Emby-Token`, endpoint/method/body contracts, live folder/rating mapping, complete-policy preservation, and every forced restriction.
- Saga tests cover name collision/suffix stability, ambiguous creation reconciliation, second collision refusal, transient retry, terminal debit/refund idempotency, and ciphertext erasure.
- HTTP/Vue tests cover exact setup price, restricted controls only, URL/rating/library states, password non-echo, retry status, and mobile/keyboard behavior.
