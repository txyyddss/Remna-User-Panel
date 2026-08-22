<script setup lang="ts">
import { computed } from 'vue'

import { useI18n } from '@/i18n'
import { selectionHaptic } from '@/utils/telegram'
import { paymentCurrencyLogo, paymentNetworkLogo, type PaymentChannelOption } from './paymentOptions'

const props = defineProps<{
  channels: readonly PaymentChannelOption[]
  selectedMethodId: string | null
}>()
const emit = defineEmits<{ choose: [id: string] }>()
const { t } = useI18n()

const currencies = computed(() => {
  const seen = new Set<string>()
  return props.channels.filter((channel) => {
    if (!channel.cryptoCurrency || seen.has(channel.cryptoCurrency)) return false
    seen.add(channel.cryptoCurrency)
    return true
  })
})
const selectedCurrency = computed(() => props.channels.find((item) => item.value === props.selectedMethodId)?.cryptoCurrency
  ?? currencies.value[0]?.cryptoCurrency)
const networks = computed(() => props.channels.filter((channel) => channel.cryptoCurrency === selectedCurrency.value))

function choose(id: string): void {
  if (id === props.selectedMethodId) return
  selectionHaptic()
  emit('choose', id)
}

function chooseCurrency(currency: 'USDT' | 'USDC'): void {
  const channel = props.channels.find((item) => item.cryptoCurrency === currency && !item.disabled)
  if (channel) choose(channel.value)
}
</script>

<template>
  <fieldset class="crypto-picker">
    <legend>{{ t('cryptoPayment.chooseCurrency') }}</legend>
    <div class="crypto-picker__currencies">
      <UButton
        v-for="item in currencies"
        :key="item.cryptoCurrency"
        class="crypto-currency"
        :class="{ 'crypto-currency--selected': selectedCurrency === item.cryptoCurrency }"
        color="neutral"
        variant="ghost"
        :disabled="item.disabled"
        :aria-pressed="selectedCurrency === item.cryptoCurrency"
        @click="item.cryptoCurrency && chooseCurrency(item.cryptoCurrency)"
      >
        <img :src="paymentCurrencyLogo(item.cryptoCurrency!)" alt="" width="34" height="34" />
        <strong>{{ item.cryptoCurrency }}</strong>
        <UIcon v-if="selectedCurrency === item.cryptoCurrency" name="i-ph-check-circle-fill" aria-hidden="true" />
      </UButton>
    </div>
  </fieldset>

  <fieldset class="crypto-picker">
    <legend>{{ t('cryptoPayment.chooseNetwork') }}</legend>
    <div class="crypto-picker__networks">
      <UButton
        v-for="item in networks"
        :key="item.value"
        class="provider-option crypto-network"
        :class="{ 'provider-option--selected': selectedMethodId === item.value }"
        color="neutral"
        variant="ghost"
        :disabled="item.disabled"
        :aria-pressed="selectedMethodId === item.value"
        @click="choose(item.value)"
      >
        <span class="provider-option__icon">
          <img :src="paymentNetworkLogo(item.network ?? '', item.cryptoCurrency!)" alt="" width="28" height="28" />
        </span>
        <span><strong>{{ item.networkName || item.label }}</strong></span>
        <UIcon v-if="selectedMethodId === item.value" class="provider-option__check" name="i-ph-check-circle-fill" aria-hidden="true" />
      </UButton>
    </div>
  </fieldset>
</template>
