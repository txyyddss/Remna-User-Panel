<script setup lang="ts">
import type { CatalogNode } from '@/api/types'
import CountryFlag from '@/components/common/CountryFlag.vue'

defineProps<{ nodes: readonly CatalogNode[] }>()
const emit = defineEmits<{ openGeocheck: [node: CatalogNode] }>()

function formatMultiplier(value: number): string {
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(value)
}
</script>

<template>
  <div v-if="nodes.length" class="squad-node-list">
    <div class="squad-node-list__grid">
      <UButton
        v-for="(node, index) in nodes"
        :key="node.uuid"
        type="button"
        color="neutral"
        variant="ghost"
        class="squad-node-list__node"
        :aria-label="$t('catalog.openNodeGeocheck', { current: index + 1, total: nodes.length })"
        data-haptic="open"
        @click.stop="emit('openGeocheck', node)"
        @keydown.stop
      >
        <CountryFlag :code="node.countryCode" />
        <span class="squad-node-list__multiplier">{{ $t('catalog.nodeMultiplier', { multiplier: formatMultiplier(node.consumptionMultiplier) }) }}</span>
      </UButton>
    </div>
  </div>
</template>

<style scoped>
.squad-node-list { min-width: 0; display: grid; gap: 0.45rem; }
.squad-node-list__grid { min-width: 0; display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 7rem), 1fr)); gap: 0.45rem; }
.squad-node-list__node { min-width: 0; min-height: 44px; display: inline-flex; align-items: center; justify-content: center; gap: 0.4rem; padding: 0.45rem 0.55rem; border: 1px solid var(--line); border-radius: var(--radius-control); color: var(--text); background: var(--surface); cursor: pointer; }
.squad-node-list__node:hover { border-color: var(--line-strong); }
.squad-node-list__node:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
.squad-node-list__node :deep(.country-flag) { width: 1.7rem; height: 1.2rem; flex: 0 0 auto; }
.squad-node-list__multiplier { color: var(--text-muted); font-family: var(--font-mono); font-size: 0.68rem; }
@media (min-width: 900px) { .squad-node-list__grid { grid-template-columns: repeat(auto-fit, minmax(7.5rem, 1fr)); } }
</style>
