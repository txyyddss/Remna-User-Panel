<script setup lang="ts">
import { computed } from 'vue'
import type { AbuseNode } from '@/api/abuse'
import InlineNotice from '@/components/common/InlineNotice.vue'
import { useI18n } from '@/i18n'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{ nodes: AbuseNode[]; statistics: Record<string, number>; busy: boolean }>()
const emit = defineEmits<{ copy: [id: string]; rotate: [id: string] }>()
const { t } = useI18n()
const statistics = computed(() => ['average', 'minimum', 'maximum'].map(name => ({ name, value: props.statistics[name] ?? 0 })))
</script>

<template>
  <section class="card nodes-card">
    <div>
      <p class="eyebrow">{{ t('adminAbuse.nodesEyebrow') }}</p>
      <h3>{{ t('adminAbuse.nodesTitle') }}</h3>
    </div>
    <div class="stats" :aria-label="t('adminAbuse.statisticsTitle')">
      <span v-for="item in statistics" :key="item.name">
        <small>{{ t(`adminAbuse.stat.${item.name}`) }}</small>
        <strong>{{ item.value }}</strong>
      </span>
    </div>
    <InlineNotice v-if="!nodes.length" tone="info">{{ t('adminAbuse.emptyNodes') }}</InlineNotice>
    <article v-for="node in nodes" :key="node.uuid" class="node">
      <div>
        <strong>{{ node.name }}</strong>
        <small>{{ node.lastReportAt ? formatDateTime(node.lastReportAt) : t('adminAbuse.neverReported') }}</small>
      </div>
      <div class="node-actions">
        <UButton color="neutral" variant="ghost" icon="i-ph-copy" :aria-label="t('adminAbuse.copyKey')" data-haptic="copy" @click="emit('copy', node.uuid)" />
        <UButton color="neutral" variant="ghost" icon="i-ph-arrows-clockwise" :loading="busy" :aria-label="t('adminAbuse.rotateKey')" data-haptic="confirm" @click="emit('rotate', node.uuid)" />
      </div>
    </article>
  </section>
</template>

<style scoped>
.card { display: grid; gap: 0.85rem; padding: 1rem; border: 1px solid var(--line); border-radius: var(--radius-panel); background: var(--surface-raised); }
.stats { display: grid; grid-template-columns: repeat(3, 1fr); gap: 0.6rem; }
.stats span { display: grid; gap: 0.15rem; }
.node { display: flex; min-height: 44px; align-items: center; justify-content: space-between; gap: 0.75rem; border-top: 1px solid var(--line); }
.node > div:first-child { display: grid; min-width: 0; gap: 0.2rem; }
.node strong { overflow-wrap: anywhere; }
.node small, .stats small { color: var(--text-muted); }
.node-actions { display: flex; gap: 0.35rem; }
</style>
