<script setup lang="ts">
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useAbuseRecords } from '@/composables/useAbuseRecords'
import { useI18n } from '@/i18n'

const { t } = useI18n(); const state = useAbuseRecords()
const when = (value: string) => new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
</script>

<template>
  <section class="abuse-page">
    <header class="page-header"><p class="eyebrow">{{ t('abuse.eyebrow') }}</p><h1>{{ t('abuse.title') }}</h1><p>{{ t('abuse.copy') }}</p></header>
    <InlineNotice v-if="state.error.value" tone="warning">{{ state.error.value }}</InlineNotice>
    <div v-if="!state.loading.value" class="abuse-page__rows"><article v-for="record in state.records.value" :key="record.id" class="abuse-row"><div><strong>{{ record.reason }}</strong><small>{{ when(record.occurredAt) }}</small></div><div><span>{{ t('abuse.qps', { measured: record.measuredQPS, limit: record.qpsLimit }) }}</span><small>{{ t(`abuse.action.${record.action}`) }}</small></div><small v-if="record.expiresAt">{{ t('abuse.expires', { time: when(record.expiresAt) }) }}</small></article><div v-if="!state.records.value.length" class="empty-inline"><div><h2>{{ t('abuse.empty') }}</h2><p>{{ t('abuse.emptyCopy') }}</p></div></div></div>
  </section>
</template>

<style scoped>
.abuse-page { display: grid; gap: 1rem; padding-bottom: max(1rem, env(safe-area-inset-bottom)); }.abuse-page__rows { display: grid; gap: .7rem; }.abuse-row { display: grid; gap: .35rem; padding: 1rem; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface-raised); }.abuse-row div { display: flex; justify-content: space-between; gap: .8rem; }.abuse-row small { color: var(--text-muted); }
</style>
