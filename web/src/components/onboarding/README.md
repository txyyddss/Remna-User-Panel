# Onboarding components

- `OnboardingFlow.vue` coordinates intro, username, and agreement steps.
- `IntroSequence.vue`, `UsernamePanel.vue`, and `AgreementPanel.vue` render each step; agreement cards render server-authored color and an optional localized page title.
- `agreementIcons.ts` maps server-owned icon keys to external Iconify names.
- `useOnboardingMainButton.ts` mirrors the decisive step action to Telegram's native MainButton with an in-app fallback.
- `useOnboardingMainButton.test.ts` verifies native-button synchronization and teardown.

Onboarding distinguishes soft navigation/opening, medium native actions, rigid completion, and semantic agreement selection.
