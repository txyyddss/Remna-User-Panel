<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useUserStore } from '@/stores/user'
import { api } from '@/api'

const userStore = useUserStore()
const loading = ref(true)
const qcCode = ref('')
const qcLoading = ref(false)
const qcMessage = ref('')
const parentalRating = ref(0)
const devices = ref<any[]>([])
const showPwdForm = ref(false)
const currentPwd = ref('')
const newPwd = ref('')

// Purchase modal state
const showPurchase = ref(false)
const purchaseMonths = ref(1)
const paymentMethod = ref('ezpay')
const paymentType = ref('alipay')
const useTXB = ref(false)
const purchasing = ref(false)

const jellyfinPrice = computed(() => {
  return userStore.appConfig?.jellyfin?.monthly_price_rmb || 2
})

const totalPrice = computed(() => {
  return jellyfinPrice.value * purchaseMonths.value
})

const txbRate = computed(() => {
  return userStore.appConfig?.credit?.txb_to_rmb_rate || 100
})

const maxTXBDiscount = computed(() => {
  if (!useTXB.value) return 0
  const userCredit = userStore.credit || 0
  const maxDiscountRMB = Math.floor(userCredit / txbRate.value)
  return Math.min(maxDiscountRMB, totalPrice.value)
})

const txbUsed = computed(() => maxTXBDiscount.value * txbRate.value)

const finalPrice = computed(() => {
  return Math.max(0, totalPrice.value - maxTXBDiscount.value)
})

onMounted(async () => {
  try {
    if (userStore.hasJellyfin) {
      const devResp = await api.jellyfinGetDevices()
      devices.value = devResp?.Items || []
      parentalRating.value = userStore.jellyfin?.parental_rating || 0
    }
  } catch (e) {}
  loading.value = false
})

async function authorizeQC() {
  if (!qcCode.value) return
  qcLoading.value = true
  qcMessage.value = ''
  try {
    await api.jellyfinQuickConnect(qcCode.value)
    qcMessage.value = '✅ 授权成功！'
    qcCode.value = ''
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    qcMessage.value = '❌ ' + e.message
  }
  qcLoading.value = false
}

async function updateRating() {
  try {
    await api.jellyfinUpdateParentalRating(parentalRating.value)
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e) {}
}

async function changePassword() {
  try {
    await api.jellyfinUpdatePassword(currentPwd.value, newPwd.value)
    showPwdForm.value = false
    currentPwd.value = ''
    newPwd.value = ''
    window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
  } catch (e: any) {
    alert(e.message)
  }
}

async function purchaseJellyfin() {
  purchasing.value = true
  try {
    const resp = await api.purchaseJellyfin({
      months: purchaseMonths.value,
      payment_method: paymentMethod.value,
      payment_type: paymentType.value,
      use_txb: useTXB.value,
    })
    if (resp.payment_url) {
      window.Telegram?.WebApp?.openLink(resp.payment_url)
    } else {
      window.Telegram?.WebApp?.HapticFeedback?.notificationOccurred('success')
    }
    showPurchase.value = false
  } catch (e: any) {
    alert(e.message)
  }
  purchasing.value = false
}
</script>

<template>
  <div class="page">
    <div class="page-header">
      <h1 class="page-title">🎬 Jellyfin</h1>
    </div>

    <div class="loading-page" v-if="loading"><div class="loading-spinner"></div></div>

    <template v-else-if="userStore.hasJellyfin">
      <div class="card">
        <div class="row-between">
          <h3>账户信息</h3>
          <span class="badge badge-success">活跃</span>
        </div>
        <div class="text-sm text-muted mt-sm">
          到期: {{ new Date(userStore.jellyfin?.expires_at).toLocaleDateString('zh-CN') }}
        </div>
      </div>

      <!-- Parental Rating -->
      <div class="card mt-md">
        <h3 class="mb-sm">🔒 内容分级</h3>
        <p class="text-xs text-muted mb-md">调整可观看内容的最高分级 (0=全部限制, 22=无限制)</p>
        <div class="row-between mb-sm">
          <span class="text-sm">分级: {{ parentalRating }}</span>
          <span class="text-xs text-muted">0 ~ 22</span>
        </div>
        <input type="range" min="0" max="22" step="1" v-model.number="parentalRating" @change="updateRating" />
      </div>

      <!-- Quick Connect -->
      <div class="card mt-md">
        <h3 class="mb-sm">⚡ Quick Connect</h3>
        <div class="row" style="gap:var(--space-sm)">
          <input class="input" v-model="qcCode" placeholder="输入授权码" style="flex:1" />
          <button class="btn btn-primary btn-sm" @click="authorizeQC" :disabled="qcLoading">授权</button>
        </div>
        <div v-if="qcMessage" class="text-sm mt-sm">{{ qcMessage }}</div>
      </div>

      <!-- Password -->
      <div class="card mt-md">
        <div class="row-between">
          <h3>🔑 密码管理</h3>
          <button class="btn btn-sm btn-secondary" @click="showPwdForm = !showPwdForm">{{ showPwdForm ? '取消' : '修改' }}</button>
        </div>
        <div v-if="showPwdForm" class="stack-sm mt-md">
          <input class="input" v-model="currentPwd" type="password" placeholder="当前密码" />
          <input class="input" v-model="newPwd" type="password" placeholder="新密码" />
          <button class="btn btn-primary btn-sm" @click="changePassword">确认修改</button>
        </div>
      </div>

      <!-- Devices -->
      <div class="card mt-md" v-if="devices.length > 0">
        <h3 class="mb-md">📱 设备列表</h3>
        <div class="stack-sm">
          <div v-for="dev in devices" :key="dev.Id" class="device-item">
            <div>
              <div class="text-sm">{{ dev.AppName || '未知应用' }}</div>
              <div class="text-xs text-muted">{{ dev.Name }}</div>
            </div>
            <span class="text-xs text-muted">{{ new Date(dev.DateLastActivity).toLocaleDateString('zh-CN') }}</span>
          </div>
        </div>
      </div>
    </template>

    <div class="empty-state" v-else>
      <span class="empty-state-icon">🎬</span>
      <p class="empty-state-text">还没有 Jellyfin 账户</p>
      <p class="text-xs text-muted mt-sm">¥{{ jellyfinPrice }}/月 · 支持多设备</p>
      <button class="btn btn-primary mt-md" @click="showPurchase = true">开通影视服务</button>
    </div>

    <!-- Purchase Modal -->
    <teleport to="body">
      <transition name="fade">
        <div class="modal-overlay" v-if="showPurchase" @click.self="showPurchase = false">
          <div class="modal card">
            <h3 class="mb-md">开通 Jellyfin 影视服务</h3>

            <div class="stack-sm">
              <label class="text-sm text-muted">购买时长</label>
              <div class="payment-grid">
                <button class="payment-option small" :class="{ active: purchaseMonths === 1 }" @click="purchaseMonths = 1">1个月</button>
                <button class="payment-option small" :class="{ active: purchaseMonths === 3 }" @click="purchaseMonths = 3">3个月</button>
                <button class="payment-option small" :class="{ active: purchaseMonths === 6 }" @click="purchaseMonths = 6">6个月</button>
                <button class="payment-option small" :class="{ active: purchaseMonths === 12 }" @click="purchaseMonths = 12">12个月</button>
              </div>

              <div class="row-between mt-sm">
                <span class="text-muted">价格</span>
                <span class="mono">¥{{ totalPrice }}</span>
              </div>

              <label class="text-sm text-muted mt-sm">支付方式</label>
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
              <button class="btn btn-secondary" style="flex:1" @click="showPurchase = false">取消</button>
              <button class="btn btn-primary" style="flex:2" @click="purchaseJellyfin" :disabled="purchasing">
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
