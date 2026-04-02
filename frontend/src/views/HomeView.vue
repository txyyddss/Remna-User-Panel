<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useUserStore } from '@/stores/user'
import { formatBytes } from '@/utils/format'

const userStore = useUserStore()
const loading = ref(true)

const sub = computed(() => userStore.liveSubInfo)
const hasSub = computed(() => sub.value?.has_subscription && sub.value.user)

const usedTraffic = computed(() =>
  sub.value?.user?.userTraffic?.usedTrafficBytes ?? sub.value?.user?.usedTrafficBytes ?? 0,
)
const trafficLimit = computed(() => sub.value?.user?.trafficLimitBytes ?? 0)
const usedPercent = computed(() => {
  if (!trafficLimit.value) return 0
  return Math.min(100, (usedTraffic.value / trafficLimit.value) * 100)
})

const greeting = computed(() => {
  const hour = new Date().getHours()
  if (hour < 6) return 'Good Night'
  if (hour < 12) return 'Good Morning'
  if (hour < 18) return 'Good Afternoon'
  return 'Good Evening'
})

const expiresText = computed(() => {
  if (!sub.value?.user?.expireAt) return ''
  const d = new Date(sub.value.user.expireAt)
  const now = new Date()
  const diffD = Math.ceil((d.getTime() - now.getTime()) / 86400000)
  if (diffD <= 0) return 'Expired'
  if (diffD <= 3) return `${diffD}d left`
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
})

onMounted(async () => {
  await userStore.refreshState({ background: true })
  loading.value = false
})
</script>

<template>
  <div class="page">
    <!-- Loading -->
    <div class="loading-page" v-if="loading">
      <div class="loading-spinner"></div>
    </div>

    <template v-else>
      <!-- Greeting -->
      <div class="hero-section stagger-enter stagger-1">
        <h1 class="hero-greeting">{{ greeting }}</h1>
        <p class="hero-name">{{ userStore.user?.telegram_name || 'User' }}</p>
      </div>

      <!-- Credit Balance -->
      <div class="credit-card card stagger-enter stagger-2">
        <div class="credit-inner">
          <div class="credit-label">{{ userStore.appConfig?.credit_name || 'TXB' }} Balance</div>
          <div class="credit-value">{{ userStore.credit.toFixed(2) }}</div>
        </div>
        <router-link to="/credits" class="credit-action">
          <span>Earn</span>
          <span class="credit-arrow">→</span>
        </router-link>
      </div>

      <!-- Subscription Status -->
      <div v-if="hasSub" class="sub-card card stagger-enter stagger-3">
        <div class="sub-header">
          <div>
            <span class="sub-label">Subscription</span>
            <span class="badge" :class="{
              'badge-success': sub?.user?.status === 'ACTIVE',
              'badge-warning': sub?.user?.status === 'LIMITED',
              'badge-danger': sub?.user?.status === 'DISABLED' || sub?.user?.status === 'EXPIRED',
            }">{{ sub?.user?.status }}</span>
          </div>
          <span class="sub-expires mono">{{ expiresText }}</span>
        </div>
        <div class="sub-progress">
          <div class="progress">
            <div
              class="progress-bar"
              :class="{ warning: usedPercent >= 70, danger: usedPercent >= 90 }"
              :style="{ width: `${usedPercent}%` }"
            ></div>
          </div>
          <div class="sub-stats">
            <span>{{ formatBytes(usedTraffic) }} used</span>
            <span>{{ trafficLimit ? formatBytes(trafficLimit) : '∞' }}</span>
          </div>
        </div>
      </div>

      <div v-else class="sub-card card stagger-enter stagger-3 empty-sub">
        <span class="empty-sub-icon">🚀</span>
        <p>No active subscription</p>
        <router-link to="/combos" class="btn btn-primary btn-sm">Browse Plans</router-link>
      </div>

      <!-- Quick Actions -->
      <div class="actions-grid stagger-enter stagger-4">
        <router-link to="/sub" class="action-tile card">
          <span class="action-icon">📡</span>
          <span class="action-label">VPN</span>
        </router-link>
        <router-link to="/jellyfin" class="action-tile card">
          <span class="action-icon">🎬</span>
          <span class="action-label">Jellyfin</span>
        </router-link>
        <router-link to="/ip" class="action-tile card">
          <span class="action-icon">🌐</span>
          <span class="action-label">Change IP</span>
        </router-link>
        <router-link to="/combos" class="action-tile card">
          <span class="action-icon">💎</span>
          <span class="action-label">Plans</span>
        </router-link>
      </div>

      <!-- Recent Activity -->
      <div v-if="userStore.recentOrders.length" class="card stagger-enter stagger-5">
        <h3 class="mb-sm">Recent Activity</h3>
        <div class="activity-list">
          <div v-for="order in userStore.recentOrders.slice(0, 3)" :key="order.uuid" class="activity-item">
            <div class="activity-info">
              <span class="text-sm">{{ order.order_type }}</span>
              <span class="text-xs text-muted">{{ order.status }}</span>
            </div>
            <span class="mono text-sm">¥{{ Number(order.final_amount || 0).toFixed(2) }}</span>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.hero-section {
  margin-bottom: var(--space-lg);
}

.hero-greeting {
  font-family: var(--font-display);
  font-size: 1.125rem;
  color: var(--text-secondary);
  font-weight: 400;
}

.hero-name {
  font-family: var(--font-display);
  font-size: 1.75rem;
  font-weight: 700;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* Credit Card */
.credit-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(135deg, rgba(124, 106, 240, 0.12), rgba(0, 210, 198, 0.08));
  border-color: rgba(124, 106, 240, 0.2);
  margin-bottom: var(--space-md);
}
.credit-card:hover {
  transform: translateY(-2px);
  border-color: rgba(124, 106, 240, 0.35);
}

.credit-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.credit-value {
  font-family: var(--font-display);
  font-size: 2rem;
  font-weight: 700;
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.credit-action {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  padding: var(--space-sm) var(--space-md);
  background: rgba(255, 255, 255, 0.05);
  border-radius: var(--radius-md);
  color: var(--text-secondary);
  font-size: 0.8125rem;
  font-weight: 500;
  text-decoration: none;
  transition: all 0.2s;
}

.credit-action:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-primary);
}

.credit-arrow {
  transition: transform 0.2s;
}

.credit-action:hover .credit-arrow {
  transform: translateX(3px);
}

/* Subscription Card */
.sub-card {
  margin-bottom: var(--space-md);
}

.sub-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-md);
}

.sub-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.08em;
  margin-right: var(--space-sm);
}

.sub-expires {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.sub-stats {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-top: var(--space-xs);
}

.empty-sub {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  text-align: center;
  padding: var(--space-xl);
}

.empty-sub-icon {
  font-size: 2rem;
}

/* Action Grid */
.actions-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: var(--space-sm);
  margin-bottom: var(--space-md);
}

.action-tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--space-sm);
  text-decoration: none;
  color: inherit;
  padding: var(--space-md) var(--space-sm);
  text-align: center;
}

.action-icon {
  font-size: 1.5rem;
}

.action-label {
  font-size: 0.6875rem;
  font-weight: 500;
  color: var(--text-secondary);
}

/* Activity List */
.activity-list {
  display: flex;
  flex-direction: column;
}

.activity-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-sm) 0;
  border-bottom: 1px solid var(--border-subtle);
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
</style>
