<script setup lang="ts">
</script>

<template>
  <div class="blocked-page">
    <div class="blocked-content">
      <div class="blocked-icon">🔒</div>
      <h1 class="blocked-title">Access Restricted</h1>
      <p class="blocked-text">This panel is only accessible via the Telegram Mini App</p>
      <p class="blocked-hint">Please open the bot in Telegram to use panel features</p>
    </div>
  </div>
</template>

<style scoped>
.blocked-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100dvh;
  padding: var(--space-lg);
}

.blocked-content {
  text-align: center;
}

.blocked-icon {
  font-size: 4rem;
  margin-bottom: var(--space-lg);
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.1); }
}

.blocked-title {
  font-family: var(--font-display);
  font-size: 1.5rem;
  margin-bottom: var(--space-md);
}

.blocked-text {
  color: var(--text-secondary);
  margin-bottom: var(--space-sm);
}

.blocked-hint {
  color: var(--text-muted);
  font-size: 0.8125rem;
}
</style>
