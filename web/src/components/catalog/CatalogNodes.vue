<script setup lang="ts">
import { computed } from 'vue'
import type { DeepReadonly } from 'vue'

import type { PurchaseQuote, RemnaNode } from '@/api/types'
import CountryFlag from '@/components/common/CountryFlag.vue'

type QuoteWithNodes = PurchaseQuote & { accessibleNodes?: readonly RemnaNode[] }

const props = defineProps<{
  quote: DeepReadonly<PurchaseQuote> | null
  loading: boolean
}>()

const nodes = computed(() => (props.quote as DeepReadonly<QuoteWithNodes> | null)?.accessibleNodes
  ?.filter((node) => node.accessible) ?? [])

function formatMultiplier(value: number): string {
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(value)
}
</script>

<template>
  <section class="catalog-step">
    <div class="section-heading section-heading--stacked">
      <h2>{{ $t('catalog.accessibleNodes') }}</h2>
      <p>{{ $t('catalog.accessibleNodesHint') }}</p>
    </div>
    <div v-if="loading" class="catalog-node-list">
      <USkeleton v-for="index in 3" :key="index" class="h-14" />
    </div>
    <div v-else-if="nodes.length" v-auto-animate class="catalog-node-list">
      <article v-for="node in nodes" :key="node.uuid" class="catalog-node">
        <CountryFlag :code="node.countryCode" />
        <div>
          <strong>{{ node.name }}</strong>
          <small>{{ $t('catalog.nodeMultiplier', { multiplier: formatMultiplier(node.consumptionMultiplier) }) }}</small>
        </div>
      </article>
    </div>
    <div v-else class="empty-inline">
      <div>
        <h3>{{ $t('catalog.noAccessibleNodes') }}</h3>
        <p>{{ $t('catalog.noAccessibleNodesHint') }}</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.catalog-step, .catalog-node-list { display: grid; gap: 0.75rem; }
.catalog-node { min-height: 58px; display: grid; grid-template-columns: auto minmax(0, 1fr); align-items: center; gap: 0.7rem; padding: 0.65rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface-raised); }
.catalog-node strong, .catalog-node small { display: block; }
.catalog-node strong { overflow: hidden; font-size: 0.8rem; text-overflow: ellipsis; white-space: nowrap; }
.catalog-node small { margin-top: 0.2rem; color: var(--text-faint); font-family: var(--font-mono); font-size: 0.65rem; }
</style>
