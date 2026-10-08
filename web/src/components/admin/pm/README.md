# Private message administration

`AdminPMPanel.vue` uses the existing paged admin inventory and durable command
tracker for conversation search, block/mute preferences and queued profile updates.
Rows are flat, mobile first, with a desktop grid. Data and all user copy come from
typed API contracts and English/Chinese localization.

`PMDeliveryDialog.vue` loads recent reference-only receipts on demand, cancels
stale reads and surfaces ambiguous delivery without offering an unsafe resend.
`PMTopicRepairDialog.vue` validates decimal Telegram references with Zod and queues
an explicit probe of an existing topic/card. `links.ts` constructs documented forum
and message links; Telegram handles private-group membership at link opening.
`modal.css` suppresses opening/closing transitions when the existing motion preference is reduced.

The backend remains authoritative for panel-admin permissions, recipient eligibility,
moderation, topic uniqueness and safe recovery. Frontend status resets start a new
explicit action; they never replay a previous uncertain message. Dialogs reuse the
existing Nuxt UI modal, notices and motion/accessibility behavior.
