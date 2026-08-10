<template>
  <main class="auth-screen" aria-live="polite" aria-busy="true">
    <div class="brand-mark brand-mark--large" aria-hidden="true">
      <UIcon name="i-ph-circles-four-fill" />
    </div>
    <p class="auth-screen__label">{{ $t('auth.securingRide') }}</p>
    <div class="auth-screen__skeleton" aria-hidden="true">
      <USkeleton />
      <USkeleton />
      <USkeleton />
    </div>
  </main>
</template>
