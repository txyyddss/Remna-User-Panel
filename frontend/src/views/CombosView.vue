<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
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

// Custom payment
const showCustomPayment = ref(false)
const customAmount = ref(10)
const customMessage = ref('')
const customSubmitting = ref(false)

const cycleName: Record<string, string> = {
  monthly: '月付',
  quarterly: '季付',
  semiannual: '半年付',
  annual: '年付',
}

function formatTraffic(gb: number): string {
  if (gb >= 1024) return `${(gb / 1024).toFixed(1)} TB`
  return `${gb} GB`
}

const txbRate = computed(() => {
  return userStore.appConfig?.credit?.txb_to_rmb_rate || 100
})

const comboPrice = computed(() => {
  return selectedCombo.value?.price_rmb || 0
})

const maxTXBDiscount = computed(() => {
  if (!useTXB.value || !selectedCombo.value) return 0
  const userCredit = userStore.credit || 0
  const maxDiscountRMB = Math.floor(userCredit / txbRate.value)
  return Math.min(maxDiscountRMB, comboPrice.value)
})

const txbUsed = computed(() => maxTXBDiscount.value * txbRate.value)

const finalPrice = computed(() => {
  return Math.max(0, comboPrice.value - maxTXBDiscount.value)
})

async function purchase() {
  if (!selectedCombo.value) return
  purchasing.value = true
  try {
    const resp = await api.purchaseCombo({
      combo_uuid: selectedCombo.value.uuid,
      payment_method: paymentMethod.value,
      payment_type: paymentType.value,
      use_txb: useTXB.value,
    })
    if (resp.payment_url) {
      window.Telegram?.WebApp?.openLink(resp.payment_url)
    } else {
      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    }
    selectedCombo.value = null
  } catch (e: any) {
    alert(e.message)
  }
  purchasing.value = false
}

async function submitCustomPayment() {
  if (customAmount.value <= 0) return
  customSubmitting.value = true
  try {
    await api.customPayment(customAmount.value, customMessage.value)
    showCustomPayment.value = false
    customAmount.value = 10
    customMessage.value = ''
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    alert('已提交，请等待管理员处理')
  } catch (e: any) {
    alert(e.message)
  }
  customSubmitting.value = false
}

onMounted(async () => {
  try { combos.value = (await api.listCombos()) || [] } catch (e) {}
  loading.value = false
})
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">🚀 套餐列表</h1>
      <p class="page-subtitle">选择适合你的方案</p>
    </div>

    <div class="loading-page" v-if="loading"><div class="loading-spinner"></div></div>

    <div class="stack" v-else>
      <div v-for="combo in combos" :key="combo.uuid" class="card combo-card" @click="selectedCombo = combo">
        <div class="row-between">
          <h3>{{ combo.name }}</h3>
          <div class="combo-price">
            <span class="price-value">¥{{ combo.price_rmb }}</span>
            <span class="price-cycle">/{{ cycleName[combo.cycle] || combo.cycle }}</span>
          </div>
        </div>
        <p class="text-sm text-muted mt-sm">{{ combo.description }}</p>
        <div class="combo-specs mt-md">
          <span class="spec-item">📦 {{ formatTraffic(combo.traffic_gb) }}</span>
          <span class="spec-item">🔄 {{ combo.strategy }}</span>
          <span class="spec-item">💰 重置 ¥{{ combo.reset_price }}</span>
        </div>
      </div>
    </div>

    <div v-if="combos.length === 0 && !loading" class="empty-state">
      <span class="empty-state-icon">📦</span>
      <p class="empty-state-text">暂无可用套餐</p>
    </div>

    <!-- Custom Payment Button -->
    <div class="card mt-md" v-if="!loading">
      <div class="row-between">
        <h3>💰 自定义充值</h3>
        <button class="btn btn-sm btn-secondary" @click="showCustomPayment = !showCustomPayment">
          {{ showCustomPayment ? '取消' : '充值' }}
        </button>
      </div>
      <div v-if="showCustomPayment" class="stack-sm mt-md">
        <input class="input" v-model.number="customAmount" type="number" placeholder="充值金额 (¥)" min="1" step="1" />
        <input class="input" v-model="customMessage" placeholder="备注 (可选)" />
        <button class="btn btn-primary btn-block" @click="submitCustomPayment" :disabled="customSubmitting">
          {{ customSubmitting ? '提交中...' : `提交 ¥${customAmount} 充值请求` }}
        </button>
        <p class="text-xs text-muted">提交后管理员将手动处理</p>
      </div>
    </div>

    <!-- Purchase Modal -->
    <teleport to="body">
      <transition name="fade">
        <div class="modal-overlay" v-if="selectedCombo" @click.self="selectedCombo = null">
          <div class="modal card">
            <h3 class="mb-md">购买 {{ selectedCombo.name }}</h3>

            <div class="row-between mb-md">
              <span class="text-muted">价格</span>
              <span class="mono">¥{{ selectedCombo.price_rmb }}</span>
            </div>

            <div class="stack-sm">
              <label class="text-sm text-muted">支付方式</label>
              <div class="payment-grid">
                <button class="payment-option" :class="{ active: paymentMethod === 'ezpay' }" @click="paymentMethod = 'ezpay'">
                  💳 易支付
                </button>
                <button class="payment-option" :class="{ active: paymentMethod === 'bepusdt' }" @click="paymentMethod = 'bepusdt'">
                  🪙 USDT
                </button>
              </div>

              <div v-if="paymentMethod === 'ezpay'" class="payment-grid mt-sm">
                <button class="payment-option small" :class="{ active: paymentType === 'alipay' }" @click="paymentType = 'alipay'">支付宝</button>
                <button class="payment-option small" :class="{ active: paymentType === 'wxpay' }" @click="paymentType = 'wxpay'">微信</button>
              </div>

              <label class="checkbox mt-md">
                <input type="checkbox" v-model="useTXB" />
                <span class="text-sm">使用 {{ userStore.appConfig?.credit_name || 'TXB' }} 折扣</span>
              </label>

              <div v-if="useTXB && maxTXBDiscount > 0" class="discount-info mt-sm">
                <div class="row-between text-sm">
                  <span class="text-muted">TXB抵扣</span>
                  <span class="text-accent">-¥{{ maxTXBDiscount }} ({{ txbUsed.toFixed(0) }} TXB)</span>
                </div>
                <div class="row-between text-sm mt-xs">
                  <span class="text-muted">实际支付</span>
                  <span class="mono price-value">¥{{ finalPrice }}</span>
                </div>
              </div>
            </div>

            <div class="row mt-lg" style="gap: var(--space-sm)">
              <button class="btn btn-secondary" style="flex:1" @click="selectedCombo = null">取消</button>
              <button class="btn btn-primary" style="flex:2" @click="purchase" :disabled="purchasing">
                {{ purchasing ? '处理中...' : `确认支付 ¥${finalPrice}` }}
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

.discount-info {
  padding: var(--space-sm);
  background: rgba(0, 206, 201, 0.08);
  border-radius: var(--radius-sm);
  border: 1px solid rgba(0, 206, 201, 0.2);
}

.text-accent {
  color: var(--accent-secondary, #00cec9);
}
</style>
