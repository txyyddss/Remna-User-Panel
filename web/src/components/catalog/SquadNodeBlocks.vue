<script setup lang="ts">
import type { CatalogNode } from '@/api/types'
import CountryFlag from '@/components/common/CountryFlag.vue'
import { t } from '@/i18n'

defineProps<{ nodes: readonly CatalogNode[] }>()

const emit = defineEmits<{ openGeocheck: [node: CatalogNode] }>()

function formatMultiplier(value: number): string {
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(value)
}

function providerLabel(node: CatalogNode): string {
  return node.providerName?.trim() || t('catalog.providerUnavailable')
}
</script>

<template>
  <section v-if="nodes.length" class="squad-nodes">
    <div class="squad-nodes__grid">
      <article v-for="node in nodes" :key="node.uuid" class="squad-node">
        <CountryFlag :code="node.countryCode" />
        <span class="squad-node__copy">
          <strong>{{ node.name }}</strong>
          <small>{{ providerLabel(node) }}</small>
        </span>
        <span class="squad-node__actions">
          <span class="squad-node__multiplier">{{ $t('catalog.nodeMultiplier', { multiplier: formatMultiplier(node.consumptionMultiplier) }) }}</span>
          <UTooltip :text="$t('statistics.geocheck.open')">
            <UButton
              type="button"
              color="neutral"
              variant="ghost"
              square
              class="squad-node__geocheck"
              icon="i-ph-globe-hemisphere-west"
              :aria-label="$t('statistics.geocheck.open')"
              data-haptic="open"
              @click.stop="emit('openGeocheck', node)"
            />
          </UTooltip>
        </span>
      </article>
    </div>
  </section>
</template>

<style scoped>
.squad-nodes { display: grid; gap: 0.4rem; }
.squad-nodes__grid { display: grid; grid-template-columns: minmax(0, 1fr); gap: 0.4rem; }
.squad-node { min-width: 0; display: grid; grid-template-columns: auto minmax(0, 1fr) auto; align-items: center; gap: 0.45rem; padding: 0.45rem; border: 1px solid var(--line); border-radius: 8px; background: color-mix(in srgb, var(--surface) 82%, transparent); }
.squad-node__copy { min-width: 0; display: grid; gap: 0.08rem; }
.squad-node strong, .squad-node small { overflow-wrap: anywhere; }
.squad-node strong { font-size: 0.7rem; line-height: 1.25; }
.squad-node small { color: var(--text-faint); font-size: 0.61rem; line-height: 1.25; }
.squad-node__actions { display: flex; align-items: center; gap: 0.2rem; }
.squad-node__multiplier { color: var(--text-muted); font-family: var(--font-mono); font-size: 0.62rem; white-space: nowrap; }
.squad-node__geocheck { flex: 0 0 auto; }
</style>
