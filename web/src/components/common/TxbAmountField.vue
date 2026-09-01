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
  integerOnly?: boolean
  slider?: boolean
  required?: boolean
  disabled?: boolean
}>(), {
  hint: undefined,
  minMinor: '0',
  maxMinor: undefined,
  integerOnly: false,
  slider: false,
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
    if (value === null || !Number.isFinite(value)) {
      model.value = ''
      return
    }
    model.value = props.integerOnly ? String(Math.trunc(value)) : value.toFixed(2)
  },
})
const minimum = computed(() => minorToNumber(props.minMinor))
const maximum = computed(() => props.maxMinor === undefined ? undefined : minorToNumber(props.maxMinor))
const showSlider = computed(() => props.slider && minimum.value !== undefined && maximum.value !== undefined && minimum.value < maximum.value)
const sliderValue = computed(() => numericModel.value ?? minimum.value ?? 0)
const sliderStep = computed(() => {
  const range = (maximum.value ?? 0) - (minimum.value ?? 0)
  if (props.integerOnly) return Math.max(1, Math.ceil(range / 200))
  return Math.max(0.01, Math.ceil((range / 200) * 100) / 100)
})
const valid = computed(() => {
  if (minor.value === '') return false
  if (props.integerOnly && !/^\d+$/.test(model.value.trim())) return false
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

function updateSlider(value: number | number[]): void {
  const next = Array.isArray(value) ? value[0] : value
  if (typeof next === 'number' && Number.isFinite(next)) numericModel.value = next
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
    <div class="amount-input amount-input--compact" :class="{ 'amount-input--slider': showSlider }">
      <UIcon name="i-ph-coins-fill" aria-hidden="true" />
      <UInputNumber
        :id="id"
        v-model="numericModel"
        class="amount-number"
        :min="minimum"
        :max="maximum"
        :step="integerOnly ? 1 : 0.01"
        :format-options="{ useGrouping: false, minimumFractionDigits: integerOnly ? 0 : 2, maximumFractionDigits: integerOnly ? 0 : 2 }"
        increment
        decrement
        :required="required"
        :disabled="disabled"
        :aria-invalid="model ? !valid : undefined"
        fixed
      />
      <span>{{ t('common.currencyTxb') }}</span>
      <USlider
        v-if="showSlider"
        class="txb-field__slider"
        :model-value="sliderValue"
        :min="minimum"
        :max="maximum"
        :step="sliderStep"
        :disabled="disabled"
        :aria-label="label"
        @update:model-value="updateSlider"
      />
    </div>
  </UFormField>
</template>

<style scoped>
.txb-field { display: flex; flex-direction: column; gap: 0.4rem; }
.amount-input--compact { width: 100%; min-height: 52px; margin: 0; }
.amount-input--slider { padding-block: 0.6rem; }
.amount-number { width: 100%; min-width: 0; }
.amount-input--compact :deep(input) { min-height: 44px; font-size: 1rem; }
.txb-field__slider { grid-column: 1 / -1; width: 100%; }
</style>
