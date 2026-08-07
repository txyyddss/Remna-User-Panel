# Accounts and onboarding module

## Ownership and interfaces

This module owns Telegram Mini App session exchange, local users and sessions, onboarding state, join-request links, membership verification, immutable username reservation, agreement acceptance, and initial Remnawave provisioning. It consumes narrow Telegram and Remnawave interfaces; it does not know provider HTTP details.

Public operations are `POST /api/v1/auth/telegram`, `GET /api/v1/me`, and the four `/api/v1/onboarding/*` operations documented in OpenAPI. Admin authorization reuses the same validated session and compares the Telegram ID to `ADMIN_TELEGRAM_ID`; no client-supplied role is trusted.

## State and invariants

- Raw `Telegram.WebApp.initData` is authenticated with Telegram's HMAC construction and a constant-time hash comparison. `auth_date` must be no older than five minutes and future-skewed data is rejected.
- A random opaque session is stored only as a hash, expires after seven days, and is delivered in a Secure, HttpOnly, SameSite cookie. Raw init data and session values are never logged.
- The resumable state machine is `intro -> membership -> username -> agreement -> complete`. Replaying an already completed transition is safe and cannot move the user backward.
- Group and channel invites request membership approval, expire after 30 minutes, and are bound to one Telegram identity. An unrelated join request is never approved. A used link is revoked after approval.
- Membership events improve responsiveness, but `getChatMember` for both configured chats is canonical when the user asks to check. Membership is enforced during signup only.
- Usernames match `^[a-z]{3,9}$`, are lowercase, carry a database unique constraint, and cannot change after reservation. Local lookup and Remnawave preflight give quick feedback; Remnawave `A019` resolves an upstream duplicate race.
- Agreement completion provisions an ACTIVE Remnawave user with `expireAt=2099-12-31T23:59:59Z`, the Telegram ID, zero traffic, `NO_RESET`, no external squad, and no internal squads.

## Failure and reconciliation

Authentication failures return the same unauthenticated envelope whether data is malformed, stale, or signed incorrectly. A Telegram outage does not infer membership from a browser claim.

Username reservation and upstream provisioning are deliberately resumable. If local reservation succeeds but Remnawave is unavailable, the account remains at the agreement/provisioning boundary. Retrying first reconciles by service username and Telegram ID and adopts only an unambiguous matching upstream user. It never creates a second local identity or silently picks a conflicting upstream user.

Invite creation/approval failures leave local membership false. Revocation must succeed before the invite is marked used, so a replay can safely complete cleanup without approving another identity. An admin can bootstrap configuration before completing ordinary membership onboarding, but all user product routes still require complete onboarding.

## Verification

- Table-driven init-data tests cover canonical sorting, percent encoding, missing fields, bad hashes, constant-time comparison path, stale/future timestamps, and accepted boundary age.
- Repository tests race two users for one username and confirm exactly one reservation.
- Telegram fake-server tests cover invite expiry, wrong-user requests, duplicate join events, partial group/channel membership, and canonical recheck.
- Remnawave tests cover the exact initial payload, `A019`, timeouts after upstream creation, reconciliation by both identifiers, and ambiguous matches.
- Handler tests verify cookie flags, session expiry/revocation, resumable transitions, admin bootstrap, onboarding guards, and redacted logs.
