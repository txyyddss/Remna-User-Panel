<script setup lang="ts">
import { computed } from 'vue'
import QrcodeVue from 'qrcode.vue'
import { usePaymentSheet } from '@/composables/usePaymentSheet'
import { useToast } from '@/composables/useToast'

const toast = useToast()
const { payment, hasPayment, closePaymentSheet } = usePaymentSheet()

const openTarget = computed(() => payment.value?.url_scheme || payment.value?.payment_url || '')
const hasQrContent = computed(() => !!payment.value?.qr_content)
const expiryLabel = computed(() => {
    const seconds = payment.value?.expires_in_seconds || 0
    if (seconds <= 0) {
        return ''
    }
    const minutes = Math.floor(seconds / 60)
    const remainder = seconds % 60
    return `${minutes}m ${String(remainder).padStart(2, '0')}s`
})

async function copyValue(value: string | undefined, label: string) {
    if (!value) {
        toast.error(`No ${label.toLowerCase()} to copy.`)
        return
    }

    try {
        await navigator.clipboard.writeText(value)
        toast.success(`${label} copied.`)
    } catch {
        toast.error(`Failed to copy ${label.toLowerCase()}.`)
    }
}

function openExternal() {
    if (!openTarget.value) {
        toast.error('No external payment link is available.')
        return
    }

    if (payment.value?.url_scheme) {
        window.location.href = payment.value.url_scheme
        return
    }

    if (window.Telegram?.WebApp?.openLink) {
        window.Telegram.WebApp.openLink(openTarget.value)
        return
    }

    window.open(openTarget.value, '_blank', 'noopener,noreferrer')
}
</script>

<template>
  <teleport to="body">
    <transition name="modal-slide">
      <div v-if="hasPayment && payment" class="modal-overlay" @click.self="closePaymentSheet()">
        <div class="modal card">
          <div class="row-between modal-head">
            <div>
              <div class="eyebrow">Payment Details</div>
              <h3 class="modal-title">Order {{ payment.order_uuid.slice(0, 8) }}</h3>
            </div>
            <button class="btn btn-secondary btn-sm" @click="closePaymentSheet()">Close</button>
          </div>

          <div class="payment-summary">
            <div class="summary-card">
              <span class="summary-label">Final RMB</span>
              <strong class="summary-value">{{ payment.final_amount.toFixed(2) }} RMB</strong>
            </div>
            <div v-if="payment.display_amount && payment.display_currency" class="summary-card">
              <span class="summary-label">Gateway Amount</span>
              <strong class="summary-value">{{ payment.display_amount }} {{ payment.display_currency }}</strong>
            </div>
          </div>

          <div class="qr-card">
            <div v-if="hasQrContent" class="qr-wrap">
              <QrcodeVue :value="payment.qr_content || ''" :size="220" level="M" render-as="svg" />
            </div>
            <div v-else class="qr-fallback">
              <div class="fallback-title">QR not available</div>
              <p class="fallback-copy">This gateway only returned a redirect page. Use the external open button below.</p>
            </div>
          </div>

          <div class="stack-sm details-list">
            <div class="detail-row">
              <span class="detail-label">Order ID</span>
              <button class="detail-action" @click="copyValue(payment.order_uuid, 'Order ID')">{{ payment.order_uuid }}</button>
            </div>
            <div v-if="payment.trade_id" class="detail-row">
              <span class="detail-label">Trade ID</span>
              <span class="detail-value mono">{{ payment.trade_id }}</span>
            </div>
            <div v-if="payment.payment_address" class="detail-row">
              <span class="detail-label">Payment Address</span>
              <button class="detail-action" @click="copyValue(payment.payment_address, 'Payment address')">{{ payment.payment_address }}</button>
            </div>
            <div v-if="payment.network" class="detail-row">
              <span class="detail-label">Network</span>
              <span class="detail-value">{{ payment.network }}</span>
            </div>
            <div v-if="payment.qr_content" class="detail-row">
              <span class="detail-label">QR Content</span>
              <button class="detail-action" @click="copyValue(payment.qr_content, 'QR content')">Copy raw payment content</button>
            </div>
            <div v-if="expiryLabel" class="detail-row">
              <span class="detail-label">Expires In</span>
              <span class="detail-value">{{ expiryLabel }}</span>
            </div>
          </div>

          <div class="action-row">
            <button class="btn btn-secondary" style="flex: 1" @click="closePaymentSheet()">Done</button>
            <button class="btn btn-primary" style="flex: 2" :disabled="!openTarget" @click="openExternal()">
              {{ payment.url_scheme ? 'Open Payment App' : 'Open Externally' }}
            </button>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(3, 10, 21, 0.76);
  backdrop-filter: blur(6px);
  -webkit-backdrop-filter: blur(6px);
  display: flex;
  align-items: flex-end;
  z-index: 300;
}

.modal {
  width: 100%;
  max-height: 88vh;
  overflow-y: auto;
  border-bottom-left-radius: 0;
  border-bottom-right-radius: 0;
}

.modal-head {
  align-items: flex-start;
  gap: var(--space-md);
}

.eyebrow {
  font-size: 0.72rem;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--accent-secondary);
}

.modal-title {
  margin-top: 6px;
}

.payment-summary {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: var(--space-sm);
  margin-top: var(--space-md);
}

.summary-card {
  padding: var(--space-md);
  border-radius: var(--radius-md);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border-subtle);
}

.summary-label {
  display: block;
  font-size: 0.72rem;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.12em;
}

.summary-value {
  display: block;
  margin-top: 8px;
  font-size: 1rem;
}

.qr-card {
  margin-top: var(--space-md);
  padding: var(--space-lg);
  border-radius: var(--radius-lg);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.04), rgba(255, 255, 255, 0.02));
  border: 1px solid var(--border-subtle);
}

.qr-wrap {
  display: flex;
  justify-content: center;
  padding: var(--space-md);
  background: #ffffff;
  border-radius: var(--radius-md);
}

.qr-fallback {
  text-align: center;
  color: var(--text-secondary);
}

.fallback-title {
  font-weight: 600;
}

.fallback-copy {
  margin-top: var(--space-sm);
  font-size: 0.875rem;
}

.details-list {
  margin-top: var(--space-md);
}

.detail-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: var(--space-sm) 0;
  border-bottom: 1px solid var(--border-subtle);
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-label {
  font-size: 0.72rem;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--text-muted);
}

.detail-value,
.detail-action {
  font-size: 0.92rem;
  color: var(--text-primary);
  word-break: break-all;
}

.detail-action {
  padding: 0;
  background: transparent;
  border: none;
  text-align: left;
}

.action-row {
  display: flex;
  gap: var(--space-sm);
  margin-top: var(--space-lg);
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

.modal-slide-enter-from,
.modal-slide-leave-to {
  opacity: 0;
}

.modal-slide-enter-from .modal,
.modal-slide-leave-to .modal {
  transform: translateY(30%);
}

@media (max-width: 520px) {
  .payment-summary {
    grid-template-columns: 1fr;
  }

  .action-row {
    flex-direction: column;
  }
}
</style>
