<template>
  <main class="auth-screen" aria-live="polite" aria-busy="true">
    <p class="auth-screen__status sr-only">{{ $t('auth.securingRide') }}</p>
    <div class="auth-screen__drive" aria-hidden="true">
      <div class="auth-screen__road">
        <span class="auth-screen__road-line" />
        <span class="auth-screen__road-line" />
        <span class="auth-screen__road-line" />
      </div>
      <div class="auth-screen__car">
        <span class="auth-screen__car-cabin">
          <span class="auth-screen__car-window auth-screen__car-window--rear" />
          <span class="auth-screen__car-window auth-screen__car-window--front" />
        </span>
        <span class="auth-screen__car-shell" />
        <span class="auth-screen__wheel auth-screen__wheel--rear" />
        <span class="auth-screen__wheel auth-screen__wheel--front" />
      </div>
    </div>
  </main>
</template>
