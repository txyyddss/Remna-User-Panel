<script setup lang="ts">
import { computed } from 'vue'
import type { IPLookupQuote, IPLookupState } from '@/api/ipLookup'
import { useI18n } from '@/i18n'
import { formatMoney } from '@/utils/format'

const ip = defineModel<string>({ required: true })
const props = defineProps<{ state: IPLookupState; quote: IPLookupQuote | null; busy: boolean; validIp: boolean }>()
defineEmits<{ check: [] }>()
const { t } = useI18n()
const cost = computed(() => props.quote?.cachedReport ? t('ipLookup.cachedFree') : props.quote?.useQuota
  ? t('ipLookup.checkIncluded') : t('ipLookup.checkPaid', { fee: props.quote ? formatMoney(props.quote.charge) : props.state.lookupFee ? formatMoney(props.state.lookupFee) : '' }))
</script>

<template>
  <section class="lookup-input">
    <div class="lookup-allowance">
      <span>{{ state.allowance ? $t('ipLookup.allowance', { remaining: state.allowance.remaining, total: state.allowance.total }) : $t('ipLookup.noAllowance') }}</span>
      <small v-if="state.allowance">{{ $t('ipLookup.until', { date: new Date(state.allowance.validUntil).toLocaleDateString() }) }}</small>
    </div>
    <form class="lookup-form" @submit.prevent="$emit('check')">
      <UFormField name="ip" :label="$t('ipLookup.address')" class="lookup-field">
        <UInput v-model.trim="ip" class="w-full" inputmode="text" autocomplete="off" autocapitalize="off" :spellcheck="false" :disabled="busy" :placeholder="$t('ipLookup.placeholder')" />
      </UFormField>
      <UButton type="submit" class="lookup-submit" icon="i-ph-magnifying-glass" :loading="busy" :disabled="busy || !validIp || !quote || !!quote.cachedReport" :aria-label="cost" :title="cost" />
    </form>
    <div class="lookup-caption"><span>{{ cost }}</span><small>{{ $t('ipLookup.billingNote') }}</small></div>
  </section>
</template>

<style scoped>
.lookup-input { display: grid; gap: 0.7rem; padding: 1rem 0; border-block: 1px solid var(--line); }
.lookup-allowance, .lookup-caption { display: flex; flex-wrap: wrap; align-items: center; justify-content: space-between; gap: 0.25rem 1rem; color: var(--text-muted); font-size: 0.8rem; }
.lookup-allowance small, .lookup-caption small { font-size: 0.75rem; }
.lookup-form { display: grid; grid-template-columns: minmax(0, 1fr) 44px; align-items: end; gap: 0.5rem; }
.lookup-field { min-width: 0; }
.lookup-submit { width: 44px; height: 44px; }
.lookup-form :deep(input) { font-size: 16px; }
</style>
