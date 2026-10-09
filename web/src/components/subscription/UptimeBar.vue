<script setup lang="ts">
import { computed, onScopeDispose, shallowRef } from 'vue'
import type { UptimeTimeline } from '@/api/subscription'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useI18n } from '@/i18n'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{ timeline: Readonly<UptimeTimeline>; label: string; host?: boolean }>()
const { t } = useI18n()
const selected = shallowRef<number | null>(null)
const now = shallowRef(Date.now())
const timer = globalThis.setInterval(() => { now.value = Date.now() }, 15_000)
onScopeDispose(() => globalThis.clearInterval(timer))
const segments = computed(() => props.timeline.segments.map(segment => ({ ...segment, duration: Math.max(0, Date.parse(segment.to) - Date.parse(segment.from)) })))
const state = computed(() => now.value - Date.parse(props.timeline.to) > 90_000 ? 'unknown' : props.timeline.state)
const tone = computed(() => state.value === 'operational' ? 'success' : state.value === 'outage' ? 'danger' : state.value === 'partial' && !props.host ? 'warning' : 'neutral')
const detail = computed(() => selected.value == null ? null : segments.value[selected.value] ?? null)

function selectAt(event: globalThis.MouseEvent): void {
  const bounds = (event.currentTarget as globalThis.HTMLElement).getBoundingClientRect()
  const at = Date.parse(props.timeline.from) + Math.min(1, Math.max(0, (event.clientX - bounds.left) / bounds.width)) * (Date.parse(props.timeline.to) - Date.parse(props.timeline.from))
  const index = segments.value.findIndex(segment => at < Date.parse(segment.to))
  selected.value = index < 0 ? segments.value.length - 1 : index
}
function selectNext(direction: number): void {
  selected.value = Math.min(segments.value.length - 1, Math.max(0, (selected.value ?? segments.value.length - 1) + direction))
}
</script>

<template>
  <div class="uptime" :class="{ 'uptime--host': host }">
    <div class="uptime__heading"><strong>{{ label }}</strong><StatusBadge :tone="tone" :label="t(`uptime.state.${state}`)" /></div>
    <UButton
      type="button" color="neutral" variant="ghost" class="uptime__bar"
      :aria-label="t('uptime.barAria', { name: label, from: formatDateTime(timeline.from), to: formatDateTime(timeline.to) })"
      @click="selectAt" @focus="selected = segments.length - 1" @blur="selected = null"
      @keydown.left.prevent="selectNext(-1)" @keydown.right.prevent="selectNext(1)" @keydown.escape="selected = null"
    >
      <span v-for="(segment, index) in segments" :key="segment.from" class="uptime__segment" :class="[`uptime__segment--${segment.state}`, { 'uptime__segment--selected': index === selected }]" :style="{ flexGrow: segment.duration }" :title="t('uptime.segment', { state: t(`uptime.state.${segment.state}`), from: formatDateTime(segment.from), to: formatDateTime(segment.to) })" aria-hidden="true" />
    </UButton>
    <div class="uptime__range"><span>{{ t('uptime.dayAgo') }}</span><span>{{ formatDateTime(timeline.to) }}</span></div>
    <p v-if="detail" class="uptime__detail" role="status">{{ t('uptime.segment', { state: t(`uptime.state.${detail.state}`), from: formatDateTime(detail.from), to: formatDateTime(detail.to) }) }}</p>
  </div>
</template>

<style scoped>
.uptime { display: grid; gap: 0.15rem; min-width: 0; }
.uptime__heading { display: flex; align-items: center; justify-content: space-between; gap: 0.75rem; }
.uptime__heading strong { font-size: 0.85rem; overflow-wrap: anywhere; }
.uptime__bar { display: flex; align-items: center; gap: 0; height: 44px; width: 100%; cursor: pointer; background: transparent; border: 0; padding: 0; border-radius: var(--radius-control); }
.uptime__bar:focus-visible { outline: 2px solid var(--accent); outline-offset: 3px; }
.uptime__segment { height: 14px; flex-basis: 0; min-width: 0; background: var(--surface-raised); }
.uptime__segment:first-child { border-radius: 3px 0 0 3px; }
.uptime__segment:last-child { border-radius: 0 3px 3px 0; }
.uptime__segment--operational { background: var(--color-success-500, #65a878); }
.uptime__segment--partial { background: var(--color-warning-500, #d2b565); }
.uptime__segment--outage { background: var(--color-error-500, #d47777); }
.uptime__segment--selected { outline: 1px solid var(--text); outline-offset: 2px; z-index: 1; }
.uptime--host .uptime__segment--partial { background: var(--surface-raised); }
.uptime__range { display: flex; justify-content: space-between; gap: 0.5rem; color: var(--text-muted); font-size: 0.7rem; }
.uptime__detail { margin: 0.4rem 0 0; color: var(--text-muted); font-size: 0.75rem; }
</style>
