<script setup lang="ts">
import type { IPLookupProvider } from '@/api/ipLookup'
import { factKeys, providerName, riskKeys } from './presentation'

defineProps<{ provider: IPLookupProvider }>()
</script>

<template>
  <details class="provider-details">
    <summary><strong>{{ providerName(provider.id) }}</strong><span>{{ $t(`ipLookup.statuses.${provider.status}`) }}</span></summary>
    <template v-if="provider.status === 'success' || provider.status === 'partial'">
      <dl class="provider-data">
        <div v-for="key in riskKeys" :key="key"><dt>{{ $t(`ipLookup.risks.${key}`) }}</dt><dd>{{ provider.signals[key] === null ? $t('ipLookup.unknown') : provider.signals[key] ? $t('ipLookup.detected') : $t('ipLookup.notDetected') }}</dd></div>
        <div v-for="key in factKeys" :key="key"><dt>{{ $t(`ipLookup.facts.${key}`) }}</dt><dd>{{ key === 'networkType' && provider.facts[key] ? $t(`ipLookup.networkTypes.${provider.facts[key]}`) : provider.facts[key] || $t('ipLookup.unknown') }}</dd></div>
        <div v-for="(value, key) in provider.scores" :key="key"><dt>{{ $t(`ipLookup.scores.${key}`) }}</dt><dd>{{ value }}</dd></div>
        <div v-if="provider.reports !== null"><dt>{{ $t('ipLookup.reports') }}</dt><dd>{{ provider.reports }}</dd></div>
      </dl>
    </template>
    <p v-else>{{ $t(`ipLookup.statusNotes.${provider.status}`) }}</p>
    <p v-if="provider.status === 'partial'">{{ $t('ipLookup.partialNote') }}</p>
  </details>
</template>

<style scoped>
.provider-details { border-top: 1px solid var(--line); }
.provider-details summary { display: flex; justify-content: space-between; align-items: center; gap: 0.75rem; min-height: 52px; cursor: pointer; }
.provider-details summary span, .provider-details p, .provider-data dt { color: var(--text-muted); font-size: 0.82rem; }
.provider-data { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 0.8rem; padding-bottom: 1rem; }
.provider-data dd { margin: 0.25rem 0 0; overflow-wrap: anywhere; }
@media (min-width: 900px) { .provider-data { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
</style>
