# Accounts and onboarding module

## Ownership and interfaces

This module owns Telegram Mini App session exchange, local users/sessions, versioned localized onboarding content, stateless signed join-request links, canonical membership verification, immutable username reservation, agreement acceptance, initial Remnawave provisioning, and recovery from a confirmed missing linked Remnawave user.

Public operations are `POST /api/v1/auth/telegram`, `GET /api/v1/me`, and the `/api/v1/onboarding/*` operations defined in OpenAPI. Administrator authorization reuses the validated session and compares the Telegram ID to `ADMIN_TELEGRAM_ID`; no browser role is trusted.

## State and invariants

- Raw `Telegram.WebApp.initData` is authenticated with Telegram's HMAC construction and constant-time comparison. `auth_date` must be no older than five minutes and future-skewed data is rejected.
- A random opaque session is stored only as a hash, expires after seven days, and is delivered in a Secure, HttpOnly, SameSite cookie. Raw init data and session values are never logged.
- The resumable state machine is `intro -> membership -> username -> agreement -> complete`. Replaying a completed transition is safe and cannot move a member backward except through the explicit missing-user recovery transition.
- Group/channel invites request approval and expire after 30 minutes. Their Telegram name fits the 32-character limit and carries base-36 user identity plus a 96-bit HMAC over user, chat, and expiry. No invite row/link is persisted; webhooks validate signature, identity, chat, expiry, and the current signing secret before approval and revocation.
- Membership events improve responsiveness, but `getChatMember` for both configured chats is canonical. Membership is enforced during signup and missing-user recovery.
- Usernames match `^[a-z]{3,9}$`, are lowercase and unique, and cannot change after reservation. Local lookup plus Remnawave preflight provide fast feedback; Remnawave `A019` resolves an upstream duplicate race.
- Draft and published welcome/agreement bundles have independent revisions. `en` and `zh-CN` are mandatory and must contain matching stable IDs/order (and matching whitelisted agreement icons). Welcome durations are derived at read time from 220 Latin words/minute plus 300 CJK characters/minute, 600 ms transition allowance, clamped to 1.8–12 seconds.
- Publishing changed agreements increments the agreement revision and routes every completed user back to `agreement`. Acceptance requires the exact current revision and every active agreement ID; stale or partial submissions return `409`. Completion provisions/reconciles the ACTIVE Remnawave user and stores the accepted revision.

## Missing linked-user recovery

Every authentication of a completed linked member rechecks the exact Remnawave user ID. Only a confirmed provider 404 begins recovery. Timeouts, transport failures, and 5xx responses return a temporary `UPSTREAM_UNAVAILABLE` response and do not mutate the local account or create a session.

Confirmed-missing recovery preserves the same local user ID, Telegram identity, immutable username, TXB balance and ledger, purchases, Activity results, Questionnaire history, and Emby linkage. It clears only the missing Remnawave identity plus membership/agreement completion and records `recoveryReason=remnawave_user_missing`.

The member repeats canonical membership verification and current agreement bundle, recreates the same Remnawave username, and then revision-aware completion transactionally queues `remna_sync_user`. That worker reapplies any still-valid local entitlement.

## Failure and reconciliation

Malformed, stale, or incorrectly signed Telegram launch data returns the same unauthenticated envelope. A Telegram outage never infers membership from a browser claim.

Username reservation and upstream provisioning are resumable. If local reservation succeeds but Remnawave is unavailable, the account remains at the agreement boundary. Retry reconciles by exact username and Telegram ID and adopts only an unambiguous match. It never creates a second local identity or silently selects a conflicting upstream user.

Invite creation/approval failures leave local membership false. Signature rotation intentionally invalidates names created with the old secret. The designated administrator can enter system setup before completing ordinary onboarding; user product operations remain guarded until that flow completes.

## Verification

- Table-driven init-data tests cover canonical sorting, percent encoding, missing fields, bad hashes, stale/future timestamps, and accepted boundary age.
- Repository tests race two members for one username and assert one reservation.
- Telegram fake-server tests cover invite expiry, wrong-user requests, duplicate join events, partial membership, and canonical recheck.
- Remnawave tests cover the exact initial payload, `A019`, timeout-after-create reconciliation, ambiguous matches, linked-user 404 recovery, and 5xx/timeout no-mutation behavior.
- Recovery integration tests assert identity, balance, ledger, purchase, Activity, Questionnaire, and Emby records remain attached to the same local user.
- Handler tests verify cookie flags, session expiry, resumable transitions, recovery reason, temporary upstream mapping, administrator bootstrap, and redacted logs.
