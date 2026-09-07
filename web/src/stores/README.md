# Stores

The session store binds the response cache to the authenticated user, role and
onboarding state. A new identity starts the bounded member/admin preload queue;
clearing or failing authentication cancels preload and discards cached responses.

- `session.ts` owns authenticated session bootstrap and derived user state.
