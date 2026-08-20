# Affiliate administration

- `AdminAffiliatesPanel.vue` owns the versioned settings form and save flow.
- `AdminAffiliateTierEditor.vue` edits one tier with accessible reorder controls.
- `affiliateForm.ts` mirrors server tier and reward-union validation with Zod.
- `useAdminAffiliates.ts` loads coupon references and handles optimistic-version saves.
- `AdminAffiliateTierEditor.test.ts` covers conditional reward-union fields.
