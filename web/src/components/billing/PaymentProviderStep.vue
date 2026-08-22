<script setup lang="ts">
import { useI18n } from '@/i18n'
import { selectionHaptic } from '@/utils/telegram'
import type { PaymentProviderOption } from './paymentOptions'

const props = defineProps<{
  options: readonly PaymentProviderOption[]
  selectedValue: string | undefined
  canContinue: boolean
}>()

const emit = defineEmits<{
  choose: [value: string]
  continue: []
}>()

const { t } = useI18n()

function choose(value: string): void {
  if (value === props.selectedValue) return
  selectionHaptic()
  emit('choose', value)
}
</script>

<template>
  <fieldset class="provider-picker">
    <legend>{{ t('payment.chooseProvider') }}</legend>
    <UButton
      v-for="item in options"
      :key="item.value"
      class="provider-option"
      :class="{ 'provider-option--selected': selectedValue === item.value }"
      color="neutral"
      variant="ghost"
      :disabled="!item.available"
      :aria-pressed="selectedValue === item.value"
      @click="choose(item.value)"
    >
      <span class="provider-option__icon"><UIcon :name="item.icon" aria-hidden="true" /></span>
      <span><strong>{{ item.label }}</strong><small v-if="item.description">{{ item.description }}</small></span>
      <UIcon v-if="selectedValue === item.value" class="provider-option__check" name="i-ph-check-circle-fill" aria-hidden="true" />
    </UButton>
  </fieldset>
  <UButton
    v-if="selectedValue !== 'coupon'"
    block
    :disabled="!canContinue"
    :label="t('payment.continueToChannel')"
    data-test="choose-channel"
    data-haptic="navigate"
    @click="emit('continue')"
  />
</template>
