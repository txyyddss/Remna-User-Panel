<script setup lang="ts">
import { computed } from 'vue'
import type { AbuseRecord } from '@/api/abuse'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useI18n } from '@/i18n'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{ records: AbuseRecord[]; hasMore: boolean; loadingMore: boolean }>()
const emit = defineEmits<{ loadMore: [] }>()
const { t } = useI18n()
const columns = computed(() => [
  { accessorKey: 'occurredAt', header: t('adminAbuse.occurredAt') },
  { accessorKey: 'username', header: t('adminAbuse.username') },
  { accessorKey: 'reason', header: t('adminAbuse.reason') },
  { accessorKey: 'qps', header: t('adminAbuse.qps') },
  { accessorKey: 'action', header: t('adminAbuse.action') },
])
const tableRows = computed(() => props.records.map(record => ({ ...record, username: record.username || t('adminAbuse.unknownUsername'), occurredAt: formatDateTime(record.occurredAt), qps: `${record.measuredQPS} / ${record.qpsLimit}`, action: t(`abuse.action.${record.action}`) })))
</script>

<template>
  <section class="card records-card">
    <div>
      <p class="eyebrow">{{ t('adminAbuse.recordsEyebrow') }}</p>
      <h3>{{ t('adminAbuse.recordsTitle') }}</h3>
    </div>
    <InlineNotice v-if="!records.length" tone="info">{{ t('adminAbuse.emptyRecords') }}</InlineNotice>
    <UTable v-else class="records-table" :data="tableRows" :columns="columns" />
    <article v-for="record in records" :key="record.id" class="record">
      <div><strong>{{ record.username || t('adminAbuse.unknownUsername') }}</strong><small>{{ formatDateTime(record.occurredAt) }}</small></div>
      <small>{{ record.reason }}</small>
      <span>{{ record.measuredQPS }} / {{ record.qpsLimit }}</span>
      <small>{{ t(`abuse.action.${record.action}`) }}</small>
    </article>
    <UButton v-if="hasMore" color="neutral" variant="outline" :loading="loadingMore" :label="t('adminAbuse.loadMoreRecords')" @click="emit('loadMore')" />
  </section>
</template>

<style scoped>
.card { display: grid; gap: 0.85rem; padding: 1rem; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface-raised); }
.record { display: none; gap: 0.35rem; border-top: 1px solid var(--line); padding-top: 0.75rem; }
.record div { display: flex; justify-content: space-between; gap: 0.5rem; }
.record small { color: var(--text-muted); }
@media (max-width: 700px) { :deep(.records-table) { display: none; } .record { display: grid; } }
</style>
