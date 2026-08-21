# Squad profile components

- `profile.ts` owns localized profile type labels, icons, and ISO country options.
- `CarrierLogo.vue` renders the bundled China Telecom, China Unicom, and China Mobile route marks with accessible labels.
- `SquadProfileSummary.vue` renders compact generated facts, a caller-owned facts slot, and safe extra Markdown. Its flat member presentation accepts the squad name and uses the profile-specific icon and restrained semantic color treatment; China route facts use carrier marks without abbreviation prefixes. Catalog occupancy uses the facts slot so its percentage inherits the existing port-speed tag styling.
- `profile.test.ts` checks locale-aware ISO country options.
- `SquadProfileSummary.test.ts` checks generated facts, unlimited ports, and Markdown.

The summary is a pure projection of the typed API profile. It does not invent
provider data or persist a second generated description.
