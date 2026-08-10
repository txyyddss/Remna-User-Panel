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
    <UInput
      :id="id"
      v-model.trim="model"
      class="amount-input amount-input--compact"
      icon="i-ph-coins-fill"
      type="text"
      inputmode="decimal"
      autocomplete="off"
      :required="required"
      :disabled="disabled"
      :aria-invalid="model ? !valid : undefined"
    >
      <template #trailing><span>{{ t('common.currencyTxb') }}</span></template>
    </UInput>
  </UFormField>
</template>

<style scoped>
.txb-field { display: flex; flex-direction: column; gap: 0.4rem; }
.amount-input--compact { width: 100%; min-height: 52px; margin: 0; }
.amount-input--compact :deep(input) { font-size: 1rem; }
</style>
