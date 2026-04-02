<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useUserStore } from '@/stores/user'
import { api } from '@/api'

const userStore = useUserStore()
const subInfo = ref<any>(null)
const loading = ref(true)

const usedPercent = computed(() => {
  if (!subInfo.value?.user?.trafficLimitBytes) return 0
  return Math.min(100, (subInfo.value.user.usedTrafficBytes / subInfo.value.user.trafficLimitBytes) * 100)
})

const daysRemaining = computed(() => {
  if (!subInfo.value?.user?.expireAt) return 0
  const diff = new Date(subInfo.value.user.expireAt).getTime() - Date.now()
  return Math.max(0, Math.ceil(diff / 86400000))
})

function formatBytes(b: number): string {
  if (b < 1024) return `${b} B`
  if (b < 1048576) return `${(b / 1024).toFixed(1)} KB`
  if (b < 1073741824) return `${(b / 1048576).toFixed(2)} MB`
  return `${(b / 1073741824).toFixed(2)} GB`
}

onMounted(async () => {
  try {
    if (userStore.hasSubscription || userStore.user?.remnawave_uuid) {
      subInfo.value = await api.getSubInfo()
    }
  } catch (e) {}
  loading.value = false
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div class="greeting">
        <span class="greeting-emoji">👋</span>
        <div>
          <h1 class="greeting-name">{{ userStore.telegramName }}</h1>
          <p class="page-subtitle">欢迎回来</p>
        </div>
      </div>
    </div>

    <!-- Credit Card -->
    <div class="credit-card card">
      <div class="credit-header">
        <span class="stat-label">{{ userStore.appConfig?.credit_name || 'TXB' }} 余额</span>
        <router-link to="/credits" class="text-sm text-accent">查看详情 →</router-link>
      </div>
      <div class="stat-value">{{ userStore.credit.toFixed(2) }}</div>
    </div>

    <!-- Quick Actions -->
    <div class="grid-2 mt-md">
      <router-link to="/combos" class="action-card card">
        <span class="action-icon">🚀</span>
        <span class="action-label">购买套餐</span>
      </router-link>
      <router-link to="/ip" class="action-card card">
        <span class="action-icon">🔄</span>
        <span class="action-label">更换IP</span>
      </router-link>
      <router-link to="/squads" class="action-card card">
        <span class="action-icon">🌐</span>
        <span class="action-label">切换线路</span>
      </router-link>
      <router-link to="/jellyfin" class="action-card card">
        <span class="action-icon">🎬</span>
        <span class="action-label">影视管理</span>
      </router-link>
    </div>

    <!-- Subscription Summary -->
    <div class="card mt-md" v-if="subInfo?.has_subscription">
      <div class="row-between mb-sm">
        <h3>📡 订阅状态</h3>
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
        <span>{{ daysRemaining }} 天</span>
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
      <p class="empty-state-text">还没有订阅</p>
      <router-link to="/combos" class="btn btn-primary btn-sm mt-md">浏览套餐</router-link>
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
