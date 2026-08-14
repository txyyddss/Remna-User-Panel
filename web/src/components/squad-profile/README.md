# Squad profile components

- `profile.ts` owns localized profile type labels, icons, and ISO country options.
- `CarrierLogo.vue` renders the bundled China Telecom, China Unicom, and China Mobile route marks with accessible labels.
- `SquadProfileSummary.vue` renders compact generated facts and safe extra Markdown. Its flat member presentation accepts the squad name and uses the profile-specific icon and restrained semantic color treatment; China route facts use carrier marks without abbreviation prefixes. Standalone catalog cards own the matching colored surface, while the default presentation remains suitable for administration.
- `profile.test.ts` checks locale-aware ISO country options.
- `SquadProfileSummary.test.ts` checks generated facts, unlimited ports, and Markdown.

The summary is a pure projection of the typed API profile. It does not invent
provider data or persist a second generated description.
