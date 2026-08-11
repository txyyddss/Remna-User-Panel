# Route views

- `ActivityView.vue` delegates to games, check-in, draws, and activity history.
- `AdminView.vue` resolves and lazy-loads the selected administrative section.
- `CatalogView.vue` delegates to plan browsing and checkout.
- `EmbyView.vue` delegates to Emby setup and preference management.
- `EntryView.vue` selects the initial authenticated destination.
- `HomeView.vue` delegates to the member dashboard, including balance funding.
- `OnboardingView.vue` hosts the resumable membership and agreement flow.
- `PaymentResultView.vue` verifies a provider-returned payment order, returns confirmed payments to Home, and offers failed/expired orders a localized reissue handoff.
- `QuestionnaireView.vue` delegates to active questionnaire participation.
