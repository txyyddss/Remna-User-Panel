<script setup lang="ts">
import { computed, onMounted, onScopeDispose, shallowRef, watch } from 'vue'

import type { CatalogNode } from '@/api/types'
import CountryFlag from '@/components/common/CountryFlag.vue'

const props = defineProps<{ nodes: readonly CatalogNode[] }>()
const emit = defineEmits<{ openGeocheck: [node: CatalogNode] }>()

const activeIndex = shallowRef(0)
const touchStartX = shallowRef<number | null>(null)
const activeNode = computed(() => props.nodes[activeIndex.value])
let timer: ReturnType<typeof globalThis.setInterval> | undefined

function formatMultiplier(value: number): string {
  return new Intl.NumberFormat(undefined, { maximumFractionDigits: 2 }).format(value)
}

function startTimer(): void {
  if (timer !== undefined) globalThis.clearInterval(timer)
  timer = props.nodes.length > 1 ? globalThis.setInterval(() => step(1, false), 2_000) : undefined
}

function step(direction: -1 | 1, restart = true): void {
  if (props.nodes.length < 2) return
  activeIndex.value = (activeIndex.value + direction + props.nodes.length) % props.nodes.length
  if (restart) startTimer()
}

function onTouchStart(event: globalThis.TouchEvent): void { touchStartX.value = event.changedTouches[0]?.clientX ?? null }
function onTouchEnd(event: globalThis.TouchEvent): void {
  const end = event.changedTouches[0]?.clientX
  if (touchStartX.value === null || end === undefined) return
  const distance = end - touchStartX.value
  touchStartX.value = null
  if (Math.abs(distance) >= 32) step(distance < 0 ? 1 : -1)
}

watch(() => props.nodes.length, () => {
  activeIndex.value = 0
  startTimer()
})
onMounted(startTimer)
onScopeDispose(() => { if (timer !== undefined) globalThis.clearInterval(timer) })
</script>

<template>
  <div v-if="activeNode" class="squad-node-carousel" @touchstart.passive="onTouchStart" @touchend.passive="onTouchEnd">
    <UButton v-if="nodes.length > 1" class="squad-node-carousel__switch" color="neutral" variant="ghost" square icon="i-ph-caret-left" :aria-label="$t('catalog.previousNode')" @click="step(-1)" />
    <UButton :key="activeNode.uuid" type="button" color="neutral" variant="ghost" class="squad-node-carousel__node" :aria-label="$t('catalog.openNodeGeocheck', { current: activeIndex + 1, total: nodes.length })" data-haptic="open" @click="emit('openGeocheck', activeNode)">
      <CountryFlag :code="activeNode.countryCode" />
      <span class="squad-node-carousel__multiplier">{{ $t('catalog.nodeMultiplier', { multiplier: formatMultiplier(activeNode.consumptionMultiplier) }) }}</span>
    </UButton>
    <UButton v-if="nodes.length > 1" class="squad-node-carousel__switch" color="neutral" variant="ghost" square icon="i-ph-caret-right" :aria-label="$t('catalog.nextNode')" @click="step(1)" />
    <span v-if="nodes.length > 1" class="squad-node-carousel__position">{{ activeIndex + 1 }}/{{ nodes.length }}</span>
  </div>
</template>

<style scoped>
.squad-node-carousel { min-width: 0; display: grid; grid-template-columns: minmax(0, 1fr); place-items: center; gap: 0.25rem; }
.squad-node-carousel__node { width: min(100%, 5.25rem); min-height: 5.25rem; display: grid; place-items: center; gap: 0.35rem; padding: 0.55rem; border: 1px solid var(--line); border-radius: var(--radius-control); color: var(--text); background: var(--surface); cursor: pointer; animation: squad-node-slide-in 180ms var(--ease-out); }
.squad-node-carousel__node:hover { border-color: var(--line-strong); }
.squad-node-carousel__node:focus-visible { outline: 2px solid var(--accent); outline-offset: 2px; }
.squad-node-carousel__node :deep(.country-flag) { width: 2.25rem; height: 1.6rem; }
.squad-node-carousel__multiplier, .squad-node-carousel__position { color: var(--text-muted); font-family: var(--font-mono); font-size: 0.66rem; }
.squad-node-carousel__switch { display: none; }
@keyframes squad-node-slide-in { from { opacity: 0; transform: translateX(0.65rem); } }
@media (min-width: 900px) {
  .squad-node-carousel { grid-template-columns: 44px minmax(0, 1fr) 44px; }
  .squad-node-carousel__switch { width: 44px; height: 44px; display: inline-flex; }
  .squad-node-carousel__position { grid-column: 1 / -1; }
  .squad-node-carousel__node { animation: none; }
}
@media (prefers-reduced-motion: reduce) { .squad-node-carousel__node { animation: none; } }
</style>
