<script setup lang="ts">
import { ref, computed } from 'vue'
import { useUserStore } from '@/stores/user'
import { api } from '@/api'
import { formatBytes } from '@/utils/format'

const userStore = useUserStore()
const subUrl = ref('')
const bindLoading = ref(false)
const bindMessage = ref('')
const subInfo = computed(() => userStore.liveSubInfo)
const loading = computed(() => userStore.loading && !userStore.user)

const usedPercent = computed(() => {
  if (!subInfo.value?.user?.trafficLimitBytes) return 0
  return Math.min(100, (subInfo.value.user.usedTrafficBytes / subInfo.value.user.trafficLimitBytes) * 100)
})

const daysRemaining = computed(() => {
  if (!subInfo.value?.user?.expireAt) return 0
  const diff = new Date(subInfo.value.user.expireAt).getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / 86400000))
})


async function bindSub() {
  if (!subUrl.value) return
  bindLoading.value = true
  bindMessage.value = ''
  try {
    const resp = await api.bindSubscription(subUrl.value)
    bindMessage.value = `✅ Binding successful! User: ${resp.rw_user}`
    subUrl.value = ''
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    await userStore.refreshState()
  } catch (e: any) {
    bindMessage.value = '❌ ' + e.message
  }
  bindLoading.value = false
}

</script>

<template>
  <div class="page">
    <div class="page-header">
      <div class="greeting">
        <span class="greeting-emoji">👋</span>
        <div>
          <h1 class="greeting-name">{{ userStore.telegramName }}</h1>
          <p class="page-subtitle">Welcome back</p>
        </div>
      </div>
    </div>

    <!-- Credit Card -->
    <div class="credit-card card">
      <div class="credit-header">
        <span class="stat-label">{{ userStore.appConfig?.credit_name || 'TXB' }} Balance</span>
        <router-link to="/credits" class="text-sm text-accent">View Details →</router-link>
      </div>
      <div class="stat-value">{{ userStore.credit.toFixed(2) }}</div>
    </div>

    <!-- Quick Actions -->
    <div class="grid-2 mt-md">
      <router-link to="/combos" class="action-card card">
        <span class="action-icon">🚀</span>
        <span class="action-label">Purchase Combo</span>
      </router-link>
      <router-link to="/ip" class="action-card card">
        <span class="action-icon">🔄</span>
        <span class="action-label">Change IP</span>
      </router-link>
      <router-link to="/squads" class="action-card card">
        <span class="action-icon">🌐</span>
        <span class="action-label">Switch Squad</span>
      </router-link>
      <router-link to="/jellyfin" class="action-card card">
        <span class="action-icon">🎬</span>
        <span class="action-label">Video Management</span>
      </router-link>
    </div>

    <!-- Subscription Summary -->
    <div class="card mt-md" v-if="subInfo?.has_subscription && subInfo.user">
      <div class="row-between mb-sm">
        <h3>📡 Subscription Status</h3>
        <span class="badge" :class="{
          'badge-success': subInfo.user.status === 'ACTIVE',
          'badge-warning': subInfo.user.status === 'LIMITED',
          'badge-danger': subInfo.user.status === 'DISABLED' || subInfo.user.status === 'EXPIRED'
        }">
          {{ subInfo.user.status }}
        </span>
      </div>

      <div class="row-between text-sm text-muted mb-sm">
        <span>{{ formatBytes(subInfo.user.usedTrafficBytes) }} / {{ subInfo.user.trafficLimitBytes ? formatBytes(subInfo.user.trafficLimitBytes) : '♾️' }}</span>
        <span>{{ daysRemaining }} days</span>
      </div>

      <div class="progress">
        <div
          class="progress-bar"
          :class="{ 'warning': usedPercent > 70, 'danger': usedPercent > 90 }"
          :style="{ width: usedPercent + '%' }"
        ></div>
      </div>
    </div>

    <div class="card mt-md empty-state" v-else-if="!loading">
      <span class="empty-state-icon">📡</span>
      <p class="empty-state-text">No subscription yet</p>
      <router-link to="/combos" class="btn btn-primary btn-sm mt-md">Browse Combos</router-link>

      <!-- Subscription Binding -->
      <div class="bind-section mt-lg">
        <p class="text-sm text-muted mb-sm">Already have a sub link? Bind directly</p>
        <div class="row" style="gap:var(--space-sm)">
          <input class="input" v-model="subUrl" placeholder="Paste subscription link" style="flex:1" />
          <button class="btn btn-sm btn-secondary" @click="bindSub" :disabled="bindLoading">
            {{ bindLoading ? '...' : 'Bind' }}
          </button>
        </div>
        <div v-if="bindMessage" class="text-sm mt-sm">{{ bindMessage }}</div>
      </div>
    </div>

    <div class="loading-page" v-if="loading">
      <div class="loading-spinner"></div>
    </div>
  </div>
</template>

<style scoped>
.greeting {
  display: flex;
  align-items: center;
  gap: var(--space-md);
}

.greeting-emoji {
  font-size: 2rem;
}

.greeting-name {
  font-size: 1.25rem;
}

.credit-card {
  background: linear-gradient(135deg, rgba(108, 92, 231, 0.15), rgba(0, 206, 201, 0.1));
  border-color: rgba(108, 92, 231, 0.2);
}

.credit-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-sm);
}

.action-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  padding: var(--space-lg) var(--space-md);
  text-decoration: none;
  color: var(--text-primary);
}

.action-icon {
  font-size: 1.75rem;
}

.action-label {
  font-size: 0.8125rem;
  font-weight: 500;
}
</style>
