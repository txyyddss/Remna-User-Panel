<script setup lang="ts">
import { computed, shallowRef, watch, onMounted, onScopeDispose, useTemplateRef } from 'vue'
import type { IPLookupBGPTopology } from '@/api/ipLookup'
import LiveSourceNote from './LiveSourceNote.vue'
import { graphLayout, pathEdges } from './bgpGraph'
import { useGraphViewport } from './useGraphViewport'
const props = defineProps<{ topology: IPLookupBGPTopology }>()
const directLimit = shallowRef(12)
const depth = shallowRef(1)
const selectedASN = shallowRef<number | null>(null)
const selectedPath = shallowRef<number | null>(null)
const viewport = useTemplateRef<globalThis.HTMLDivElement>('viewport')
const { scale, offset, reset, zoom, pointerDown, pointerMove, pointerUp, keyboard } = useGraphViewport()
const layout = computed(() => graphLayout(props.topology, directLimit.value, depth.value))
const paths = computed(() => props.topology.paths.map((path, index) => ({ ...path, index })).filter(path => selectedASN.value === null || path.asPath.includes(selectedASN.value)))
const highlighted = computed(() => selectedPath.value !== null && props.topology.paths[selectedPath.value] ? pathEdges(props.topology.paths[selectedPath.value]!.asPath) : null)
const selectedNode = computed(() => props.topology.nodes.find(node => node.asn === selectedASN.value))
function resetView() { reset(); scale.value = Math.min(1, Math.max(0.35, (viewport.value?.clientWidth ?? layout.value.width) / layout.value.width)) }
let observer: InstanceType<typeof globalThis.ResizeObserver> | undefined
onMounted(() => { resetView(); if (viewport.value && typeof globalThis.ResizeObserver === 'function') { observer = new globalThis.ResizeObserver(resetView); observer.observe(viewport.value) } })
onScopeDispose(() => observer?.disconnect())
watch(() => props.topology, () => { directLimit.value = 12; depth.value = 1; selectedASN.value = selectedPath.value = null; resetView() })
function select(asn: number) { selectedASN.value = selectedASN.value === asn ? null : asn; selectedPath.value = null }
function showPath(index: number) { selectedPath.value = selectedPath.value === index ? null : index; if (selectedPath.value !== null) { depth.value = 300; directLimit.value = props.topology.nodes.length; resetView() } }
</script>

<template>
  <section class="bgp-topology">
    <header><h3>{{ $t('ipLookupLive.bgpTitle') }}</h3><code v-if="topology.prefix">{{ topology.prefix }}</code></header>
    <template v-if="topology.nodes.length">
      <div class="graph-toolbar">
        <UButton color="neutral" variant="ghost" size="sm" icon="i-ph-plus" :aria-label="$t('ipLookupLive.graph.zoomIn')" @click="zoom(1.2)" />
        <UButton color="neutral" variant="ghost" size="sm" icon="i-ph-minus" :aria-label="$t('ipLookupLive.graph.zoomOut')" @click="zoom(1 / 1.2)" />
        <UButton color="neutral" variant="ghost" size="sm" :label="$t('ipLookupLive.graph.reset')" @click="resetView" />
        <UButton v-if="directLimit < layout.directCount" color="neutral" variant="soft" size="sm" :label="$t('ipLookupLive.graph.moreDirect')" @click="directLimit += 12" />
        <UButton v-if="layout.hidden && depth < layout.maxDepth" color="neutral" variant="soft" size="sm" :label="$t('ipLookupLive.graph.expand')" @click="depth++" />
        <UButton v-if="depth > 1 || directLimit > 12" color="neutral" variant="ghost" size="sm" :label="$t('ipLookupLive.graph.collapse')" @click="depth = 1; directLimit = 12; resetView()" />
      </div>
      <div ref="viewport" class="graph-viewport" tabindex="0" role="region" :aria-label="$t('ipLookupLive.graph.instructions')" @pointerdown="pointerDown" @pointermove="pointerMove" @pointerup="pointerUp" @pointercancel="pointerUp" @keydown="keyboard">
        <svg :viewBox="`0 0 ${layout.width} ${layout.height}`" :style="{ width: `${layout.width}px`, height: `${layout.height}px`, transform: `translate(${offset.x}px, ${offset.y}px) scale(${scale})` }" :aria-label="$t('ipLookupLive.bgpTitle')">
          <g class="graph-edges"><path v-for="edge in layout.edges" :key="edge.id" :d="edge.d" :class="[`edge-${edge.relationship}`, { 'edge-dimmed': highlighted && !highlighted.has(edge.id), 'edge-highlighted': highlighted?.has(edge.id) }]"><title>{{ $t(`ipLookupLive.graph.relationship.${edge.relationship}`) }}</title></path></g>
          <g v-for="node in layout.nodes" :key="node.asn" :data-asn="node.asn" :transform="`translate(${node.x}, ${node.y})`" class="graph-node" :class="{ 'graph-origin': node.layer === 0, 'graph-selected': node.asn === selectedASN }" role="button" tabindex="0" :aria-label="`AS${node.asn} ${node.name}`" @click="select(node.asn)" @keydown.enter.stop="select(node.asn)" @keydown.space.stop.prevent="select(node.asn)">
            <title>{{ `AS${node.asn} ${node.name}` }}</title><rect width="184" height="48" rx="5" /><text x="10" y="18">AS{{ node.asn }}</text><text x="10" y="35" class="graph-node-name">{{ node.name.length > 24 ? `${node.name.slice(0, 23)}…` : node.name || $t('ipLookupLive.graph.unknownName') }}</text>
          </g>
        </svg>
      </div>
      <p class="graph-note">{{ $t('ipLookupLive.graph.instructions') }} <span v-if="layout.hidden">{{ $t('ipLookupLive.graph.hidden', { count: layout.hidden }) }}</span><span v-if="topology.truncated">{{ $t('ipLookupLive.graph.truncated') }}</span></p>
      <p class="graph-legend"><span>{{ $t('ipLookupLive.graph.relationship.upstream') }}</span><span>{{ $t('ipLookupLive.graph.relationship.peer') }}</span><span>{{ $t('ipLookupLive.graph.relationship.observed') }}</span></p>
      <details class="graph-routes" :open="selectedASN !== null"><summary>{{ selectedNode ? `AS${selectedNode.asn} ${selectedNode.name}` : $t('ipLookupLive.graph.routes') }} <span>{{ paths.length }}</span></summary><ul><li v-for="path in paths.slice(0, 60)" :key="path.index"><UButton color="neutral" variant="ghost" :aria-pressed="selectedPath === path.index" @click="showPath(path.index)"><code>{{ path.asPath.map(asn => `AS${asn}`).join(' → ') }}</code><small>{{ path.collector }}<span v-if="path.exchange">{{ $t('ipLookupLive.graph.exchangeContext', { name: path.exchange }) }}</span><time v-if="path.observedAt" :datetime="path.observedAt">{{ path.observedAt }}</time></small></UButton></li></ul><p v-if="paths.length > 60" class="graph-note">{{ $t('ipLookupLive.graph.routeLimit', { count: paths.length }) }}</p></details>
    </template>
    <LiveSourceNote :meta="topology" /><LiveSourceNote v-if="topology.edges.length" :meta="topology.relationshipMeta" />
    <p class="graph-note">{{ $t('ipLookupLive.graph.context') }}</p>
  </section>
</template>

<style scoped>
.bgp-topology { min-width: 0; padding-block: 0.6rem 1.2rem; }
.bgp-topology header { display: flex; flex-wrap: wrap; align-items: baseline; gap: 0.5rem 1rem; }
.bgp-topology h3 { font-size: 1rem; margin: 0; }
.bgp-topology header code { font-size: 0.82rem; }
.graph-toolbar { display: flex; flex-wrap: wrap; gap: 0.3rem; margin-block: 0.8rem; }
.graph-viewport { height: 340px; overflow: hidden; position: relative; touch-action: none; background: var(--surface); border: 1px solid var(--line); border-radius: 6px; cursor: grab; }
.graph-viewport svg { max-width: none; transform-origin: 0 0; user-select: none; }
.graph-edges path { fill: none; stroke: var(--text-muted); stroke-width: 1.1; opacity: 0.5; }
.graph-edges .edge-upstream { stroke: var(--ui-info); opacity: 0.8; }
.graph-edges .edge-peer { stroke: var(--ui-success); }
.graph-edges .edge-observed { stroke-dasharray: 4 4; }
.graph-edges .edge-highlighted { stroke-width: 2.8; opacity: 1; }
.graph-edges .edge-dimmed { opacity: 0.12; }
.graph-node { cursor: pointer; }
.graph-node rect { fill: var(--surface-elevated, var(--surface)); stroke: var(--line); }
.graph-origin rect, .graph-selected rect, .graph-node:focus rect { stroke: var(--ui-info); stroke-width: 2; }
.graph-node text { fill: var(--text); font: 12px var(--font-mono, monospace); pointer-events: none; }
.graph-node .graph-node-name { fill: var(--text-muted); font-family: var(--font-body, sans-serif); font-size: 10px; }
.graph-note, .graph-legend { font-size: 0.74rem; color: var(--text-muted); line-height: 1.5; }
.graph-note span { margin-left: 0.5rem; }
.graph-legend { display: flex; flex-wrap: wrap; gap: 0.6rem; }
.graph-legend span { border-bottom: 2px dashed var(--text-muted); padding-bottom: 0.15rem; }
.graph-legend span:first-child { border-bottom: 2px solid var(--ui-info); }
.graph-legend span:nth-child(2) { border-bottom: 2px solid var(--ui-success); }
.graph-routes summary { font-size: 0.82rem; cursor: pointer; overflow-wrap: anywhere; }
.graph-routes summary span { color: var(--text-muted); margin-left: 0.4rem; }
.graph-routes ul { list-style: none; padding: 0; margin: 0.5rem 0; max-height: 280px; overflow-y: auto; }
.graph-routes button { display: block; width: 100%; text-align: left; padding: 0.6rem 0; border: 0; border-bottom: 1px solid var(--line); background: transparent; color: var(--text); cursor: pointer; }
.graph-routes code { display: block; font-size: 0.76rem; overflow-wrap: anywhere; }
.graph-routes small { display: flex; flex-wrap: wrap; gap: 0.25rem 0.6rem; margin-top: 0.3rem; color: var(--text-muted); font-size: 0.7rem; }
.graph-routes button[aria-pressed=true] code { color: var(--ui-info); }
@media (min-width: 900px) { .graph-viewport { height: 420px; } }
</style>
