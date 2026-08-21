# Composables

- `useActivity.ts` loads games, draws, and check-in actions.
- `useAffiliateCentre.ts` loads member metrics and fixed five-row referral pages.
- `useAdminDatabase.ts` coordinates protected database review and mutation flows.
- `useAdminSection.ts` provides latest-request-safe administrator list operations and cursor accumulation.
- `useCatalog.ts` prepares catalog selection, quote freshness, full paid add-on guards, automatic-renewal catalog blocking, and user-scoped session drafts.
- `catalogQuoteState.ts` owns the request-versioned authoritative catalog quote and its selection fingerprint.
- `catalogDraft.ts` reads, validates through callers, writes, and clears user-scoped catalog drafts in session storage.
- `catalogLoader.ts` loads catalog, balance, coupon wallet, and validated user-scoped drafts while reporting whether the latest response was applied.
- `useCatalog.test.ts` covers catalog selection behavior.
- `useCouponRedemption.ts` redeems a coupon code for the guided purchase flow.
- `useCouponRedemption.test.ts` covers guided redemption normalization and idempotency reuse.
- `useAutoRenewal.ts` owns the owner-scoped automatic-renewal status load, eligibility, and server toggle update.
- `useAutoRenewal.test.ts` covers the server-ineligible toggle constraint.
- `useClipboard.ts` provides safe clipboard feedback.
- `useConnectionScan.ts` creates one idempotent provider scan and polls its owner-scoped progress every second without repeating an ambiguous start mutation.
- `useConnectionScan.test.ts` covers accepted scan polling and ambiguous-start idempotency reuse.
- `useConnectionBlocks.ts` loads active blocks, retains per-target block/unblock idempotency keys across ambiguous failures, and delegates receipt polling without replaying mutations.
- `useConnectionBlocks.test.ts` covers key reuse, unblock polling, and active-list refresh.
- `useCoupons.ts` manages coupon-wallet redemption and server-confirmed grant discards.
- `useCoupons.test.ts` covers confirmed and failed coupon-wallet discards.
- `useDashboard.ts` loads dashboard summaries and provides disposal-safe bounded per-node UTC traffic state to its Home descendants.
- `useRolloverDetail.ts` fetches fresh aggregate rollover details when the active Home ride opens, with cancellation-safe retry/reset state and localized errors.
- `useEmby.ts` manages Emby setup and preferences.
- `useImageZoom.ts` owns bounded focal-point pinch, pan, wheel, double-tap, button, interaction, and reset state for image canvases.
- `useImageZoom.test.ts` covers zoom bounds, reset, and two-finger gesture state.
- `useIntroSequence.ts` drives the onboarding intro sequence.
- `useIntroSequence.test.ts` covers intro timing and completion.
- `onboardingState.ts` resolves the next onboarding step from the server-owned
  membership response; `onboardingState.test.ts` covers returning-user state.
- `useOnboarding.ts` coordinates onboarding steps and persistence, including
  refreshing the server-owned state before moving a returning user through a
  revised agreement.
- `usePaymentOrder.ts` coordinates durable checkout/cancellation receipts with payment-sheet presentation states.
- `usePaymentConfiguration.ts` owns exact TXB bounds, amount conversion, and selected payment method state.
- `usePaymentOrderOperations.ts` retains stable mutation keys and polls payment provider-operation receipts without replaying provider calls.
- `usePaymentTarget.ts` opens accepted provider targets and read-only polls authoritative payment status.
- `paymentOrderHelpers.ts` validates exact payment bounds and replacement candidates, classifies terminal orders, and renders provider QR payloads.
- `usePaymentOrder.test.ts` covers payment-order state transitions.
- `usePaymentReturn.ts` polls a provider-returned payment order until it is confirmed or terminal.
- `usePaymentReturn.test.ts` covers durable payment-return confirmation and pending polling.
- `useOperationReceipt.ts` performs bounded, read-only polling for durable member operation receipts and stops on every terminal or review state.
- `useDurableCommand.ts` retains stable UUID keys for ambiguous command retries, invalidates disposed submissions, and delegates accepted receipt polling to `useOperationReceipt.ts`.
- `usePurchaseOperations.ts` owns authoritative paid-reset/refund quotes, the account-wide reset automation load/save state, semantic feedback for actual preference changes and outcomes, conflict refresh, mutation idempotency, and receipt state for an active purchase.
- `usePurchaseOperations.test.ts` covers quote conflict refresh, idempotency reuse, server-owned refund eligibility, and immediate automation persistence.
- `useQuestionnaireImport.ts` previews and settles questionnaire imports with reset-safe mutation and polling generations.
- `useQuestionnaireImport.test.ts` covers import analysis and settlement.
- `useQuestionnaires.ts` manages member questionnaire participation.
- `useNodeGeocheck.ts` loads an image-only cached node Geocheck for a minimal UUID/name target shared by Statistics and Catalog, only after explicit selection, and clears its state on close.
- `useNodeGeocheck.test.ts` covers selected-target loading, localized unavailability, and close reset behavior.
- `useStatistics.ts` independently loads the 30-minute aggregate snapshot and
  ten-second node cache while preserving last-good data across partial failures.
- `useTelegramBackButton.ts` coordinates one native Telegram BackButton across route and overlay owners, adds soft navigation feedback, and removes every handler on teardown.
- `useTelegramProtection.ts` reference-counts active destructive, busy, or dirty workflows and enables native closing confirmation plus vertical-swipe suppression only while needed.
