# Onboarding components

- `OnboardingFlow.vue` coordinates intro, membership, username, and agreement steps.
- `IntroSequence.vue`, `MembershipPanel.vue`, `UsernamePanel.vue`, and `AgreementPanel.vue` render each step.
- `agreementIcons.ts` maps server-owned icon keys to external Iconify names.
- `useOnboardingMainButton.ts` mirrors the decisive step action to Telegram's native MainButton with an in-app fallback.
- `useOnboardingMainButton.test.ts` verifies native-button synchronization and teardown.
