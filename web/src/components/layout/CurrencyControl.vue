<script setup lang="ts">
import { computed, shallowRef } from 'vue'

import { useDisplayCurrency } from '@/composables/useDisplayCurrency'
import { useI18n } from '@/i18n'
import { selectionHaptic } from '@/utils/telegram'

interface Props {
  showLabel?: boolean
}

const props = withDefaults(defineProps<Props>(), { showLabel: false })
const open = shallowRef(false)
const { t } = useI18n()
const { currency, currencies, error, saving, select } = useDisplayCurrency()
const options = computed(() => currencies.value.map(value => ({ value, label: currencyLabel(value) })))

function currencyLabel(value: typeof currency.value): string {
  return t({ TXB: 'app.currencyTxb', CNY: 'app.currencyCny', USD: 'app.currencyUsd' }[value])
}

async function choose(value: typeof currency.value): Promise<void> {
  if (await select(value)) {
    selectionHaptic()
    open.value = false
  }
}
</script>

<template>
  <div class="currency-control">
    <span v-if="props.showLabel" class="currency-control__label">{{ $t('app.currency') }}</span>
    <UPopover v-model:open="open">
      <UButton color="neutral" variant="ghost" size="sm" icon="i-ph-currency-circle-dollar" :label="currencyLabel(currency)" :aria-label="$t('app.currency')" />
      <template #content>
        <div class="currency-control__menu">
          <UButton
            v-for="option in options"
            :key="option.value"
            color="neutral"
            :variant="option.value === currency ? 'soft' : 'ghost'"
            :label="option.label"
            :disabled="saving"
            :aria-pressed="option.value === currency"
            @click="choose(option.value)"
          />
          <p v-if="error" class="currency-control__error" role="status">{{ error }}</p>
        </div>
      </template>
    </UPopover>
  </div>
</template>

<style scoped>
.currency-control { display: inline-flex; align-items: center; gap: 0.5rem; }
.currency-control__label { color: var(--text-muted); font-size: 0.75rem; font-weight: 700; }
.currency-control__menu { display: grid; min-width: 9rem; gap: 0.2rem; padding: 0.35rem; }
.currency-control__menu :deep(button) { justify-content: flex-start; min-height: 44px; }
.currency-control__error { margin: 0.25rem 0 0; padding: 0.3rem; color: var(--warning); font-size: 0.7rem; line-height: 1.35; }
</style>
