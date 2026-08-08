<script setup lang="ts">
import { computed } from 'vue'
import { PhCoins } from '@phosphor-icons/vue'

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
  hint: 'Enter TXB in major units, including cents when needed.',
  minMinor: '0',
  maxMinor: undefined,
  required: false,
  disabled: false,
})

const model = defineModel<string>({ required: true })
const minor = computed(() => moneyFromTxbInput(model.value))
const valid = computed(() => {
  if (minor.value === '') return false
  const value = BigInt(minor.value)
  return value >= BigInt(props.minMinor)
    && (props.maxMinor === undefined || value <= BigInt(props.maxMinor))
})
const message = computed(() => {
  if (!model.value || valid.value) return props.hint
  if (minor.value === '') return 'Use a number with no more than two decimal places.'
  if (BigInt(minor.value) < BigInt(props.minMinor)) return 'Enter a larger TXB amount.'
  return 'Enter a smaller TXB amount.'
})

defineExpose({ minor, valid })
</script>

<template>
  <label class="txb-field" :for="id">
    <span class="field-label">{{ label }}</span>
    <span class="amount-input amount-input--compact" :class="{ 'amount-input--invalid': model && !valid }">
      <PhCoins :size="20" weight="fill" />
      <input
        :id="id"
        v-model.trim="model"
        type="text"
        inputmode="decimal"
        autocomplete="off"
        :required="required"
        :disabled="disabled"
        :aria-invalid="model ? !valid : undefined"
        :aria-describedby="`${id}-help`"
      />
      <span>TXB</span>
    </span>
    <small :id="`${id}-help`" :class="valid || !model ? 'field-hint' : 'field-error'">{{ message }}</small>
  </label>
</template>

<style scoped>
.txb-field {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.amount-input--compact {
  min-height: 52px;
  margin: 0;
}

.amount-input--compact input {
  font-size: 1rem;
}

.amount-input--invalid {
  border-color: var(--danger);
}
</style>
