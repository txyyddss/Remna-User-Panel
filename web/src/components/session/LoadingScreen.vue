<template>
  <main class="auth-screen" aria-live="polite" aria-busy="true">
    <p class="auth-screen__label">{{ $t('auth.securingRide') }}</p>
    <div class="auth-screen__skeleton" aria-hidden="true">
      <USkeleton />
      <USkeleton />
      <USkeleton />
    </div>
  </main>
</template>
