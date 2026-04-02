<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'

const combos = ref<any[]>([])
const loading = ref(true)
const purchasing = ref(false)
const selectedCombo = ref<any>(null)
const paymentMethod = ref('ezpay')
const paymentType = ref('alipay')
const useTXB = ref(false)

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
                <span class="text-sm">使用 TXB 折扣</span>
              </label>
            </div>

            <div class="row mt-lg" style="gap: var(--space-sm)">
              <button class="btn btn-secondary" style="flex:1" @click="selectedCombo = null">取消</button>
              <button class="btn btn-primary" style="flex:2" @click="purchase" :disabled="purchasing">
                {{ purchasing ? '处理中...' : '确认购买' }}
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
</style>
