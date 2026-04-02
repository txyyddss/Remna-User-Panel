<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { api } from '@/api'

const userStore = useUserStore()
const history = ref<any[]>([])
const loading = ref(false)
const signupResult = ref<{ value: number; balance: number } | null>(null)
const betAmount = ref('')
const betResult = ref<{ result: number; balance: number } | null>(null)
const error = ref('')

async function doSignup() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.creditSignup()
    signupResult.value = { value: data.value, balance: data.new_balance }
    await userStore.refreshCredit()
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    error.value = e.message
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('error')
  }
  loading.value = false
}

async function doBet() {
  const amt = parseFloat(betAmount.value)
  if (!amt || amt <= 0) { error.value = '请输入有效金额'; return }
  loading.value = true
  error.value = ''
  try {
    const data = await api.creditBet(amt)
    betResult.value = { result: data.result, balance: data.new_balance }
    await userStore.refreshCredit()
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred(data.result > 0 ? 'success' : 'error')
  } catch (e: any) {
    error.value = e.message
  }
  loading.value = false
}

async function loadHistory() {
  try {
    history.value = await api.getCreditHistory(50) || []
  } catch (e) {}
}

onMounted(loadHistory)
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">💎 TXB 积分</h1>
    </div>

    <!-- Balance -->
    <div class="card credit-hero">
      <span class="stat-label">当前余额</span>
      <div class="stat-value">{{ userStore.credit.toFixed(2) }}</div>
      <span class="text-xs text-muted">100 TXB = 1 RMB 折扣</span>
    </div>

    <!-- Actions -->
    <div class="grid-2 mt-md">
      <div class="card action-section">
        <h4 class="mb-sm">🎁 每日签到</h4>
        <button class="btn btn-primary btn-block btn-sm" @click="doSignup" :disabled="loading">
          签到
        </button>
        <div v-if="signupResult" class="result-text text-success mt-sm text-sm">
          +{{ signupResult.value.toFixed(2) }} TXB
        </div>
      </div>

      <div class="card action-section">
        <h4 class="mb-sm">🎲 赌博</h4>
        <input class="input" v-model="betAmount" type="number" placeholder="下注金额" step="0.01" />
        <button class="btn btn-secondary btn-block btn-sm mt-sm" @click="doBet" :disabled="loading">
          下注
        </button>
        <div v-if="betResult" class="result-text mt-sm text-sm" :class="{ 'text-success': betResult.result > 0, 'text-danger': betResult.result < 0 }">
          {{ betResult.result > 0 ? '+' : '' }}{{ betResult.result.toFixed(2) }} TXB
        </div>
      </div>
    </div>

    <div v-if="error" class="text-danger text-sm mt-sm text-center">{{ error }}</div>

    <!-- History -->
    <div class="card mt-md">
      <h3 class="mb-md">📜 积分记录</h3>
      <div v-if="history.length === 0" class="text-muted text-sm text-center">暂无记录</div>
      <div v-else class="history-list">
        <div v-for="log in history" :key="log.id" class="history-item">
          <div class="history-info">
            <span class="text-sm">{{ log.reason }}</span>
            <span class="text-xs text-muted">{{ new Date(log.created_at).toLocaleString('zh-CN') }}</span>
          </div>
          <div class="history-amount" :class="{ positive: log.amount > 0, negative: log.amount < 0 }">
            {{ log.amount > 0 ? '+' : '' }}{{ log.amount.toFixed(2) }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.credit-hero {
  text-align: center;
  background: linear-gradient(135deg, rgba(108, 92, 231, 0.12), rgba(0, 206, 201, 0.08));
  border-color: rgba(108, 92, 231, 0.2);
}

.credit-hero .stat-value {
  font-size: 2.5rem;
}

.action-section {
  display: flex;
  flex-direction: column;
}

.action-section:hover {
  transform: none;
}

.result-text {
  text-align: center;
  font-family: var(--font-display);
  font-weight: 700;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  max-height: 400px;
  overflow-y: auto;
}

.history-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-sm) 0;
  border-bottom: 1px solid var(--border-subtle);
}

.history-item:last-child {
  border-bottom: none;
}

.history-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.history-amount {
  font-family: var(--font-display);
  font-weight: 700;
  font-size: 0.875rem;
}

.history-amount.positive { color: var(--accent-success); }
.history-amount.negative { color: var(--accent-danger); }
</style>
