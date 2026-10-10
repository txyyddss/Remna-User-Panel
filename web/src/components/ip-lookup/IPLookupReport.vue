<script setup lang="ts">
import { computed } from 'vue'
import type { IPLookupCheck } from '@/api/ipLookup'
import { formatMoney } from '@/utils/format'
import ProviderDetails from './ProviderDetails.vue'
import { factKeys, providerName, reasonText } from './presentation'

const props = defineProps<{ check: IPLookupCheck }>()
const report = computed(() => props.check.report)
</script>

<template>
  <section v-if="report" class="lookup-report" aria-live="polite">
    <header class="report-heading">
      <div><p class="eyebrow">{{ report.ip }}</p><h2 :class="`verdict-${report.verdict}`">{{ $t(`ipLookup.verdicts.${report.verdict}`) }}</h2></div>
      <span>{{ check.cached ? $t('ipLookup.cached') : $t('ipLookup.fresh') }}</span>
    </header>
    <p class="report-note">{{ $t('ipLookup.predictionNote') }}</p>
    <ul v-if="report.reasons.length" class="report-reasons"><li v-for="reason in report.reasons" :key="reason">{{ reasonText(reason) }}</li></ul>
    <dl class="report-facts">
      <div v-for="key in factKeys" :key="key">
        <dt>{{ $t(`ipLookup.facts.${key}`) }}</dt>
        <dd>{{ key === 'networkType' && report.facts[key] ? $t(`ipLookup.networkTypes.${report.facts[key]}`) : report.facts[key] || $t('ipLookup.unknown') }}<small v-if="report.sources[key]">{{ providerName(report.sources[key]) }}</small></dd>
      </div>
    </dl>
    <p class="report-note">{{ $t('ipLookup.checkedAt', { date: new Date(report.checkedAt).toLocaleString() }) }} · {{ check.refunded ? $t('ipLookup.outageRefund') : check.cached && check.charge.minor === '0' ? $t('ipLookup.cachedFree') : check.usedQuota ? $t('ipLookup.usedIncluded') : formatMoney(check.charge) }}</p>
    <ProviderDetails v-for="provider in report.providers" :key="provider.id" :provider="provider" />
  </section>
</template>

<style scoped>
.lookup-report { padding: 1.6rem 0; }
.report-heading { display: flex; flex-wrap: wrap; align-items: start; justify-content: space-between; gap: 0.6rem; }
.report-heading span, .report-note, .report-facts dt, .report-facts small { color: var(--text-muted); font-size: 0.82rem; }
.report-heading h2 { margin: 0.2rem 0; font-size: 1.45rem; }
.verdict-unsuitable { color: var(--ui-error); }
.verdict-suitable { color: var(--ui-success); }
.report-reasons { padding-left: 1.2rem; margin-block: 1rem; }
.report-facts { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 1rem; padding-block: 1rem; }
.report-facts dd { margin: 0.25rem 0 0; overflow-wrap: anywhere; }
.report-facts small { display: block; margin-top: 0.2rem; }
@media (min-width: 900px) { .report-facts { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
</style>
