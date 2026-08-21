<script setup lang="ts">
import { computed, onMounted, shallowRef, watch } from 'vue'

import type { MemberRefundQuote, OperationStatus, Purchase, TrafficResetQuote } from '@/api/types'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { type PurchaseOperationKind, usePurchaseOperations } from '@/composables/usePurchaseOperations'
import { useTelegramBackButton } from '@/composables/useTelegramBackButton'
import { t } from '@/i18n'
import { formatDateTime, formatMoney } from '@/utils/format'
import AutoRenewalControl from './AutoRenewalControl.vue'
import TrafficResetAutomationControl from './TrafficResetAutomationControl.vue'

const props = defineProps<{ purchase: Purchase }>()
const emit = defineEmits<{ changed: [] }>()
const dialogKind = shallowRef<PurchaseOperationKind | null>(null)
const operation = usePurchaseOperations(() => props.purchase.id)
let emittedReceiptId = ''

const dialogOpen = computed({
  get: () => dialogKind.value !== null,
  set: (open) => { if (!open && !operation.mutating.value) dialogKind.value = null },
})
const ownsBack = computed(() => dialogOpen.value)
const currentQuote = computed<TrafficResetQuote | MemberRefundQuote | null>(() => (
  dialogKind.value === 'reset' ? operation.resetQuote.value : operation.refundQuote.value
))
const quoteEligible = computed(() => currentQuote.value?.eligible === true)
const refundEligible = computed(() => operation.refundQuote.value?.eligible === true)
const operationTone = computed(() => operation.receipt.value?.status === 'succeeded' ? 'success' : 'warning')
const operationMessage = computed(() => t(`purchaseOperations.operation.${operation.activeKind.value ?? 'reset'}.${operation.receipt.value?.status ?? 'queued'}`))

function quoteAmount(quote: TrafficResetQuote | MemberRefundQuote): string {
  return formatMoney('price' in quote ? quote.price : quote.refund)
}

function reasonMessage(reason: string | null | undefined): string {
  if (!reason) return t('purchaseOperations.ineligible')
  const key = `purchaseOperations.reason.${reason}`
  const message = t(key)
  return message === key ? t('purchaseOperations.ineligible') : message
}

function statusLabel(status: OperationStatus): string {
  return t(`operations.status.${status}`)
}

async function openQuote(kind: PurchaseOperationKind): Promise<void> {
  if (operation.blocksMutations.value) {
    dialogKind.value = operation.activeKind.value ?? kind
    return
  }
  operation.dismissOperation()
  dialogKind.value = kind
  await operation.loadQuote(kind)
}

function closeDialog(): void {
  if (!operation.mutating.value) dialogKind.value = null
}

watch(() => operation.receipt.value, (receipt) => {
  if (!receipt || receipt.id === emittedReceiptId || !['succeeded', 'compensated'].includes(receipt.status)) return
  emittedReceiptId = receipt.id
  emit('changed')
  void operation.loadRefundEligibility()
})
watch(() => props.purchase.id, () => {
  operation.reset()
  void operation.loadRefundEligibility()
})

onMounted(() => void operation.loadRefundEligibility())
useTelegramBackButton(ownsBack, closeDialog)
</script>

<template>
  <div class="purchase-actions">
    <div class="purchase-actions__row">
      <AutoRenewalControl class="purchase-actions__primary" :purchase="purchase" @changed="emit('changed')" />
      <USkeleton v-if="operation.refundEligibilityLoading.value" class="purchase-actions__refund h-11" />
      <UButton
        v-else-if="refundEligible"
        class="purchase-actions__refund"
        color="warning"
        variant="soft"
        icon="i-ph-arrow-u-up-left"
        :disabled="operation.blocksMutations.value"
        :label="$t('purchaseOperations.refundAction')"
        data-haptic
        @click="openQuote('refund')"
      />
      <UTooltip :text="$t('purchaseOperations.resetAction')">
        <UButton
          class="purchase-actions__reset"
          color="neutral"
          variant="outline"
          square
          icon="i-ph-arrows-clockwise"
          :disabled="operation.blocksMutations.value"
          :aria-label="$t('purchaseOperations.resetAction')"
          :title="$t('purchaseOperations.resetAction')"
          data-haptic
          @click="openQuote('reset')"
        />
      </UTooltip>
    </div>

    <InlineNotice v-if="operation.receipt.value && !dialogOpen" :tone="operationTone" :title="statusLabel(operation.receipt.value.status)">
      {{ operationMessage }}
    </InlineNotice>
    <UButton
      v-if="operation.receipt.value && operation.error.value && !dialogOpen"
      color="neutral"
      variant="outline"
      icon="i-ph-arrow-clockwise"
      :loading="operation.checking.value"
      :label="$t('operations.checkStatus')"
      data-haptic
      @click="operation.refresh"
    />

    <UModal
      v-model:open="dialogOpen"
      :title="$t(`purchaseOperations.${dialogKind ?? 'reset'}.title`)"
      :description="$t(`purchaseOperations.${dialogKind ?? 'reset'}.description`)"
      :dismissible="!operation.mutating.value"
      :close="{ 'data-haptic': '' }"
      :ui="{ footer: 'justify-end' }"
    >
      <template #body>
        <div class="purchase-operation-dialog">
          <TrafficResetAutomationControl
            v-if="dialogKind === 'reset'"
            :enabled="operation.resetAutomation.value?.enabled ?? null"
            :loading="operation.resetAutomationLoading.value"
            :saving="operation.resetAutomationSaving.value"
            :error="operation.resetAutomationError.value"
            @update="operation.setResetAutomation"
          />
          <USkeleton v-if="operation.quoteLoading.value" class="h-28" />
          <template v-else-if="currentQuote">
            <dl class="purchase-operation-quote">
              <div><dt>{{ $t('purchaseOperations.amount') }}</dt><dd>{{ quoteAmount(currentQuote) }}</dd></div>
              <div v-if="'resetStrategy' in currentQuote"><dt>{{ $t('purchaseOperations.cadence') }}</dt><dd>{{ $t(`home.reset.${currentQuote.resetStrategy}`) }}</dd></div>
              <div v-if="'eligibilityExpiresAt' in currentQuote && currentQuote.eligibilityExpiresAt"><dt>{{ $t('purchaseOperations.eligibleUntil') }}</dt><dd>{{ formatDateTime(currentQuote.eligibilityExpiresAt) }}</dd></div>
            </dl>
            <InlineNotice v-if="!currentQuote.eligible" tone="warning">{{ reasonMessage(currentQuote.reasonCode) }}</InlineNotice>
          </template>
          <InlineNotice v-if="operation.receipt.value" :tone="operationTone" :title="statusLabel(operation.receipt.value.status)">
            {{ operationMessage }}
          </InlineNotice>
          <InlineNotice v-if="operation.error.value" tone="warning">{{ operation.error.value }}</InlineNotice>
          <UButton
            v-if="operation.receipt.value && operation.error.value"
            color="neutral"
            variant="outline"
            icon="i-ph-arrow-clockwise"
            :loading="operation.checking.value"
            :label="$t('operations.checkStatus')"
            data-haptic
            @click="operation.refresh"
          />
        </div>
      </template>
      <template #footer>
        <UButton color="neutral" variant="outline" :disabled="operation.mutating.value" :label="$t('common.close')" data-haptic @click="closeDialog" />
        <UButton
          v-if="!operation.receipt.value"
          :color="dialogKind === 'refund' ? 'warning' : 'primary'"
          :icon="dialogKind === 'refund' ? 'i-ph-arrow-u-up-left' : 'i-ph-arrows-clockwise'"
          :loading="operation.mutating.value"
          :disabled="operation.quoteLoading.value || !quoteEligible"
          :label="$t(`purchaseOperations.${dialogKind ?? 'reset'}.confirm`)"
          data-haptic="heavy"
          @click="dialogKind && operation.start(dialogKind)"
        />
      </template>
    </UModal>
  </div>
</template>

<style scoped>
.purchase-actions, .purchase-operation-dialog { display: grid; gap: 0.75rem; }
.purchase-actions__row { display: flex; flex-wrap: wrap; align-items: stretch; gap: 0.55rem; }
.purchase-actions__primary { min-width: 0; flex: 1 1 auto; }
.purchase-actions__refund { min-width: 0; flex: 1 1 8rem; }
.purchase-actions__reset { width: 44px; min-width: 44px; height: 44px; display: inline-flex; align-items: center; justify-content: center; padding: 0; }
.purchase-operation-quote { display: grid; gap: 0.5rem; margin: 0; }
.purchase-operation-quote > div { display: flex; align-items: baseline; justify-content: space-between; gap: 1rem; padding-bottom: 0.5rem; border-bottom: 1px solid var(--line); }
.purchase-operation-quote dt { color: var(--text-faint); font-size: 0.72rem; }
.purchase-operation-quote dd { margin: 0; font-family: var(--font-mono); font-size: 0.78rem; font-weight: 700; text-align: right; overflow-wrap: anywhere; }
</style>
