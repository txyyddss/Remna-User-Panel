<script setup lang="ts">
import { computed } from 'vue'
import type { AbuseNode, AbuseRecord } from '@/api/abuse'
import { useI18n } from '@/i18n'

defineProps<{ nodes: AbuseNode[]; records: AbuseRecord[]; statistics: Record<string, number>; busy: boolean }>()
const emit = defineEmits<{ copy: [id: string]; rotate: [id: string] }>()
const { t } = useI18n()
const columns = computed(() => [
  { accessorKey: 'reason', header: t('adminAbuse.reason') },
  { accessorKey: 'measuredQPS', header: t('adminAbuse.qps') },
  { accessorKey: 'action', header: t('adminAbuse.action') },
])
</script>

<template>
  <div class="nodes-records-stack">
    <section class="card">
      <h3>{{ t('adminAbuse.statisticsTitle') }}</h3>
      <div class="stats">
        <span v-for="(value, name) in statistics" :key="name">
          <small>{{ t(`adminAbuse.stat.${name}`) }}</small>
          <strong>{{ value }}</strong>
        </span>
      </div>
    </section>

    <section class="card">
      <h3>{{ t('adminAbuse.nodesTitle') }}</h3>
      <article v-for="node in nodes" :key="node.uuid" class="node">
        <div>
          <strong>{{ node.name }}</strong>
          <small>{{ node.lastReportAt || t('adminAbuse.neverReported') }}</small>
        </div>
        <div class="node__actions">
          <UButton color="neutral" variant="ghost" icon="i-ph-copy" :aria-label="t('adminAbuse.copyKey')" @click="emit('copy', node.uuid)" />
          <UButton color="neutral" variant="ghost" icon="i-ph-arrows-clockwise" :loading="busy" :aria-label="t('adminAbuse.rotateKey')" @click="emit('rotate', node.uuid)" />
        </div>
      </article>
    </section>

    <section class="card">
      <h3>{{ t('adminAbuse.recordsTitle') }}</h3>
      <UTable class="records-table" :data="records" :columns="columns" />
      <article v-for="record in records" :key="record.id" class="record">
        <strong>{{ record.reason }}</strong>
        <span>{{ record.measuredQPS }} / {{ record.qpsLimit }}</span>
        <small>{{ t(`abuse.action.${record.action}`) }}</small>
      </article>
    </section>
  </div>
</template>

<style scoped>
.nodes-records-stack,
.card {
  display: grid;
  gap: 0.75rem;
}

.card {
  padding: 1rem;
  border: 1px solid var(--line);
  border-radius: var(--radius-panel);
  background: var(--surface-raised);
}

.stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.6rem;
}

.stats span,
.node {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

.node,
.record {
  padding: 0.65rem 0;
  border-top: 1px solid var(--line);
}

.node {
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
}

.node__actions {
  display: flex;
  gap: 0.35rem;
}

.node small,
.record small,
.stats small {
  color: var(--text-muted);
}

.record {
  display: none;
}

@media (max-width: 700px) {
  .records-table {
    display: none;
  }

  .record {
    display: grid;
    gap: 0.25rem;
  }
}
</style>
