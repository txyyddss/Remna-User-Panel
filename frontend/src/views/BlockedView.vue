<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const isGroupBlock = computed(() => route.query.reason === 'group')
const title = computed(() => (isGroupBlock.value ? 'Group Membership Required' : 'Telegram Mini App Only'))
const message = computed(() =>
  isGroupBlock.value
    ? 'Join the required Telegram group first, then reopen the mini app.'
    : 'Open this panel from the Telegram mini app to continue.',
)
</script>

<template>
  <div class="blocked-page">
    <div class="blocked-card card">
      <div class="blocked-kicker">{{ isGroupBlock ? 'Access Policy' : 'Launch Required' }}</div>
      <h1 class="blocked-title">{{ title }}</h1>
      <p class="blocked-text">{{ message }}</p>
      <p class="blocked-hint">If you already joined the group, close the mini app and open it again from Telegram.</p>
    </div>
  </div>
</template>

<style scoped>
.blocked-page {
  min-height: 100dvh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-lg);
}

.blocked-card {
  max-width: 420px;
  text-align: center;
}

.blocked-kicker {
  margin-bottom: var(--space-sm);
  color: var(--accent-secondary);
  font-size: 0.75rem;
  letter-spacing: 0.2em;
  text-transform: uppercase;
}

.blocked-title {
  margin-bottom: var(--space-md);
}

.blocked-text {
  color: var(--text-secondary);
  margin-bottom: var(--space-sm);
}

.blocked-hint {
  color: var(--text-muted);
  font-size: 0.875rem;
}
</style>
