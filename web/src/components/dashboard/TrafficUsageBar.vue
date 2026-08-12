<script setup lang="ts">
import { computed, shallowRef } from 'vue'

import type { CatalogNode, TopNode } from '@/api/types'
import CountryFlag from '@/components/common/CountryFlag.vue'
import { getLocale, t } from '@/i18n'
import { formatBytes } from '@/utils/format'

const props = defineProps<{
  nodes: readonly TopNode[]
  totalBytes: string
  catalogNodes: readonly CatalogNode[]
}>()

const scale = 10_000n
const pointerScale = 1_000_000n
const paletteSize = 5
const hoveredKey = shallowRef<string | null>(null)
const selectedKey = shallowRef<string | null>(null)
const trackRef = shallowRef<globalThis.HTMLElement | null>(null)

interface Segment { key: string; name: string; countryCode: string; bytes: bigint; startBytes: bigint; widthBasis: bigint; color: string; bytesLabel: string; percentageLabel: string; multiplierLabel: string; isOther: boolean; ariaLabel: string }

function byteCount(value: string): bigint {
  try {
    return /^\d+$/.test(value.trim()) ? BigInt(value.trim()) : 0n
  } catch {
    return 0n
  }
}

function percentageLabel(value: bigint, total: bigint): string {
  const tenths = total > 0n ? (value * 1000n) / total : 0n
  return new Intl.NumberFormat(getLocale(), { minimumFractionDigits: 1, maximumFractionDigits: 1 }).format(Number(tenths) / 10)
}

function multiplierLabel(value: number | null): string {
  if (value === null || !Number.isFinite(value)) return t('home.trafficMultiplierUnavailable')
  const formatted = new Intl.NumberFormat(getLocale(), { maximumFractionDigits: 2 }).format(value)
  return t('home.trafficMultiplierValue', { value: formatted })
}

const sourceNodes = computed(() => props.nodes.map((node, index) => ({ node, index, bytes: byteCount(node.totalBytes) })).filter((entry) => entry.bytes > 0n))

const totalForBar = computed(() => {
  const reported = byteCount(props.totalBytes)
  const listed = sourceNodes.value.reduce((sum, entry) => sum + entry.bytes, 0n)
  return reported > listed ? reported : listed
})

const catalogByUuid = computed(() => new Map(props.catalogNodes.map((node) => [node.uuid, node])))
const colorByUuid = computed(() => {
  const uuids = [...new Set(sourceNodes.value.map(({ node }) => node.uuid))].sort()
  return new Map(uuids.map((uuid, index) => [uuid, index % paletteSize]))
})

const segments = computed<Segment[]>(() => {
  const total = totalForBar.value
  if (total <= 0n || sourceNodes.value.length === 0) return []
  const listed = sourceNodes.value.reduce((sum, entry) => sum + entry.bytes, 0n)
  const entries = sourceNodes.value.map(({ node, index, bytes }) => {
    const metadata = catalogByUuid.value.get(node.uuid)
    const multiplier = metadata && Number.isFinite(metadata.consumptionMultiplier) ? metadata.consumptionMultiplier : null
    return {
      key: node.uuid + '-' + index,
      name: metadata?.name || node.name,
      countryCode: metadata?.countryCode || node.countryCode,
      bytes,
      rawWidthBasis: (bytes * scale) / total,
      color: 'var(--traffic-node-' + ((colorByUuid.value.get(node.uuid) ?? index % paletteSize) + 1) + ')',
      multiplierLabel: multiplierLabel(multiplier),
      isOther: false,
    }
  })
  if (total > listed) {
    entries.push({
      key: 'other-usage',
      name: t('home.trafficOther'),
      countryCode: '',
      bytes: total - listed,
      rawWidthBasis: ((total - listed) * scale) / total,
      color: 'var(--traffic-node-other)',
      multiplierLabel: t('home.trafficMultiplierUnavailable'),
      isOther: true,
    })
  }
  let startBytes = 0n
  let widthBasisUsed = 0n
  return entries.map((entry, index) => {
    const widthBasis = index === entries.length - 1 ? scale - widthBasisUsed : entry.rawWidthBasis
    widthBasisUsed += widthBasis
    const bytesLabel = formatBytes(entry.bytes.toString())
    const percentage = percentageLabel(entry.bytes, total)
    const ariaLabel = entry.isOther
      ? t('home.trafficOtherAria', { usage: bytesLabel, percentage })
      : t('home.trafficSegmentAria', { name: entry.name, country: entry.countryCode, usage: bytesLabel, percentage, multiplier: entry.multiplierLabel })
    const segment: Segment = { ...entry, startBytes, widthBasis, bytesLabel, percentageLabel: percentage, ariaLabel }
    startBytes += entry.bytes
    return segment
  })
})

const activeSegment = computed(() => {
  const key = hoveredKey.value ?? selectedKey.value
  return segments.value.find((segment) => segment.key === key) ?? null
})

function segmentAt(clientX: number): Segment | null {
  const rect = trackRef.value?.getBoundingClientRect()
  const total = totalForBar.value
  if (!rect || rect.width <= 0 || total <= 0n) return null
  const relative = Math.min(0.999999, Math.max(0, (clientX - rect.left) / rect.width))
  const position = BigInt(Math.floor(relative * Number(pointerScale)))
  return segments.value.find((segment) => position * total < (segment.startBytes + segment.bytes) * pointerScale) ?? null
}

function previewAt(event: globalThis.PointerEvent): void { hoveredKey.value = segmentAt(event.clientX)?.key ?? null }
function selectAt(event: globalThis.MouseEvent): void { const segment = segmentAt(event.clientX); if (segment) toggleSelection(segment.key) }
function toggleSelection(key: string): void {
  if (selectedKey.value === key) { selectedKey.value = null; hoveredKey.value = null; return }
  selectedKey.value = key
}
function clearHover(): void { hoveredKey.value = null }
function clearSelection(): void { selectedKey.value = null; hoveredKey.value = null }
</script>

<template>
  <div v-if="segments.length" class="traffic-usage-bar">
    <div ref="trackRef" class="traffic-usage-bar__track" role="group" :aria-label="$t('home.trafficBarLabel')" @click="selectAt" @pointermove="previewAt" @pointerleave="clearHover" @pointercancel="clearHover" @keydown.escape.stop.prevent="clearSelection">
      <span v-for="segment in segments" :key="'visual-' + segment.key" class="traffic-usage-bar__visual" :style="{ width: Number(segment.widthBasis) / 100 + '%', backgroundColor: segment.color }" aria-hidden="true" />
      <UButton v-for="segment in segments" :key="segment.key" type="button" color="neutral" variant="ghost" class="traffic-usage-bar__segment" :class="{ 'traffic-usage-bar__segment--selected': selectedKey === segment.key }" :style="{ left: Number(segment.startBytes * scale / totalForBar) / 100 + '%', width: Number(segment.widthBasis) / 100 + '%' }" :aria-label="segment.ariaLabel" :aria-pressed="selectedKey === segment.key" @click.stop="toggleSelection(segment.key)" @focus="hoveredKey = segment.key" @blur="clearHover" />
    </div>
    <article v-if="activeSegment" class="traffic-usage-bar__details" role="status" aria-live="polite">
      <header class="traffic-usage-bar__header">
        <span class="traffic-usage-bar__swatch" :style="{ backgroundColor: activeSegment.color }" aria-hidden="true" />
        <div>
          <strong>{{ activeSegment.name }}</strong>
          <span v-if="!activeSegment.isOther" class="traffic-usage-bar__country">
            <CountryFlag :code="activeSegment.countryCode" />
            {{ activeSegment.countryCode }}
          </span>
        </div>
      </header>
      <p v-if="activeSegment.isOther" class="traffic-usage-bar__description">{{ $t('home.trafficOtherDescription') }}</p>
      <dl class="traffic-usage-bar__stats">
        <div><dt>{{ $t('home.trafficNodeUsage') }}</dt><dd>{{ activeSegment.bytesLabel }}</dd></div>
        <div><dt>{{ $t('home.trafficNodeShare') }}</dt><dd>{{ activeSegment.percentageLabel }}%</dd></div>
        <div><dt>{{ $t('home.trafficNodeMultiplier') }}</dt><dd>{{ activeSegment.multiplierLabel }}</dd></div>
      </dl>
    </article>
  </div>
</template>

<style scoped>
.traffic-usage-bar { --traffic-node-1: var(--accent); --traffic-node-2: #83a6b8; --traffic-node-3: #d1b578; --traffic-node-4: #c1918e; --traffic-node-5: #9d9bc2; --traffic-node-other: var(--line-strong); margin-top: 0.75rem; }
.traffic-usage-bar__track { position: relative; display: flex; min-height: 44px; overflow: hidden; border: 1px solid var(--line); border-radius: 12px; background: var(--surface-muted); isolation: isolate; }
.traffic-usage-bar__visual { display: block; min-width: 0; height: 100%; min-height: 44px; opacity: 0.9; }
.traffic-usage-bar__segment { position: absolute; inset-block: 0; z-index: 1; min-width: 0; padding: 0; border: 0; border-radius: 0; background: transparent; cursor: pointer; }
.traffic-usage-bar__segment:focus-visible { outline: 2px solid var(--text); outline-offset: -3px; }
.traffic-usage-bar__segment--selected { box-shadow: inset 0 0 0 2px var(--text); }
.traffic-usage-bar__details { margin-top: 0.65rem; padding: 0.8rem 0.9rem; border: 1px solid var(--line); border-radius: 12px; background: var(--surface-muted); }
.traffic-usage-bar__header { display: flex; align-items: center; gap: 0.65rem; }
.traffic-usage-bar__header strong { display: block; color: var(--text); font-size: 0.88rem; }
.traffic-usage-bar__country { display: inline-flex; align-items: center; gap: 0.2rem; color: var(--text-muted); font-size: 0.74rem; }
.traffic-usage-bar__swatch { width: 0.65rem; height: 0.65rem; flex: 0 0 auto; border-radius: 3px; }
.traffic-usage-bar__description { margin: 0.55rem 0 0; color: var(--text-muted); font-size: 0.76rem; }
.traffic-usage-bar__stats { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 0.65rem; margin: 0.7rem 0 0; }
.traffic-usage-bar__stats div { min-width: 0; }
.traffic-usage-bar__stats dt { color: var(--text-muted); font-size: 0.68rem; }
.traffic-usage-bar__stats dd { margin: 0.15rem 0 0; color: var(--text); font-size: 0.78rem; font-weight: 700; overflow-wrap: anywhere; }
@media (max-width: 360px) { .traffic-usage-bar__stats { gap: 0.35rem; } .traffic-usage-bar__stats dt, .traffic-usage-bar__stats dd { font-size: 0.66rem; } }
@media (prefers-reduced-motion: no-preference) { .traffic-usage-bar__details { animation: traffic-details-in 160ms ease-out; } }
@keyframes traffic-details-in { from { opacity: 0; transform: translateY(-3px); } to { opacity: 1; transform: translateY(0); } }
</style>
