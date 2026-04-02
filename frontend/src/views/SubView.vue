<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const loading = ref(true)
const showKeys = ref(false)
const copied = ref(false)

const subInfo = computed(() => userStore.liveSubInfo)
const keys = computed(() => userStore.subKeys)

const usedTraffic = computed(() => subInfo.value?.user?.userTraffic?.usedTrafficBytes ?? subInfo.value?.user?.usedTrafficBytes ?? 0)
const lifetimeTraffic = computed(() => subInfo.value?.user?.userTraffic?.lifetimeUsedTrafficBytes ?? subInfo.value?.user?.lifetimeUsedTrafficBytes ?? 0)
const trafficLimit = computed(() => subInfo.value?.user?.trafficLimitBytes ?? 0)
const activeSquads = computed(() => subInfo.value?.user?.activeInternalSquads ?? [])

const usedPercent = computed(() => {
  if (!trafficLimit.value) {
    return 0
  }
  return Math.min(100, (usedTraffic.value / trafficLimit.value) * 100)
})

function formatBytes(bytes: number): string {
  if (!bytes) {
    return '0 B'
  }
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  return `${(bytes / 1024 ** index).toFixed(index === 0 ? 0 : 2)} ${units[index]}`
}

async function copyValue(value?: string) {
  if (!value) {
    return
  }
  await navigator.clipboard.writeText(value)
  copied.value = true
  window.setTimeout(() => {
    copied.value = false
  }, 1800)
}

onMounted(async () => {
  await userStore.refreshState({ background: true })
  loading.value = false
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">Subscription</h1>
      <p class="page-subtitle">Connection keys, plan status, and usage details from Remnawave.</p>
    </div>

    <div class="loading-page" v-if="loading">
      <div class="loading-spinner"></div>
    </div>

    <template v-else-if="subInfo?.has_subscription">
      <div class="card">
        <div class="row-between mb-sm">
          <h3>Overview</h3>
          <span class="badge" :class="{
            'badge-success': subInfo.user.status === 'ACTIVE',
            'badge-warning': subInfo.user.status === 'LIMITED',
            'badge-danger': subInfo.user.status === 'DISABLED' || subInfo.user.status === 'EXPIRED',
          }">
            {{ subInfo.user.status }}
          </span>
        </div>

        <div class="stat-grid">
          <div class="metric">
            <span class="stat-value text-sm">{{ formatBytes(usedTraffic) }}</span>
            <span class="stat-label">Used</span>
          </div>
          <div class="metric">
            <span class="stat-value text-sm">{{ trafficLimit ? formatBytes(trafficLimit) : 'Unlimited' }}</span>
            <span class="stat-label">Limit</span>
          </div>
          <div class="metric">
            <span class="stat-value text-sm">{{ formatBytes(lifetimeTraffic) }}</span>
            <span class="stat-label">Lifetime</span>
          </div>
        </div>

        <div class="progress mt-md">
          <div class="progress-bar" :class="{ warning: usedPercent >= 70, danger: usedPercent >= 90 }" :style="{ width: `${usedPercent}%` }"></div>
        </div>

        <div class="row-between text-xs text-muted mt-sm">
          <span>{{ usedPercent.toFixed(1) }}%</span>
          <span>Expires {{ new Date(subInfo.user.expireAt).toLocaleString('en-US') }}</span>
        </div>

        <div class="combo-specs mt-md">
          <span v-for="squad in activeSquads" :key="squad.uuid" class="spec-item">{{ squad.name }}</span>
          <span v-if="subInfo.user.trafficLimitStrategy" class="spec-item">{{ subInfo.user.trafficLimitStrategy }}</span>
        </div>
      </div>

      <div class="card mt-md" v-if="keys">
        <div class="row-between">
          <div>
            <h3>Connection Keys</h3>
            <p class="text-sm text-muted mt-sm">Show the subscription URL and manual connection values.</p>
          </div>
          <button class="btn btn-secondary btn-sm" @click="showKeys = !showKeys">{{ showKeys ? 'Hide' : 'Show' }}</button>
        </div>

        <div v-if="showKeys" class="stack-sm mt-md">
          <button class="key-card" @click="copyValue(keys.subscription_url)">
            <span class="key-label">Subscription URL</span>
            <code class="key-value">{{ keys.subscription_url }}</code>
          </button>

          <button class="key-card" v-if="keys.vless_uuid" @click="copyValue(keys.vless_uuid)">
            <span class="key-label">VLESS UUID</span>
            <code class="key-value">{{ keys.vless_uuid }}</code>
          </button>

          <button class="key-card" v-if="keys.trojan_password" @click="copyValue(keys.trojan_password)">
            <span class="key-label">Trojan Password</span>
            <code class="key-value">{{ keys.trojan_password }}</code>
          </button>

          <button class="key-card" v-if="keys.ss_password" @click="copyValue(keys.ss_password)">
            <span class="key-label">Shadowsocks Password</span>
            <code class="key-value">{{ keys.ss_password }}</code>
          </button>

          <div class="text-xs text-success" v-if="copied">Copied to clipboard.</div>

          <div class="card instruction-card">
            <h4 class="mb-sm">Client Instructions</h4>
            <ul class="instruction-list">
              <li v-for="instruction in keys.instructions" :key="instruction">{{ instruction }}</li>
            </ul>
          </div>
        </div>
      </div>

      <div class="grid-2 mt-md">
        <router-link to="/info" class="action-card card">
          <strong>Usage</strong>
          <span class="text-sm text-muted">Bandwidth, devices, and history</span>
        </router-link>
        <router-link to="/ip" class="action-card card">
          <strong>Change IP</strong>
          <span class="text-sm text-muted">Standalone reconnect tool</span>
        </router-link>
        <router-link to="/squads" class="action-card card">
          <strong>Route Group</strong>
          <span class="text-sm text-muted">Switch external route group</span>
        </router-link>
        <router-link to="/combos" class="action-card card">
          <strong>Plans</strong>
          <span class="text-sm text-muted">Renew or switch subscription</span>
        </router-link>
      </div>
    </template>

    <div class="empty-state" v-else>
      <p class="empty-state-text">No subscription is currently bound to this account.</p>
      <router-link to="/combos" class="btn btn-primary mt-md">Browse Plans</router-link>
    </div>
  </div>
</template>

<style scoped>
.stat-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: var(--space-sm);
  margin-top: var(--space-md);
}

.metric {
  padding: var(--space-sm);
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.03);
}

.combo-specs {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
}

.spec-item {
  padding: 6px 10px;
  border-radius: 999px;
  background: var(--bg-glass-strong);
  color: var(--text-secondary);
  font-size: 0.75rem;
}

.key-card {
  width: 100%;
  text-align: left;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--space-md);
  background: rgba(255, 255, 255, 0.02);
  color: inherit;
}

.key-label {
  display: block;
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.12em;
  color: var(--text-muted);
  margin-bottom: var(--space-xs);
}

.key-value {
  display: block;
  white-space: normal;
  word-break: break-all;
}

.instruction-card {
  background: rgba(34, 197, 94, 0.06);
  border-color: rgba(34, 197, 94, 0.16);
}

.instruction-list {
  padding-left: 1rem;
  color: var(--text-secondary);
}

.instruction-list li + li {
  margin-top: var(--space-xs);
}

.action-card {
  text-decoration: none;
  color: inherit;
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}
</style>
