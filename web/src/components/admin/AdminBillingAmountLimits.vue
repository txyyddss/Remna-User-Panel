<script setup lang="ts">
import { computed, onMounted, reactive, shallowRef } from 'vue'

import { adminBillingApi } from '@/api/adminBilling'
import { api } from '@/api/client'
import InlineNotice from '@/components/common/InlineNotice.vue'
import TxbAmountField from '@/components/common/TxbAmountField.vue'
import { localizedError, useI18n } from '@/i18n'
import { formatDateTime, moneyFromTxbInput, txbInputFromMinor } from '@/utils/format'
import { notifyHaptic } from '@/utils/telegram'

const maximumInt64Minor = '9223372036854775807'
const draft = reactive({ minimum: '', maximum: '' })
const loading = shallowRef(true)
const busy = shallowRef(false)
const error = shallowRef<string | null>(null)
const saved = shallowRef(false)
const updatedAt = shallowRef<string | null>(null)
const { t } = useI18n()

const minimumMinor = computed(() => moneyFromTxbInput(draft.minimum))
const maximumMinor = computed(() => moneyFromTxbInput(draft.maximum))
const rangeInvalid = computed(() => Boolean(
  minimumMinor.value && maximumMinor.value
  && BigInt(minimumMinor.value) > BigInt(maximumMinor.value),
))
const canSave = computed(() => {
  if (!minimumMinor.value || !maximumMinor.value || rangeInvalid.value) return false
  return BigInt(minimumMinor.value) > 0n
    && BigInt(maximumMinor.value) <= BigInt(maximumInt64Minor)
})

async function load(): Promise<void> {
  loading.value = true
  error.value = null
  saved.value = false
  try {
    const response = await api.getBalance()
    draft.minimum = txbInputFromMinor(response.addAmountLimits.minimum.minor)
    draft.maximum = txbInputFromMinor(response.addAmountLimits.maximum.minor)
    updatedAt.value = response.addAmountLimits.updatedAt
  } catch (caught) {
    error.value = localizedError(caught, 'errors.adminLoad')
  } finally {
    loading.value = false
  }
}

async function save(): Promise<void> {
  if (busy.value || !canSave.value) return
  busy.value = true
  error.value = null
  saved.value = false
  try {
    const limits = await adminBillingApi.updateAmountLimits(minimumMinor.value, maximumMinor.value)
    draft.minimum = txbInputFromMinor(limits.minimum.minor)
    draft.maximum = txbInputFromMinor(limits.maximum.minor)
    updatedAt.value = limits.updatedAt
    saved.value = true
    notifyHaptic('success')
  } catch (caught) {
    error.value = localizedError(caught, 'errors.adminAction')
    notifyHaptic('error')
  } finally {
    busy.value = false
  }
}

defineExpose({ save, loading })

onMounted(() => void load())
</script>

<template>
  <section class="billing-limits" aria-labelledby="billing-limits-title">
    <div class="admin-subsection-heading">
      <div>
        <h3 id="billing-limits-title">{{ t('adminSettings.amountLimits.title') }}</h3>
        <p>{{ t('adminSettings.amountLimits.copy') }}</p>
      </div>
    </div>
    <USkeleton v-if="loading" class="billing-limits__skeleton" />
    <form v-else class="billing-limits__form" @submit.prevent>
      <div class="billing-limits__fields">
        <TxbAmountField
          id="billing-minimum-txb"
          v-model="draft.minimum"
          :label="t('adminSettings.amountLimits.minimum')"
          :hint="t('adminSettings.amountLimits.minimumHint')"
          min-minor="1"
          :max-minor="maximumInt64Minor"
          required
          :disabled="busy"
        />
        <TxbAmountField
          id="billing-maximum-txb"
          v-model="draft.maximum"
          :label="t('adminSettings.amountLimits.maximum')"
          :hint="t('adminSettings.amountLimits.maximumHint')"
          min-minor="1"
          :max-minor="maximumInt64Minor"
          required
          :disabled="busy"
        />
      </div>
      <InlineNotice v-if="rangeInvalid" tone="warning">{{ t('adminSettings.amountLimits.rangeInvalid') }}</InlineNotice>
      <InlineNotice v-if="error" tone="warning">{{ error }}</InlineNotice>
      <InlineNotice v-if="saved" tone="success">{{ t('adminSettings.amountLimits.saved') }}</InlineNotice>
      <div v-if="error" class="button-row billing-limits__actions">
        <UButton v-if="error" type="button" color="neutral" variant="outline" icon="i-ph-arrow-clockwise" :label="t('adminSection.retry')" @click="load" />
      </div>
      <small v-if="updatedAt" class="billing-limits__updated">{{ t('adminSettings.amountLimits.updated', { date: formatDateTime(updatedAt) }) }}</small>
    </form>
  </section>
</template>

<style scoped>
.billing-limits { display: grid; gap: 0.8rem; padding: 1rem; border-bottom: 1px solid var(--line); }
.billing-limits__skeleton { width: 100%; height: 7rem; }
.billing-limits__form, .billing-limits__fields { display: grid; gap: 0.8rem; }
.billing-limits__fields { grid-template-columns: repeat(auto-fit, minmax(min(100%, 15rem), 1fr)); }
.billing-limits__actions { justify-content: flex-end; }
.billing-limits__updated { color: var(--text-muted); text-align: right; }
</style>
