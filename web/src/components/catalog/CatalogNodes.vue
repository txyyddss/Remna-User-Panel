<script setup lang="ts">
import { computed, shallowRef } from 'vue'
import type { DeepReadonly } from 'vue'

import type { PurchaseQuote, RemnaNode } from '@/api/types'
import CountryFlag from '@/components/common/CountryFlag.vue'
import { t } from '@/i18n'

type QuoteWithNodes = PurchaseQuote & { accessibleNodes?: readonly RemnaNode[] }

const props = defineProps<{
  quote: DeepReadonly<PurchaseQuote> | null
  loading: boolean
}>()

const nodes = computed(() => (props.quote as DeepReadonly<QuoteWithNodes> | null)?.accessibleNodes
  ?.filter((node) => node.accessible) ?? [])
const failedProviderFaviconIDs = shallowRef<readonly string[]>([])

function formatMultiplier(value: number): string {
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(value)
}

function providerLabel(node: RemnaNode): string {
  return node.providerName?.trim() || t('catalog.providerUnavailable')
}

function providerFaviconAvailable(node: RemnaNode): boolean {
  return Boolean(node.providerFaviconUrl) && !failedProviderFaviconIDs.value.includes(node.uuid)
}

function useProviderFallback(node: RemnaNode): void {
  if (!failedProviderFaviconIDs.value.includes(node.uuid)) {
    failedProviderFaviconIDs.value = [...failedProviderFaviconIDs.value, node.uuid]
  }
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
        <div class="catalog-node__provider">
          <img v-if="providerFaviconAvailable(node)" class="catalog-node__provider-icon" :src="node.providerFaviconUrl ?? undefined" :alt="$t('catalog.providerIconAlt', { name: providerLabel(node) })" @error="useProviderFallback(node)" />
          <UIcon v-else class="catalog-node__provider-icon" name="i-ph-buildings-fill" role="img" :aria-label="$t('catalog.providerIconFallback')" />
          <span>{{ providerLabel(node) }}</span>
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
.catalog-node { min-height: 58px; display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: 0.7rem; padding: 0.65rem; border: 1px solid var(--line); border-radius: var(--radius-control); background: var(--surface-raised); }
.catalog-node strong, .catalog-node small { display: block; }
.catalog-node strong { overflow: hidden; font-size: 0.8rem; text-overflow: ellipsis; white-space: nowrap; }
.catalog-node small { margin-top: 0.2rem; color: var(--text-faint); font-family: var(--font-mono); font-size: 0.65rem; }
.catalog-node__provider { display: flex; align-items: center; justify-content: flex-end; gap: 0.35rem; max-width: 7rem; color: var(--text-faint); font-size: 0.65rem; text-align: right; }
.catalog-node__provider-icon { width: 1rem; height: 1rem; flex: 0 0 auto; border-radius: 3px; object-fit: cover; }
.catalog-node__provider span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
