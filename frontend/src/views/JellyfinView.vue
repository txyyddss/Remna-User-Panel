<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useUserStore } from '@/stores/user'
import { api } from '@/api'
import { useToast } from '@/composables/useToast'

const userStore = useUserStore()
const toast = useToast()
const loading = ref(true)
const qcCode = ref('')
const qcLoading = ref(false)
const qcMessage = ref('')
const parentalRating = ref(0)
const lastSavedRating = ref(0)
const devices = ref<any[]>([])
const showPwdForm = ref(false)
const currentPwd = ref('')
const newPwd = ref('')

const showPurchase = ref(false)
const purchaseMonths = ref(1)
const paymentMethod = ref('ezpay')
const paymentType = ref('alipay')
const useTXB = ref(false)
const discountRMB = ref(0)
const purchasing = ref(false)

const jellyfinPrice = computed(() => userStore.appConfig?.jellyfin?.monthly_price_rmb ?? 2)
const totalPrice = computed(() => jellyfinPrice.value * purchaseMonths.value)
const txbRate = computed(() => userStore.appConfig?.credit?.txb_to_rmb_rate ?? userStore.appConfig?.txb_to_rmb_rate ?? 100)
const maxTXBDiscount = computed(() => {
  if (!useTXB.value) return 0
  return Math.max(0, Math.min(totalPrice.value, Math.floor((userStore.credit / txbRate.value) * 100) / 100))
})
const txbUsed = computed(() => discountRMB.value * txbRate.value)
const finalPrice = computed(() => Math.max(0, totalPrice.value - discountRMB.value))

watch(maxTXBDiscount, (value) => {
  if (discountRMB.value > value) discountRMB.value = value
})
watch(useTXB, (enabled) => {
  if (!enabled) discountRMB.value = 0
})

async function loadDevices() {
  if (!userStore.hasJellyfin) {
    devices.value = []
    return
  }
  try {
    const devResp = await api.jellyfinGetDevices()
    devices.value = devResp?.Items || []
  } catch (e) {
    devices.value = []
  }
}

async function loadViewState() {
  await userStore.refreshState({ background: true })
  parentalRating.value = userStore.jellyfin?.parental_rating || 0
  lastSavedRating.value = parentalRating.value
  await loadDevices()
  loading.value = false
}

async function authorizeQC() {
  if (!qcCode.value) return
  qcLoading.value = true
  qcMessage.value = ''
  try {
    await api.jellyfinQuickConnect(qcCode.value)
    qcMessage.value = 'Quick Connect authorized successfully.'
    qcCode.value = ''
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    toast.success('Quick Connect authorized!')
  } catch (e: any) {
    qcMessage.value = e.message
    toast.error(e.message)
  } finally {
    qcLoading.value = false
  }
}

async function updateRating() {
  try {
    await api.jellyfinUpdateParentalRating(parentalRating.value)
    lastSavedRating.value = parentalRating.value
    await userStore.refreshState({ background: true })
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    toast.success('Content rating updated.')
  } catch (e: any) {
    parentalRating.value = lastSavedRating.value
    toast.error(e.message)
  }
}

async function changePassword() {
  try {
    await api.jellyfinUpdatePassword(currentPwd.value, newPwd.value)
    showPwdForm.value = false
    currentPwd.value = ''
    newPwd.value = ''
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    toast.success('Password updated successfully.')
  } catch (e: any) {
    toast.error(e.message)
  }
}

async function purchaseJellyfin() {
  purchasing.value = true
  try {
    const resp = await api.purchaseJellyfin({
      months: purchaseMonths.value,
      payment_method: paymentMethod.value,
      payment_type: paymentMethod.value === 'bepusdt' ? 'usdt' : paymentType.value,
      use_txb: useTXB.value,
      discount_rmb: discountRMB.value,
    })
    await userStore.refreshState({ background: true })
    if (resp.payment_url) {
      window.Telegram?.WebApp?.openLink(resp.payment_url)
    } else {
      await loadViewState()
      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    }
    showPurchase.value = false
    toast.success('Order created successfully.')
  } catch (e: any) {
    toast.error(e.message)
  } finally {
    purchasing.value = false
  }
}

onMounted(loadViewState)

watch(() => userStore.jellyfin, async (nextValue) => {
  parentalRating.value = nextValue?.parental_rating || 0
  lastSavedRating.value = parentalRating.value
  await loadDevices()
}, { deep: true })
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">Jellyfin</h1>
    </div>

    <div class="loading-page" v-if="loading"><div class="loading-spinner"></div></div>

    <template v-else-if="userStore.hasJellyfin">
      <div class="card">
        <div class="row-between">
          <div>
            <h3>Account</h3>
            <div class="text-sm text-muted mt-sm">Username: {{ userStore.jellyfin?.username || userStore.user?.jellyfin_user_id }}</div>
            <div class="text-sm text-muted mt-sm">Jellyfin ID: {{ userStore.jellyfin?.jellyfin_user_id }}</div>
          </div>
          <span class="badge badge-success">Active</span>
        </div>
        <div class="text-sm text-muted mt-sm">
          Expires: {{ userStore.jellyfin?.expires_at ? new Date(userStore.jellyfin.expires_at).toLocaleDateString('en-US') : '' }}
        </div>
        <button class="btn btn-primary btn-sm mt-md" @click="showPurchase = true">Renew Service</button>
      </div>

      <div class="card mt-md">
        <h3 class="mb-sm">Content Rating</h3>
        <p class="text-xs text-muted mb-md">Choose the highest allowed rating for this Jellyfin account.</p>
        <div class="row-between mb-sm">
          <span class="text-sm">Current setting: {{ parentalRating }}</span>
          <span class="text-xs text-muted">0 to 22</span>
        </div>
        <input class="slider" type="range" min="0" max="22" step="1" v-model.number="parentalRating" @change="updateRating" />
      </div>

      <div class="card mt-md">
        <h3 class="mb-sm">Quick Connect</h3>
        <div class="row" style="gap:var(--space-sm)">
          <input class="input" v-model="qcCode" placeholder="Enter the 6-character code" style="flex:1" />
          <button class="btn btn-primary btn-sm" @click="authorizeQC" :disabled="qcLoading">{{ qcLoading ? 'Authorizing...' : 'Authorize' }}</button>
        </div>
        <div v-if="qcMessage" class="text-sm mt-sm">{{ qcMessage }}</div>
      </div>

      <div class="card mt-md">
        <div class="row-between">
          <h3>Password</h3>
          <button class="btn btn-sm btn-secondary" @click="showPwdForm = !showPwdForm">{{ showPwdForm ? 'Cancel' : 'Change Password' }}</button>
        </div>
        <div v-if="showPwdForm" class="stack-sm mt-md">
          <input class="input" v-model="currentPwd" type="password" placeholder="Current password" />
          <input class="input" v-model="newPwd" type="password" placeholder="New password" />
          <button class="btn btn-primary btn-sm" @click="changePassword">Save Password</button>
        </div>
      </div>

      <div class="card mt-md">
        <div class="row-between mb-md">
          <h3>Device List</h3>
          <span class="text-xs text-muted">{{ devices.length }} device(s)</span>
        </div>
        <div v-if="devices.length === 0" class="text-sm text-muted">No active devices are recorded for this account.</div>
        <div v-else class="stack-sm">
          <div v-for="dev in devices" :key="dev.Id" class="device-item">
            <div>
              <div class="text-sm">{{ dev.AppName || 'Unknown App' }}</div>
              <div class="text-xs text-muted">{{ dev.Name || dev.Id }}</div>
            </div>
            <span class="text-xs text-muted">{{ new Date(dev.DateLastActivity).toLocaleDateString('en-US') }}</span>
          </div>
        </div>
      </div>
    </template>

    <div class="empty-state" v-else>
      <span class="empty-state-icon">🎬</span>
      <p class="empty-state-text">No Jellyfin account yet</p>
      <p class="text-xs text-muted mt-sm">¥{{ jellyfinPrice.toFixed(2) }} / month</p>
      <button class="btn btn-primary mt-md" @click="showPurchase = true">Activate Jellyfin</button>
    </div>

    <teleport to="body">
      <transition name="modal-slide">
        <div class="modal-overlay" v-if="showPurchase" @click.self="showPurchase = false">
          <div class="modal card">
            <h3 class="mb-md">{{ userStore.hasJellyfin ? 'Renew Jellyfin' : 'Activate Jellyfin' }}</h3>

            <div class="stack-sm">
              <label class="text-sm text-muted">Duration</label>
              <div class="payment-grid">
                <button class="payment-option small" :class="{ active: purchaseMonths === 1 }" @click="purchaseMonths = 1">1 Month</button>
                <button class="payment-option small" :class="{ active: purchaseMonths === 3 }" @click="purchaseMonths = 3">3 Months</button>
                <button class="payment-option small" :class="{ active: purchaseMonths === 6 }" @click="purchaseMonths = 6">6 Months</button>
                <button class="payment-option small" :class="{ active: purchaseMonths === 12 }" @click="purchaseMonths = 12">12 Months</button>
              </div>

              <div class="row-between mt-sm">
                <span class="text-muted">Original price</span>
                <span class="mono">¥{{ totalPrice.toFixed(2) }}</span>
              </div>

              <label class="text-sm text-muted mt-sm">Payment Method</label>
              <div class="payment-grid">
                <button class="payment-option" :class="{ active: paymentMethod === 'ezpay' }" @click="paymentMethod = 'ezpay'">EZPay</button>
                <button class="payment-option" :class="{ active: paymentMethod === 'bepusdt' }" @click="paymentMethod = 'bepusdt'">USDT</button>
              </div>

              <div v-if="paymentMethod === 'ezpay'" class="payment-grid mt-sm">
                <button class="payment-option small" :class="{ active: paymentType === 'alipay' }" @click="paymentType = 'alipay'">Alipay</button>
                <button class="payment-option small" :class="{ active: paymentType === 'wxpay' }" @click="paymentType = 'wxpay'">WeChat</button>
              </div>

              <label class="checkbox mt-md">
                <input type="checkbox" v-model="useTXB" />
                <span class="text-sm">Use {{ userStore.appConfig?.credit_name || 'TXB' }} as a discount</span>
              </label>

              <div v-if="useTXB && maxTXBDiscount > 0" class="discount-card">
                <div class="row-between text-sm">
                  <span class="text-muted">Discount</span>
                  <span>¥{{ discountRMB.toFixed(2) }}</span>
                </div>
                <input class="slider" type="range" min="0" :max="maxTXBDiscount" step="0.01" v-model.number="discountRMB" />
                <div class="row-between text-xs text-muted">
                  <span>0</span>
                  <span>{{ txbUsed.toFixed(0) }} {{ userStore.appConfig?.credit_name || 'TXB' }}</span>
                  <span>¥{{ maxTXBDiscount.toFixed(2) }}</span>
                </div>
              </div>

              <div class="row-between text-sm mt-xs">
                <span class="text-muted">Final price</span>
                <span class="mono price-value">¥{{ finalPrice.toFixed(2) }}</span>
              </div>
            </div>

            <div class="row mt-lg" style="gap: var(--space-sm)">
              <button class="btn btn-secondary" style="flex:1" @click="showPurchase = false">Cancel</button>
              <button class="btn btn-primary" style="flex:2" @click="purchaseJellyfin" :disabled="purchasing">
                {{ purchasing ? 'Processing...' : `Confirm Payment ¥${finalPrice.toFixed(2)}` }}
              </button>
            </div>
          </div>
        </div>
      </transition>
    </teleport>
  </div>
</template>

<style scoped>
.device-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: var(--space-sm) 0;
  border-bottom: 1px solid var(--border-subtle);
}

.device-item:last-child {
  border-bottom: none;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(3, 10, 21, 0.72);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
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

.payment-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: var(--space-sm);
}

.payment-option {
  padding: var(--space-md);
  background: var(--bg-glass);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  cursor: pointer;
  font-family: var(--font-body);
  font-size: 0.875rem;
  transition: all 0.2s;
}

.payment-option.small {
  padding: var(--space-sm);
  font-size: 0.8125rem;
}

.payment-option.active {
  border-color: var(--accent-primary);
  background: rgba(108, 92, 231, 0.1);
}

.checkbox {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  cursor: pointer;
}

.discount-card {
  padding: var(--space-sm);
  background: rgba(0, 206, 201, 0.08);
  border-radius: var(--radius-sm);
  border: 1px solid rgba(0, 206, 201, 0.2);
}

.slider {
  width: 100%;
}

.modal-slide-enter-active {
  transition: opacity 0.3s ease;
}
.modal-slide-enter-active .modal {
  transition: transform 0.35s cubic-bezier(0.4, 0, 0.2, 1);
}
.modal-slide-leave-active {
  transition: opacity 0.25s ease;
}
.modal-slide-leave-active .modal {
  transition: transform 0.25s ease-in;
}
.modal-slide-enter-from {
  opacity: 0;
}
.modal-slide-enter-from .modal {
  transform: translateY(100%);
}
.modal-slide-leave-to {
  opacity: 0;
}
.modal-slide-leave-to .modal {
  transform: translateY(30%);
}
</style>
