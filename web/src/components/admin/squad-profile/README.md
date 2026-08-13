# Admin squad profile editor

- `AdminSquadProfileEditor.vue` owns the type selector, validation state, and preview.
- `BroadbandProfileFields.vue` edits ISP, port, access mode, and detailed location.
- `ChinaOptimizedProfileFields.vue` edits CT, CU, CM, port, and country.
- `InternationalNetworkProfileFields.vue` edits port, country, and upstream carriers.
- `SquadPortField.vue` provides the shared Mbps and unlimited-port control.
- `profileForm.ts` maps API profiles to the compact editor draft and back.
- `profileForm.test.ts` checks type-specific mapping, unlimited ports, and required fields.

The editor emits only a complete typed profile. Incomplete drafts remain local
to the form and cannot be submitted to the admin API.
