<script setup lang="ts">
import InlineNotice from '@/components/common/InlineNotice.vue'
import SkeletonBlock from '@/components/common/SkeletonBlock.vue'
import { useAbuseRecords } from '@/composables/useAbuseRecords'
import { useI18n } from '@/i18n'
import { formatDateTime } from '@/utils/format'

const { t } = useI18n()
const { records, loading, error, load } = useAbuseRecords()
</script>

<template>
  <section class="abuse-page">
    <header class="page-header"><p class="eyebrow">{{ t('abuse.eyebrow') }}</p><h1>{{ t('abuse.title') }}</h1><p>{{ t('abuse.copy') }}</p></header>
    <div v-if="loading" class="abuse-page__rows" :aria-label="t('common.loading')" aria-busy="true">
      <SkeletonBlock v-for="index in 3" :key="index" height="6rem" />
    </div>
    <div v-else-if="error" class="abuse-page__rows">
      <InlineNotice tone="warning">{{ error }}</InlineNotice>
      <UButton :label="t('common.tryAgain')" data-haptic="retry" @click="load" />
    </div>
    <div v-else-if="records.length" class="abuse-page__rows">
      <article v-for="record in records" :key="record.id" class="abuse-row">
        <div>
          <strong>{{ record.reason }}</strong>
          <small>{{ formatDateTime(record.occurredAt) }}</small>
        </div>
        <div>
          <span>{{ t('abuse.qps', { measured: record.measuredQPS, limit: record.qpsLimit }) }}</span>
          <small>{{ t(`abuse.action.${record.action}`) }}</small>
        </div>
        <small v-if="record.expiresAt">{{ t('abuse.expires', { time: formatDateTime(record.expiresAt) }) }}</small>
      </article>
    </div>
    <div v-else class="empty-inline">
      <div><h2>{{ t('abuse.empty') }}</h2><p>{{ t('abuse.emptyCopy') }}</p></div>
    </div>
  </section>
</template>

<style scoped>
.abuse-page { display: grid; gap: 1rem; padding-bottom: max(1rem, env(safe-area-inset-bottom)); }
.abuse-page__rows { min-width: 0; display: grid; gap: 0.7rem; }
.abuse-row { min-width: 0; display: grid; gap: 0.5rem; padding: 1rem; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface-raised); overflow-wrap: anywhere; }
.abuse-row div { min-width: 0; display: flex; flex-wrap: wrap; justify-content: space-between; gap: 0.35rem 0.8rem; }
.abuse-row small { color: var(--text-muted); }
</style>
