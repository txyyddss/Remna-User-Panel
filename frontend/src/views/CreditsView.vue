<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api } from '@/api'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const history = ref<any[]>([])
const loading = ref(false)
const signupResult = ref<{ value: number; balance: number } | null>(null)
const betAmount = ref('')
const betResult = ref<{ result: number; balance: number } | null>(null)
const error = ref('')
const selectedOrder = ref<any>(null)
const orderLoading = ref(false)

async function doSignup() {
  loading.value = true
  error.value = ''
  try {
    const data = await api.creditSignup()
    signupResult.value = { value: data.value, balance: data.new_balance }
    await userStore.refreshState({ background: true })
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    error.value = e.message
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('error')
  } finally {
    loading.value = false
  }
}

async function doBet() {
  const amount = Number.parseFloat(betAmount.value)
  if (!amount || amount <= 0) {
    error.value = 'Please enter a valid amount'
    return
  }

  loading.value = true
  error.value = ''
  try {
    const data = await api.creditBet(amount)
    betResult.value = { result: data.result, balance: data.new_balance }
    await userStore.refreshState({ background: true })
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred(data.result > 0 ? 'success' : 'error')
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function loadHistory() {
  try {
    history.value = (await api.getCreditHistory(50)) || []
  } catch {
    history.value = []
  }
}

async function openOrder(uuid: string) {
  orderLoading.value = true
  try {
    selectedOrder.value = await api.getOrder(uuid)
  } catch (e: any) {
    alert(e.message)
  } finally {
    orderLoading.value = false
  }
}

onMounted(async () => {
  await Promise.all([
    loadHistory(),
    userStore.refreshOrders(20),
  ])
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">{{ userStore.appConfig?.credit_name || 'TXB' }} Credits</h1>
    </div>

    <div class="card credit-hero">
      <span class="stat-label">Current Balance</span>
      <div class="stat-value">{{ userStore.credit.toFixed(2) }}</div>
      <span class="text-xs text-muted">{{ userStore.appConfig?.rmb_to_txb_rate || 100 }} {{ userStore.appConfig?.credit_name || 'TXB' }} = 1 RMB Discount</span>
    </div>

    <div class="grid-2 mt-md">
      <div class="card action-section">
        <h4 class="mb-sm">Daily Check-in</h4>
        <button class="btn btn-primary btn-block btn-sm" @click="doSignup" :disabled="loading">
          Check-in
        </button>
        <div v-if="signupResult" class="result-text text-success mt-sm text-sm">
          +{{ signupResult.value.toFixed(2) }} {{ userStore.appConfig?.credit_name || 'TXB' }}
        </div>
      </div>

      <div class="card action-section">
        <h4 class="mb-sm">Bet</h4>
        <input class="input" v-model="betAmount" type="number" placeholder="Bet Amount" step="0.01" />
        <button class="btn btn-secondary btn-block btn-sm mt-sm" @click="doBet" :disabled="loading">
          Bet
        </button>
        <div v-if="betResult" class="result-text mt-sm text-sm" :class="{ 'text-success': betResult.result > 0, 'text-danger': betResult.result < 0 }">
          {{ betResult.result > 0 ? '+' : '' }}{{ betResult.result.toFixed(2) }} {{ userStore.appConfig?.credit_name || 'TXB' }}
        </div>
      </div>
    </div>

    <div v-if="error" class="text-danger text-sm mt-sm text-center">{{ error }}</div>

    <div class="card mt-md">
      <h3 class="mb-md">Credit History</h3>
      <div v-if="history.length === 0" class="text-muted text-sm text-center">No history available</div>
      <div v-else class="history-list">
        <div v-for="log in history" :key="log.id" class="history-item">
          <div class="history-info">
            <span class="text-sm">{{ log.reason }}</span>
            <span class="text-xs text-muted">{{ new Date(log.created_at).toLocaleString('en-US') }}</span>
          </div>
          <div class="history-amount" :class="{ positive: log.amount > 0, negative: log.amount < 0 }">
            {{ log.amount > 0 ? '+' : '' }}{{ log.amount.toFixed(2) }}
          </div>
        </div>
      </div>
    </div>

    <div class="card mt-md">
      <h3 class="mb-md">Bills</h3>
      <div v-if="userStore.recentOrders.length === 0" class="text-muted text-sm text-center">No bills yet</div>
      <div v-else class="history-list">
        <button v-for="order in userStore.recentOrders" :key="order.uuid" class="history-item order-item" @click="openOrder(order.uuid)">
          <div class="history-info">
            <span class="text-sm">{{ order.order_type }} · {{ new Date(order.created_at).toLocaleString('en-US') }}</span>
            <span class="text-xs text-muted">Payment: {{ order.status }} · Service: {{ order.service_status }}</span>
          </div>
          <div class="history-amount" :class="{ positive: order.status === 'paid', negative: order.status === 'refunded' }">
            ¥{{ Number(order.final_amount || 0).toFixed(2) }}
          </div>
        </button>
      </div>
    </div>

    <teleport to="body">
      <transition name="fade">
        <div v-if="selectedOrder || orderLoading" class="modal-overlay" @click.self="selectedOrder = null">
          <div class="modal card">
            <h3 class="mb-md">Bill Detail</h3>
            <div v-if="orderLoading" class="loading-page"><div class="loading-spinner"></div></div>
            <div v-else-if="selectedOrder" class="stack-sm">
              <div class="text-xs text-muted">{{ selectedOrder.uuid }}</div>
              <div class="row-between text-sm">
                <span class="text-muted">Type</span>
                <span>{{ selectedOrder.order_type }}</span>
              </div>
              <div class="row-between text-sm">
                <span class="text-muted">Created</span>
                <span>{{ new Date(selectedOrder.created_at).toLocaleString('en-US') }}</span>
              </div>
              <div class="row-between text-sm">
                <span class="text-muted">Payment</span>
                <span>{{ selectedOrder.status }}</span>
              </div>
              <div class="row-between text-sm">
                <span class="text-muted">Service</span>
                <span>{{ selectedOrder.service_status }}</span>
              </div>
              <div class="row-between text-sm">
                <span class="text-muted">Final amount</span>
                <span class="mono">¥{{ Number(selectedOrder.final_amount || 0).toFixed(2) }}</span>
              </div>

              <div class="card muted-card">
                <h4 class="mb-sm">Timeline</h4>
                <div v-if="!selectedOrder.events?.length" class="text-xs text-muted">No bill events yet.</div>
                <div v-else class="history-list">
                  <div v-for="event in selectedOrder.events" :key="event.id" class="history-item">
                    <div class="history-info">
                      <span class="text-sm">{{ event.event_type }}</span>
                      <span class="text-xs text-muted">{{ event.message }}</span>
                    </div>
                    <div class="text-xs text-muted">{{ new Date(event.created_at).toLocaleString('en-US') }}</div>
                  </div>
                </div>
              </div>
            </div>
            <div class="row mt-lg" style="gap: var(--space-sm)">
              <button class="btn btn-primary btn-block" @click="selectedOrder = null">Close</button>
            </div>
          </div>
        </div>
      </transition>
    </teleport>
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

.order-item {
  width: 100%;
  border: none;
  background: transparent;
  color: inherit;
  text-align: left;
  cursor: pointer;
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

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: flex-end;
  z-index: 200;
}

.modal {
  width: 100%;
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
  max-height: 80vh;
  overflow-y: auto;
}
</style>
