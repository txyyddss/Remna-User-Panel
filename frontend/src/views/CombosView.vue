<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { api } from '@/api'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

const combos = ref<any[]>([])
const loading = ref(true)
const purchasing = ref(false)
const selectedCombo = ref<any>(null)
const paymentMethod = ref('ezpay')
const paymentType = ref('alipay')
const useTXB = ref(false)
const discountRMB = ref(0)

const showCustomPayment = ref(false)
const customAmount = ref(10)
const customMessage = ref('')
const customPaymentMethod = ref('ezpay')
const customPaymentType = ref('alipay')
const customUseTXB = ref(false)
const customDiscountRMB = ref(0)
const customSubmitting = ref(false)

const cycleName: Record<string, string> = {
  monthly: 'Monthly',
  quarterly: 'Quarterly',
  semiannual: 'Semi-Annual',
  annual: 'Annual',
}

const txbRate = computed(() => userStore.appConfig?.credit?.txb_to_rmb_rate ?? userStore.appConfig?.txb_to_rmb_rate ?? 100)
const comboPrice = computed(() => selectedCombo.value?.price_rmb || 0)
const maxComboDiscount = computed(() => {
  if (!useTXB.value || !selectedCombo.value) return 0
  return Math.max(0, Math.min(comboPrice.value, Math.floor((userStore.credit / txbRate.value) * 100) / 100))
})
const comboFinalPrice = computed(() => Math.max(0, comboPrice.value - discountRMB.value))
const comboTXBUsed = computed(() => discountRMB.value * txbRate.value)

const maxCustomDiscount = computed(() => {
  if (!customUseTXB.value || customAmount.value <= 0) return 0
  return Math.max(0, Math.min(customAmount.value, Math.floor((userStore.credit / txbRate.value) * 100) / 100))
})
const customFinalPrice = computed(() => Math.max(0, customAmount.value - customDiscountRMB.value))
const customTXBUsed = computed(() => customDiscountRMB.value * txbRate.value)

watch(useTXB, (enabled) => {
  if (!enabled) discountRMB.value = 0
})
watch(customUseTXB, (enabled) => {
  if (!enabled) customDiscountRMB.value = 0
})
watch(maxComboDiscount, (value) => {
  if (discountRMB.value > value) discountRMB.value = value
})
watch(maxCustomDiscount, (value) => {
  if (customDiscountRMB.value > value) customDiscountRMB.value = value
})
watch(selectedCombo, (value) => {
  if (!value) {
    paymentMethod.value = 'ezpay'
    paymentType.value = 'alipay'
    useTXB.value = false
    discountRMB.value = 0
  }
})

function formatTraffic(gb: number): string {
  if (gb >= 1024) return `${(gb / 1024).toFixed(1)} TB`
  return `${gb} GB`
}

async function loadCombos() {
  try {
    combos.value = (await api.listCombos()) || []
  } catch (e) {
    combos.value = []
  } finally {
    loading.value = false
  }
}

async function purchase() {
  if (!selectedCombo.value) return
  purchasing.value = true
  try {
    const resp = await api.purchaseCombo({
      combo_uuid: selectedCombo.value.uuid,
      payment_method: paymentMethod.value,
      payment_type: paymentMethod.value === 'bepusdt' ? 'usdt' : paymentType.value,
      use_txb: useTXB.value,
      discount_rmb: discountRMB.value,
    })
    await userStore.refreshState({ background: true })
    if (resp.payment_url) {
      window.Telegram?.WebApp?.openLink(resp.payment_url)
    } else {
      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    }
    selectedCombo.value = null
  } catch (e: any) {
    alert(e.message)
  } finally {
    purchasing.value = false
  }
}

async function submitCustomPayment() {
  if (customAmount.value <= 0) return
  customSubmitting.value = true
  try {
    const resp = await api.customPayment({
      amount: customAmount.value,
      message: customMessage.value,
      payment_method: customPaymentMethod.value,
      payment_type: customPaymentMethod.value === 'bepusdt' ? 'usdt' : customPaymentType.value,
      use_txb: customUseTXB.value,
      discount_rmb: customDiscountRMB.value,
    })
    await userStore.refreshState({ background: true })
    if (resp.payment_url) {
      window.Telegram?.WebApp?.openLink(resp.payment_url)
      alert('Payment created. After it is paid, admins will be notified and can apply the top-up.')
    } else {
      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
      alert('Payment completed. Admins have been notified and can apply the top-up.')
    }
    showCustomPayment.value = false
    customAmount.value = 10
    customMessage.value = ''
    customUseTXB.value = false
    customDiscountRMB.value = 0
  } catch (e: any) {
    alert(e.message)
  } finally {
    customSubmitting.value = false
  }
}

onMounted(loadCombos)
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">Plans</h1>
      <p class="page-subtitle">Buy, renew, or switch to another subscription package</p>
    </div>

    <div class="loading-page" v-if="loading"><div class="loading-spinner"></div></div>

    <div class="stack" v-else>
      <div v-for="combo in combos" :key="combo.uuid" class="card combo-card" @click="selectedCombo = combo">
        <div class="row-between">
          <h3>{{ combo.name }}</h3>
          <div class="combo-price">
            <span class="price-value">¥{{ combo.price_rmb.toFixed(2) }}</span>
            <span class="price-cycle">/{{ cycleName[combo.cycle] || combo.cycle }}</span>
          </div>
        </div>
        <p class="text-sm text-muted mt-sm">{{ combo.description }}</p>
        <div class="combo-specs mt-md">
          <span class="spec-item">Traffic {{ formatTraffic(combo.traffic_gb) }}</span>
          <span class="spec-item">Reset {{ combo.strategy }}</span>
          <span class="spec-item">Reset Fee ¥{{ combo.reset_price.toFixed(2) }}</span>
        </div>
      </div>
    </div>

    <div v-if="combos.length === 0 && !loading" class="empty-state">
      <span class="empty-state-icon">📦</span>
      <p class="empty-state-text">No plans are available right now</p>
    </div>

    <div class="card mt-md" v-if="!loading">
      <div class="row-between">
        <h3>Custom Top-up</h3>
        <button class="btn btn-sm btn-secondary" @click="showCustomPayment = !showCustomPayment">
          {{ showCustomPayment ? 'Cancel' : 'Top-up' }}
        </button>
      </div>

      <div v-if="showCustomPayment" class="stack-sm mt-md">
        <input class="input" v-model.number="customAmount" type="number" min="1" step="0.01" placeholder="Amount (RMB)" />
        <input class="input" v-model="customMessage" placeholder="Note for the admin (optional)" />

        <label class="text-sm text-muted">Payment Method</label>
        <div class="payment-grid">
          <button class="payment-option" :class="{ active: customPaymentMethod === 'ezpay' }" @click="customPaymentMethod = 'ezpay'">
            EZPay
          </button>
          <button class="payment-option" :class="{ active: customPaymentMethod === 'bepusdt' }" @click="customPaymentMethod = 'bepusdt'">
            USDT
          </button>
        </div>

        <div v-if="customPaymentMethod === 'ezpay'" class="payment-grid mt-sm">
          <button class="payment-option small" :class="{ active: customPaymentType === 'alipay' }" @click="customPaymentType = 'alipay'">Alipay</button>
          <button class="payment-option small" :class="{ active: customPaymentType === 'wxpay' }" @click="customPaymentType = 'wxpay'">WeChat</button>
        </div>

        <label class="checkbox mt-sm">
          <input type="checkbox" v-model="customUseTXB" />
          <span class="text-sm">Use {{ userStore.appConfig?.credit_name || 'TXB' }} as a discount</span>
        </label>

        <div v-if="customUseTXB && maxCustomDiscount > 0" class="discount-card">
          <div class="row-between text-sm">
            <span class="text-muted">Discount</span>
            <span>¥{{ customDiscountRMB.toFixed(2) }}</span>
          </div>
          <input class="slider" type="range" min="0" :max="maxCustomDiscount" step="0.01" v-model.number="customDiscountRMB" />
          <div class="row-between text-xs text-muted">
            <span>0</span>
            <span>{{ customTXBUsed.toFixed(0) }} {{ userStore.appConfig?.credit_name || 'TXB' }}</span>
            <span>¥{{ maxCustomDiscount.toFixed(2) }}</span>
          </div>
        </div>

        <div class="row-between text-sm">
          <span class="text-muted">Final payment</span>
          <span class="mono">¥{{ customFinalPrice.toFixed(2) }}</span>
        </div>

        <button class="btn btn-primary btn-block" @click="submitCustomPayment" :disabled="customSubmitting">
          {{ customSubmitting ? 'Submitting...' : `Create Payment ¥${customFinalPrice.toFixed(2)}` }}
        </button>
        <p class="text-xs text-muted">Admins are notified only after the payment succeeds.</p>
      </div>
    </div>

    <teleport to="body">
      <transition name="fade">
        <div class="modal-overlay" v-if="selectedCombo" @click.self="selectedCombo = null">
          <div class="modal card">
            <h3 class="mb-md">Purchase {{ selectedCombo.name }}</h3>

            <div class="row-between mb-md">
              <span class="text-muted">Original price</span>
              <span class="mono">¥{{ comboPrice.toFixed(2) }}</span>
            </div>

            <div class="stack-sm">
              <label class="text-sm text-muted">Payment Method</label>
              <div class="payment-grid">
                <button class="payment-option" :class="{ active: paymentMethod === 'ezpay' }" @click="paymentMethod = 'ezpay'">
                  EZPay
                </button>
                <button class="payment-option" :class="{ active: paymentMethod === 'bepusdt' }" @click="paymentMethod = 'bepusdt'">
                  USDT
                </button>
              </div>

              <div v-if="paymentMethod === 'ezpay'" class="payment-grid mt-sm">
                <button class="payment-option small" :class="{ active: paymentType === 'alipay' }" @click="paymentType = 'alipay'">Alipay</button>
                <button class="payment-option small" :class="{ active: paymentType === 'wxpay' }" @click="paymentType = 'wxpay'">WeChat</button>
              </div>

              <label class="checkbox mt-md">
                <input type="checkbox" v-model="useTXB" />
                <span class="text-sm">Use {{ userStore.appConfig?.credit_name || 'TXB' }} as a discount</span>
              </label>

              <div v-if="useTXB && maxComboDiscount > 0" class="discount-card">
                <div class="row-between text-sm">
                  <span class="text-muted">Discount</span>
                  <span>¥{{ discountRMB.toFixed(2) }}</span>
                </div>
                <input class="slider" type="range" min="0" :max="maxComboDiscount" step="0.01" v-model.number="discountRMB" />
                <div class="row-between text-xs text-muted">
                  <span>0</span>
                  <span>{{ comboTXBUsed.toFixed(0) }} {{ userStore.appConfig?.credit_name || 'TXB' }}</span>
                  <span>¥{{ maxComboDiscount.toFixed(2) }}</span>
                </div>
              </div>

              <div class="row-between text-sm">
                <span class="text-muted">Final price</span>
                <span class="mono price-value">¥{{ comboFinalPrice.toFixed(2) }}</span>
              </div>
            </div>

            <div class="row mt-lg" style="gap: var(--space-sm)">
              <button class="btn btn-secondary" style="flex:1" @click="selectedCombo = null">Cancel</button>
              <button class="btn btn-primary" style="flex:2" @click="purchase" :disabled="purchasing">
                {{ purchasing ? 'Processing...' : `Confirm Payment ¥${comboFinalPrice.toFixed(2)}` }}
              </button>
            </div>
          </div>
        </div>
      </transition>
    </teleport>
  </div>
</template>

<style scoped>
.combo-card {
  cursor: pointer;
}

.combo-price {
  text-align: right;
}

.price-value {
  font-family: var(--font-display);
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--accent-primary);
}

.price-cycle {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.combo-specs {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-sm);
}

.spec-item {
  font-size: 0.75rem;
  color: var(--text-secondary);
  background: var(--bg-glass-strong);
  padding: 2px var(--space-sm);
  border-radius: 100px;
}

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

.checkbox input[type="checkbox"] {
  width: 18px;
  height: 18px;
  accent-color: var(--accent-primary);
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
</style>
