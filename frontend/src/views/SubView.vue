<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const loading = ref(true)
const showKeys = ref(false)
const copied = ref(false)
const subInfo = computed(() => userStore.liveSubInfo)
const keys = computed(() => userStore.subKeys)

const usedPercent = computed(() => {
  if (!subInfo.value?.user?.trafficLimitBytes) return 0
  return Math.min(100, (subInfo.value.user.usedTrafficBytes / subInfo.value.user.trafficLimitBytes) * 100)
})

function formatBytes(b: number): string {
  if (b < 1073741824) return `${(b / 1048576).toFixed(2)} MB`
  return `${(b / 1073741824).toFixed(2)} GB`
}

async function copySubUrl() {
  if (keys.value?.subscription_url) {
    await navigator.clipboard.writeText(keys.value.subscription_url)
    copied.value = true
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    setTimeout(() => copied.value = false, 2000)
  }
}

onMounted(async () => {
  await userStore.refreshState({ background: true })
  loading.value = false
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">📡 My Subscription</h1>
    </div>

    <div class="loading-page" v-if="loading">
      <div class="loading-spinner"></div>
    </div>

    <template v-else-if="subInfo?.has_subscription">
      <!-- Status Card -->
      <div class="card">
        <div class="row-between mb-sm">
          <h3>Subscription Overview</h3>
          <span class="badge" :class="{
            'badge-success': subInfo.user.status === 'ACTIVE',
            'badge-warning': subInfo.user.status === 'LIMITED',
            'badge-danger': subInfo.user.status === 'DISABLED'
          }">{{ subInfo.user.status }}</span>
        </div>

        <div class="stat-row">
          <div class="stat-item">
            <span class="stat-value text-sm">{{ formatBytes(subInfo.user.usedTrafficBytes) }}</span>
            <span class="stat-label">Traffic Used</span>
          </div>
          <div class="stat-item">
            <span class="stat-value text-sm">{{ subInfo.user.trafficLimitBytes ? formatBytes(subInfo.user.trafficLimitBytes) : '♾️' }}</span>
            <span class="stat-label">Total Traffic</span>
          </div>
        </div>

        <div class="progress mt-md">
          <div class="progress-bar" :class="{ 'warning': usedPercent > 70, 'danger': usedPercent > 90 }" :style="{ width: usedPercent + '%' }"></div>
        </div>
        <div class="row-between text-xs text-muted mt-sm">
          <span>{{ usedPercent.toFixed(1) }}%</span>
          <span>Expires: {{ new Date(subInfo.user.expireAt).toLocaleDateString('en-US') }}</span>
        </div>
      </div>

      <!-- Connection Keys -->
      <div class="card mt-md" v-if="keys">
        <div class="row-between mb-sm">
          <h3>🔑 Connection Info</h3>
          <button class="btn btn-sm btn-secondary" @click="showKeys = !showKeys">
            {{ showKeys ? 'Hide' : 'Show' }}
          </button>
        </div>

        <div v-if="showKeys" class="key-section">
          <div class="key-item">
            <span class="text-sm text-muted">Subscription URL</span>
            <div class="key-value" @click="copySubUrl">
              <code class="mono text-xs">{{ keys.subscription_url }}</code>
              <span class="copy-icon">{{ copied ? '✅' : '📋' }}</span>
            </div>
          </div>

          <div class="key-item mt-sm">
            <span class="text-sm text-muted">Short UUID</span>
            <code class="mono text-sm">{{ keys.short_uuid }}</code>
          </div>
        </div>
      </div>

      <!-- Quick Actions -->
      <div class="grid-2 mt-md">
        <router-link to="/info" class="action-card card">
          <span>📊</span> <span class="text-sm">Traffic Details</span>
        </router-link>
        <router-link to="/ip" class="action-card card">
          <span>🔄</span> <span class="text-sm">Change IP</span>
        </router-link>
        <router-link to="/squads" class="action-card card">
          <span>🌐</span> <span class="text-sm">Switch Squad</span>
        </router-link>
        <router-link to="/combos" class="action-card card">
          <span>🔄</span> <span class="text-sm">Renew/Change</span>
        </router-link>
      </div>
    </template>

    <div class="empty-state" v-else>
      <span class="empty-state-icon">📡</span>
      <p class="empty-state-text">No subscription</p>
      <router-link to="/combos" class="btn btn-primary mt-md">Browse Combos</router-link>
    </div>
  </div>
</template>

<style scoped>
.stat-row {
  display: flex;
  gap: var(--space-lg);
  margin-top: var(--space-md);
}

.stat-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.key-section {
  background: var(--bg-glass);
  border-radius: var(--radius-md);
  padding: var(--space-md);
}

.key-item {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
}

.key-value {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  cursor: pointer;
  padding: var(--space-sm);
  background: var(--bg-glass);
  border-radius: var(--radius-sm);
  word-break: break-all;
}

.key-value:hover {
  background: var(--bg-glass-strong);
}

.copy-icon {
  flex-shrink: 0;
}

.action-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  text-decoration: none;
  color: var(--text-primary);
}
</style>
