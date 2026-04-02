<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useUserStore } from '@/stores/user'
import { api } from '@/api'

const userStore = useUserStore()
const subInfo = ref<any>(null)
const keys = ref<any>(null)
const loading = ref(true)
const showKeys = ref(false)
const copied = ref(false)

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
  try {
    subInfo.value = await api.getSubInfo()
    if (subInfo.value?.has_subscription) {
      keys.value = await api.getSubKeys()
    }
  } catch (e) {}
  loading.value = false
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">📡 我的订阅</h1>
    </div>

    <div class="loading-page" v-if="loading">
      <div class="loading-spinner"></div>
    </div>

    <template v-else-if="subInfo?.has_subscription">
      <!-- Status Card -->
      <div class="card">
        <div class="row-between mb-sm">
          <h3>订阅概览</h3>
          <span class="badge" :class="{
            'badge-success': subInfo.user.status === 'ACTIVE',
            'badge-warning': subInfo.user.status === 'LIMITED',
            'badge-danger': subInfo.user.status === 'DISABLED'
          }">{{ subInfo.user.status }}</span>
        </div>

        <div class="stat-row">
          <div class="stat-item">
            <span class="stat-value text-sm">{{ formatBytes(subInfo.user.usedTrafficBytes) }}</span>
            <span class="stat-label">已用流量</span>
          </div>
          <div class="stat-item">
            <span class="stat-value text-sm">{{ subInfo.user.trafficLimitBytes ? formatBytes(subInfo.user.trafficLimitBytes) : '♾️' }}</span>
            <span class="stat-label">总流量</span>
          </div>
        </div>

        <div class="progress mt-md">
          <div class="progress-bar" :class="{ 'warning': usedPercent > 70, 'danger': usedPercent > 90 }" :style="{ width: usedPercent + '%' }"></div>
        </div>
        <div class="row-between text-xs text-muted mt-sm">
          <span>{{ usedPercent.toFixed(1) }}%</span>
          <span>到期: {{ new Date(subInfo.user.expireAt).toLocaleDateString('zh-CN') }}</span>
        </div>
      </div>

      <!-- Connection Keys -->
      <div class="card mt-md" v-if="keys">
        <div class="row-between mb-sm">
          <h3>🔑 连接信息</h3>
          <button class="btn btn-sm btn-secondary" @click="showKeys = !showKeys">
            {{ showKeys ? '隐藏' : '显示' }}
          </button>
        </div>

        <div v-if="showKeys" class="key-section">
          <div class="key-item">
            <span class="text-sm text-muted">订阅链接</span>
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
          <span>📊</span> <span class="text-sm">流量详情</span>
        </router-link>
        <router-link to="/ip" class="action-card card">
          <span>🔄</span> <span class="text-sm">更换IP</span>
        </router-link>
        <router-link to="/squads" class="action-card card">
          <span>🌐</span> <span class="text-sm">切换线路</span>
        </router-link>
        <router-link to="/combos" class="action-card card">
          <span>🔄</span> <span class="text-sm">续费/换套餐</span>
        </router-link>
      </div>
    </template>

    <div class="empty-state" v-else>
      <span class="empty-state-icon">📡</span>
      <p class="empty-state-text">还没有订阅</p>
      <router-link to="/combos" class="btn btn-primary mt-md">浏览套餐</router-link>
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
