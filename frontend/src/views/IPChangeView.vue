<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api } from '@/api'
import { useUserStore } from '@/stores/user'
import { useToast } from '@/composables/useToast'
import type { IPChangeStatus } from '@/types'

const userStore = useUserStore()
const toast = useToast()
const loading = ref(true)
const changing = ref(false)
const error = ref('')

const status = ref<IPChangeStatus | null>(null)

const timeLeft = computed(() => {
  if (!status.value?.next_available) return ''
  const diff = new Date(status.value.next_available).getTime() - Date.now()
  if (diff <= 0) return ''
  const h = Math.floor(diff / 3600000)
  const m = Math.floor((diff % 3600000) / 60000)
  return `${h}h ${m}m`
})

const progressPercent = computed(() => {
  if (!status.value || status.value.can_change) return 100
  const cd = status.value.cooldown_hours * 3600000
  const elapsed = Date.now() - new Date(status.value.last_change).getTime()
  return Math.min(100, (elapsed / cd) * 100)
})

async function loadStatus() {
  try {
    status.value = await api.getIPStatus()
  } catch {
    status.value = { can_change: true, last_change: '', next_available: '', cooldown_hours: 6 }
  } finally {
    loading.value = false
  }
}

async function doIPChange() {
  changing.value = true
  error.value = ''
  try {
    await api.changeIP()
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    toast.success('Connection dropped. Reconnect to get a new IP.')
    await loadStatus()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : 'Failed to change IP'
    error.value = msg
    toast.error(msg)
  } finally {
    changing.value = false
  }
}

onMounted(loadStatus)
</script>

<template>
  <div class="page">
    <div class="page-header stagger-enter stagger-1">
      <h1 class="page-title">Change IP</h1>
      <p class="page-subtitle">Drop your active connection to rotate the exit IP address.</p>
    </div>

    <div class="loading-page" v-if="loading">
      <div class="loading-spinner"></div>
    </div>

    <template v-else>
      <!-- Status Ring -->
      <div class="ip-hero stagger-enter stagger-2">
        <div
          class="ip-ring"
          :class="{ ready: status?.can_change, cooldown: !status?.can_change }"
        >
          <div class="ip-ring-inner">
            <span v-if="status?.can_change" class="ip-ring-label gradient-text">READY</span>
            <template v-else>
              <span class="ip-ring-time">{{ timeLeft }}</span>
              <span class="ip-ring-sublabel">cooldown</span>
            </template>
          </div>
        </div>
      </div>

      <!-- Action -->
      <div class="ip-action stagger-enter stagger-3">
        <button
          class="btn btn-primary btn-lg btn-block"
          :disabled="!status?.can_change || changing"
          @click="doIPChange"
        >
          {{ changing ? 'Dropping connection...' : status?.can_change ? 'Drop Connection' : 'On Cooldown' }}
        </button>
      </div>

      <!-- Cooldown Progress -->
      <div v-if="!status?.can_change" class="card stagger-enter stagger-4">
        <div class="row-between mb-sm">
          <span class="text-sm text-muted">Cooldown Progress</span>
          <span class="text-sm mono">{{ progressPercent.toFixed(0) }}%</span>
        </div>
        <div class="progress">
          <div class="progress-bar" :style="{ width: `${progressPercent}%` }"></div>
        </div>
        <p class="text-xs text-muted mt-sm">
          Next available: {{ status?.next_available ? new Date(status.next_available).toLocaleString('en-US') : '—' }}
        </p>
      </div>

      <!-- Error -->
      <div v-if="error" class="card stagger-enter stagger-5 error-card">
        <span class="text-sm text-danger">{{ error }}</span>
      </div>

      <!-- Info Card -->
      <div class="card stagger-enter stagger-5">
        <h3 class="mb-sm">How It Works</h3>
        <ul class="info-list">
          <li>Clicking "Drop Connection" disconnects all active VPN sessions.</li>
          <li>Reconnect through your proxy client to get a new exit IP.</li>
          <li>Cooldown period is <strong>{{ status?.cooldown_hours || 6 }} hours</strong> between changes.</li>
          <li>Frequently used by users who need to bypass regional IP blocks.</li>
        </ul>
      </div>
    </template>
  </div>
</template>

<style scoped>
.ip-hero {
  display: flex;
  justify-content: center;
  padding: var(--space-xl) 0;
}

.ip-ring {
  width: 160px;
  height: 160px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
}

.ip-ring::before {
  content: '';
  position: absolute;
  inset: 0;
  border-radius: 50%;
  padding: 3px;
  background: conic-gradient(
    from 0deg,
    var(--accent-primary),
    var(--accent-secondary),
    var(--accent-primary)
  );
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  mask-composite: exclude;
}

.ip-ring.ready::before {
  animation: ringRotate 3s linear infinite;
}

.ip-ring.cooldown::before {
  background: conic-gradient(
    from 0deg,
    var(--text-muted) 0%,
    var(--accent-warning) 50%,
    var(--text-muted) 100%
  );
  opacity: 0.5;
}

@keyframes ringRotate {
  to {
    transform: rotate(360deg);
  }
}

.ip-ring-inner {
  width: calc(100% - 8px);
  height: calc(100% - 8px);
  border-radius: 50%;
  background: var(--bg-secondary);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.ip-ring-label {
  font-family: var(--font-display);
  font-size: 1.25rem;
  font-weight: 700;
}

.ip-ring-time {
  font-family: var(--font-display);
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--accent-warning);
}

.ip-ring-sublabel {
  font-size: 0.6875rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.1em;
}

.ip-action {
  margin-bottom: var(--space-lg);
}

.error-card {
  border-color: rgba(255, 107, 122, 0.3);
  background: rgba(255, 107, 122, 0.06);
}

.info-list {
  list-style: none;
  padding: 0;
}

.info-list li {
  position: relative;
  padding-left: var(--space-lg);
  font-size: 0.8125rem;
  color: var(--text-secondary);
  line-height: 1.6;
}

.info-list li + li {
  margin-top: var(--space-sm);
}

.info-list li::before {
  content: '→';
  position: absolute;
  left: 0;
  color: var(--accent-primary);
  font-weight: 600;
}
</style>
