# Squad profile components

- `profile.ts` owns localized profile type labels, icons, and ISO country options.
- `CarrierLogo.vue` renders the bundled China Telecom, China Unicom, and China Mobile route marks with accessible labels.
- `SquadProfileSummary.vue` renders compact generated facts, caller-owned heading/facts slots, and safe extra Markdown. Its member presentation combines localized type and port speed beneath the squad name, enlarges the unframed profile icon, and suppresses redundant location facts; China routes keep carrier marks without abbreviation prefixes.
- `profile.test.ts` checks locale-aware ISO country options.
- `SquadProfileSummary.test.ts` checks generated facts, unlimited ports, and Markdown.

The summary is a pure projection of the typed API profile. It does not invent
provider data or persist a second generated description.
