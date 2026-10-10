<script setup lang="ts">
import type { IPLookupQuote, IPLookupState } from '@/api/ipLookup'
import { formatMoney } from '@/utils/format'

const ip = defineModel<string>({ required: true })
defineProps<{ state: IPLookupState; quote: IPLookupQuote | null; refreshQuote: IPLookupQuote | null; busy: boolean; validIp: boolean }>()
defineEmits<{ check: []; refresh: [] }>()
</script>

<template>
  <section class="lookup-input">
    <div class="lookup-allowance">
      <strong>{{ state.allowance ? $t('ipLookup.allowance', { remaining: state.allowance.remaining, total: state.allowance.total }) : $t('ipLookup.noAllowance') }}</strong>
      <small v-if="state.allowance">{{ $t('ipLookup.until', { date: new Date(state.allowance.validUntil).toLocaleDateString() }) }}</small>
    </div>
    <form class="lookup-form" @submit.prevent="$emit('check')">
      <UFormField name="ip" :label="$t('ipLookup.address')" :hint="$t('ipLookup.addressHint')">
        <UInput v-model.trim="ip" class="w-full" inputmode="text" autocomplete="off" :disabled="busy" :placeholder="$t('ipLookup.placeholder')" />
      </UFormField>
      <div class="lookup-actions">
        <UButton type="submit" :loading="busy" :disabled="busy || !validIp || !quote" :label="quote?.useQuota ? $t('ipLookup.checkIncluded') : $t('ipLookup.checkPaid', { fee: quote ? formatMoney(quote.charge) : state.lookupFee ? formatMoney(state.lookupFee) : '' })" />
        <UButton v-if="refreshQuote" type="button" color="neutral" variant="outline" :disabled="busy" :label="$t('ipLookup.refreshPaid', { fee: formatMoney(refreshQuote.charge) })" @click="$emit('refresh')" />
      </div>
    </form>
    <p class="lookup-note">{{ $t('ipLookup.billingNote') }}</p>
  </section>
</template>

<style scoped>
.lookup-input { display: grid; gap: 1rem; padding: 1.25rem 0; border-block: 1px solid var(--line); }
.lookup-allowance { display: grid; gap: 0.25rem; }
.lookup-allowance small, .lookup-note { color: var(--text-muted); font-size: 0.82rem; }
.lookup-form { display: grid; gap: 1rem; }
.lookup-actions { display: flex; flex-wrap: wrap; gap: 0.6rem; }
.lookup-actions :deep(button) { min-height: 44px; }
.lookup-form :deep(input) { font-size: 16px; }
@media (min-width: 900px) { .lookup-form { grid-template-columns: minmax(0, 1fr) auto; align-items: end; } }
</style>
