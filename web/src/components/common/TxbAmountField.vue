<script setup lang="ts">
import { computed } from 'vue'

import { useI18n } from '@/i18n'
import { moneyFromTxbInput } from '@/utils/format'

const props = withDefaults(defineProps<{
  id: string
  label: string
  hint?: string
  minMinor?: string
  maxMinor?: string
  required?: boolean
  disabled?: boolean
}>(), {
  hint: undefined,
  minMinor: '0',
  maxMinor: undefined,
  required: false,
  disabled: false,
})

const model = defineModel<string>({ required: true })
const { t } = useI18n()
const minor = computed(() => moneyFromTxbInput(model.value))
const numericModel = computed<number | null>({
  get: () => {
    if (!model.value) return null
    const value = Number(model.value)
    return Number.isFinite(value) ? value : null
  },
  set: (value) => {
    model.value = value === null || !Number.isFinite(value) ? '' : value.toFixed(2)
  },
})
const minimum = computed(() => minorToNumber(props.minMinor))
const maximum = computed(() => props.maxMinor === undefined ? undefined : minorToNumber(props.maxMinor))
const valid = computed(() => {
  if (minor.value === '') return false
  const value = BigInt(minor.value)
  return value >= BigInt(props.minMinor)
    && (props.maxMinor === undefined || value <= BigInt(props.maxMinor))
})
const message = computed(() => {
  if (!model.value || valid.value) return props.hint ?? t('amount.hint')
  if (minor.value === '') return t('amount.invalidPrecision')
  if (BigInt(minor.value) < BigInt(props.minMinor)) return t('amount.tooSmall')
  return t('amount.tooLarge')
})
const error = computed(() => model.value && !valid.value ? message.value : undefined)

defineExpose({ minor, valid })

function minorToNumber(value: string): number | undefined {
  if (!/^\d+$/.test(value)) return undefined
  const amount = Number(BigInt(value)) / 100
  return Number.isFinite(amount) ? amount : undefined
}
</script>

<template>
  <UFormField
    :name="id"
    :label="label"
    :description="error ? undefined : message"
    :error="error"
    :required="required"
    class="txb-field"
  >
    <div class="amount-input amount-input--compact">
      <UIcon name="i-ph-coins-fill" aria-hidden="true" />
      <UInputNumber
        :id="id"
        v-model="numericModel"
        class="amount-number"
        :min="minimum"
        :max="maximum"
        :step="0.01"
        :format-options="{ useGrouping: false, minimumFractionDigits: 2, maximumFractionDigits: 2 }"
        increment
        decrement
        :required="required"
        :disabled="disabled"
        :aria-invalid="model ? !valid : undefined"
        fixed
      />
      <span>{{ t('common.currencyTxb') }}</span>
    </div>
  </UFormField>
</template>

<style scoped>
.txb-field { display: flex; flex-direction: column; gap: 0.4rem; }
.amount-input--compact { width: 100%; min-height: 52px; margin: 0; }
.amount-number { width: 100%; min-width: 0; }
.amount-input--compact :deep(input) { min-height: 44px; font-size: 1rem; }
</style>
