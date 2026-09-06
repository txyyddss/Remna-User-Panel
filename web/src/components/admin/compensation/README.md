# Compensation components

- `CompensationConfigCard.vue` edits the revisioned nullable observation policy, with shrinkable grid columns so number fields fit 320px phones.
- `CompensationEventCard.vue` presents recipient-safe outage facts and operation state; long node and squad labels wrap within the card.
- `CompensationReviewModal.vue` keeps snapshots read-only while collecting minutes and reason.
- `format.ts` owns basis-point and observed-duration display conversion.
- `format.test.ts` covers exact factor formatting and duration flooring.
- `statusFilter.ts` maps the non-empty Nuxt UI selection sentinel to the API's empty status query.
- `statusFilter.test.ts` prevents empty select items from reintroducing a render failure.
