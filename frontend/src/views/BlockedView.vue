<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const isGroupBlock = computed(() => route.query.reason === 'group')
const title = computed(() =>
  isGroupBlock.value ? 'Group Membership Required' : 'Telegram Mini App Only',
)
const message = computed(() =>
  isGroupBlock.value
    ? 'Join the required Telegram group first, then reopen the mini app.'
    : 'Open this panel from the Telegram mini app to continue.',
)
</script>

<template>
  <div class="blocked-page">
    <div class="blocked-orb"></div>
    <div class="blocked-card card stagger-enter stagger-1">
      <div class="blocked-kicker">{{ isGroupBlock ? 'Access Policy' : 'Launch Required' }}</div>
      <h1 class="blocked-title gradient-text">{{ title }}</h1>
      <p class="blocked-text">{{ message }}</p>
      <p class="blocked-hint">
        If you already joined the group, close the mini app and open it again from Telegram.
      </p>
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
  position: relative;
  overflow: hidden;
}

.blocked-orb {
  position: absolute;
  width: 280px;
  height: 280px;
  border-radius: 50%;
  background: radial-gradient(circle, rgba(var(--accent-primary-rgb), 0.15), transparent 70%);
  filter: blur(60px);
  top: 20%;
  left: 50%;
  transform: translateX(-50%);
  animation: orbFloat 6s ease-in-out infinite alternate;
}

@keyframes orbFloat {
  0% { transform: translateX(-50%) translateY(0) scale(1); }
  100% { transform: translateX(-50%) translateY(-20px) scale(1.1); }
}

.blocked-card {
  max-width: 420px;
  text-align: center;
  position: relative;
  z-index: 1;
}

.blocked-kicker {
  margin-bottom: var(--space-sm);
  color: var(--accent-secondary);
  font-size: 0.75rem;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  font-weight: 600;
}

.blocked-title {
  font-size: 1.5rem;
  margin-bottom: var(--space-md);
}

.blocked-text {
  color: var(--text-secondary);
  margin-bottom: var(--space-sm);
  font-size: 0.875rem;
}

.blocked-hint {
  color: var(--text-muted);
  font-size: 0.8125rem;
}
</style>
