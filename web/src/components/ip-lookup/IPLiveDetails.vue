<script setup lang="ts">
import type { IPLookupReport } from '@/api/ipLookup'
import { useLiveIPDetails } from './useLiveIPDetails'
import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import IPBGPTopology from './IPBGPTopology.vue'
import IPASNDetails from './IPASNDetails.vue'
import LiveSourceNote from './LiveSourceNote.vue'
const props = defineProps<{ requestedIp: string; report: IPLookupReport }>()
const { details, loading, errorCode, refresh } = useLiveIPDetails(() => props.requestedIp, () => props.report.status === 'processing' ? '' : `${props.report.id}:${props.report.checkedAt}`)
</script>

<template>
  <section class="live-details">
    <header class="live-heading"><div><h2>{{ $t('ipLookupLive.title') }}</h2><p>{{ $t('ipLookupLive.description') }}</p></div><UButton color="neutral" variant="ghost" icon="i-ph-arrow-clockwise" :loading="loading" :disabled="loading" :label="$t('ipLookupLive.refresh')" @click="refresh" /></header>
    <SkeletonBlock v-if="loading" height="14rem" />
    <InlineNotice v-else-if="errorCode" tone="warning">{{ $t(`ipLookupLive.errors.${errorCode}`) }}</InlineNotice>
    <template v-else-if="details">
      <IPBGPTopology :topology="details.topology" />
      <IPASNDetails :asns="details.asns" />
      <div class="live-ip-facts">
        <section><h3>{{ $t('ipLookupLive.scores') }}</h3><dl><div v-for="kind in ['company', 'asn'] as const" :key="kind"><dt>{{ $t(`ipLookup.scores.${kind}_abuse_ratio`) }}</dt><dd>{{ details.scores[kind].ratio === null ? $t('ipLookupLive.status.empty') : `${(details.scores[kind].ratio! * 100).toFixed(3)}%` }}<span v-if="details.scores[kind].label">{{ details.scores[kind].label }}</span></dd></div></dl><LiveSourceNote :meta="details.scores" /></section>
        <section><h3>{{ $t('ipLookupLive.blockTitle') }}</h3><code v-if="details.block.prefix">{{ details.block.prefix }}</code><p>{{ $t('ipLookupLive.blockCount', { reported: details.block.reportedAddresses ?? $t('ipLookupLive.status.empty'), capacity: details.block.addressCapacity || $t('ipLookupLive.status.empty'), days: details.block.windowDays }) }}</p><LiveSourceNote :meta="details.block" /></section>
        <section><h3>{{ $t('ipLookupLive.ptrTitle') }}</h3><ul v-if="details.ptr.domains.length"><li v-for="domain in details.ptr.domains" :key="domain"><code>{{ domain }}</code></li></ul><LiveSourceNote :meta="details.ptr" /></section>
      </div>
      <p class="live-fetch-time">{{ $t('ipLookupLive.fetchedAt', { date: new Date(details.fetchedAt).toLocaleString() }) }}</p>
      <section v-if="details.shodan" class="live-shodan"><h3>{{ $t('ipLookupLive.shodanTitle') }}</h3><dl><div><dt>{{ $t('ipLookupLive.shodanHostnames') }}</dt><dd>{{ details.shodan.hostnames.length ? details.shodan.hostnames.join(', ') : $t('ipLookupLive.status.empty') }}</dd></div><div><dt>{{ $t('ipLookupLive.shodanPorts') }}</dt><dd><code>{{ details.shodan.ports.length ? details.shodan.ports.join(', ') : $t('ipLookupLive.status.empty') }}</code></dd></div></dl><LiveSourceNote :meta="details.shodan" /><p>{{ $t('ipLookupLive.shodanContext') }}</p></section>
    </template>
  </section>
</template>

<style scoped>
.live-details { min-width: 0; padding-top: 1.2rem; border-top: 1px solid var(--line); }
.live-heading { display: flex; align-items: start; justify-content: space-between; gap: 0.8rem; margin-bottom: 1rem; }
.live-heading h2, .live-ip-facts h3 { font-size: 1rem; margin: 0; }
.live-heading p, .live-fetch-time { font-size: 0.78rem; color: var(--text-muted); margin: 0.35rem 0 0; line-height: 1.5; }
.live-ip-facts { display: grid; gap: 1.3rem; padding-block: 1rem; border-top: 1px solid var(--line); }
.live-ip-facts section { min-width: 0; }
.live-ip-facts dl { margin: 0.6rem 0 0; display: grid; gap: 0.6rem; }
.live-ip-facts dt { color: var(--text-muted); font-size: 0.78rem; }
.live-ip-facts dd { margin: 0.25rem 0 0; font-size: 0.88rem; }
.live-ip-facts dd span { margin-left: 0.5rem; color: var(--text-muted); }
.live-ip-facts p, .live-ip-facts code { font-size: 0.82rem; overflow-wrap: anywhere; }
.live-ip-facts ul { margin: 0.7rem 0; padding-left: 1rem; }
.live-shodan { padding-block: 1rem; border-top: 1px solid var(--line); }
.live-shodan h3 { font-size: 1rem; margin: 0; }
.live-shodan dl { display: grid; gap: 0.9rem; margin-block: 0.75rem; }
.live-shodan dt, .live-shodan p { font-size: 0.78rem; color: var(--text-muted); }
.live-shodan dd { margin: 0.25rem 0 0; font-size: 0.85rem; overflow-wrap: anywhere; }
@media (min-width: 900px) { .live-ip-facts { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
</style>
