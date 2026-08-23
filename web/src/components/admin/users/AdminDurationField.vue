<script setup lang="ts">
import { computed } from 'vue'

import { useI18n } from '@/i18n'
import { maxDurationAmount, type DurationDraft, type DurationUnit } from './duration'

const value = defineModel<DurationDraft>({ required: true })
const { t } = useI18n()
const unitItems = computed(() => (['minutes', 'hours', 'days'] as const).map((unit) => ({
  label: t(`adminBulkExtension.units.${unit}`), value: unit,
})))
const amount = computed({
  get: () => value.value.amount,
  set: (amount: number) => { value.value = { ...value.value, amount } },
})
const unit = computed({
  get: () => value.value.unit,
  set: (unit: DurationUnit) => {
    value.value = { amount: Math.min(value.value.amount, maxDurationAmount(unit)), unit }
  },
})
const maximum = computed(() => maxDurationAmount(unit.value))
</script>

<template>
  <div class="duration-field">
    <UFormField name="extensionAmount" :label="t('adminBulkExtension.amount')" required>
      <UInputNumber v-model="amount" class="w-full" :min="1" :max="maximum" :step="1" />
    </UFormField>
    <UFormField name="extensionUnit" :label="t('adminBulkExtension.unit')" required>
      <USelect v-model="unit" class="w-full" :items="unitItems" value-key="value" label-key="label" />
    </UFormField>
  </div>
</template>

<style scoped>
.duration-field {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(7rem, 0.55fr);
  gap: 0.75rem;
}

@media (max-width: 420px) {
  .duration-field { grid-template-columns: 1fr; }
}
</style>
