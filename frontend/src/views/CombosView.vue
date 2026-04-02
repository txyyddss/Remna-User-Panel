<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { api } from '@/api'
import { useUserStore } from '@/stores/user'
import { parseSanitizedDecimal, sanitizeDecimalInput } from '@/utils/number'
import type { Combo } from '@/types'

const userStore = useUserStore()

const combos = ref<Combo[]>([])
const loading = ref(true)
const purchasing = ref(false)
const selectedCombo = ref<Combo | null>(null)
const error = ref('')

const paymentMethod = ref('ezpay')
const paymentType = ref('alipay')
const usdtNetwork = ref('usdt.polygon')
const useTXB = ref(false)
const discountRMB = ref(0)

const showCustomPayment = ref(false)
const customAmount = ref('10')
const customMessage = ref('')
const customPaymentMethod = ref('ezpay')
const customPaymentType = ref('alipay')
const customUsdtNetwork = ref('usdt.polygon')
const customUseTXB = ref(false)
const customDiscountRMB = ref(0)
const customSubmitting = ref(false)

const cycleName: Record<string, string> = {
  monthly: 'Monthly',
  quarterly: 'Quarterly',
  semiannual: 'Semi-Annual',
  annual: 'Annual',
}

const usdtNetworks = computed(() => userStore.appConfig?.payments?.usdt_networks ?? [
  { value: 'usdt.aptos', label: 'USDT Aptos' },
  { value: 'usdt.arbitrum', label: 'USDT Arbitrum' },
  { value: 'usdt.polygon', label: 'USDT Polygon' },
])

const txbRate = computed(() => userStore.appConfig?.credit?.txb_to_rmb_rate ?? userStore.appConfig?.txb_to_rmb_rate ?? 100)
const comboPrice = computed(() => selectedCombo.value?.price_rmb || 0)
const customAmountValue = computed(() => parseSanitizedDecimal(customAmount.value))

const maxComboDiscount = computed(() => {
  if (!useTXB.value || !selectedCombo.value) {
    return 0
  }
  return Math.max(0, Math.min(comboPrice.value, Math.floor((userStore.credit / txbRate.value) * 100) / 100))
})

const comboFinalPrice = computed(() => Math.max(0, comboPrice.value - discountRMB.value))
const comboTXBUsed = computed(() => discountRMB.value * txbRate.value)

const maxCustomDiscount = computed(() => {
  if (!customUseTXB.value || customAmountValue.value <= 0) {
    return 0
  }
  return Math.max(0, Math.min(customAmountValue.value, Math.floor((userStore.credit / txbRate.value) * 100) / 100))
})

const customFinalPrice = computed(() => Math.max(0, customAmountValue.value - customDiscountRMB.value))
const customTXBUsed = computed(() => customDiscountRMB.value * txbRate.value)

watch(useTXB, (enabled) => {
  if (!enabled) {
    discountRMB.value = 0
  }
})

watch(customUseTXB, (enabled) => {
  if (!enabled) {
    customDiscountRMB.value = 0
  }
})

watch(maxComboDiscount, (value) => {
  if (discountRMB.value > value) {
    discountRMB.value = value
  }
})

watch(maxCustomDiscount, (value) => {
  if (customDiscountRMB.value > value) {
    customDiscountRMB.value = value
  }
})

watch(selectedCombo, (value) => {
  if (!value) {
    paymentMethod.value = 'ezpay'
    paymentType.value = 'alipay'
    usdtNetwork.value = 'usdt.polygon'
    useTXB.value = false
    discountRMB.value = 0
  }
})

function formatTraffic(gb: number): string {
  if (gb >= 1024) {
    return `${(gb / 1024).toFixed(1)} TB`
  }
  return `${gb} GB`
}

function getSelectedPaymentType(method: string, fallbackType: string, network: string): string {
  return method === 'bepusdt' ? network : fallbackType
}

function onCustomAmountInput(event: Event) {
  customAmount.value = sanitizeDecimalInput((event.target as HTMLInputElement).value)
}

async function loadCombos() {
  try {
    combos.value = (await api.listCombos()) || []
  } catch (e: any) {
    error.value = e.message || 'Failed to load plans.'
    combos.value = []
  } finally {
    loading.value = false
  }
}

async function purchase() {
  if (!selectedCombo.value) {
    return
  }

  purchasing.value = true
  error.value = ''
  try {
    const resp = await api.purchaseCombo({
      combo_uuid: selectedCombo.value.uuid,
      payment_method: paymentMethod.value,
      payment_type: getSelectedPaymentType(paymentMethod.value, paymentType.value, usdtNetwork.value),
      use_txb: useTXB.value,
      discount_rmb: discountRMB.value,
    })

    await userStore.refreshState({ background: true })
    if (resp.payment_url) {
      window.Telegram?.WebApp?.openLink(resp.payment_url)
    }
    selectedCombo.value = null
  } catch (e: any) {
    error.value = e.message
  } finally {
    purchasing.value = false
  }
}

async function submitCustomPayment() {
  if (customAmountValue.value <= 0) {
    error.value = 'Enter a valid top-up amount.'
    return
  }

  customSubmitting.value = true
  error.value = ''
  try {
    const resp = await api.customPayment({
      amount: customAmountValue.value,
      message: customMessage.value.trim(),
      payment_method: customPaymentMethod.value,
      payment_type: getSelectedPaymentType(customPaymentMethod.value, customPaymentType.value, customUsdtNetwork.value),
      use_txb: customUseTXB.value,
      discount_rmb: customDiscountRMB.value,
    })

    await userStore.refreshState({ background: true })
    if (resp.payment_url) {
      window.Telegram?.WebApp?.openLink(resp.payment_url)
    }

    showCustomPayment.value = false
    customAmount.value = '10'
    customMessage.value = ''
    customUseTXB.value = false
    customDiscountRMB.value = 0
  } catch (e: any) {
    error.value = e.message
  } finally {
    customSubmitting.value = false
  }
}

onMounted(() => {
  void loadCombos()
})
</script>

<template>
  <div class="page">
    <div class="page-header stagger-enter stagger-1">
      <h1 class="page-title">Plans</h1>
      <p class="page-subtitle">Buy, renew, or top up your account. Pending payments expire automatically after 30 minutes.</p>
    </div>

    <div v-if="error" class="card text-danger text-sm mb-md">{{ error }}</div>
    <div class="loading-page" v-if="loading"><div class="loading-spinner"></div></div>

    <div class="stack stagger-enter stagger-2" v-else>
      <div v-for="combo in combos" :key="combo.uuid" class="card combo-card" @click="selectedCombo = combo">
        <div class="row-between combo-head">
          <div>
            <h3>{{ combo.name }}</h3>
            <p class="text-sm text-muted mt-sm">{{ combo.description }}</p>
          </div>
          <div class="combo-price">
            <span class="price-value">{{ combo.price_rmb.toFixed(2) }}</span>
            <span class="price-cycle">RMB / {{ cycleName[combo.cycle] || combo.cycle }}</span>
          </div>
        </div>
        <div class="combo-specs mt-md">
          <span class="spec-item">{{ formatTraffic(combo.traffic_gb) }}</span>
          <span class="spec-item">{{ combo.strategy }}</span>
          <span class="spec-item">Reset fee {{ combo.reset_price.toFixed(2) }} RMB</span>
        </div>
      </div>

      <div v-if="combos.length === 0" class="empty-state">
        <p class="empty-state-text">No plans are available right now.</p>
      </div>

      <div class="card">
        <div class="row-between">
          <div>
            <h3>Custom Top-up</h3>
            <p class="text-sm text-muted mt-sm">Create a manual top-up order for your account.</p>
          </div>
          <button class="btn btn-secondary btn-sm" @click="showCustomPayment = !showCustomPayment">
            {{ showCustomPayment ? 'Close' : 'Open' }}
          </button>
        </div>

        <div v-if="showCustomPayment" class="stack-sm mt-md">
          <input
            class="input"
            :value="customAmount"
            inputmode="decimal"
            placeholder="Amount in RMB"
            @input="onCustomAmountInput"
          />
          <input class="input" v-model="customMessage" placeholder="Order note for admins (optional)" />

          <div class="payment-panel">
            <label class="field-label">Payment method</label>
            <div class="payment-grid">
              <button class="payment-option" :class="{ active: customPaymentMethod === 'ezpay' }" @click="customPaymentMethod = 'ezpay'">EZPay</button>
              <button class="payment-option" :class="{ active: customPaymentMethod === 'bepusdt' }" @click="customPaymentMethod = 'bepusdt'">USDT</button>
            </div>

            <div v-if="customPaymentMethod === 'ezpay'" class="payment-grid compact mt-sm">
              <button class="payment-option" :class="{ active: customPaymentType === 'alipay' }" @click="customPaymentType = 'alipay'">Alipay</button>
              <button class="payment-option" :class="{ active: customPaymentType === 'wxpay' }" @click="customPaymentType = 'wxpay'">WeChat Pay</button>
            </div>

            <div v-else class="payment-grid compact mt-sm">
              <button
                v-for="network in usdtNetworks"
                :key="network.value"
                class="payment-option"
                :class="{ active: customUsdtNetwork === network.value }"
                @click="customUsdtNetwork = network.value"
              >
                {{ network.label }}
              </button>
            </div>
          </div>

          <label class="checkbox-row">
            <input type="checkbox" v-model="customUseTXB" />
            <span>Use {{ userStore.appConfig?.credit_name || 'TXB' }} as discount</span>
          </label>

          <div v-if="customUseTXB && maxCustomDiscount > 0" class="discount-card">
            <div class="row-between text-sm">
              <span class="text-muted">Discount</span>
              <span>{{ customDiscountRMB.toFixed(2) }} RMB</span>
            </div>
            <input class="slider" type="range" min="0" :max="maxCustomDiscount" step="0.01" v-model.number="customDiscountRMB" />
            <div class="row-between text-xs text-muted">
              <span>0</span>
              <span>{{ customTXBUsed.toFixed(0) }} {{ userStore.appConfig?.credit_name || 'TXB' }}</span>
              <span>{{ maxCustomDiscount.toFixed(2) }} RMB</span>
            </div>
          </div>

          <div class="summary-row">
            <span class="text-muted">Final payment</span>
            <span class="mono">{{ customFinalPrice.toFixed(2) }} RMB</span>
          </div>

          <button class="btn btn-primary btn-block" @click="submitCustomPayment" :disabled="customSubmitting">
            {{ customSubmitting ? 'Creating order...' : `Create payment for ${customFinalPrice.toFixed(2)} RMB` }}
          </button>
        </div>
      </div>
    </div>

    <teleport to="body">
      <transition name="fade">
        <div v-if="selectedCombo" class="modal-overlay" @click.self="selectedCombo = null">
          <div class="modal card">
            <div class="row-between mb-md">
              <h3>{{ selectedCombo.name }}</h3>
              <span class="mono">{{ comboPrice.toFixed(2) }} RMB</span>
            </div>

            <div class="payment-panel">
              <label class="field-label">Payment method</label>
              <div class="payment-grid">
                <button class="payment-option" :class="{ active: paymentMethod === 'ezpay' }" @click="paymentMethod = 'ezpay'">EZPay</button>
                <button class="payment-option" :class="{ active: paymentMethod === 'bepusdt' }" @click="paymentMethod = 'bepusdt'">USDT</button>
              </div>

              <div v-if="paymentMethod === 'ezpay'" class="payment-grid compact mt-sm">
                <button class="payment-option" :class="{ active: paymentType === 'alipay' }" @click="paymentType = 'alipay'">Alipay</button>
                <button class="payment-option" :class="{ active: paymentType === 'wxpay' }" @click="paymentType = 'wxpay'">WeChat Pay</button>
              </div>

              <div v-else class="payment-grid compact mt-sm">
                <button
                  v-for="network in usdtNetworks"
                  :key="network.value"
                  class="payment-option"
                  :class="{ active: usdtNetwork === network.value }"
                  @click="usdtNetwork = network.value"
                >
                  {{ network.label }}
                </button>
              </div>
            </div>

            <label class="checkbox-row mt-md">
              <input type="checkbox" v-model="useTXB" />
              <span>Use {{ userStore.appConfig?.credit_name || 'TXB' }} as discount</span>
            </label>

            <div v-if="useTXB && maxComboDiscount > 0" class="discount-card mt-md">
              <div class="row-between text-sm">
                <span class="text-muted">Discount</span>
                <span>{{ discountRMB.toFixed(2) }} RMB</span>
              </div>
              <input class="slider" type="range" min="0" :max="maxComboDiscount" step="0.01" v-model.number="discountRMB" />
              <div class="row-between text-xs text-muted">
                <span>0</span>
                <span>{{ comboTXBUsed.toFixed(0) }} {{ userStore.appConfig?.credit_name || 'TXB' }}</span>
                <span>{{ maxComboDiscount.toFixed(2) }} RMB</span>
              </div>
            </div>

            <div class="summary-row mt-md">
              <span class="text-muted">Final payment</span>
              <span class="mono">{{ comboFinalPrice.toFixed(2) }} RMB</span>
            </div>

            <div class="row mt-lg action-row">
              <button class="btn btn-secondary" style="flex: 1" @click="selectedCombo = null">Cancel</button>
              <button class="btn btn-primary" style="flex: 2" @click="purchase" :disabled="purchasing">
                {{ purchasing ? 'Creating order...' : `Continue with ${comboFinalPrice.toFixed(2)} RMB` }}
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

.combo-head {
  align-items: flex-start;
}

.combo-price {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 2px;
}

.price-value {
  font-family: var(--font-display);
  font-size: 1.4rem;
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
  padding: 6px 10px;
  border-radius: 999px;
  background: var(--bg-glass-strong);
  color: var(--text-secondary);
  font-size: 0.75rem;
}

.field-label {
  display: block;
  font-size: 0.75rem;
  color: var(--text-muted);
  margin-bottom: var(--space-sm);
  text-transform: uppercase;
  letter-spacing: 0.12em;
}

.payment-panel {
  padding: var(--space-md);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.02);
}

.payment-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-sm);
}

.payment-grid.compact {
  grid-template-columns: 1fr;
}

.payment-option {
  min-height: 44px;
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  background: #121a2c;
  color: var(--text-secondary);
}

.payment-option.active {
  color: var(--text-primary);
  border-color: rgba(91, 141, 239, 0.55);
  background: linear-gradient(180deg, rgba(91, 141, 239, 0.18), rgba(34, 197, 94, 0.1));
}

.checkbox-row {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
}

.discount-card {
  padding: var(--space-md);
  border-radius: var(--radius-md);
  background: rgba(34, 197, 94, 0.08);
  border: 1px solid rgba(34, 197, 94, 0.2);
}

.summary-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(3, 10, 21, 0.72);
  display: flex;
  align-items: flex-end;
  z-index: 200;
}

.modal {
  width: 100%;
  max-height: 86vh;
  overflow-y: auto;
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
}

.action-row {
  gap: var(--space-sm);
}
</style>
